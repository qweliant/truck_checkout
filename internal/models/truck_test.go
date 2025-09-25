package models

import (
	"database/sql"
	"os"
	"testing"
	"time"
	db "truck-checkout/internal/database"

	"github.com/google/uuid"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	if err := db.CreateTables(testDB); err != nil {
		panic(err)
	}

	// Set the global DB to our test database
	originalDB := db.DB
	db.DB = testDB

	// Run tests
	code := m.Run()

	// Clean up AFTER tests run
	testDB.Close()

	// Restore original DB (though it was nil)
	db.DB = originalDB

	os.Exit(code)
}

func TestInsertTruck(t *testing.T) {
	ResetTestDB(t)
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
	if !truck.IsCheckedOut {
		t.Error("expected truck to be free")
	}
}

func TestInsertTruck_InvalidName(t *testing.T) {
	ResetTestDB(t)
	calendarID := uuid.New()
	team := "beltline"
	err := InsertTruck("BananaBoat", &team, calendarID, true)
	if err == nil {
		t.Fatal("expected error for invalid truck name")
	}
}

func TestUpdateTruck(t *testing.T) {
	ResetTestDB(t)
	team := "floaters"
	calendarID := uuid.New()
	err := InsertTruck("Libby", &team, calendarID, false)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	truck, err := GetTruckByName("Libby")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// update Libby to be inactive and change team
	newTeam := "beltline"
	truck.DefaultTeam = &newTeam
	truck.IsCheckedOut = true

	err = UpdateTruck(*truck)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := GetTruckByName("Libby")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if !updated.IsCheckedOut {
		t.Errorf("expected IsCheckedOut to be true, got false")
	}

	if *updated.DefaultTeam != "beltline" {
		t.Errorf("expected DefaultTeam to be 'beltline', got %s", *updated.DefaultTeam)
	}
}

func TestGetAvailableTrucksForToday(t *testing.T) {
	ResetTestDB(t)

	// Create test trucks
	team1 := "forest_restoration"
	team2 := "beltline"

	// Create available trucks (active)
	err := InsertTruck("Tulip", &team1, uuid.New(), true)
	if err != nil {
		t.Fatalf("failed to insert truck Tulip: %v", err)
	}

	err = InsertTruck("Andre350", &team2, uuid.New(), true)
	if err != nil {
		t.Fatalf("failed to insert truck Andre350: %v", err)
	}

	// Create unavailable truck (unavaialble)
	err = InsertTruck("Magnolia", &team1, uuid.New(), false)
	if err != nil {
		t.Fatalf("failed to insert truck Magnolia: %v", err)
	}

	// Get Magnolia truck for checkout
	magnolia, err := GetTruckByName("Magnolia")
	if err != nil {
		t.Fatalf("failed to get Magnolia truck: %v", err)
	}

	// Create a checkout for today that overlaps with the query day
	today := time.Now()
	startOfToday := time.Date(today.Year(), today.Month(), today.Day(), 9, 0, 0, 0, today.Location())
	endOfToday := startOfToday.Add(8 * time.Hour) // 9 AM to 5 PM

	checkout := Checkout{
		ID:        uuid.New(),
		TruckID:   magnolia.ID,
		UserID:    uuid.New().String(),
		UserName:  "Test User",
		TeamName:  "neighborwoods",
		StartDate: startOfToday,
		EndDate:   endOfToday,
		Purpose:   "Testing checkout overlap",
	}

	err = InsertCheckout(checkout)
	if err != nil {
		t.Fatalf("failed to insert checkout: %v", err)
	}

	// Test: Get available trucks for today
	availableTrucks, err := GetTrucksByCheckoutStatus(today, true)
	if err != nil {
		t.Fatalf("failed to get available trucks: %v", err)
	}

	t.Logf("Available trucks: %v", getTruckNames(availableTrucks))

	// Should have 2 available trucks (Tulip and Andre350)
	// Magnolia should be excluded (checked out)
	// InactiveTruck should be excluded (inactive)
	expectedCount := 2
	if len(availableTrucks) != expectedCount {
		t.Errorf("Expected %d available trucks, got %d", expectedCount, len(availableTrucks))
	}

	// Check that the right trucks are available
	truckNames := getTruckNames(availableTrucks)
	expectedTrucks := []string{"Tulip", "Andre350"}

	for _, expected := range expectedTrucks {
		found := false
		for _, actual := range truckNames {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected truck %s to be available, but it wasn't found", expected)
		}
	}

	// Verify Magnolia is NOT in the list
	for _, truckName := range truckNames {
		if truckName == "Magnolia" {
			t.Error("Magnolia should not be available (it's checked out)")
		}
	}
}

func getTruckNames(trucks []Truck) []string {
	names := make([]string, len(trucks))
	for i, truck := range trucks {
		names[i] = truck.Name
	}
	return names
}
