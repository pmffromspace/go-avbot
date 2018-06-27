// Package gitea implements a Service capable of processing webhooks from Gitea
package gitea

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"../../database"
	"../../types"

	"git.aventer.biz/AVENTER/gomatrix"
	"git.aventer.biz/AVENTER/util"
	log "github.com/sirupsen/logrus"
)

// ServiceType of the Gitea service.
const ServiceType = "gitea"

// DefaultTemplate contains the template that will be used if none is supplied.
const DefaultTemplate = (`%{boardsitory}#%{build_number} (%{branch} - %{commit} : %{author}): %{message}
	Change view : %{compare_url}
	Build details : %{build_url}`)

var httpClient = &http.Client{}

// Service contains the Config fields for the Gitea service.
//
// This service will send notifications into a Matrix room when Gitea sends
// webhook events to it. It requires a public domain which Gitea can reach.
// Notices will be sent as the service user ID.
//
// Example JSON request:
//   {
//       rooms: {
//           "!ewfug483gsfe:localhost": {
//               repos: {
//                   "1" {
//                   }
//               }
//           }
//       }
//   }
type Service struct {
	types.DefaultService
	webhookEndpointURL string
	// The URL which should be added to .gitea.yml - Populated by Go-NEB after Service registration.
	WebhookURL string `json:"webhook_url"`
	// A map from Matrix room ID to Github-style owner/board boardsitories.
	Rooms map[string]struct {
		// A map of "repos" to configuration information
		Repos map[string]struct {
			Template string `json:"template"`
		} `json:"repos"`
	} `json:"rooms"`
}

// The payload from Gitea
type webhookNotification struct {
	Secret     string `json:"secret"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	CompareURL string `json:"compare_url"`
	Commits    []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
		Timestamp string `json:"timestamp"`
	} `json:"commits"`
	Repository struct {
		ID    int `json:"id"`
		Owner struct {
			ID        int    `json:"id"`
			Login     string `json:"login"`
			FullName  string `json:"full_name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
			Username  string `json:"username"`
		} `json:"owner"`
		Name            string `json:"name"`
		FullName        string `json:"full_name"`
		Description     string `json:"description"`
		Private         bool   `json:"private"`
		Fork            bool   `json:"fork"`
		HTMLURL         string `json:"html_url"`
		SSHURL          string `json:"ssh_url"`
		CloneURL        string `json:"clone_url"`
		Website         string `json:"website"`
		StarsCount      int    `json:"stars_count"`
		ForksCount      int    `json:"forks_count"`
		WatchersCount   int    `json:"watchers_count"`
		OpenIssuesCount int    `json:"open_issues_count"`
		DefaultBranch   string `json:"default_branch"`
		CreatedAt       string `json:"created_at"`
		UpdatedAt       string `json:"updated_at"`
	} `json:"repository"`
	Pusher struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		FullName  string `json:"full_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Username  string `json:"username"`
	} `json:"pusher"`
	Sender struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		FullName  string `json:"full_name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
		Username  string `json:"username"`
	} `json:"sender"`
}

func outputForTemplate(giteaTmpl string, tmpl map[string]string) (out string) {
	if giteaTmpl == "" {
		giteaTmpl = DefaultTemplate
	}
	out = giteaTmpl
	for tmplVar, tmplValue := range tmpl {
		out = strings.Replace(out, "%{"+tmplVar+"}", tmplValue, -1)
	}
	return out
}

// OnReceiveWebhook receives requests from gitea and possibly sends requests to Matrix as a result.
//
// If the boardsitory matches a known gitea board, a notification will be formed from the
// template for that boardsitory and a notice will be sent to Matrix.
//
// Go-AVBOT cannot register with gitea for webhooks automatically. The user must manually add the
// webhook endpoint URL to their .gitea.yml file:
//    notifications:
//        webhooks: http://go-avbot-endpoint.com/gitea_webhook_service
// 	Gitea webhook definition: https://docs.gitea.io/en-us/webhooks/
func (s *Service) OnReceiveWebhook(w http.ResponseWriter, req *http.Request, cli *gomatrix.Client) {
	if err := req.ParseForm(); err != nil {
		log.WithError(err).Error("Failed to read incoming Gitea webhook form")
		w.WriteHeader(400)
		return
	}
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Error("Gitea webhook is missing payload= form value", err)
		w.WriteHeader(400)
		return
	}

	var notif webhookNotification
	if err := json.Unmarshal([]byte(payload), &notif); err != nil {
		log.WithError(err).Error("Gitea webhook received an invalid JSON payload=", payload)
		w.WriteHeader(400)
		return
	}

	repositoryName := notif.Repository.FullName

	logger := log.WithFields(log.Fields{
		"repo": repositoryName,
	})

	for roomID, roomData := range s.Rooms {
		for boardData := range roomData.Repos {
			if boardData != repositoryName && boardData != notif.Repository.Owner.Login+"/*" {
				continue
			}

			var message string
			message = fmt.Sprintf("Gitea commit from User **%s** in Repo **%s**\n", notif.Commits[0].Author.Name, notif.Repository.FullName)
			message = message + fmt.Sprintf("*%s* \n", notif.Commits[0].Message)
			message = message + fmt.Sprintf("[Commit](%s) \n", notif.Commits[0].URL)

			msg := gomatrix.HTMLMessage{
				Body:          message,
				MsgType:       "m.text",
				Format:        "org.matrix.custom.html",
				FormattedBody: util.MarkdownRender(message),
			}

			logger.WithFields(log.Fields{
				"message": msg,
				"room_id": roomID,
			}).Print("Sending Gitea notification to room")
			if _, e := cli.SendMessageEvent(roomID, "m.room.message", msg); e != nil {
				logger.WithError(e).WithField("room_id", roomID).Print(
					"Failed to send Gitea notification to room.")
			}
		}
	}
	w.WriteHeader(200)
}

// Register makes sure the Config information supplied is valid.
func (s *Service) Register(oldService types.Service, client *gomatrix.Client) error {
	s.WebhookURL = s.webhookEndpointURL
	log.Info("Gitea WebhookURL: ", s.WebhookURL)
	s.joinRooms(client)
	return nil
}

// PostRegister deletes this service if there are no registered boards.
func (s *Service) PostRegister(oldService types.Service) {
	for _, roomData := range s.Rooms {
		for range roomData.Repos {
			return // at least 1 board exists
		}
	}
	// Delete this service since no boards are configured
	logger := log.WithFields(log.Fields{
		"service_type": s.ServiceType(),
		"service_id":   s.ServiceID(),
	})
	logger.Info("Removing service as no boardsitories are registered.")
	if err := database.GetServiceDB().DeleteService(s.ServiceID()); err != nil {
		logger.WithError(err).Error("Failed to delete service")
	}
}

func (s *Service) joinRooms(client *gomatrix.Client) {
	for roomID := range s.Rooms {
		if _, err := client.JoinRoom(roomID, "", nil); err != nil {
			log.WithFields(log.Fields{
				log.ErrorKey: err,
				"room_id":    roomID,
				"user_id":    client.UserID,
			}).Error("Failed to join room")
		}
	}
}

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService:     types.NewDefaultService(serviceID, serviceUserID, ServiceType),
			webhookEndpointURL: webhookEndpointURL,
		}
	})
}
