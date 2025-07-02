package models

import (
	"database/sql"
	"os"
	"testing"
	"time"

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
	err := InsertTruck("Tulip", &team, calendarID, false)
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
	if truck.IsActive {
		t.Error("expected truck to be free")
	}
}

func TestInsertTruck_InvalidName(t *testing.T) {
	calendarID := uuid.New()
	team := "beltline"
	err := InsertTruck("BananaBoat", &team, calendarID, true)
	if err == nil {
		t.Fatal("expected error for invalid truck name")
	}
}

func TestCheckoutDatabaseOperations(t *testing.T) {
	// First create a truck
	team := "forest_restoration"
	err := InsertTruck("Magnolia", &team, uuid.New(), false)
	if err != nil {
		t.Fatalf("failed to insert truck: %v", err)
	}

	truck, err := GetTruckByName("Magnolia")
	if err != nil {
		t.Fatalf("failed to get truck: %v", err)
	}

	// Test checkout insertion
	checkout := Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    uuid.New(),
		UserName:  "Qwelian Tanner",
		TeamName:  "beltline",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(4 * time.Hour),
		Purpose:   "Testing This truck was checked out digitally",
	}

	err = InsertCheckout(checkout)
	if err != nil {
		t.Fatalf("failed to insert checkout: %v", err)
	}

	// Test retrieval
	retrievedCheckout, err := GetCheckoutByID(checkout.ID)
	if err != nil {
		t.Fatalf("failed to get checkout: %v", err)
	}

	if retrievedCheckout.UserName != checkout.UserName {
		t.Errorf("expected user name %s, got %s", checkout.UserName, retrievedCheckout.UserName)
	}
	if retrievedCheckout.TeamName != checkout.TeamName {
		t.Errorf("expected team name %s, got %s", checkout.TeamName, retrievedCheckout.TeamName)
	}
	if retrievedCheckout.Purpose != checkout.Purpose {
		t.Errorf("expected purpose %s, got %s", checkout.Purpose, retrievedCheckout.Purpose)
	}
}
