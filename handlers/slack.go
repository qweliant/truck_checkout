package handlers

import (
	"log"
	"strings"

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
		args := strings.Fields(cmd.Text)
		switch len(args) {
		case 0:
			// Later: open Block Kit modal for truck selection
			client.Ack(*evt.Request, map[string]string{
				"text": "ℹ️ Use `/checkout [truck-name]` to check out a truck.",
			})
			return
		case 1:
			HandleQuickCheckout(client, evt.Request, args[0], cmd.UserID, cmd.UserName)
			return
		default:
			client.Ack(*evt.Request, map[string]string{
				"text": "⚠️ Too many arguments. Try `/checkout Tulip`",
			})
			return
		}
	case "/trucks":
		args := strings.Fields(cmd.Text)
		if len(args) > 0 {
			switch args[0] {
			case "available":
				HandleTrucksAvailable(client, evt.Request)
				return
			case "checked-out":
				HandleTrucksCheckedOut(client, evt.Request)
				return
			}
		}
		// fallback
		client.Ack(*evt.Request, map[string]string{
			"text": "ℹ️ Try `/trucks available` to see today's available trucks.",
		})
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
