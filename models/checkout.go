package models

import (
	"database/sql"
	"fmt"
	"time"
	db "truck-checkout/database"

	"github.com/google/uuid"
)

type Checkout struct {
	ID              uuid.UUID
	TruckID         uuid.UUID
	UserID          string
	UserName        string
	TeamName        string
	StartDate       time.Time
	EndDate         time.Time
	Purpose         string
	CalendarEventID string
	CreatedAt       time.Time
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

	// Update the truck's status to not checked out
	_, err = tx.Exec(`
		UPDATE trucks SET is_checked_out = FALSE WHERE id = ?
	`, truckID.String())
	if err != nil {
		return fmt.Errorf("failed to update truck status: %w", err)
	}

	// Archive checkouts instead of deleting (for audit trail)
	_, err = tx.Exec(`
		UPDATE checkouts 
		SET 
			end_date = ?,
			released_by = ?,
			released_at = ?
		WHERE truck_id = ? AND end_date > ?
	`, time.Now(), releasedBy, time.Now(), truckID.String(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update checkouts: %w", err)
	}

	return tx.Commit()
}

func GetActiveCheckoutByTruckID(truckID uuid.UUID) (*Checkout, error) {
	var checkout Checkout
	err := db.DB.QueryRow(`
		SELECT id, truck_id, user_id, user_name, team_name, start_date, end_date, purpose
		FROM checkouts 
		WHERE truck_id = ? AND end_date > ?
		ORDER BY start_date DESC
		LIMIT 1
	`, truckID.String(), time.Now()).Scan(
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


