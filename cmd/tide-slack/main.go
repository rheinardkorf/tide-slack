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
)

func handleOauth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "oAuth Handler")
}

func handleTideCommand(w http.ResponseWriter, r *http.Request) {

	//s := slack.New(token)
	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//
	//if !s.ValidateToken(token) {
		//w.WriteHeader(http.StatusUnauthorized)
		//fmt.Fprintf(w, "awwwww" )
		//return
	//}

	fmt.Fprintf(w, "/tide handler: " + s.TeamDomain )
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
