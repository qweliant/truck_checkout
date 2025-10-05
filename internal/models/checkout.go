package models

import (
	"database/sql"
	"fmt"
	"time"
	db "truck-checkout/internal/database"

	"github.com/google/uuid"
)

type Checkout struct {
	ID              uuid.UUID  `json:"id"`
	TruckID         uuid.UUID  `json:"truck_id"`
	UserID          string     `json:"user_id"`
	UserName        string     `json:"user_name"`
	TeamName        string     `json:"team_name"`
	StartDate       time.Time  `json:"start_date"`
	EndDate         time.Time  `json:"end_date"`
	Purpose         string     `json:"purpose,omitempty"`
	CalendarEventID string     `json:"calendar_event_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ReleasedBy      *string    `json:"released_by,omitempty"`
	ReleasedAt      *time.Time `json:"released_at,omitempty"`
}

func InsertCheckout(checkout Checkout) error {
	if !IsValidTeam(checkout.TeamName) {
		return fmt.Errorf("invalid team name: %s", checkout.TeamName)
	}
	_, err := db.DB.Exec(`
		INSERT INTO checkouts (id, truck_id, user_id, user_name, team_name, start_date, end_date, purpose)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, checkout.ID.String(), checkout.TruckID.String(), checkout.UserID,
		checkout.UserName, checkout.TeamName, checkout.StartDate, checkout.EndDate, checkout.Purpose)
	return err
}

func CreateCheckout(checkout Checkout) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer a rollback in case of an error
	defer tx.Rollback()

	// Step 1: Insert the checkout record
	_, err = tx.Exec(`
		INSERT INTO checkouts (id, truck_id, user_id, user_name, team_name, start_date, end_date, purpose)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, checkout.ID.String(), checkout.TruckID.String(), checkout.UserID,
		checkout.UserName, checkout.TeamName, checkout.StartDate, checkout.EndDate, checkout.Purpose)
	if err != nil {
		return fmt.Errorf("failed to insert checkout: %w", err)
	}

	// Step 2: Update the truck's status to checked out
	_, err = tx.Exec(`UPDATE trucks SET is_checked_out = true WHERE id = ?`, checkout.TruckID.String())
	if err != nil {
		return fmt.Errorf("failed to update truck status: %w", err)
	}

	return tx.Commit()
}

func GetCheckoutByID(id uuid.UUID) (*Checkout, error) {
	var checkout Checkout
	var purpose sql.NullString

	row := db.DB.QueryRow(`
		SELECT id, truck_id, user_id, user_name, team_name, start_date, end_date, purpose
		FROM checkouts WHERE id = ?
	`, id.String())

	err := row.Scan(&checkout.ID, &checkout.TruckID, &checkout.UserID,
		&checkout.UserName, &checkout.TeamName, &checkout.StartDate, &checkout.EndDate, &purpose)
	if err != nil {
		return nil, err
	}

	if purpose.Valid {
		checkout.Purpose = purpose.String
	}

	return &checkout, nil
}

func ReleaseTruckFromCheckout(truckID uuid.UUID, releasedBy string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	var currentCheckoutID string
	err = tx.QueryRow(`
        SELECT id FROM checkouts 
        WHERE truck_id = ? AND released_at IS NULL
        ORDER BY start_date DESC 
        LIMIT 1
    `, truckID.String()).Scan(&currentCheckoutID)

	if err != nil {
		if err == sql.ErrNoRows {
			// This is not an error, it just means the truck is already available.
			// We can double-check the truck's status and fix it if it's inconsistent.
			if err == sql.ErrNoRows {
				return fmt.Errorf("no active checkout found for this truck")
			}
		}
		return fmt.Errorf("failed to find current checkout: %w", err)
	}

	// Update the active checkout to mark it as released
	_, err = tx.Exec(`
        UPDATE checkouts 
        SET released_at = ?, released_by = ?
        WHERE id = ?
    `, now, releasedBy, currentCheckoutID)
	if err != nil {
		return fmt.Errorf("failed to update current checkout: %w", err)
	}

	_, err = tx.Exec(`
        UPDATE trucks SET is_checked_out = false WHERE id = ?
    `, truckID.String())
	if err != nil {
		return fmt.Errorf("failed to update truck status: %w", err)
	}

	return tx.Commit()
}

func GetActiveCheckoutByTruckID(truckID uuid.UUID) (*Checkout, error) {
	var checkout Checkout
	now := time.Now()

	query := `
        SELECT id, truck_id, user_id, user_name, team_name, start_date, end_date, purpose
        FROM checkouts
        WHERE truck_id = ?
          AND start_date <= ?
          AND end_date > ?
          AND released_at IS NULL
        ORDER BY start_date DESC
        LIMIT 1
    `
	err := db.DB.QueryRow(query, truckID.String(), now, now).Scan(
		&checkout.ID,
		&checkout.TruckID,
		&checkout.UserID,
		&checkout.UserName,
		&checkout.TeamName,
		&checkout.StartDate,
		&checkout.EndDate,
		&checkout.Purpose,
	)

	if err != nil {
		return nil, err
	}

	return &checkout, nil
}
