// Package gitea implements a Service capable of processing webhooks from Gitea
package gitea

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"../../database"
	"../../types"

	"git.aventer.biz/AVENTER/gomatrix"
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
//               boards: {
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
		// A map of "boardID's" to configuration information
		Boards map[string]struct {
			Template string `json:"template"`
		} `json:"boards"`
	} `json:"rooms"`
}

// The payload from Gitea
type webhookNotification struct {
	ID          string `json:"cardId"`
	Text        string `json:"text"`
	ListID      string `json:"listId"`
	OldListID   string `json:"oldListId"`
	BoardID     string `json:"boardId"`
	User        string `json:"user"`
	Card        string `json:"card"`
	Description string `json:"description"`
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
//
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

	whForBoard := notif.BoardID

	logger := log.WithFields(log.Fields{
		"board": whForBoard,
	})

	for roomID, roomData := range s.Rooms {
		for boardData := range roomData.Boards {
			if boardData != whForBoard {
				continue
			}
			msg := gomatrix.TextMessage{
				Body:    notif.Text,
				MsgType: "m.notice",
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
		for range roomData.Boards {
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
