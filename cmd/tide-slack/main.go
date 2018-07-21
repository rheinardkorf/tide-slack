package main

import (
	"log"
	_ "github.com/heroku/x/hmetrics/onload"
	"net/http"
	"fmt"
	"os"
	"github.com/nlopes/slack"
	"github.com/wptide/pkg/tide/api"
	"github.com/wptide/pkg/env"
	"github.com/wptide/pkg/tide"
)

var (
	token                                  = os.Getenv("TOKEN")
	teamDomain                             = os.Getenv("TEAM")
	environment                            = os.Getenv("ENVIRONMENT")
	slackApi                               = slack.New(token)
	tideClient        tide.ClientInterface = &api.Client{}
	tideBaseURL       string
	tideClientEnabled bool
)

func handleOauth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "oAuth Handler")
}

func handleSlashCommand(w http.ResponseWriter, r *http.Request) {

	sCmd, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if sCmd.TeamDomain != teamDomain && environment != "debug" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch sCmd.Command {
	case "/tide":
		hnd := &tideCommand{
			&sCmd,
			slackApi,
			tideClient,
			tideBaseURL,
			tideClientEnabled,
		}
		hnd.handle(w, r)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func main() {

	port := os.Getenv("PORT")
	conf := getConfig()["tide"]

	tideBaseURL = fmt.Sprintf("%s://%s/api/tide/%s", conf["protocol"], conf["host"], conf["version"])

	// Use authenticated client or standard net libraries.
	tideClientEnabled = conf["enabled"] == "yes" || conf["enabled"] == "1"

	// Authenticated client required for some information, but basic information
	// can be retrieved without a Tide Client.
	if tideClientEnabled {
		if err := tideClient.Authenticate(conf["key"], conf["secret"], conf["auth"]); err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/oauth", handleOauth)
	http.HandleFunc("/tide", handleSlashCommand)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Nothing to see here.")
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getConfig() map[string]map[string]string {
	return map[string]map[string]string{
		"tide": {
			"key":      env.GetEnv("API_KEY", ""),
			"secret":   env.GetEnv("API_SECRET", ""),
			"auth":     env.GetEnv("API_AUTH_URL", ""),
			"host":     env.GetEnv("API_HTTP_HOST", ""),
			"protocol": env.GetEnv("API_PROTOCOL", ""),
			"version":  env.GetEnv("API_VERSION", ""),
			"enabled":  env.GetEnv("CLIENT_ENABLED", ""),
		},
	}
}
