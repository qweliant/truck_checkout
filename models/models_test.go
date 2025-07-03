package models

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	db "truck-checkout/database"

	"github.com/google/uuid"
)

var TestDB *sql.DB

func TestMain(m *testing.M) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

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
	if truck.IsCheckedOut {
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
	truck.IsCheckedOut = false

	err = UpdateTruck(*truck)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	updated, err := GetTruckByName("Libby")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if updated.IsCheckedOut {
		t.Errorf("expected IsCheckedOut to be false, got true")
	}

	if *updated.DefaultTeam != "beltline" {
		t.Errorf("expected DefaultTeam to be 'beltline', got %s", *updated.DefaultTeam)
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
	availableTrucks, err := GetTrucksByCheckoutStatus(today, false)
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
		UserID:    uuid.New().String(),
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

func TestGetActiveCheckoutByTruckID(t *testing.T) {
	resetTestDB(t)

	// Create a truck
	team := "forest_restoration"
	err := InsertTruck("Magnolia", &team, uuid.New(), false)
	if err != nil {
		t.Fatalf("failed to insert truck: %v", err)
	}

	truck, err := GetTruckByName("Magnolia")
	if err != nil {
		t.Fatalf("failed to get truck: %v", err)
	}
	// No active checkout, should return error
	_, err = GetActiveCheckoutByTruckID(truck.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows when no checkout exists, got %v", err)
	}

	// Test 2: Create an active checkout
	now := time.Now()
	activeCheckout := Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    "user123",
		UserName:  "John Doe",
		TeamName:  "beltline",
		StartDate: now.Add(-1 * time.Hour), // Started 1 hour ago
		EndDate:   now.Add(2 * time.Hour),  // Ends in 2 hours
		Purpose:   "Active checkout test",
	}
	log.Printf("to be inserted checkout: %v", activeCheckout)

	err = InsertCheckout(activeCheckout)
	if err != nil {
		t.Fatalf("failed to insert checkout: %v", err)
	}

	// Should find the active checkout
	foundCheckout, err := GetActiveCheckoutByTruckID(truck.ID)
	log.Printf("Found checkout: %v", foundCheckout)
	if err != nil {
		t.Fatalf("failed to get active checkout: %v", err)
	}

	if foundCheckout.ID != activeCheckout.ID {
		t.Errorf("expected checkout ID %s, got %s", activeCheckout.ID, foundCheckout.ID)
	}
	if foundCheckout.UserName != activeCheckout.UserName {
		t.Errorf("expected user name %s, got %s", activeCheckout.UserName, foundCheckout.UserName)
	}
	if foundCheckout.Purpose != activeCheckout.Purpose {
		t.Errorf("expected purpose %s, got %s", activeCheckout.Purpose, foundCheckout.Purpose)
	}

	// Test 3: Create an expired checkout
	expiredCheckout := Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    "user456",
		UserName:  "Jane Smith",
		TeamName:  "floaters",
		StartDate: now.Add(-3 * time.Hour), // Started 3 hours ago
		EndDate:   now.Add(-1 * time.Hour), // Ended 1 hour ago
		Purpose:   "Expired checkout test",
	}

	err = InsertCheckout(expiredCheckout)
	if err != nil {
		t.Fatalf("failed to insert expired checkout: %v", err)
	}

	// Should still find the active checkout (not the expired one)
	foundCheckout, err = GetActiveCheckoutByTruckID(truck.ID)
	if err != nil {
		t.Fatalf("failed to get active checkout: %v", err)
	}

	if foundCheckout.ID != activeCheckout.ID {
		t.Errorf("expected to find active checkout, but got expired one")
	}
}
func TestReleaseTruckFromCheckout(t *testing.T) {
	resetTestDB(t)

	team := "forest_restoration"
	err := InsertTruck("Andre350", &team, uuid.New(), true) // Start as checked out
	if err != nil {
		t.Fatalf("failed to insert truck: %v", err)
	}

	truck, err := GetTruckByName("Andre350")
	if err != nil {
		t.Fatalf("failed to get truck: %v", err)
	}

	now := time.Now()
	checkout := Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    "user123",
		UserName:  "John Doe",
		TeamName:  "beltline",
		StartDate: now.Add(-1 * time.Hour), // Started 1 hour ago
		EndDate:   now.Add(8 * time.Hour),  // Should end in 8 hours
		Purpose:   "Test checkout for release",
	}

	err = InsertCheckout(checkout)
	if err != nil {
		t.Fatalf("failed to insert checkout: %v", err)
	}

	// Verify truck is checked out initially
	if !truck.IsCheckedOut {
		t.Fatal("truck should be checked out initially")
	}

	// Release the truck
	releasedBy := "admin123"
	err = ReleaseTruckFromCheckout(truck.ID, releasedBy)
	if err != nil {
		t.Fatalf("failed to release truck: %v", err)
	}

	// Verify truck is no longer checked out
	updatedTruck, err := GetTruckByName("Andre350")
	if err != nil {
		t.Fatalf("failed to get updated truck: %v", err)
	}

	if updatedTruck.IsCheckedOut {
		t.Error("truck should not be checked out after release")
	}

	// Verify checkout was updated (no longer active)
	_, err = GetActiveCheckoutByTruckID(truck.ID)
	if err != sql.ErrNoRows {
		t.Errorf("expected no active checkout after release, got %v", err)
	}

	// Verify checkout record still exists but with updated end_date
	var endDate time.Time
	var releasedByDB sql.NullString
	var releasedAtDB sql.NullTime

	err = db.DB.QueryRow(`
		SELECT end_date, released_by, released_at 
		FROM checkouts 
		WHERE id = ?
	`, checkout.ID.String()).Scan(&endDate, &releasedByDB, &releasedAtDB)

	if err != nil {
		t.Fatalf("failed to query checkout record: %v", err)
	}

	// end_date should be updated to current time (within reasonable margin)
	if time.Since(endDate) > 5*time.Second {
		t.Errorf("end_date should be updated to current time, got %v", endDate)
	}

	// Check audit fields (if your schema supports them)
	if releasedByDB.Valid && releasedByDB.String != releasedBy {
		t.Errorf("expected released_by to be %s, got %s", releasedBy, releasedByDB.String)
	}

	if releasedAtDB.Valid && time.Since(releasedAtDB.Time) > 5*time.Second {
		t.Errorf("released_at should be current time, got %v", releasedAtDB.Time)
	}
}

func TestReleaseTruckFromCheckout_InvalidTruckID(t *testing.T) {
	resetTestDB(t)

	fakeID := uuid.New()
	err := ReleaseTruckFromCheckout(fakeID, "admin123")

	if err != nil {
		t.Fatalf("release with invalid truck ID should not error: %v", err)
	}
}

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
