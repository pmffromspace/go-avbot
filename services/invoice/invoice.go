// Package invoice implements a Service which create invoices
package invoice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"go-avbot/types"

	"github.com/matrix-org/gomatrix"
	"github.com/russross/blackfriday"
	log "github.com/sirupsen/logrus"
)

// ServiceType of the Invoice service
const ServiceType = "invoice"

// Service represents the Invoice service. It has no Config fields.
type Service struct {
	types.DefaultService
	// The Users who are allowed to use the invoice service
	AllowedUsers string
}

// Commands supported:
//    !invoice help
func (s *Service) Commands(cli *gomatrix.Client) []types.Command {
	return []types.Command{
		{
			Path: []string{"invoice", "help"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				var message string
				message = fmt.Sprintf("##### Help \n")
				message = message + fmt.Sprintf("```\n")
				message = message + fmt.Sprintf("create\n======\n \t customer : customer name\n \t unitprice : unitprice in format NNN.NN\n \t quantity : the quantity of the unit\n \t period : [once|month] how often the invoice should be recur\n \t description : A very short description of the invoice\n\n")
				message = message + fmt.Sprintf("get\n======\n \t customer : customer name\n\n")
				message = message + fmt.Sprintf("acl\n======\n \t Give a list of all allowed users\n\n")
				message = message + fmt.Sprintf("```\n")

				return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
			},
		},
		{
			Path: []string{"invoice", "get"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) {
					return s.cmdInvoiceGet(roomID, userID, args)
				} else {
					return &gomatrix.TextMessage{MsgType: "m.notice", Body: "U are not allowed to use this function"}, nil
				}
			},
		},
		{
			Path: []string{"invoice", "acl"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if strings.Contains(s.AllowedUsers, userID) {
					return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Allowed Users: %s", s.AllowedUsers)}, nil
				} else {
					return &gomatrix.TextMessage{MsgType: "m.notice", Body: "U are not allowed to use this function"}, nil
				}
			},
		},
		{
			Path: []string{"invoice", "create"},
			Command: func(roomID, userID string, args []string) (interface{}, error) {
				if len(args) < 5 {
					return &gomatrix.TextMessage{
						MsgType: "m.notice",
						Body:    fmt.Sprintf(`Missing parameters. Have a look with !invoice help"`),
					}, nil
				}
				return s.cmdCreateInvoice(roomID, userID, args)
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

// command to create a invoice for a customer
func (s *Service) cmdCreateInvoice(roomID, userID string, args []string) (interface{}, error) {
	customer := args[0]
	unitprice := args[1]
	quantity := args[2]
	period := args[3]
	description := args[4]

	log.Info("Service: Invoice: createInvoice: Customer=", customer)

	reader := strings.NewReader(fmt.Sprintf(`{"func":"createInvoice","customer":"%s","unitprice":"%s","quantity":"%s","period":"%s","description":"%s"}`, customer, unitprice, quantity, period, description))
	request, err := http.NewRequest("POST", "http://localhost:8888/jsonrpc.php", reader)

	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		fmt.Errorf("Failed to create issue. HTTP %d", res.StatusCode)
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}
	defer res.Body.Close()

	// read the body and convert it to json
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Dont know whats wrong. But I could not decode the gateways body.")}, nil
	}
	var str Invoice
	err = json.Unmarshal(body, &str)

	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}

	// create a list of invoices
	var message string
	message = fmt.Sprintf("##### Create Invoice: %s %s \n", customer, str.Data)
	return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
}

// command to get all invoices of a customer
func (s *Service) cmdInvoiceGet(roomID, userID string, args []string) (interface{}, error) {
	var customer string
	customer = args[0]

	log.Info("Service: Invoice: getInvoiceOfClient: Customer=", customer)

	reader := strings.NewReader(fmt.Sprintf(`{"func":"getInvoicesOfClient","customer":"%s"}`, customer))
	request, err := http.NewRequest("POST", "http://localhost:8888/jsonrpc.php", reader)

	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		fmt.Errorf("Failed to create issue. HTTP %d", res.StatusCode)
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}
	defer res.Body.Close()

	// read the body and convert it to json
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: fmt.Sprintf("Dont know whats wrong. But I could not decode the gateways body.")}, nil
	}
	var str Invoice
	err = json.Unmarshal(body, &str)

	if err != nil {
		return &gomatrix.TextMessage{MsgType: "m.notice", Body: "There is a connection error to the gateway"}, nil
	}

	// create a list of invoices
	var message string
	message = fmt.Sprintf("##### Invoice: %s %s \n", str.Data[0].CompanyName, str.Data[0].ContactName)
	message = message + fmt.Sprintf("```\nNr     Date        Paid Amount\n==============================\n")
	for i := 0; i < len(str.Data); i++ {
		message = message + fmt.Sprintf("%s  %s  %s    %s \n", str.Data[i].InvoiceNumber, str.Data[i].InvoiceDate, str.Data[i].StatusPaid, str.Data[i].InvoiceAmount)
	}
	message = message + "\n```"
	return &gomatrix.HTMLMessage{message, "m.text", "org.matrix.custom.html", markdownRender(message)}, nil
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
