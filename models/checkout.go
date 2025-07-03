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
