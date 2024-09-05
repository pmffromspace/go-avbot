// Package unifi_protect_alarm implements a Service capable of processing webhooks from Wekan
package unifi_protect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AVENTER-UG/gomatrix"
	"github.com/AVENTER-UG/util/util"
	"github.com/sirupsen/logrus"
)

type webhookNotification struct {
	Alarm struct {
		Name       string        `json:"name"`
		Sources    []interface{} `json:"sources"`
		Conditions []struct {
			Condition struct {
				Type   string `json:"type"`
				Source string `json:"source"`
			} `json:"condition"`
		} `json:"conditions"`
		Triggers []struct {
			Key    string `json:"key"`
			Device string `json:"device"`
		} `json:"triggers"`
	} `json:"alarm"`
	Timestamp int64 `json:"timestamp"`
}

// OnReceiveWebhook receives requests from unifi protect and possibly sends requests to Matrix as a result.
// Go-AVBOT cannot register with unifi_protect_alarm for webhooks automatically. The user must manually add the
// webhook endpoint URL.
//
//	notifications:
//	    webhooks: http://go-avbot-endpoint.com/unifi_protect_alarm_webhook_service
func (e *Service) OnReceiveWebhook(w http.ResponseWriter, req *http.Request, client *gomatrix.Client) {
	logrus.Info("Receive Unifi Protect WebHook")

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		logrus.Error("Unifi webhook is missing payload= form value", err)
		w.WriteHeader(400)
		return
	}

	logrus.Info(string(payload))

	var notif webhookNotification
	if err := json.Unmarshal([]byte(payload), &notif); err != nil {
		logrus.WithError(err).Error("Unifi webhook received an invalid JSON payload=", payload)
		w.WriteHeader(400)
		return
	}

	message := fmt.Sprintf("<i>%s</i> triggerd by: ", notif.Alarm.Name)
	for _, key := range notif.Alarm.Triggers {
		message += key.Key + " "
	}

	msg := gomatrix.HTMLMessage{
		Body:          message,
		MsgType:       "m.notice",
		Format:        "org.matrix.custom.html",
		FormattedBody: util.MarkdownRender(message),
	}

	if _, err := client.SendMessageEvent(e.RoomID, "m.room.message", msg); err != nil {
		logrus.WithField("room_id", e.RoomID).Error("Failed to send unifi ring notification to room.")
	}

	w.WriteHeader(200)
}
