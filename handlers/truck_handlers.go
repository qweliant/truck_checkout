package handlers

import (
	"fmt"
	"time"
	"truck-checkout/models"

	"github.com/slack-go/slack/socketmode"
)

func HandleTrucksAvailable(client *socketmode.Client, req *socketmode.Request) {
	trucks, err := models.GetTrucksByCheckoutStatus(time.Now(), false)
	if err != nil {
		client.Ack(*req, map[string]string{"text": "❌ Could not retrieve available trucks."})
		return
	}
	if len(trucks) == 0 {
		client.Ack(*req, map[string]string{"text": "🚫 No trucks are currently available today."})
		return
	}

	msg := "🟢 *Available Trucks Today:*\n"
	for _, t := range trucks {
		team := "unassigned"
		if t.DefaultTeam != nil {
			team = *t.DefaultTeam
		}
		msg += fmt.Sprintf("• %s (%s)\n", t.Name, team)
	}

	client.Ack(*req, map[string]string{"text": msg})
}

func HandleTrucksCheckedOut(client *socketmode.Client, req *socketmode.Request) {
	trucks, err := models.GetTrucksByCheckoutStatus(time.Now(), true)
	if err != nil {
		client.Ack(*req, map[string]string{"text": "❌ Could not retrieve checked-out trucks."})
		return
	}
	if len(trucks) == 0 {
		client.Ack(*req, map[string]string{"text": "✅ All trucks are currently available!"})
		return
	}

	msg := "🔴 *Checked Out Trucks Today:*\n"
	for _, t := range trucks {
		team := "unassigned"
		if t.DefaultTeam != nil {
			team = *t.DefaultTeam
		}
		msg += fmt.Sprintf("• %s (%s)\n", t.Name, team)
	}

	client.Ack(*req, map[string]string{"text": msg})
}
