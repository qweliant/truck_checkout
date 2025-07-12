package models

import (
	"testing"
	db "truck-checkout/database"
)

// ResetTestDB clears all data from the 'checkouts' and 'trucks' tables in the test database.
// It is intended to be used in test setup or teardown to ensure a clean database state.
// If the operation fails, the test is immediately failed with a fatal error.
func ResetTestDB(t *testing.T) {
	if db.DB == nil {
		t.Fatal("db.DB is nil in ResetTestDB")
	}
	_, err := db.DB.Exec(`DELETE FROM checkouts; DELETE FROM trucks; DELETE FROM users;`)
	if err != nil {
		t.Fatalf("failed to reset database: %v", err)
	}
}
