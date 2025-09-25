package main

import (
	"log"
	"os"
	"time"

	db "truck-checkout/internal/database"
	"truck-checkout/internal/models"

	"github.com/google/uuid"
)

func main() {
	// Determine the database path, with a fallback for local development.
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "truckbot.db" // Default database file name
	}

	// Initialize the database connection.
	log.Printf("🔌 Connecting to the database... %s", dbPath)
	db.InitDB(dbPath)
	log.Println("✅ Database connection established.")

	// --- Data Cleanup ---
	log.Println("🗑️  Clearing existing data...")
	if _, err := db.DB.Exec("DELETE FROM checkouts"); err != nil {
		log.Fatalf("❌ Failed to clear checkouts table: %v", err)
	}
	if _, err := db.DB.Exec("DELETE FROM trucks"); err != nil {
		log.Fatalf("❌ Failed to clear trucks table: %v", err)
	}
	log.Println("✅ All tables cleared.")

	// --- Seed Trucks ---
	log.Println("🌱 Seeding trucks with assigned default teams...")
	truckSeedData := []struct {
		Name        string
		DefaultTeam string
	}{
		{"Libby", "floaters"},
		{"Tulip", "beltline"},
		{"Watson", "forest_restoration"},
		{"Andre350", "urban_trees"},
	}

	for _, truckData := range truckSeedData {
		// A unique calendar ID for each truck is good practice for Phase 2.
		calendarID := uuid.New()

		// Insert the truck with its default state. IsAvailable is false by default.
		err := models.InsertTruck(truckData.Name, &truckData.DefaultTeam, calendarID, true) // Initially, all trucks are available.
		if err != nil {
			log.Fatalf("❌ Failed to insert truck %s: %v", truckData.Name, err)
		}
		log.Printf("   🚚 Inserted truck: %s (Default Team: %s)", truckData.Name, truckData.DefaultTeam)
	}
	log.Println("✅ Trucks seeded successfully.")

	// --- Seed Checkouts ---
	log.Println("🚛 Seeding example checkouts for Tulip and Libby...")

	// Define common details for the seeded checkouts.
	userID := uuid.New().String()
	start := time.Now()
	end := start.Add(8 * time.Hour) // A standard 8-hour checkout

	checkoutTrucks := []string{"Libby", "Tulip"}

	for _, truckName := range checkoutTrucks {
		// Retrieve the truck from the database to ensure we have the correct ID.
		truck, err := models.GetTruckByName(truckName)
		if err != nil {
			log.Fatalf("❌ Failed to retrieve truck %s for checkout: %v", truckName, err)
		}

		// Create the checkout record.
		checkout := models.Checkout{
			ID:              uuid.New(),
			TruckID:         truck.ID, // Use the actual truck ID from the database.
			UserID:          userID,
			UserName:        "Seeder McSeedface",
			TeamName:        *truck.DefaultTeam,
			StartDate:       start,
			EndDate:         end,
			Purpose:         "Automated seed data",
			CalendarEventID: uuid.New().String(), // Placeholder for Phase 2.
			CreatedAt:       time.Now(),
		}

		// Insert the checkout record into the database.
		if err = models.InsertCheckout(checkout); err != nil {
			log.Fatalf("❌ Failed to insert checkout for %s: %v", truckName, err)
		}

		// When a truck is checked out, its IsAvailable status must be set to false.
		truck.IsCheckedOut = false
		if err = models.UpdateTruck(*truck); err != nil {
			log.Fatalf("❌ Failed to update availability for truck %s: %v", truckName, err)
		}

		log.Printf("   🟡 Checked out %s, status set to UNAVAILABLE.", truckName)
	}

	log.Println("🌳 Seed complete.")
}
