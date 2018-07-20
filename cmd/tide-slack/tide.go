package main

import (
	"github.com/nlopes/slack"
	"net/http"
	"strings"
	"github.com/wptide/pkg/tide"
)

type tideCommand struct {
	slash   *slack.SlashCommand
	client  *slack.Client
	tideApi tide.ClientInterface
}

func (h *tideCommand) handle(w http.ResponseWriter, r *http.Request) {

	commands := strings.Split(h.slash.Text, " ")
	command := commands[0]
	commands = commands[1:]

	switch command {
	case "theme":
		fallthrough
	case "plugin":
		fallthrough
	case "phpcompat":
		fallthrough
	case "help":
		fallthrough
	default:
		w.Write([]byte("Getting the info... please wait."))
		go func() {
			if h.tideApi != nil {
				// @todo complete this
				result, _ := h.tideApi.SendPayload("GET", "https://wptide.org/api/tide/v1/audit/", "" )
				h.client.PostMessage(h.slash.ChannelID, result, slack.PostMessageParameters{})
			}
		}()
	}
}
