package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"

	"truck-checkout/models"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func HandleQuickCheckout(client *socketmode.Client, req *socketmode.Request, truckName string, userId string, userName string) {
	truckName = cases.Title(language.English).String(strings.ToLower(truckName))

	truck, err := models.GetTruckByName(truckName)
	if err != nil {
		client.Ack(*req, map[string]string{"text": fmt.Sprintf("❌ Truck `%s` not found.", truckName)})
		return
	}

	if truck.IsCheckedOut {
		client.Ack(*req, map[string]string{"text": fmt.Sprintf("🚫 Truck `%s` is already checked out.", truckName)})
		return
	}

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 7, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 15, 30, 0, 0, now.Location())

	checkout := models.Checkout{
		ID:        uuid.New(),
		TruckID:   truck.ID,
		UserID:    userId,
		UserName:  string(userName),
		TeamName:  *truck.DefaultTeam,
		StartDate: start,
		EndDate:   end,
		Purpose:   "Quick checkout via slash command",
	}

	if err := models.InsertCheckout(checkout); err != nil {
		log.Printf("InsertCheckout failed: %v", err)
		client.Ack(*req, map[string]string{
			"text": "❌ Could not check out the truck.",
		})
		return
	}
	truck.IsCheckedOut = true
	// Mark truck as checked out

	if err = models.UpdateTruck(*truck); err != nil {
		log.Printf("Truck status update error: %v", err)
	}

	// Send message to #vehicleupdates channel
	channelID := "C0944357V1A" // You might need to use the actual channel ID instead
	message := fmt.Sprintf("🚛 **%s** checked out truck **%s** (7:00 AM - 3:30 PM)", userName, truckName)

	_, _, err = client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to post message to #vehicleupdates: %v", err)
	}

	client.Ack(*req, map[string]string{
		"text": fmt.Sprintf("✅ Truck `%s` checked out from 7:00 AM to 3:30 PM today!", truckName),
	})
}
