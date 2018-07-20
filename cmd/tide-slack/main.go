package main

import (
	"log"
	_ "github.com/heroku/x/hmetrics/onload"
	"net/http"
	"fmt"
	"os"
	"github.com/nlopes/slack"
)

var (
	token = os.Getenv("TOKEN")
	teamDomain = os.Getenv("TEAM")
	environment = os.Getenv("ENVIRONMENT")
)

func handleOauth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "oAuth Handler")
}

func handleTideCommand(w http.ResponseWriter, r *http.Request) {

	sCmd, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if sCmd.TeamDomain != teamDomain && environment != "debug" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	slackApi := slack.New(token)
	slackApi.PostMessage( sCmd.ChannelID, "oh, hai", slack.PostMessageParameters{})

	//fmt.Fprintf(w, "/tide handler: " + sCmd.TeamDomain )
}

func main() {

	port := os.Getenv("PORT")

	http.HandleFunc("/oauth", handleOauth)
	http.HandleFunc("/tide", handleTideCommand)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Nothing to see here.")
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
