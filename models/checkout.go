package models

import (
	"time"

	"github.com/google/uuid"
)

type Checkout struct {
	ID              uuid.UUID
	TruckID         uuid.UUID
	UserID          uuid.UUID
	UserName        string
	TeamName        string
	StartDate       time.Time
	EndDate         time.Time
	Purpose         string
	CalendarEventID string
	CreatedAt       time.Time
}
