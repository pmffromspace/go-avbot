// Package aws implements a Service to manage aws
package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"

	"../../types"
	log "github.com/sirupsen/logrus"

	"git.aventer.biz/AVENTER/gomatrix"
	"github.com/aws/aws-sdk-go/aws"
)

// ServiceType of the AWS service
const ServiceType = "aws"

// Service represents the AWS service. It has no Config fields.
type Service struct {
	types.DefaultService
	// All Region we want to work
	Regions string
	// The Usermaping to the AWS Keys
	Users map[string]struct {
		// AWS AccessKey
		AccessKey string
		// AWS Secret Key
		SecretAccessKey string
		// AWS Token
		AccessToken string
		// AWS Default Region
		DefaultRegion string
	}
}

// Commands supported:
//    !aws help
func (s *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		{
			Path: []string{"aws", "help"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				var message string
				message = fmt.Sprintf("##### Help \n")
				message = message + fmt.Sprintf("```\n")
				message = message + fmt.Sprintf("instance start\n==============\n \t InstanceId : Start the Instance \n \t Region : The AWS Region \n\n")
				message = message + fmt.Sprintf("instance stop\n==============\n \t InstanceId : Stop the Instance \n \t Region : The AWS Region \n\n")
				message = message + fmt.Sprintf("instance show\n==============\n \t Show you a list of your instances \n\n")
				message = message + fmt.Sprintf("instance run\n==============\n \t AmiID : AMI ID of the Image to run \n \t InstanceType : Instance Type like t2.micro  \n\n")
				message = message + fmt.Sprintf("image search\n==============\n \t store : [marketplace|amazon|microsoft|all] where so search\n\t name : query string to search a image (case sensitive) \n\n")
				message = message + fmt.Sprintf("```\n")

				return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
			},
		},
		{
			Path: []string{"aws", "instance", "start"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceStart(roomID, userID, args)
			},
		},
		{
			Path: []string{"aws", "instance", "stop"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceStop(roomID, userID, args)
			},
		},
		{
			Path: []string{"aws", "instance", "show"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceShow(roomID, userID, args)
			},
		},
		{
			Path: []string{"aws", "instance", "run"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsInstanceRun(roomID, userID, args)
			},
		},
		{
			Path: []string{"aws", "image", "search", "amazon"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "amazon", args)
			},
		},
		{
			Path: []string{"aws", "image", "search", "marketplace"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "aws-marketplace", args)
			},
		},
		{
			Path: []string{"aws", "image", "search", "microsoft"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "microsoft", args)
			},
		},
		{
			Path: []string{"aws", "image", "search", "all"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				return s.cmdAwsImageSearch(roomID, userID, "*", args)
			},
		},
	}
}

// Show a list of amazon images (ami)
// Give me a list of all instances
func (s *Service) cmdAwsImageSearch(roomID, userID, ownerAlias string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Show Images")

	if len(args) < 2 {
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
