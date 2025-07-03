package main

import (
	"log"
	"os"
	"time"

	db "truck-checkout/database"
	"truck-checkout/models"

	"github.com/google/uuid"
)

func main() {
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "truckbot.db" // fallback
	}

	db.InitDB(dbPath)

	log.Println("🗑️ Clearing existing data...")

	// Clear existing data
	_, err := db.DB.Exec("DELETE FROM checkouts")
	if err != nil {
		log.Fatalf("❌ Failed to clear checkouts: %v", err)
	}

	_, err = db.DB.Exec("DELETE FROM trucks")
	if err != nil {
		log.Fatalf("❌ Failed to clear trucks: %v", err)
	}

	log.Println("✅ Database cleared")

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

	// Insert trucks and get their actual IDs
	for _, truck := range truckSeedData {
		calendarID := uuid.New()

		err := models.InsertTruck(truck.Name, &truck.DefaultTeam, calendarID, false)
		if err != nil {
			log.Fatalf("❌ Failed to insert truck %s: %v", truck.Name, err)
		}

		log.Printf("✅ Inserted truck: %s (%s)", truck.Name, truck.DefaultTeam)
	}

	log.Println("🚛 Seeding 2 checkouts for Tulip and Libby...")

	userID := uuid.New()
	start := time.Now()
	end := start.Add(8 * time.Hour)

	checkoutTrucks := []string{"Libby", "Tulip"}

	for _, truckName := range checkoutTrucks {
		// Get the actual truck from the database (with its real ID)
		truck, err := models.GetTruckByName(truckName)
		if err != nil {
			log.Fatalf("❌ Failed to get truck %s: %v", truckName, err)
		}

		// Create checkout with the real truck ID
		checkout := models.Checkout{
			ID:              uuid.New(),
			TruckID:         truck.ID, // Use the actual truck ID from database
			UserID:          userID.String(),
			UserName:        "Seeder McSeedface",
			TeamName:        *truck.DefaultTeam,
			StartDate:       start,
			EndDate:         end,
			Purpose:         "Seeding test run",
			CalendarEventID: uuid.New().String(),
			CreatedAt:       time.Now(),
		}

		err = models.InsertCheckout(checkout)
		if err != nil {
			log.Fatalf("❌ Failed to checkout %s: %v", truckName, err)
		}

		// Mark truck as checked out
		truck.IsCheckedOut = true
		err = models.UpdateTruck(*truck)
		if err != nil {
			log.Fatalf("❌ Failed to update truck %s: %v", truckName, err)
		}

		log.Printf("🟡 Checked out %s (%s)", truckName, *truck.DefaultTeam)
	}

	log.Println("🌳 Seed complete.")
}
