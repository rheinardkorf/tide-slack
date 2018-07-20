package main

import (
	"log"
	_ "github.com/heroku/x/hmetrics/onload"
	"net/http"
	"fmt"
)

func handleOauth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "oAuth Handler")
}

func handleTideCommand(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "/tide handler")
}

func main() {
	http.HandleFunc("/oauth", handleOauth)
	http.HandleFunc("/tide", handleTideCommand)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Nothing to see here.")
	})
	log.Fatal(http.ListenAndServe(":", nil))
}
