package handlers

import (
	"fmt"
	"strings"

	"log"
	"truck-checkout/models"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// releaseas a single vehicle based on its name
func HandleReleaseTruck(client *socketmode.Client, req *socketmode.Request, truckName string, userId string, userName string) {
	truckName = cases.Title(language.English).String(strings.ToLower(truckName))

	// Find the truck by name
	truck, err := models.GetTruckByName(truckName)
	if err != nil {
		client.Ack(*req, map[string]string{"text": fmt.Sprintf("❌ Truck `%s` not found.", truckName)})
		return
	}

	if !truck.IsCheckedOut {
		client.Ack(*req, map[string]string{"text": fmt.Sprintf("ℹ️ Truck `%s` is not currently checked out.", truckName)})
		return
	}

	checkout, err := models.GetActiveCheckoutByTruckID(truck.ID)
	if err != nil {
		log.Printf("Warning: Could not get checkout info for truck %s: %v", truckName, err)
		// Continue with release even if we can't get checkout info
	}

	err = models.ReleaseTruckFromCheckout(truck.ID, userId)
	if err != nil {
		log.Printf("Failed to release truck %s: %v", truckName, err)
		client.Ack(*req, map[string]string{"text": "❌ Failed to release the truck."})
		return
	}

	channelID := "vehicleupdates"
	var message string
	if checkout != nil {
		message = fmt.Sprintf("🚛 *%s* released truck *%s* (previously checked out by %s)", userName, truckName, checkout.UserName)
	} else {
		message = fmt.Sprintf("🚛 *%s* released truck *%s*", userName, truckName)
	}

	_, _, err = client.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Printf("Failed to post message to #vehicleupdates: %v", err)
	}

	client.Ack(*req, map[string]string{
		"text": fmt.Sprintf("✅ Truck `%s` has been released successfully!", truckName),
	})
}
