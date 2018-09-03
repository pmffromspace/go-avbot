// Package echo implements a Service which echoes back !commands.
package echo

import (
	"strings"

	"../../types"
	"git.aventer.biz/AVENTER/gomatrix"
)

// ServiceType of the Echo service
const ServiceType = "echo"

// Service represents the Echo service. It has no Config fields.
type Service struct {
	types.DefaultService
}

// Commands supported:
//    !echo some message
// Responds with a notice of "some message".
func (e *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		{
			Path: []string{"echo"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return &gomatrix.TextMessage{"m.notice", strings.Join(args, " ")}, nil
			},
		},
		{
			Path: []string{"echo", "widget_1"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return cli.CreateWidget(roomID, "widget_1", "grafana", "https://www.aventer.biz", "AVENTER")
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
