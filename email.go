package main

import (
	"fmt"
	"github.com/go-mail/mail"
)

func playWithEmail() {
   scd := &SecretConfigData
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
