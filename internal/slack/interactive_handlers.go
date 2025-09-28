package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"truck-checkout/internal/models"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func showTeamSelectionModal(client *socketmode.Client, req *socketmode.Request, triggerID string, truckName string, businessDays int, userId string, userName string, channelId string) {
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
	metadata := fmt.Sprintf("%s|%d|%s|%s|%s", truckName, businessDays, userId, userName, channelId)
	
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

// This function builds a simple modal that just displays an error message.
func buildErrorModal(errorMessage string) slack.ModalViewRequest {
	// The \` characters create a multi-line string in Go.
	errorText := fmt.Sprintf(`:warning: *An error occurred:*\n\n%s\n\nPlease try again.`, errorMessage)

	return slack.ModalViewRequest{
		Type:  slack.ViewType("modal"),
		Title: slack.NewTextBlockObject("plain_text", "Error", true, false),
		Close: slack.NewTextBlockObject("plain_text", "Close", true, false),
		Blocks: slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewSectionBlock(
					slack.NewTextBlockObject("mrkdwn", errorText, false, false),
					nil,
					nil,
				),
			},
		},
	}
}

func handleTeamSelectionModal(client *socketmode.Client, req *socketmode.Request, callback *slack.InteractionCallback) {
	teamValue := callback.View.State.Values["team_block"]["team_select"].SelectedOption.Value
	metadata := callback.View.PrivateMetadata
	parts := strings.Split(metadata, "|")
	if len(parts) != 5 {
		client.Ack(*req, map[string]string{
			"text": "❌ Error processing team selection.",
		})
		return
	}

	truckName := parts[0]
	businessDays, _ := strconv.Atoi(parts[1])
	userId := parts[2]
	userName := parts[3]
	channelId := parts[4]

	log.Printf("User %s selected team %s for truck %s", userName, teamValue, truckName)

	user, err := models.GetOrCreateUserBySlackID(userId, userName, teamValue)
	if err != nil {
		log.Printf("Failed to create user %s (%s) with team %s: %v", userName, userId, teamValue, err)
		client.Ack(*req, map[string]string{
			"text": "❌ Error creating user profile. Please try again.",
		})
		return
	}

	log.Printf("User %s (%s) with team %s is checking out the truck %s", userName, userId, teamValue, truckName)

	responseText, err := performCheckout(client, user, truckName, businessDays, userId, userName)
	if err != nil {
		log.Printf("Checkout error: %v", err)
		errorView := buildErrorModal(err.Error())
		response := map[string]interface{}{
			"response_action": "update",
			"view":            errorView,
		}

		client.Ack(*req, response)
		return
	}

	combinedMessage := fmt.Sprintf("👋 Welcome! Created your profile with team %s. %s", teamValue, responseText)
	log.Printf("Final response to user %s: %s in %s", userName, combinedMessage, channelId)
	client.Ack(*req, map[string]interface{}{
        "response_action": "clear",
    })

	_, err = client.PostEphemeral(
        channelId,
        callback.User.ID, 
        slack.MsgOptionText(combinedMessage, false),
    )
    if err != nil {
        log.Printf("Failed to send ephemeral message: %v", err)
    }
}
