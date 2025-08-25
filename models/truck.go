package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	db "truck-checkout/database"

	"github.com/google/uuid"
)

type Truck struct {
	ID           uuid.UUID
	Name         string
	DefaultTeam  *string
	CalendarID   uuid.UUID
	IsCheckedOut bool
}

func InsertTruck(name string, team *string, calendarID uuid.UUID, isCheckedOut bool) error {
	// anyway for the enum to be auto validated?
	if !IsValidTruck(name) {
		return fmt.Errorf("invalid truck name: %s", name)
	}
	if team != nil && !IsValidTeam(*team) {
		return fmt.Errorf("invalid default team: %s", *team)
	}

	id := uuid.New()
	_, err := db.DB.Exec(`
		INSERT INTO trucks (id, name, default_team, calendar_id, is_checked_out)
		VALUES (?, ?, ?, ?, ?);
	`, id, name, team, calendarID, isCheckedOut)
	return err
}

func GetTruckByName(name string) (*Truck, error) {
	var truck Truck
	var defaultTeam sql.NullString

	row := db.DB.QueryRow("SELECT id, name, default_team, calendar_id, is_checked_out FROM trucks WHERE name = ?", name)
	err := row.Scan(&truck.ID, &truck.Name, &defaultTeam, &truck.CalendarID, &truck.IsCheckedOut)
	if err != nil {
		return nil, err
	}

	if defaultTeam.Valid {
		truck.DefaultTeam = &defaultTeam.String
	}

	return &truck, nil
}

func UpdateTruck(truck Truck) error {
	if !IsValidTruck(truck.Name) {
		return fmt.Errorf("invalid truck name: %s", truck.Name)
	}
	if truck.DefaultTeam != nil && !IsValidTeam(*truck.DefaultTeam) {
		return fmt.Errorf("invalid default team: %s", *truck.DefaultTeam)
	}

	_, err := db.DB.Exec(`
		UPDATE trucks
		SET name = ?, default_team = ?, calendar_id = ?, is_checked_out = ?
		WHERE id = ?;
	`, truck.Name, truck.DefaultTeam, truck.CalendarID, truck.IsCheckedOut, truck.ID)
	return err
}

func GetTrucksByCheckoutStatus(day time.Time, isCheckedOut bool) ([]Truck, error) {
	// startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	// endOfDay := startOfDay.Add(24 * time.Hour)
	currentTime := time.Now()

	var query string
	if isCheckedOut {
		// Forunavailable trucks: should be marked as checked out AND have active checkout records
		query = `
			SELECT t.id, t.name, t.default_team, t.calendar_id, t.is_checked_out
			FROM trucks t
			WHERE t.is_checked_out = true
			AND EXISTS (
				SELECT 1 FROM checkouts c
				WHERE c.truck_id = t.id
				AND c.start_date <= ?
				AND c.end_date > ?
			)
		`
	} else {
		// For available trucks: should be marked as available AND have no active checkout records
		query = `
			SELECT t.id, t.name, t.default_team, t.calendar_id, t.is_checked_out
			FROM trucks t
			WHERE t.is_checked_out = false
			AND NOT EXISTS (
				SELECT 1 FROM checkouts c
				WHERE c.truck_id = t.id
				AND c.start_date <= ?
				AND c.end_date > ?
			)
		`
	}

	rows, err := db.DB.Query(query, currentTime, currentTime)
	if err != nil {
		return nil, fmt.Errorf("querying trucks by checkout status: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	var trucks []Truck

	for rows.Next() {
		var t Truck
		var defaultTeam sql.NullString
		var idStr, calendarIDStr string

		if err := rows.Scan(&idStr, &t.Name, &defaultTeam, &calendarIDStr, &t.IsCheckedOut); err != nil {
			return nil, fmt.Errorf("scanning truck row: %w", err)
		}

		var parseErr error
		if t.ID, parseErr = uuid.Parse(idStr); parseErr != nil {
			return nil, fmt.Errorf("parsing truck UUID: %w", parseErr)
		}

		if t.CalendarID, parseErr = uuid.Parse(calendarIDStr); parseErr != nil {
			return nil, fmt.Errorf("parsing calendar UUID: %w", parseErr)
		}

		if defaultTeam.Valid {
			t.DefaultTeam = &defaultTeam.String
		}
		trucks = append(trucks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("after row iteration: %w", err)
	}

	return trucks, nil
}
