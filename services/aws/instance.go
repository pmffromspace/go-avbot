package aws

import (
	"fmt"
	"strings"

	"github.com/AVENTER-UG/gomatrix"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (s *Service) cmdAwsInstanceRun(roomID, userID string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Instance Run")
	amiId := args[0]
	instanceType := args[1]
	//hostname := args[1]
	//subnet := args[2]
	//sshKeyName := args[2]

	if len(args) < 1 {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf(`Missing parameters. Have a look with !invoice help`)}, nil
	}

	// Have to login first
	sess := s.awsLogin(userID)

	if sess != nil {
		ec := ec2.New(sess)

		input := &ec2.RunInstancesInput{
			ImageId:      aws.String(amiId),
			InstanceType: aws.String(instanceType),
			MinCount:     aws.Int64(1),
			MaxCount:     aws.Int64(1),

			/*
				TagSpecifications: []*ec2.TagSpecification{
					Tags: []*ec2.Tag{
						Key:   aws.String("name"),
						Value: aws.String(hostname),
					},
				},
			*/
		}

		instances, err := ec.RunInstances(input)

		if err != nil {
			log.Info(instances)
			return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("There is sth wrong. %s", err)}, nil
		}

	}
	return &gomatrix.TextMessage{MsgType: "m.notice", Body: "Cannot login into aws"}, nil
}

// Stop the aws instance
func (s *Service) cmdAwsInstanceStop(roomID, userID string, args []string) (interface{}, error) {
	if len(args) < 1 {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf(`Missing parameters. Have a look with !invoice help"`)}, nil
	}

	instanceId := args[0]

	// Which region the user is workign
	region := s.getDefaultRegion(userID)
	if len(args) == 2 {
		// stpp the instance in this region
		region = args[1]
	}

	log.Info("Service: Aws: Stop Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLoginRegion(userID, region)

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
			return &gomatrix.TextMessage{MsgType: "m.notice", Body: "Dont know whats wrong. But I cannot stop the instance."}, nil
		}

		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Instance is down")}, nil
	}

	return &gomatrix.TextMessage{MsgType: "m.notice", Body: "Cannot login into aws"}, nil
}

// Start the aws instance
func (s *Service) cmdAwsInstanceStart(roomID, userID string, args []string) (interface{}, error) {
	if len(args) < 1 {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf(`Missing parameters. Have a look with !invoice help"`)}, nil
	}

	instanceId := args[0]

	// Which region the user is workign
	region := s.getDefaultRegion(userID)
	if len(args) == 2 {
		// start the instance in this region
		region = args[1]
	}

	log.Info("Service: Aws: Start Instance: ", instanceId)

	// have to login first to get a aws session
	sess := s.awsLoginRegion(userID, region)

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
			return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Dont know whats wrong. But I cannot start the instance. %s", err)}, nil
		}

		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Instance is running")}, nil
	}

	return &gomatrix.TextMessage{MsgType: "m.notice", Body: "Cannot login into aws"}, nil
}

// Wrapper function to get all instances of every region
func (s *Service) cmdAwsInstanceShow(roomID, userID string, args []string) (interface{}, error) {
	log.Info("Service: Aws: Show Instance")

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

	region := strings.Split(s.Regions, ",")

	for v, b := range region {
		// Have to login first
		sess := s.awsLoginRegion(userID, b)

		if sess != nil {
			// create me a new ecs session and get out a list of all instances
			ec := ec2.New(sess)
			instances, err := ec.DescribeInstances(nil)
			if err != nil {
				return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Didnt got a list of instances: %s %d", err, v)}, nil
			}

			message = message + s.cmdAwsInstanceShowRegion(b, instances)
		} else {
			return &gomatrix.TextMessage{MsgType: "m.notice", Body: "Cannot login into aws"}, nil
		}
	}
	message = message + fmt.Sprintf("\n```\n")
	return &gomatrix.HTMLMessage{Body: message, MsgType: "m.text", Format: "org.matrix.custom.html", FormattedBody: markdownRender(message)}, nil
}

// Give me a list of all instances from a region
func (s *Service) cmdAwsInstanceShowRegion(region string, instances *ec2.DescribeInstancesOutput) string {
	log.Info("Service: Aws: Show Instance of Region: ", region)

	var message string

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

	return message
}
