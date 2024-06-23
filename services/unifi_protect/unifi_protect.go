// Package unifi read out event stream from unifi to notify room members
package unifi_protect

import (
	"encoding/json"
	"io"
	"strings"

	"go-avbot/types"

	"github.com/AVENTER-UG/gomatrix"
	"github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// ServiceType of the Unifi service
const ServiceType = "unifi_protect"

// Service represents the Echo service. It has no Config fields.
type Service struct {
	types.DefaultService
	connected bool
	Host      string
	Port      int
	User      string
	Password  string
	csrfToken string
	cookies   string
	RoomID    string
}

// Register makes sure that the given realm ID maps to a github realm.
func (e *Service) Register(oldService types.Service, client *gomatrix.Client) error {
	// Setup the NVR
	nvr := NewNVR(
		e.Host,
		e.Port,
		e.User,
		e.Password)

	// Start the NVR Livefeed
	if err := nvr.Authenticate(); err != nil {
		logrus.Fatal(err)
	}

	events, err := NewWebsocketEvent(nvr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case message := <-events.Events:
				action, err := message.GetAction()
				payload := message.Payload.GetRAWData()

				if action.Action == "update" && action.ModelKey == "event" && action.RecordModel == "camera" {
					var smart SmartDetectTypes
					json.Unmarshal(payload, &smart)

					logrus.Info(action)
					//https://192.168.178.244/proxy/protect/api/events/<action.event>/thumbnail
					logrus.Info(string(payload))

					if len(smart.SmartDetectTypes) == 0 {
						continue
					}

					message := "Detect: " + strings.Join(smart.SmartDetectTypes, "")

					msg := gomatrix.HTMLMessage{
						Body:          message,
						MsgType:       "m.notice",
						Format:        "org.matrix.custom.html",
						FormattedBody: util.MarkdownRender(message),
					}

					if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
						logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect notification to room.")
						continue
					}

					var out io.ReadCloser
					var length int64
					url := "/proxy/protect/api/events/" + action.ID + "/thumbnail"
					out, length, err = nvr.httpDoIO("GET", url)
					if err != nil {
						logrus.WithField("room_id", e.RoomID).Errorf("Could not get thumbnail of %s. Error: %s", url, err.Error())
						continue
					}

					rmu, err := client.UploadToContentRepo(out, "image/jpeg", length)
					if err != nil {
						logrus.WithField("room_id", e.RoomID).Error("Could not upload thumbnail.", err.Error())
						continue
					}
					log.Info(rmu.ContentURI)

					if _, err := client.SendImage(e.RoomID, "file"+action.ID+".jpg", rmu.ContentURI); err != nil {
						logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect thumbnail to room.")
					}

				}

				if err != nil {
					logrus.Warningf("Skipping message due to err: %s", err)
					continue
				}
			}
		}
	}()

	return nil
}

// Commands supported:
//
//	!unifi some message
//
// Responds with a notice of "some message".
func (e *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		{
			Path: []string{"unifi_protect"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return &gomatrix.TextMessage{MsgType: "m.notice", Body: strings.Join(args, " ")}, nil
			},
		},
	}
}

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
