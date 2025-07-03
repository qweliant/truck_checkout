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

func TestInsertTruck(t *testing.T) {
	resetTestDB(t)
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
	resetTestDB(t)
	calendarID := uuid.New()
	team := "beltline"
	err := InsertTruck("BananaBoat", &team, calendarID, true)
	if err == nil {
		t.Fatal("expected error for invalid truck name")
	}
}
func TestUpdateTruck(t *testing.T) {
	resetTestDB(t)
	team := "floaters"
	calendarID := uuid.New()
	err := InsertTruck("Libby", &team, calendarID, true)
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
	truck.IsActive = false

	err = UpdateTruck(*truck)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := GetTruckByName("Libby")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if updated.IsActive {
		t.Errorf("expected IsActive to be false, got true")
	}

	if *updated.DefaultTeam != "beltline" {
		t.Errorf("expected DefaultTeam to be 'beltline', got %s", *updated.DefaultTeam)
	}
}

func TestCheckoutDatabaseOperations(t *testing.T) {
	resetTestDB(t)
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

func TestGetAvailableTrucksForToday(t *testing.T) {
	resetTestDB(t)

	// Create test trucks
	team1 := "forest_restoration"
	team2 := "beltline"

	// Create available trucks (active)
	err := InsertTruck("Tulip", &team1, uuid.New(), false)
	if err != nil {
		t.Fatalf("failed to insert truck Tulip: %v", err)
	}

	err = InsertTruck("Andre350", &team2, uuid.New(), false)
	if err != nil {
		t.Fatalf("failed to insert truck Andre350: %v", err)
	}

	// Create unavailable truck (checked out)
	err = InsertTruck("Magnolia", &team1, uuid.New(), true)
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
		UserID:    uuid.New(),
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
	availableTrucks, err := GetAvailableTrucksForToday(today)
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

// func TestGetAvailableTrucksForDifferentDay(t *testing.T) {
// 	resetTestDB(t)

// 	// Create test truck
// 	team := "test_team"
// 	err := InsertTruck("TestTruck", &team, uuid.New(), true)
// 	if err != nil {
// 		t.Fatalf("failed to insert truck: %v", err)
// 	}

// 	truck, err := GetTruckByName("TestTruck")
// 	if err != nil {
// 		t.Fatalf("failed to get truck: %v", err)
// 	}

// 	// Create checkout for tomorrow
// 	tomorrow := time.Now().Add(24 * time.Hour)
// 	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, tomorrow.Location())
// 	endOfTomorrow := startOfTomorrow.Add(8 * time.Hour)

// 	checkout := Checkout{
// 		ID:        uuid.New(),
// 		TruckID:   truck.ID,
// 		UserID:    uuid.New(),
// 		UserName:  "Test User",
// 		TeamName:  "test_team",
// 		StartDate: startOfTomorrow,
// 		EndDate:   endOfTomorrow,
// 		Purpose:   "Tomorrow's checkout",
// 	}

// 	err = InsertCheckout(checkout)
// 	if err != nil {
// 		t.Fatalf("failed to insert checkout: %v", err)
// 	}

// 	// Test: Get available trucks for today (should be available since checkout is tomorrow)
// 	today := time.Now()
// 	availableTrucks, err := GetAvailableTrucksForToday(today)
// 	if err != nil {
// 		t.Fatalf("failed to get available trucks: %v", err)
// 	}

// 	if len(availableTrucks) != 1 {
// 		t.Errorf("Expected 1 available truck for today, got %d", len(availableTrucks))
// 	}

// 	// Test: Get available trucks for tomorrow (should be unavailable)
// 	availableTrucksTomorrow, err := GetAvailableTrucksForToday(tomorrow)
// 	if err != nil {
// 		t.Fatalf("failed to get available trucks for tomorrow: %v", err)
// 	}

// 	if len(availableTrucksTomorrow) != 0 {
// 		t.Errorf("Expected 0 available trucks for tomorrow, got %d", len(availableTrucksTomorrow))
// 	}
// }

// Helper function to extract truck names
func getTruckNames(trucks []Truck) []string {
	names := make([]string, len(trucks))
	for i, truck := range trucks {
		names[i] = truck.Name
	}
	return names
}

func resetTestDB(t *testing.T) {
	_, err := db.DB.Exec(`DELETE FROM checkouts; DELETE FROM trucks;`)
	if err != nil {
		t.Fatalf("failed to reset database: %v", err)
	}
}
