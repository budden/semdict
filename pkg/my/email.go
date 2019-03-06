package my

import (
	"fmt"

	"github.com/budden/a/pkg/shared"
	"github.com/go-mail/mail"
)

// PlayWithEmail sends an email
func PlayWithEmail() {
	scd := &shared.SecretConfigData
	m := mail.NewMessage()
	m.SetHeader("From", scd.SenderEMail)
	m.SetHeader("To", scd.RecieverEMail)
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello world!")

	d := mail.NewDialer(scd.SMTPServer, 587, scd.SMTPUser, scd.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("Failed to send an E-mail, err = %#v\n", err)
	}
}
