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
	log.Println("🟢 Database connection established.")

	// --- Data Cleanup ---
	log.Println("🗑️  Clearing existing data...")
	_, err := db.DB.Exec(`DELETE FROM checkouts; DELETE FROM trucks; DELETE FROM users;`)
	if err != nil {
		log.Fatalf("❌ Failed to reset database: %v", err)
	}
	log.Println("🟢 All tables cleared.")

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
		err := models.InsertTruck(truckData.Name, &truckData.DefaultTeam, calendarID, false) // Initially, all trucks are available.
		if err != nil {
			log.Fatalf("❌ Failed to insert truck %s: %v", truckData.Name, err)
		}
		log.Printf("   🚚 Inserted truck: %s (Default Team: %s) 🟢 AVAILABLE", truckData.Name, truckData.DefaultTeam)
	}
	log.Println("🟢 Trucks seeded successfully.")

	// --- Seed Checkouts ---
	log.Println("🚛 Seeding example checkouts for Tulip and Libby...")

	// --- Seed User ---
	testUserSlackID := "U000SEEDER" // Use a fake but valid-looking Slack ID
	testUserName := "Seeder McSeedface"
	testUserTeam := "seeders" // Arbitrary team, since team doesn't restrict checkout

	user, err := models.CreateUser(testUserSlackID, testUserName, testUserTeam)
	if err != nil {
		log.Fatalf("❌ Failed to create seed user: %v", err)
	}
	log.Printf("   👤 Created seed user: %s (ID: %s, Team: %s)", testUserName, testUserSlackID, testUserTeam)
	userID := user.ID
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
		if err = models.CreateCheckout(checkout); err != nil {
			log.Fatalf("❌ Failed to insert checkout for %s: %v", truckName, err)
		}

		// Update the truck's availability status to checked out.
		truck.IsCheckedOut = true
		if err = models.UpdateTruck(*truck); err != nil {
			log.Fatalf("❌ Failed to update availability for truck %s: %v", truckName, err)
		}

		log.Printf("   🟡 Checked out %s, status: UNAVAILABLE", truckName)
	}

	log.Println("🌳 Seed complete. 🟢")
}
