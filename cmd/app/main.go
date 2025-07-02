package main

import (
	"log"
	"os"

	db "truck-checkout/database"
	"truck-checkout/handlers"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func main() {
	dbPath := os.Getenv("DATABASE_URL")
	db.InitDB(dbPath)

	api := slack.New(
		os.Getenv("SLACK_BOT_TOKEN"),
		slack.OptionDebug(true),
		slack.OptionAppLevelToken(os.Getenv("SLACK_APP_TOKEN")),
	)
	client := socketmode.New(api)

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				log.Println("Connecting to Slack...")
			case socketmode.EventTypeConnected:
				log.Println("Connected to Slack!")
			case socketmode.EventTypeConnectionError:
				log.Printf("Connection error: %v\n", evt)
			case socketmode.EventTypeSlashCommand:
				log.Println("Slash command received")
				handlers.HandleSlashCommand(client, evt)
			case socketmode.EventTypeInteractive:
				log.Println("Interactive event received")
				handlers.HandleInteractive(client, evt)
			default:
				log.Printf("Unhandled event: %+v\n", evt.Type)
			}
		}
	}()

	client.Run()
}
