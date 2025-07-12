package handlers

import (
	"log"
	"strconv"
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
			// TODO: open Block Kit modal for truck selection
			client.Ack(*evt.Request, map[string]string{
				"text": "ℹ️ Use `/checkout [truck-name]` or `/checkout [truck-name] [days]` to check out a truck.",
			})
			return
		case 1:
			HandleCheckout(client, evt.Request, args[0], 1, cmd.UserID, cmd.UserName)
			return
		case 2:
			days, err := strconv.Atoi(args[1])
			if err != nil || days < 1 {
				log.Printf("Warning: User %s tried to check out for an invalid number of days: %s", cmd.UserID, args[1])
				client.Ack(*evt.Request, map[string]string{
					"text": "⚠️ Invalid number of days. Use a positive integer like `/checkout Tulip 4`",
				})
				return
			}
			if days > 6 { // ? Assuming max checkout period is 6 days
				log.Printf("Warning: User %s tried to check out for %d days, which exceeds the limit.", cmd.UserID, days)
				client.Ack(*evt.Request, map[string]string{
					"text": "⚠️ Maximum checkout period is 6 days.",
				})
				return
			}
			HandleCheckout(client, evt.Request, args[0], days, cmd.UserID, cmd.UserName)
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
		args := strings.Fields(cmd.Text)
		switch len(args) {
		case 0:
			// Later: open Block Kit modal for truck selection
			client.Ack(*evt.Request, map[string]string{
				"text": "ℹ️ Use `/release [truck-name]` to release a truck.",
			})
			return
		case 1:
			HandleReleaseTruck(client, evt.Request, args[0], cmd.UserID, cmd.UserName)
			return
		default:
			client.Ack(*evt.Request, map[string]string{
				"text": "⚠️ Too many arguments. Try `/release Tulip`",
			})
			return
		}
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
