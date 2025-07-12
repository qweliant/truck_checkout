package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	db "truck-checkout/database"

	"github.com/google/uuid"
)

type User struct {
	ID          string    `json:"id"`
	SlackUserID string    `json:"slack_user_id"`
	Username    string    `json:"username"`
	Team        string    `json:"team"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetUserBySlackID(slackUserID string) (*User, error) {
	query := `
		SELECT id, slack_user_id, username, team, created_at 
		FROM users 
		WHERE slack_user_id = ?
	`

	var user User
	err := db.DB.QueryRow(query, slackUserID).Scan(
		&user.ID,
		&user.SlackUserID,
		&user.Username,
		&user.Team,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found, return nil without error
		}
		return nil, err
	}

	return &user, nil
}

func CreateUser(slackUserID, username, team string) (*User, error) {
	if strings.TrimSpace(slackUserID) == "" {
		return nil, fmt.Errorf("slack_user_id cannot be empty")
	}
	if strings.TrimSpace(username) == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if strings.TrimSpace(team) == "" {
		return nil, fmt.Errorf("team cannot be empty")
	}
	user := User{
		ID:          uuid.New().String(),
		SlackUserID: slackUserID,
		Username:    username,
		Team:        team,
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO users (id, slack_user_id, username, team, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := db.DB.Exec(query, user.ID, user.SlackUserID, user.Username, user.Team, user.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(user User) error {
	query := `
		UPDATE users 
		SET username = ?, team = ?
		WHERE slack_user_id = ?
	`

	_, err := db.DB.Exec(query, user.Username, user.Team, user.SlackUserID)
	return err
}

func GetAllUsers() ([]User, error) {
	query := `
		SELECT id, slack_user_id, username, team, created_at 
		FROM users 
		ORDER BY username
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.SlackUserID,
			&user.Username,
			&user.Team,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}
