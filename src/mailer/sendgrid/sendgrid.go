package sendgrid

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var (
	mailer    *sendgrid.Client
	fromemail = ""
)

func Start(sendgrid_api_key, from string) *sendgrid.Client {
	mailer = sendgrid.NewSendClient(sendgrid_api_key)
	fromemail = from
	return mailer
}

func Send(toemail, subject, plain string) error {
	from := mail.NewEmail(fromemail, fromemail)
	to := mail.NewEmail("Hey Player", toemail)
	plainTextContent := plain
	htmlContent := ""
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := mailer.Send(message)
	return err
}
