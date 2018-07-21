package main

import (
	"github.com/nlopes/slack"
	"net/http"
	"strings"
	"github.com/wptide/pkg/tide"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"reflect"
)

type tideCommand struct {
	slash   *slack.SlashCommand
	client  *slack.Client
	tideApi tide.ClientInterface
	baseUrl string
	clientEnabled bool
}

func (h *tideCommand) handle(w http.ResponseWriter, r *http.Request) {

	commands := strings.Split(h.slash.Text, " ")
	command := commands[0]
	commands = commands[1:]

	switch command {
	case "theme":
		fallthrough
	case "plugin":
		go func() {
			var project, slug, version, endpoint string = "", "", "", ""

			if len(commands) > 0 {
				project = commands[0]
				commands = commands[1:]
			}

			if len(commands) > 0 {
				slug = commands[0]
				commands = commands[1:]
			}

			if len(commands) > 0 {
				version = commands[0]
				commands = commands[1:]
			}

			endpoint = fmt.Sprintf("%s/audit/%s/%s/%s", h.baseUrl, project, command, slug)

			if version != "" {
				endpoint = fmt.Sprintf("%s?version=%s", endpoint, version)
			}

			results := make(map[string]interface{})

			var err, jsonErr error

			if h.clientEnabled {
				// Use Tide client.
				response, errClient := h.tideApi.SendPayload("GET", endpoint, "")
				err = errClient
				jsonErr = json.Unmarshal([]byte(response),&results)
			} else {
				// Grab from public accessible data.
				resp, errGet := http.Get( endpoint )
				defer resp.Body.Close()

				if errGet == nil {
					bBody, _ := ioutil.ReadAll( resp.Body )
					jsonErr = json.Unmarshal(bBody,&results)
				}
				err = errGet
			}

			if err != nil {
				h.client.PostMessage(h.slash.ChannelID, err.Error(), slack.PostMessageParameters{})
				return;
			}

			if jsonErr != nil {
				h.client.PostMessage(h.slash.ChannelID, jsonErr.Error(), slack.PostMessageParameters{})
				return;
			}

			params := slack.PostMessageParameters{
				Markdown: true,
			}

			projectTitle := slug

			if results["title"] != nil {
				projectTitle = results["title"].(string)
			}

			if results["reports"] != nil {
				h.client.PostMessage(h.slash.ChannelID, "No reports for: " + projectTitle, slack.PostMessageParameters{})
				return
			}

			// Add theme/plugin info.
			params.Attachments = append(params.Attachments, slack.Attachment{
				Color: "#0B35F5",
				Title: strings.Title( command ) + ": " + projectTitle,
			} )

			// Add phpcs_wordpress
			if results["reports"].(map[string]interface{})["phpcs_phpcompatibility"] != nil &&
				reflect.TypeOf(results["reports"].(map[string]interface{})["phpcs_phpcompatibility"]).Kind() == reflect.Map {

				phpcsWordPress := results["reports"].(map[string]interface{})["phpcs_wordpress"].(map[string]interface{})

				if phpcsWordPress["summary"] == nil {

					params.Attachments = append(params.Attachments, slack.Attachment{
						Color: "#3C91E6",
						Text: "Audit resulted in errors.",
						Title: "PHPCS: WordPress",
					} )

				} else {
					summary := phpcsWordPress["summary"].(map[string]interface{})

					fields := []slack.AttachmentField{
						{
							Title: "Errors",
							Value: fmt.Sprintf("%.0f", summary["errors_count"]),
							Short: true,
						},
						{
							Title: "Warnings",
							Value: fmt.Sprintf("%.0f", summary["warnings_count"]),
							Short: true,
						},
						{
							Title: "Files",
							Value: fmt.Sprintf("%.0f", summary["files_count"]),
							Short: true,
						},
					}

					params.Attachments = append(params.Attachments, slack.Attachment{
						Color:  "#3C91E6",
						Title:  "PHPCS: WordPress",
						Fields: fields,
					})
				}
			}

			// Add phpcs_phpcompatibility.
			if results["reports"].(map[string]interface{})["phpcs_phpcompatibility"] != nil &&
				reflect.TypeOf(results["reports"].(map[string]interface{})["phpcs_phpcompatibility"]).Kind() == reflect.Map {

				phpcsPhpCompatibility := results["reports"].(map[string]interface{})["phpcs_phpcompatibility"].(map[string]interface{})

				if phpcsPhpCompatibility["compatible_versions"] == nil {
					params.Attachments = append(params.Attachments, slack.Attachment{
						Color:  "#5ACEF4",
						Text:   "Audit resulted in errors.",
						Title:  "PHP Compatibility",
						Footer: "Tide API",
					})
				} else {
					compatibleVersions := phpcsPhpCompatibility["compatible_versions"].([]interface{})

					versionString := ""
					for _, compatibleVersion := range compatibleVersions {
						versionString = fmt.Sprintf("%s %s,", versionString, compatibleVersion)
					}
					versionString = strings.TrimRight(versionString, ",")

					params.Attachments = append(params.Attachments, slack.Attachment{
						Color: "#5ACEF4",
						Title: "PHP Compatibility",
						Text:  versionString,
					})
				}
			}

			//EBED82
			// Add phpcs_wordpress
			if results["reports"].(map[string]interface{})["lighthouse"] != nil &&
				reflect.TypeOf(results["reports"].(map[string]interface{})["lighthouse"]).Kind() == reflect.Map {

				lighthouse := results["reports"].(map[string]interface{})["lighthouse"].(map[string]interface{})

				if lighthouse != nil {

					lighthouseCategories := lighthouse["summary"].(map[string]interface{})["categories"].(map[string]interface{})

					scores := make(map[string]float64)
					for catId, category := range lighthouseCategories {
						cat := category.(map[string]interface{})
						catScore := cat["score"].(float64)
						scores[catId] = catScore * 100.0
					}

					fields := []slack.AttachmentField {
						{
							Title: "PWA",
							Value: fmt.Sprintf("%.0f", scores["pwa"]),
							Short: true,
						},
						{
							Title: "Performance",
							Value: fmt.Sprintf("%.0f", scores["performance"]),
							Short: true,
						},
						{
							Title: "Accessibility",
							Value: fmt.Sprintf("%.0f", scores["accessibility"]),
							Short: true,
						},
						{
							Title: "Practices",
							Value: fmt.Sprintf("%.0f", scores["best-practices"]),
							Short: true,
						},
					}

					params.Attachments = append(params.Attachments, slack.Attachment{
						Color: "#fbb30b",
						Title: "Lighthouse Audit",
						Fields: fields,
						Footer: "Powered by Google Lighthouse",
						FooterIcon: "https://developers.google.com/web/tools/lighthouse/images/lighthouse-icon-128.png",
					} )
				}
			}

				// Add "Requested by".
			params.Attachments = append(params.Attachments, slack.Attachment{
				//Color: "#9F43E0",
				Footer: "Results powered by wptide.org\nRequested by @" + h.slash.UserName,
				//Ts: json.Number(time.Now().String()),
			} )

			h.client.PostMessage(h.slash.ChannelID, " ", params)
		}()

	case "help":
		fallthrough
	default:
		w.Write([]byte("Please provide request:\n\tFormat: `/tide plugin|theme <project> <slug> <version>`"))
	}
}