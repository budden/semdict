package user

import (
	"fmt"

	"github.com/budden/semdict/pkg/shared"
	"github.com/go-mail/mail"
)

// if fakeEmail() is true, email is printed to stdout
func fakeEmail() bool {
	return shared.SecretConfigData.SMTPServer == ""
}

// SendEmail sends an email, or, if fakeEmail() is true, prints it to stdout
// Sender, SMTP server and credentials are taken from semdict.config.json
// (loaded when program starts)
func SendEmail(recieverEMail, subj, html string) (err error) {
	if fakeEmail() {
		fmt.Printf(
			"user.fakeEmail() is true, so printing this EMail:\nTo:«%s»\nSubj:«%s»\n«%s»\n",
			recieverEMail, subj, html)
		return
	}
	scd := shared.SecretConfigData
	m := mail.NewMessage()
	m.SetHeader("From", scd.SenderEMail)
	m.SetHeader("To", recieverEMail)
	m.SetHeader("Subject", subj)
	m.SetBody("text/html", html)

	d := mail.NewDialer(scd.SMTPServer, 25, scd.SMTPUser, scd.SMTPPassword)

	err = d.DialAndSend(m)
	return
}

// PlayWithEmail sends an example email (of fakes it)
func PlayWithEmail() {
	scd := shared.SecretConfigData
	SendEmail(scd.RecieverEMail, "Hello!", "Hello, world!")
}
