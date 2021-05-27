package user

import (
	"fmt"

	"github.com/budden/semdict/pkg/shared"
	"github.com/go-mail/mail"
)

// если fakeEmail() равен true, письмо выводится в stdout
func fakeEmail() bool {
	return shared.SecretConfigData.SMTPServer == ""
}

// SendEmail отправляет электронное письмо или, если fakeEmail() равен true, печатает его на stdout
// Отправитель, SMTP-сервер и учетные данные берутся из файла semdict.config.json
// (загружается при запуске программы)
func SendEmail(recieverEMail, subj, html string) (err error) {
	if fakeEmail() {
		fmt.Printf(
			"user.fakeEmail() равно true, поэтому печатается этот EMail:\nTo:«%s»\nSubj:«%s»\n«%s»\n",
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
	if err != nil {
		switch err.(type) {
		case *mail.SendError: fmt.Printf("Причина ошибки: %#v", err.(*mail.SendError).Cause)
		}
	}
	return
}
