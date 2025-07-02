package handlers

import (
	"log"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func HandleSlashCommand(client *socketmode.Client, evt socketmode.Event) {
	cmd, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		log.Printf("Ignored unknown command event")
		return
	}

	log.Printf("Received slash command: %s", cmd.Command)

	switch cmd.Command {
	case "/checkout":
		client.Ack(*evt.Request, map[string]string{"text": "🚛 Handled /checkout!"})
	case "/release":
		client.Ack(*evt.Request, map[string]string{"text": "🔁 Handled /release!"})
	case "/swap":
		client.Ack(*evt.Request, map[string]string{"text": "🔀 Handled /swap!"})
	default:
		client.Ack(*evt.Request, map[string]string{"text": "Unknown command"})
	}
}

func HandleInteractive(client *socketmode.Client, evt socketmode.Event) {
	client.Ack(*evt.Request)
	log.Println("Handled interactive event")
}
