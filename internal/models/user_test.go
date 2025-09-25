package models

import (
	"testing"
)

func TestGetUserBySlackID(t *testing.T) {
	// Test case 1: User not found
	t.Run("UserNotFound", func(t *testing.T) {
		ResetTestDB(t)

		user, err := GetUserBySlackID("nonexistent")
		if err != nil {
			t.Errorf("Expected no error for non-existent user, got: %v", err)
		}
		if user != nil {
			t.Errorf("Expected nil user for non-existent user, got: %v", user)
		}
	})

	// Test case 2: Create user and then find them
	t.Run("UserFound", func(t *testing.T) {
		ResetTestDB(t)

		// Create a test user
		createdUser, err := CreateUser("U123456", "testuser", "forest_restoration")
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}

		// Try to find the user
		foundUser, err := GetUserBySlackID("U123456")
		if err != nil {
			t.Errorf("Expected no error when finding user, got: %v", err)
		}
		if foundUser == nil {
			t.Fatal("Expected to find user, got nil")
		}

		// Verify user data
		if foundUser.SlackUserID != "U123456" {
			t.Errorf("Expected SlackUserID 'U123456', got '%s'", foundUser.SlackUserID)
		}
		if foundUser.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", foundUser.Username)
		}
		if foundUser.Team != "forest_restoration" {
			t.Errorf("Expected Team 'forest_restoration', got '%s'", foundUser.Team)
		}
		if foundUser.ID != createdUser.ID {
			t.Errorf("Expected ID to match created user ID")
		}
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("ValidUser", func(t *testing.T) {
		ResetTestDB(t)

		user, err := CreateUser("U789012", "newuser", "road_maintenance")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Verify user was created correctly
		if user.SlackUserID != "U789012" {
			t.Errorf("Expected SlackUserID 'U789012', got '%s'", user.SlackUserID)
		}
		if user.Username != "newuser" {
			t.Errorf("Expected Username 'newuser', got '%s'", user.Username)
		}
		if user.Team != "road_maintenance" {
			t.Errorf("Expected Team 'road_maintenance', got '%s'", user.Team)
		}
		if user.ID == "" {
			t.Error("Expected non-empty ID")
		}
		if user.CreatedAt.IsZero() {
			t.Error("Expected non-zero CreatedAt timestamp")
		}

		// Verify user was actually inserted into database
		foundUser, err := GetUserBySlackID("U789012")
		if err != nil {
			t.Errorf("Failed to retrieve created user: %v", err)
		}
		if foundUser == nil {
			t.Error("Created user not found in database")
		}
	})

	t.Run("DuplicateSlackUserID", func(t *testing.T) {
		ResetTestDB(t)

		// Create first user
		_, err := CreateUser("U111111", "user1", "team1")
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create second user with same slack_user_id
		_, err = CreateUser("U111111", "user2", "team2")
		if err == nil {
			t.Error("Expected error when creating user with duplicate slack_user_id")
		}
	})

	t.Run("EmptyValues", func(t *testing.T) {
		ResetTestDB(t)

		// Test with empty slack_user_id
		_, err := CreateUser("", "username", "team")
		if err == nil {
			t.Error("Expected error when creating user with empty slack_user_id")
		}

		// Test with empty username
		_, err = CreateUser("U222222", "", "team")
		if err == nil {
			t.Error("Expected error when creating user with empty username")
		}

		// Test with empty team
		_, err = CreateUser("U333333", "username", "")
		if err == nil {
			t.Error("Expected error when creating user with empty team")
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("ValidUpdate", func(t *testing.T) {
		ResetTestDB(t)

		// Create a user
		user, err := CreateUser("U444444", "originaluser", "originalteam")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// Update the user
		user.Username = "updateduser"
		user.Team = "updatedteam"
		err = UpdateUser(*user)
		if err != nil {
			t.Errorf("Failed to update user: %v", err)
		}

		// Verify the update
		updatedUser, err := GetUserBySlackID("U444444")
		if err != nil {
			t.Errorf("Failed to retrieve updated user: %v", err)
		}
		if updatedUser.Username != "updateduser" {
			t.Errorf("Expected updated username 'updateduser', got '%s'", updatedUser.Username)
		}
		if updatedUser.Team != "updatedteam" {
			t.Errorf("Expected updated team 'updatedteam', got '%s'", updatedUser.Team)
		}
		// ID and CreatedAt should remain unchanged
		if updatedUser.ID != user.ID {
			t.Error("User ID should not change during update")
		}
		if !updatedUser.CreatedAt.Equal(user.CreatedAt) {
			t.Error("CreatedAt should not change during update")
		}
	})

	t.Run("NonexistentUser", func(t *testing.T) {
		ResetTestDB(t)

		nonexistentUser := User{
			SlackUserID: "U999999",
			Username:    "ghost",
			Team:        "phantom",
		}
		err := UpdateUser(nonexistentUser)
		// This should not return an error in SQLite (it just affects 0 rows)
		if err != nil {
			t.Errorf("Unexpected error when updating nonexistent user: %v", err)
		}
	})
}

func TestGetAllUsers(t *testing.T) {
	t.Run("EmptyDatabase", func(t *testing.T) {
		ResetTestDB(t)

		users, err := GetAllUsers()
		if err != nil {
			t.Errorf("Failed to get all users from empty database: %v", err)
		}
		if len(users) != 0 {
			t.Errorf("Expected 0 users in empty database, got %d", len(users))
		}
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		ResetTestDB(t)

		// Create multiple users
		testUsers := []struct {
			slackID  string
			username string
			team     string
		}{
			{"U111", "alice", "forest_restoration"},
			{"U222", "bob", "road_maintenance"},
			{"U333", "charlie", "forest_restoration"},
		}

		for _, tu := range testUsers {
			_, err := CreateUser(tu.slackID, tu.username, tu.team)
			if err != nil {
				t.Fatalf("Failed to create test user %s: %v", tu.username, err)
			}
		}

		// Get all users
		users, err := GetAllUsers()
		if err != nil {
			t.Errorf("Failed to get all users: %v", err)
		}

		if len(users) != len(testUsers) {
			t.Errorf("Expected %d users, got %d", len(testUsers), len(users))
		}

		// Verify users are sorted by username (alice, bob, charlie)
		expectedOrder := []string{"alice", "bob", "charlie"}
		for i, expectedUsername := range expectedOrder {
			if i >= len(users) {
				t.Errorf("Missing user at index %d", i)
				continue
			}
			if users[i].Username != expectedUsername {
				t.Errorf("Expected user %d to be '%s', got '%s'", i, expectedUsername, users[i].Username)
			}
		}
	})
}

func TestUserModelIntegration(t *testing.T) {
	t.Run("CompleteUserLifecycle", func(t *testing.T) {
		ResetTestDB(t)

		slackID := "U555555"

		// 1. User should not exist initially
		user, err := GetUserBySlackID(slackID)
		if err != nil {
			t.Errorf("Unexpected error checking for non-existent user: %v", err)
		}
		if user != nil {
			t.Error("User should not exist initially")
		}

		// 2. Create user
		_, err = CreateUser(slackID, "lifecycle_user", "test_team")
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		// 3. Verify user can be found
		foundUser, err := GetUserBySlackID(slackID)
		if err != nil {
			t.Errorf("Failed to find created user: %v", err)
		}
		if foundUser == nil {
			t.Fatal("Created user not found")
		}

		// 4. Update user
		foundUser.Username = "updated_lifecycle_user"
		foundUser.Team = "updated_team"
		err = UpdateUser(*foundUser)
		if err != nil {
			t.Errorf("Failed to update user: %v", err)
		}

		// 5. Verify update
		finalUser, err := GetUserBySlackID(slackID)
		if err != nil {
			t.Errorf("Failed to retrieve updated user: %v", err)
		}
		if finalUser.Username != "updated_lifecycle_user" {
			t.Errorf("Username not updated correctly")
		}
		if finalUser.Team != "updated_team" {
			t.Errorf("Team not updated correctly")
		}

		// 6. Verify user appears in GetAllUsers
		allUsers, err := GetAllUsers()
		if err != nil {
			t.Errorf("Failed to get all users: %v", err)
		}
		found := false
		for _, u := range allUsers {
			if u.SlackUserID == slackID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Updated user not found in GetAllUsers result")
		}
	})
}
