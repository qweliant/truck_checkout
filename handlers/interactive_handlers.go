package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"truck-checkout/models"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func showTeamSelectionModal(client *socketmode.Client, req *socketmode.Request, triggerID string, truckName string, businessDays int, userId string, userName string) {
	// Create options for team selection
	var options []*slack.OptionBlockObject
	for _, team := range models.ValidTeams {
		// Make team names more readable (replace underscores with spaces)
		displayName := strings.ReplaceAll(team, "_", " ")
		displayName = cases.Title(language.English).String(displayName)

		options = append(options, slack.NewOptionBlockObject(
			team, // value (what gets submitted)
			slack.NewTextBlockObject("plain_text", displayName, true, false), // display text
			nil,
		))
	}

	// Store checkout parameters in metadata so we can retrieve them later
	metadata := fmt.Sprintf("%s|%d|%s|%s", truckName, businessDays, userId, userName)

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType("modal"),
		Title:           slack.NewTextBlockObject("plain_text", "Select Your Team", true, false),
		Close:           slack.NewTextBlockObject("plain_text", "Cancel", true, false),
		Submit:          slack.NewTextBlockObject("plain_text", "Continue Checkout", true, false),
		CallbackID:      "team_selection", // We'll use this to identify the modal response
		PrivateMetadata: metadata,
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn",
						fmt.Sprintf("👋 Welcome! To checkout *%s*, please select your team:", truckName), true, false),
					nil, nil,
				),
				slack.NewInputBlock(
					"team_block", // blockID
					slack.NewTextBlockObject("plain_text", "Your Team", true, false),                     // label
					slack.NewTextBlockObject("plain_text", "Select the team you belong to", true, false), // hint
					slack.NewOptionsSelectBlockElement(
						"static_select",
						slack.NewTextBlockObject("plain_text", "Choose your team...", true, false),
						"team_select",
						options...,
					),
				),
			},
		},
	}

	// Show the modal
	_, err := client.OpenView(triggerID, modalRequest)
	if err != nil {
		log.Printf("Failed to open team selection modal: %v", err)
		client.Ack(*req, map[string]string{
			"text": "❌ Error showing team selection. Please try again.",
		})
		return
	}

	// Acknowledge the slash command (modal is now open)
	client.Ack(*req, map[string]string{})
}

func handleButtonActions(client *socketmode.Client, req *socketmode.Request, callback *slack.InteractionCallback) {
	if len(callback.ActionCallback.BlockActions) == 0 {
		client.Ack(*req)
		return
	}

	action := callback.ActionCallback.BlockActions[0]

	switch action.ActionID {
	case "ask_permission":
		// Handle "Ask in #vehicleupdates" button
	case "continue_anyway":
		// Handle "Continue anyway" button
	}

	client.Ack(*req)
}


func handleTeamSelectionModal(client *socketmode.Client, req *socketmode.Request, callback *slack.InteractionCallback) {
	teamValue := callback.View.State.Values["team_block"]["team_select"].SelectedOption.Value
	metadata := callback.View.PrivateMetadata
	parts := strings.Split(metadata, "|")
	if len(parts) != 4 {
		client.Ack(*req, map[string]string{
			"text": "❌ Error processing team selection.",
		})
		return
	}

	truckName := parts[0]
	businessDays, _ := strconv.Atoi(parts[1])
	userId := parts[2]
	userName := parts[3]

	log.Printf("User %s selected team %s for truck %s", userName, teamValue, truckName)

	user, err := models.CreateUser(userId, userName, teamValue)
	if err != nil {
		log.Printf("Failed to create user %s (%s) with team %s: %v", userName, userId, teamValue, err)
		client.Ack(*req, map[string]string{
			"text": "❌ Error creating user profile. Please try again.",
		})
		return
	}

	log.Printf("Created new user %s (%s) with team %s for truck %s", userName, userId, teamValue, truckName)

	responseText, err := performCheckout(client, user, truckName, businessDays, userId, userName)
	if err != nil {
		log.Printf("Checkout error: %v", err)
		client.Ack(*req, map[string]string{"text": err.Error()})
		return
	}

	combinedMessage := fmt.Sprintf("👋 Welcome! Created your profile with team %s. %s", teamValue, responseText)
	client.Ack(*req, map[string]string{"text": combinedMessage})
}
