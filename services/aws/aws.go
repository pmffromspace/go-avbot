// Package aws implements a Service to manage aws
package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"

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
	// The Usermaping to the AWS Keys
	Users map[string]struct {
		// AWS AccessKey
		AccessKey string
		// AWS Secret Key
		SecretAccessKey string
		// AWS Token
		AccessToken string
		// AWS Region
		Region string
	}
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
				message = message + fmt.Sprintf("```\n")
				message = message + fmt.Sprintf("instance start\n==============\n \t Instance Id: Start the Instance \n\n")
				message = message + fmt.Sprintf("instance stop\n==============\n \t Instance Id: Stop the Instance \n\n")
				message = message + fmt.Sprintf("instance show\n==============\n \t Show you a list of your instances \n\n")
				message = message + fmt.Sprintf("image search\n==============\n \t store : [marketplace|amazon|microsoft|all] where so search\n\t name : query string to search a image (case sensitive) \n\n")
				message = message + fmt.Sprintf("```\n")

				return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "start"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceStart(roomID, userID, args)
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "stop"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceStop(roomID, userID, args)
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "show"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceShow(roomID, userID, args)
			},
		},
		types.Command{
			Path: []string{"aws", "instance", "create"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceCreate(roomID, userID, args)
			},
		},
		types.Command{
			Path: []string{"aws", "image", "search", "amazon"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "amazon", args)
			},
		},
		types.Command{
			Path: []string{"aws", "image", "search", "marketplace"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "aws-marketplace", args)
			},
		},
		types.Command{
			Path: []string{"aws", "image", "search", "microsoft"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "microsoft", args)
			},
		},
		types.Command{
			Path: []string{"aws", "image", "search", "all"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "*", args)
			},
		},
	}
}

func (s *Service) cmdAwsInstanceCreate(roomID, userID string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Instance Create")

	if len(args) < 1 {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf(`Missing parameters. Have a look with !invoice help`)}, nil
	}

	// Have to login first
	sess := s.awsLogin(userID)

	if sess != nil {
		ops := opsworks.New(sess)

		input := &opsworks.CreateInstanceInput{}

		instances, err := ops.CreateInstance(input)

		if err != nil {
			log.Info(instances)
			return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("There is sth wrong. %s", err)}, nil
		}

	}
	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

// Stop the aws instance
func (s *Service) cmdAwsInstanceStop(roomID, userID string, args []string) (interface{}, error) {
	instanceId := args[0]

	if len(args) < 1 {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf(`Missing parameters. Have a look with !invoice help"`)}, nil
	}

	log.Info("Service: Aws: Stop Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLogin(userID)

	if sess != nil {
		// to start a instance, we need a ec2 session
		ec := ec2.New(sess)
		input := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(instanceId),
			},
		}
		instances, err := ec.StopInstances(input)

		if err != nil {
			log.Info(instances)
			return &gomatrix.TextMessage{"m.notice", "Dont know whats wrong. But I cannot stop the instance."}, nil
		}

		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Instance is down")}, nil
	}

	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

// Start the aws instance
func (s *Service) cmdAwsInstanceStart(roomID, userID string, args []string) (interface{}, error) {
	if len(args) < 1 {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf(`Missing parameters. Have a look with !invoice help"`)}, nil
	}

	instanceId := args[0]

	log.Info("Service: Aws: Start Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLogin(userID)

	if sess != nil {
		// to start a instance, we need a ec2 session
		ec := ec2.New(sess)
		// set a filter to get only the instance we want
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
	sess := s.awsLogin(userID)

	if sess != nil {
		// create me a new ecs session and get out a list of all instances
		ec := ec2.New(sess)

		instances, err := ec.DescribeInstances(nil)
		if err != nil {
			return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Didnt go a list of instances: %s", err)}, nil
		}
		// Well, now we have all instances in a nice structure, so print them out
		var message string
		message = fmt.Sprintf("##### Instance List")
		message = message + fmt.Sprintf("\n```\n")

		// To make it more pretty, we nead a header
		message = message + fmt.Sprintf("| %s ", printValueStr("INSTANCEID", 22))
		message = message + fmt.Sprintf("| %s ", printValueStr("NAME", 20))
		message = message + fmt.Sprintf("| %s ", printValueStr("TYPE", 11))
		message = message + fmt.Sprintf("| %s ", printValueStr("STATE", 9))
		message = message + fmt.Sprintf("| %s ", printValueStr("PUBLICDNS", 55))
		message = message + fmt.Sprintf("| %s ", printValueStr("PUBLICIP", 20))
		message = message + fmt.Sprintf("| %s ", printValueStr("REGION", 20))
		message = message + fmt.Sprintf("| %s |", printValueStr("LAUNCHTIME", 20))
		message = message + fmt.Sprintf("\n\n")

		for i := 0; i < len(instances.Reservations); i++ {
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].InstanceId, 22))
			if len(instances.Reservations[i].Instances[0].Tags) > 0 {
				message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].Tags[0].Value, 20))
			} else {
				message = message + fmt.Sprintf("| %s ", printValue(nil, 20))
			}
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].InstanceType, 11))
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].State.Name, 9))
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].PublicDnsName, 55))
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].PublicIpAddress, 20))
			message = message + fmt.Sprintf("| %s ", printValue(instances.Reservations[i].Instances[0].Placement.AvailabilityZone, 20))
			message = message + fmt.Sprintf("| %s |", instances.Reservations[i].Instances[0].LaunchTime)
			message = message + fmt.Sprintf("\n")
		}
		message = message + fmt.Sprintf("\n```\n")
		return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
	}
	return &gomatrix.TextMessage{"m.notice", "Cannot login into aws"}, nil
}

