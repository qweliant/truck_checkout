package main

import (
	"log"
	"os"
	db "truck-checkout/database"
)

func main() {
	dbPath := os.Getenv("DATABASE_URL")
	db.InitDB(dbPath)
	// calendarId = new uuid.UUID{} // Replace with actual calendar ID if needed
	// err := db.InsertTruck("Truck-1", nil, calendarId)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	log.Println("Seeded Truck-1")
}
