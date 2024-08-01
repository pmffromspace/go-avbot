// Package unifi read out event stream from unifi to notify room members
package unifi_protect

import (
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

// Service represents the unifi_protext service. It has no Config fields.
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

func (e *Service) Register(oldService types.Service, client *gomatrix.Client) error {
	// Start the NVR Livefeed
	if err := e.Authenticate(); err != nil {
		logrus.Fatal(err)
	}

	events, err := NewWebsocketEvent(e)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case message := <-events.Events:
				action, err := message.GetAction()

				if action.ModelKey == "event" {
					event := Event{}
					err := message.Payload.GetJSON(&event)
					if err != nil {
						logrus.Warningf("Error during unmarshal event %s", err)
					}

					smart := SmartDetectTypes{}
					err = message.Payload.GetJSON(&smart)
					if err != nil {
						logrus.Warningf("Error during unmarshal smart %s", err)
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

					if len(smart.SmartDetectTypes) > 0 {
						if len(smart.Metadata.DetectedThumbnails) > 0 {
							go e.SmartDetect(strings.Join(event.SmartDetectTypes, " "), client, action)
						}
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

func (e *Service) SmartDetect(types string, client *gomatrix.Client, action *WsAction) {
	//	time.Sleep( * time.Second)

	message := "Detect: " + types

	msg := gomatrix.HTMLMessage{
		Body:          message,
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: util.MarkdownRender(message),
	}

	var out io.ReadCloser
	var length int64
	url := "/proxy/protect/api/events/" + action.ID + "/thumbnail"
	out, length, err := e.CallIO("GET", url)
	if err != nil {
		logrus.WithField("room_id", e.RoomID).Errorf("Could not get thumbnail of %s. Error: %s", url, err.Error())
		return
	}

	if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
		logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi_protect smartDetectZone notification to room.")
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

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}
