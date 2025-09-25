package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	if err := CreateTables(DB); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}
}

func CreateTables(database *sql.DB) error {
	userSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		slack_user_id TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		team TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	truckSQL := `
	CREATE TABLE IF NOT EXISTS trucks (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL UNIQUE,
		default_team TEXT,
		calendar_id TEXT,
		is_checked_out BOOLEAN DEFAULT FALSE
	);`

	checkoutSQL := `
	CREATE TABLE IF NOT EXISTS checkouts (
		id TEXT PRIMARY KEY,
		truck_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		user_name TEXT NOT NULL,
		team_name TEXT NOT NULL,
		start_date DATETIME NOT NULL,
		end_date DATETIME NOT NULL,
		purpose TEXT,
		calendar_event_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		released_by TEXT,
		released_at TIMESTAMP,
		FOREIGN KEY(truck_id) REFERENCES trucks(id)
	);	
	`

	if _, err := database.Exec(userSQL); err != nil {
		return err
	}
	if _, err := database.Exec(truckSQL); err != nil {
		return err
	}
	if _, err := database.Exec(checkoutSQL); err != nil {
		return err
	}
	return nil
}
