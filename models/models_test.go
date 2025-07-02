package models

import (
	"database/sql"
	"os"
	"testing"

	db "truck-checkout/database"

	"github.com/google/uuid"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	// Setup test database
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	// Create tables
	if err := db.CreateTables(testDB); err != nil {
		panic(err)
	}

	// Set the global DB to our test database
	originalDB := db.DB
	db.DB = testDB

	// Run tests
	code := m.Run()

	// Restore original DB (though it was nil)
	db.DB = originalDB

	os.Exit(code)
}

func TestTruckModel(t *testing.T) {
	team := "beltline"
	calendarID := uuid.New()
	err := InsertTruck("Tulip", &team, calendarID, true)
	if err != nil {
		t.Fatalf("failed to insert truck: %v", err)
	}

	truck, err := GetTruckByName("Tulip")
	if err != nil {
		t.Fatalf("failed to get truck: %v", err)
	}
	if truck.Name != "Tulip" {
		t.Errorf("expected truck name to be 'Tulip', got '%s'", truck.Name)
	}
	if *truck.DefaultTeam != team {
		t.Errorf("expected default team to be 'beltline', got '%s'", *truck.DefaultTeam)
	}
	if truck.CalendarID != calendarID {
		t.Errorf("expected calendar ID to match, got '%s'", truck.CalendarID)
	}
	if !truck.IsActive {
		t.Error("expected truck to be active")
	}
}

// func TestCheckoutModel(t *testing.T) {
// 	truckID, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
// 	if err != nil {
// 		t.Fatalf("failed to parse UUID: %v", err)
// 	}
// 	checkout := Checkout{
// 		ID:        uuid.New(),
// 		TruckID:   truckID,
// 		UserID:    uuid.New(),
// 		UserName:  "Qwelian",
// 		TeamName:  "forest_restoration",
// 		StartDate: time.Now(),
// 		EndDate:   time.Now().Add(8 * time.Hour),
// 		Purpose:   "Watering trees",
// 	}

// 	if checkout.TeamName != "forest_restoration" {
// 		t.Errorf("expected team name to match")
// 	}
// }
