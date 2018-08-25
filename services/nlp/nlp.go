// Package nlp implements a gateeway service so IKY
package nlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"../../types"
	"git.aventer.biz/AVENTER/gomatrix"
	"github.com/russross/blackfriday"
)

// ServiceType of the Echo service
const ServiceType = "nlp"

// Service represents the Echo service. It has no Config fields.
type Service struct {
	types.DefaultService
}

// NLPResponse represents the IKY chat response
type NLPResponse struct {
	CurrentNode         string        `json:"currentNode"`
	Complete            bool          `json:"complete"`
	Parameters          []interface{} `json:"parameters"`
	ExtractedParameters struct {
	} `json:"extractedParameters"`
	SpeechResponse []string  `json:"speechResponse"`
	Date           time.Time `json:"date"`
	Intent         struct {
		Confidence float64 `json:"confidence"`
		ID         string  `json:"id"`
		ObjectID   string  `json:"object_id"`
	} `json:"intent"`
	Context struct {
	} `json:"context"`
	Owner             string        `json:"owner"`
	Input             string        `json:"input"`
	MissingParameters []interface{} `json:"missingParameters"`
}

// ObjectID is the current nlp intend
var ObjectID map[string]NLPResponse

func CmdForwardToNLP(roomID, userID string, message string) interface{} {

	var decodeResp NLPResponse

	// if the user object already exists, get the data for the next response, if not, use default values
	_, ok := ObjectID[userID]
	if ok {
		decodeResp = ObjectID[userID]
	} else {
		decodeResp.Complete = true
	}

	// remember the message of the user
	decodeResp.Input = message

	// set the username
	decodeResp.Owner = userID

	// convert the map to json
	str, err := json.Marshal(decodeResp)

	// send the message to the IKY service
	buf := bytes.NewBuffer(str)

	log.Println(buf)

	resp, err := http.Post("http://localhost:8080/gateway/api/v1", "application/json", buf)
	if err != nil {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("nlp: Could not talk with the IKY: %s", err)}
	}
	defer resp.Body.Close()

	if err != nil {
		return &gomatrix.TextMessage{"m.notice", fmt.Sprintf("nlp: The IKY ignore me: %s", err)}
	}

	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &decodeResp)
	ObjectID[userID] = decodeResp

	// if i got nothing from IKY, just ignore it
	if err != nil {
		return nil
	}

	var strBuffer bytes.Buffer
	strBuffer.WriteString("")
	for _, value := range decodeResp.SpeechResponse {
		strBuffer.WriteString(value)
	}

	return &gomatrix.HTMLMessage{strBuffer.String(), "m.text", "org.matrix.custom.html", markdownRender(strBuffer.String())}
}

func init() {

	ObjectID = make(map[string]NLPResponse)

	types.RegisterService(func(serviceID, serviceUserID, webhookEndpointURL string) types.Service {
		return &Service{
			DefaultService: types.NewDefaultService(serviceID, serviceUserID, ServiceType),
		}
	})
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
