package models

import (
	"database/sql"
	"fmt"

	db "truck-checkout/database"

	"github.com/google/uuid"
)

type Truck struct {
	ID          uuid.UUID
	Name        string
	DefaultTeam *string
	CalendarID  uuid.UUID
	IsActive    bool
}

func InsertTruck(name string, team *string, calendarID uuid.UUID, isActive bool) error {
	// anyway for the enum to be auto validated?
	if !IsValidTruck(name) {
		return fmt.Errorf("invalid truck name: %s", name)
	}
	if team != nil && !IsValidTeam(*team) {
		return fmt.Errorf("invalid default team: %s", *team)
	}

	id := uuid.New()
	_, err := db.DB.Exec(`
		INSERT INTO trucks (id, name, default_team, calendar_id, is_active)
		VALUES (?, ?, ?, ?, ?);
	`, id, name, team, calendarID, isActive)
	return err
}

func GetTruckByName(name string) (*Truck, error) {
	var truck Truck
	var defaultTeam sql.NullString

	row := db.DB.QueryRow("SELECT id, name, default_team, calendar_id, is_active FROM trucks WHERE name = ?", name)
	err := row.Scan(&truck.ID, &truck.Name, &defaultTeam, &truck.CalendarID, &truck.IsActive)
	if err != nil {
		return nil, err
	}

	if defaultTeam.Valid {
		truck.DefaultTeam = &defaultTeam.String
	}

	return &truck, nil
}
