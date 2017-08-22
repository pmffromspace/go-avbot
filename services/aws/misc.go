package aws

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/russross/blackfriday"
)

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
func (s *Service) getDefaultRegion(userID string) string {
	userConfig := s.Users[userID]
	return userConfig.DefaultRegion
}

// get the default region
func (s *Service) getCredentials(userID string) (string, string, string, string) {
	for localUserID, userConfig := range s.Users {
		if localUserID == userID {
			return userConfig.AccessKey, userConfig.SecretAccessKey, userConfig.AccessToken, userConfig.DefaultRegion
		}
	}

	return "", "", "", ""
}

// loginto the aws without region
func (s *Service) awsLogin(userID string) *session.Session {
	return s.awsLoginRegion(userID, "")
}

// loginto the aws with a region
func (s *Service) awsLoginRegion(userID, region string) *session.Session {
	AccessKey, SecretAccessKey, AccessToken, Region := s.getCredentials(userID)

	if region != "" {
		Region = region
	}

	if AccessKey == "" {
		return nil
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewStaticCredentials(AccessKey, SecretAccessKey, AccessToken),
	})

	if err != nil {
		log.Info("Service: Aws: Start Login Error: ", err)
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
