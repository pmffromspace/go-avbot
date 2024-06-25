// Package unifi read out event stream from unifi to notify room members
package unifi_protect

import (
	"encoding/json"
	"io"
	"strings"
	"time"

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
	nvr       *NVR
}

// Register makes sure that the given realm ID maps to a github realm.
func (e *Service) Register(oldService types.Service, client *gomatrix.Client) error {
	// Setup the NVR
	e.nvr = NewNVR(
		e.Host,
		e.Port,
		e.User,
		e.Password)

	// Start the NVR Livefeed
	if err := e.nvr.Authenticate(); err != nil {
		logrus.Fatal(err)
	}

	events, err := NewWebsocketEvent(e.nvr)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case message := <-events.Events:
				action, err := message.GetAction()
				payload := message.Payload.GetRAWData()

				if action.ModelKey == "event" {
					var event Event
					err := json.Unmarshal(payload, &event)
					if err != nil {
						logrus.Warningf("Error during unmarshal %s", err)
					}

					if event.Type == "ring" {
						message := "<b>RING RING</b>"

						msg := gomatrix.HTMLMessage{
							Body:          message,
							MsgType:       "m.notice",
							Format:        "org.matrix.custom.html",
							FormattedBody: util.MarkdownRender(message),
						}

						if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
							logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect ring notification to room.")
							continue
						}
					}

					if event.Type == "smartDetectZone" {
						go e.SmartDetect(strings.Join(event.SmartDetectTypes, ""), client, action)
					}

				}

				//				if action.Action == "update" && action.ModelKey == "event" && action.RecordModel == "camera" {
				//					var smart SmartDetectTypes
				//					json.Unmarshal(payload, &smart)
				//
				//					logrus.Info(action)
				//					logrus.Info(string(payload))
				//
				//					if len(smart.SmartDetectTypes) == 0 {
				//						continue
				//					}
				//
				//					message := "Detect: " + strings.Join(smart.SmartDetectTypes, "")
				//
				//					msg := gomatrix.HTMLMessage{
				//						Body:          message,
				//						MsgType:       "m.notice",
				//						Format:        "org.matrix.custom.html",
				//						FormattedBody: util.MarkdownRender(message),
				//					}
				//
				//					if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
				//						logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect notification to room.")
				//						continue
				//					}
				//
				//					i := 0
				//				retry:
				//					var out io.ReadCloser
				//					var length int64
				//					url := "/proxy/protect/api/events/" + action.ID + "/thumbnail"
				//					out, length, err = nvr.httpDoIO("GET", url)
				//					if err != nil {
				//						logrus.WithField("room_id", e.RoomID).Errorf("Could not get thumbnail of %s. Error: %s", url, err.Error())
				//						if i <= 3 {
				//							logrus.WithField("room_id", e.RoomID).Errorf("Retry authentication %d", i)
				//							i += 1
				//							goto retry
				//						}
				//						continue
				//					}
				//
				//					rmu, err := client.UploadToContentRepo(out, "image/jpeg", length)
				//					if err != nil {
				//						logrus.WithField("room_id", e.RoomID).Error("Could not upload thumbnail.", err.Error())
				//						continue
				//					}
				//					log.Info(rmu.ContentURI)
				//
				//					if _, err := client.SendImage(e.RoomID, "file"+action.ID+".jpg", rmu.ContentURI); err != nil {
				//						logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect thumbnail to room.")
				//					}
				//
				//				}

				if err != nil {
					logrus.Warningf("Skipping message due to err: %s", err)
					continue
				}
			}
		}
	}()

	return nil
}

func (e *Service) SmartDetect(types string, client *gomatrix.Client, action *WsAction) {
	message := "Detect: " + types

	msg := gomatrix.HTMLMessage{
		Body:          message,
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: util.MarkdownRender(message),
	}

	if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
		logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect smartDetectZone notification to room.")
		return
	}

	time.Sleep(5 * time.Second)

	i := 0
retry:
	var out io.ReadCloser
	var length int64
	url := "/proxy/protect/api/events/" + action.ID + "/thumbnail"
	out, length, err := e.nvr.httpDoIO("GET", url)
	if err != nil {
		logrus.WithField("room_id", e.RoomID).Errorf("Could not get thumbnail of %s. Error: %s", url, err.Error())
		if i <= 3 {
			logrus.WithField("room_id", e.RoomID).Errorf("Retry authentication %d", i)
			i += 1
			goto retry
		}
		return
	}

	rmu, err := client.UploadToContentRepo(out, "image/jpeg", length)
	if err != nil {
		logrus.WithField("room_id", e.RoomID).Error("Could not upload thumbnail.", err.Error())
		return
	}
	log.Info(rmu.ContentURI)

	if _, err := client.SendImage(e.RoomID, "file"+action.ID+".jpg", rmu.ContentURI); err != nil {
		logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect thumbnail to room.")
	}
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