// Show a list of amazon images (ami)
// Give me a list of all instances
func (s *Service) cmdAwsImageSearch(roomID, userID, ownerAlias string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Show Images")

	if len(args) < 1 {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf(`Missing parameters. Have a look with !invoice help`)}, nil
	}

	searchString := strings.Replace(args[0], "*", "", -1)
	if len(searchString) < 4 {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf(`Your search string is to short my friend!`)}, nil
	}
	// Have to login first
	sess := s.awsLogin(userID)

	if sess != nil {
		// create me a new ecs session and get out a list of all instances
		ec := ec2.New(sess)
		input := &ec2.DescribeImagesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("image-type"),
					Values: []*string{
						aws.String("machine"),
					},
				},
				{
					Name: aws.String("owner-alias"),
					Values: []*string{
						aws.String(ownerAlias),
					},
				},
				{
					Name: aws.String("state"),
					Values: []*string{
						aws.String("available"),
					},
				},
				{
					Name: aws.String("is-public"),
					Values: []*string{
						aws.String("true"),
					},
				},
				{
					Name: aws.String("description"),
					Values: []*string{
						aws.String("*" + searchString + "*"),
					},
				},
			},
		}
		images, err := ec.DescribeImages(input)
		if err != nil {
			return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("Didnt go a list of instances: %s", err)}, nil
		}

		// Well, now we have all images in a nice structure, so print them out
		var message string
		message = fmt.Sprintf("##### Images List")
		message = message + fmt.Sprintf("\n```\n")

		// To make it more pretty, we nead a header
		message = message + fmt.Sprintf("| %s ", printValueStr("IMAGEID", 13))
		message = message + fmt.Sprintf("| %s ", printValueStr("DESCRIPTION", 100))
		message = message + fmt.Sprintf("| %s ", printValueStr("NAME", 100))
		message = message + fmt.Sprintf("| %s ", printValueStr("HYPERVISOR", 15))
		message = message + fmt.Sprintf("| %s ", printValueStr("ARCHITECTURE", 15))
		message = message + fmt.Sprintf("\n")

		length := len(images.Images)
		log.Info(fmt.Sprintf("%d", length))
		for i := 0; i < length && i < 30; i++ {
			message = message + fmt.Sprintf("| %s ", printValue(images.Images[i].ImageId, 13))
			message = message + fmt.Sprintf("| %s ", printValue(images.Images[i].Description, 100))
			message = message + fmt.Sprintf("| %s ", printValue(images.Images[i].Name, 100))
			message = message + fmt.Sprintf("| %s ", printValue(images.Images[i].Hypervisor, 15))
			message = message + fmt.Sprintf("| %s ", printValue(images.Images[i].Architecture, 15))
			message = message + fmt.Sprintf("\n")
		}

		message = message + fmt.Sprintf("\n```\n")

		if length > 30 {
			message = message + fmt.Sprintf("\n")
			message = message + fmt.Sprintf("There are %d more results. Please use a better query string.", length-30)
			message = message + fmt.Sprintf("\n")
		}

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

// a wrapper fo printValue to use strings and not string pointers
func printValueStr(message string, length int) string {
	return printValue(&message, length)
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

// get the aws credentials of a user
func (s *Service) getCredentials(userID string) (string, string, string, string) {
	for localUserID, userConfig := range s.Users {
		if localUserID == userID {
			return userConfig.AccessKey, userConfig.SecretAccessKey, userConfig.AccessToken, userConfig.Region
		}
	}

	return "", "", "", ""
}

// loginto the aws
func (s *Service) awsLogin(userID string) *session.Session {
	AccessKey, SecretAccessKey, AccessToken, Region := s.getCredentials(userID)

	if AccessKey == "" {
		return nil
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewStaticCredentials(AccessKey, SecretAccessKey, AccessToken),
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
