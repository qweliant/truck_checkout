package models

import (
	"database/sql"
	"strings"
	"testing"
	"time"
	db "truck-checkout/database"

	"github.com/google/uuid"
)

func TestCheckoutDatabaseOperations(t *testing.T) {
	ResetTestDB(t)
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
	ResetTestDB(t)

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

	err = InsertCheckout(activeCheckout)
	if err != nil {
		t.Fatalf("failed to insert checkout: %v", err)
	}

	// Should find the active checkout
	foundCheckout, err := GetActiveCheckoutByTruckID(truck.ID)
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
	ResetTestDB(t)

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
	ResetTestDB(t)

	fakeID := uuid.New()
	err := ReleaseTruckFromCheckout(fakeID, "admin123")

	if err == nil {
		t.Fatal("expected error when releasing checkout for non-existent truck")
	}

	expectedMsg := "no active checkout found for truck"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing '%s', got: %v", expectedMsg, err)
	}
}
