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

	log.Println("🌱 Seeding trucks with assigned default teams...")

	truckSeedData := []struct {
		Name        string
		DefaultTeam string
	}{
		{"Libby", "floaters"},
		{"Tulip", "beltline"},
		{"Watson", "forest_restoration"}, // update if a better default team becomes clear
		{"Andre350", "urban_trees"},
	}

	truckIDs := make(map[string]uuid.UUID)

	for _, truck := range truckSeedData {
		truckID := uuid.New()
		calendarID := uuid.New()

		err := models.InsertTruck(truck.Name, &truck.DefaultTeam, calendarID, false)
		if err != nil {
			log.Fatalf("❌ Failed to insert truck %s: %v", truck.Name, err)
		}

		truckIDs[truck.Name] = truckID
		log.Printf("✅ Inserted truck: %s (%s)", truck.Name, truck.DefaultTeam)
	}

	log.Println("🚛 Seeding 2 checkouts for Tulip and Libby...")

	userID := uuid.New()
	start := time.Now()
	end := start.Add(8 * time.Hour)

	checkouts := []struct {
		TruckName string
		Team      string
	}{
		{"Tulip", "beltline"},
		{"Libby", "floaters"},
	}

	//set the trucks libby and tulip to have isActive = true
	for _, truckName := range []string{"Libby", "Tulip"} {
		truck, err := models.GetTruckByName(truckName)
		if err != nil {
			log.Fatalf("❌ Failed to get truck %s: %v", truckName, err)
		}
		truck.IsActive = true
		err = models.UpdateTruck(*truck)
		if err != nil {
			log.Fatalf("❌ Failed to update truck %s: %v", truckName, err)
		}
		log.Printf("✅ Updated truck %s to active", truckName)
	}

	for _, entry := range checkouts {
		checkout := models.Checkout{
			ID:              uuid.New(),
			TruckID:         truckIDs[entry.TruckName],
			UserID:          userID,
			UserName:        "Seeder McSeedface",
			TeamName:        entry.Team,
			StartDate:       start,
			EndDate:         end,
			Purpose:         "Seeding test run",
			CalendarEventID: uuid.New().String(),
			CreatedAt:       time.Now(),
		}

		err := models.InsertCheckout(checkout)
		if err != nil {
			log.Fatalf("❌ Failed to checkout %s: %v", entry.TruckName, err)
		}
		log.Printf("🟡 Checked out %s (%s)", entry.TruckName, entry.Team)
	}

	log.Println("🌳 Seed complete.")
}
