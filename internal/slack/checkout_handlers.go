package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"truck-checkout/internal/models"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Helper function to check if a day is a business day (Monday-Friday)
func isValidDay(t time.Time) bool {
	weekday := t.Weekday()
	return weekday >= time.Monday && weekday <= time.Saturday
}

// Helper function to add business days to a date
func addBusinessDays(start time.Time, businessDays int) time.Time {
	current := start
	daysAdded := 0

	for daysAdded < businessDays {
		current = current.AddDate(0, 0, 1)
		if isValidDay(current) {
			daysAdded++
		}
	}

	return current
}

// Helper function to calculate the end date for a multi-day checkout
func calculateEndDate(startDate time.Time, businessDays int) time.Time {
	if businessDays == 1 {
		// Single day checkout - end same day at 3:30 PM
		return time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 15, 30, 0, 0, startDate.Location())
	}

	// Multi-day checkout - find the last business day and set end time to 3:30 PM
	endDate := addBusinessDays(startDate, businessDays-1) // -1 because start day counts as day 1
	return time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 15, 30, 0, 0, endDate.Location())
}

// Helper function to format date range for display
func formatDateRange(start, end time.Time) string {
	if start.Year() == end.Year() && start.Month() == end.Month() && start.Day() == end.Day() {
		// Same day
		return fmt.Sprintf("%s (7:00 AM - 3:30 PM)", start.Format("Jan 2, 2006"))
	}
	// Different days
	return fmt.Sprintf("%s 7:00 AM - %s 3:30 PM", start.Format("Jan 2"), end.Format("Jan 2, 2006"))
}

func performCheckout(client *socketmode.Client, user *models.User, truckName string, businessDays int, slackUserId string, userName string) (string, error) {
	truck, err := models.GetTruckByName(truckName)
	if err != nil {
		return "", fmt.Errorf("❌ Truck `%s` not found", truckName)
	}

	if truck.IsCheckedOut {
		return "", fmt.Errorf("🚫 Truck `%s` is already checked out", truckName)
	}

	if truck.DefaultTeam != nil && user.Team != *truck.DefaultTeam {
		// TODO: Show cross-team warning
		// showCrossTeamWarning(client, req, user, truck, truckName, businessDays)
		return "", fmt.Errorf("⚠️ Warning: %s is typically used by %s team, but you're on %s team. Cross-team warning will be implemented next", truckName)
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())
	end := calculateEndDate(start, businessDays)

	checkout := models.Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    slackUserId,
		UserName:  string(userName),
		TeamName:  user.Team, // Use the user's actual team
		StartDate: start,
		EndDate:   end,
		Purpose:   fmt.Sprintf("Quick checkout via slash command (%d business days)", businessDays),
	}

	if err := models.InsertCheckout(checkout); err != nil {
		log.Printf("InsertCheckout failed: %v", err)
		return "", fmt.Errorf("❌ Could not check out the truck")
	}
	truck.IsCheckedOut = true

	if err = models.UpdateTruck(*truck); err != nil {
		log.Printf("Truck status update error: %v", err)
	}

	// Send message to #vehicleupdates channel
	channelID := "vehicleupdates"
	dateRange := formatDateRange(start, end)
	message := fmt.Sprintf("🚛 *%s* checked out truck *%s* (%s)", userName, truckName, dateRange)

	_, _, err = client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to post message to #vehicleupdates: %v", err)
		return "", fmt.Errorf("❌ Could not post update to #vehicleupdates channel")
	}

	var responseText string
	if businessDays == 1 {
		responseText = fmt.Sprintf("✅ Truck `%s` checked out for today (7:00 AM - 3:30 PM)!", truckName)
	} else {
		responseText = fmt.Sprintf("✅ Truck `%s` checked out for %d business days (%s)!", truckName, businessDays, dateRange)
	}

	return responseText, nil
}

func HandleCheckout(client *socketmode.Client, req *socketmode.Request, truckName string, businessDays int, slackUserId string, userName string, triggerId string) {
	truckName = cases.Title(language.English).String(strings.ToLower(truckName))
	_, err := models.GetTruckByName(truckName)
	if err != nil {
		client.Ack(*req, map[string]string{"text": fmt.Sprintf("❌ Truck `%s` not found.", truckName)})
		return
	}

	user, err := models.GetUserBySlackID(slackUserId)
	if err != nil {
		client.Ack(*req, map[string]string{"text": "❌ Error retrieving user information."})
		return
	}

	if user == nil {
		showTeamSelectionModal(client, req, triggerId, truckName, businessDays, slackUserId, userName)
		client.Ack(*req, map[string]string{"text": "👋 Please select your team to continue with checkout."})
		return
	}

	responseText, err := performCheckout(client, user, truckName, businessDays, slackUserId, userName)
	if err != nil {
		log.Printf("Checkout error: %v", err)
		client.Ack(*req, map[string]string{"text": err.Error()})
		return
	}

	client.Ack(*req, map[string]string{"text": responseText})
}
