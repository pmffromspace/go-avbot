// Package aws implements a Service to manage aws
package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"

	"../../types"
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/matrix-org/gomatrix"
	"github.com/russross/blackfriday"
)

// ServiceType of the AWS service
const ServiceType = "aws"

// Service represents the AWS service. It has no Config fields.
type Service struct {
	types.DefaultService
	// The Users who are allowed to use the invoice service
	AllowedUsers string
	// AWS AccessKey
	AccessKey string
	// AWS Secret Key
	SecretAccessKey string
	// AWS Token
	AccessToken string
}

// Commands supported:
//    !aws help
func (s *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		types.Command{
			Path: []string{"aws", "help"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				var message string
				message = fmt.Sprintf("##### Help \n")
				message = message + fmt.Sprintf("```\n ")
				message = message + fmt.Sprintf("acl\n======\n \t Give a list of all allowed users\n\n")
				message = message + fmt.Sprintf("instance start\n======\n \t Instance Id: Start the Instance \n\n")
				message = message + fmt.Sprintf("instance stop\n======\n \t Instance Id: Stop the Instance \n\n")
				message = message + fmt.Sprintf("instance show\n======\n \t Give out a list of all instances \n\n")
				message = message + fmt.Sprintf("```\n ")
				return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
			},
		},
		types.Command{
			Path: []string{"aws", "acl"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) {
					return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Allowed Users: %s", s.AllowedUsers)}, nil
				}
				return &gomatrix.TextMessage{"m.notice", "U are not allowed to use this function"}, nil
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "start"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) && len(args) == 1 {
					return s.cmdAwsInstanceStart(roomID, userID, args)
				}
				return &gomatrix.TextMessage{"m.notice", "U are not allowed to use this function or u forgot some arguments"}, nil
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "stop"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) && len(args) == 1 {
					return s.cmdAwsInstanceStop(roomID, userID, args)
				}
				return &gomatrix.TextMessage{"m.notice", "U are not allowed to use this function or u forgot some arguments"}, nil
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "show"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) {
					return s.cmdAwsInstanceShow(roomID, userID, args)
				}
				return &gomatrix.TextMessage{"m.notice", "U are not allowed to use this function"}, nil
			},
		},
	}
}

// Stop the aws instance
func (s *Service) cmdAwsInstanceStop(roomID, userID string, args []string) (interface{}, error) {
	instanceId := args[0]

	log.Info("Service: Aws: Stop Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLogin()

	if sess != nil {
		// to start a instance, we need a ec2 session
		ec := ec2.New(sess)
		input := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceId),
			},
		}
		instances, err := ec.StopInstances(input)

		if err != nil && instances != nil {
			return &gomatrix.TextMessage{"m.notice", "Dont know whats wrong. But I cannot stop the instance."}, nil
		}

		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Instance is down")}, nil
	}

	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

// Start the aws instance
func (s *Service) cmdAwsInstanceStart(roomID, userID string, args []string) (interface{}, error) {
	instanceId := args[0]

	log.Info("Service: Aws: Start Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLogin()

	if sess != nil {
		// to start a instance, we need a ec2 session
		ec := ec2.New(sess)
		input := &ec2.StartInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceId),
			},
		}
		instances, err := ec.StartInstances(input)

		if err != nil && instances != nil {
			return &gomatrix.TextMessage{"m.notice", "Dont know whats wrong. But I cannot start the instance."}, nil
		}

		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Instance is running")}, nil
	}

	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

// Give me a list of all instances
func (s *Service) cmdAwsInstanceShow(roomID, userID string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Show Instance")

	// Have to login first
	sess := s.awsLogin()

	if sess != nil {
		// create me a new ecs session and get out a list of all instances
		ec := ec2.New(sess)

		instances, err := ec.DescribeInstances(nil)
		if err != nil {
			return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Didnt go a list of instances: %s", err)}, nil
		}
		// Well, now we have all instances in a nice structure
		var message string
		message = fmt.Sprintf("##### Instance List")
		message = message + fmt.Sprintf("\n```\n")
		for i := 0; i < len(instances.Reservations); i++ {
			if len(instances.Reservations[i].Instances[0].Tags) > 0 {
				message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].Tags[0].Value, 20))
			} else {
				message = message + fmt.Sprintf("%s", printValue(nil, 20))
			}
			message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].InstanceId, 22))
			message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].InstanceType, 11))
			message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].State.Name, 9))
			message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].PublicDnsName, 55))
			message = message + fmt.Sprintf("%s", printValue(instances.Reservations[i].Instances[0].PublicIpAddress, 20))
			message = message + fmt.Sprintf("%s", instances.Reservations[i].Instances[0].LaunchTime)
			message = message + fmt.Sprintf("\n")
		}
		message = message + fmt.Sprintf("\n```\n")
		return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
	}
	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

func init() {
	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
}

// this function will add spaces to a string, until the length of the string is like we need it
// thats usefull to make the output more pretty
func printValue(message *string, length int) string {
	if message != nil {
		if len(*message) < length {
			*message = *message + " "
			return printValue(message, length)
		}
	} else {
		newMsg := " "
		return printValue(&newMsg, length)
	}
	return *message
}

func (s *Service) awsLogin() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials(s.AccessKey, s.SecretAccessKey, s.AccessToken),
	})

	if err != nil {
		log.Info("Service: Aws: Start Instance Error: ", err)
		return nil
	}
	return sess
}

func markdownRender(content string) string {
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS

	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	extensions := 0
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS

	return string(blackfriday.Markdown([]byte(content), renderer, extensions))
}
