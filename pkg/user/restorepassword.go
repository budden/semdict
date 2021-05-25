package user

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
)

// RestorePasswordFormPageHandler отображает страницу /restorepasswordform
func RestorePasswordFormPageHandler(c *gin.Context) {
	sduserID := GetSDUserIdOrZero(c)
	if sduserID > 0 {
		c.Redirect(http.StatusFound, "/changepasswordform")
		return
	}
	c.HTML(http.StatusOK, "restorepasswordform.t.html", nil)
}

// RestorePasswordSubmitPageHandler обрабатывает пост-запрос формы /restorepasswordsubmit
func RestorePasswordSubmitPageHandler(c *gin.Context) {
	email := c.PostForm("email")

	sduserID := GetSDUserIdOrZero(c)
	if sduserID > 0 {
		c.Redirect(http.StatusFound, "/changepasswordform")
		return
	}

	err := doRestorePasswordSubmit(c, &restorePasswordData{
		Email:           email,
		Confirmationkey: GenNonce(16),
	})
	if err != nil {
		log.Println("failed restore password submit err: ", err)
	}
	c.HTML(http.StatusOK,
		"general.t.html",
		shared.GeneralTemplateParams{
			Message: "Если вы указали свой действительный регистрационный E-mail, то мы отправили вам сообщение со ссылкой для сброса пароля."})
}

type restorePasswordData struct {
	Email           string
	Confirmationkey string
}

func doRestorePasswordSubmit(c *gin.Context, d *restorePasswordData) (err error) {
	err = processRestorePasswordSubmitWithDb(d)
	if err == nil {
		// sendConfirmationEmail производит только 500 в случае неудачи
		sendConfirmationRestorePassword(c, d)
	}
	return
}

func processRestorePasswordSubmitWithDb(d *restorePasswordData) error {
	return sddb.WithTransaction(func(trans *sddb.TransactionType) (err error) {
		sddb.CheckDbAlive()
		_, err = trans.Tx.NamedExec(
			`
    DELETE FROM registrationattempt WHERE registrationemail = :email;
`,
			d)
		if err != nil {
			return
		}
		_, err = trans.Tx.NamedExec(
			`
    INSERT INTO  registrationattempt(nickname, registrationemail, confirmationkey, salt, hash)
    VALUES ((SELECT nickname FROM sduser WHERE registrationemail = :email), :email, :confirmationkey, '', '');
`,
			d)
		if err == nil {
			sddb.CheckDbAlive()
			err = trans.Tx.Commit()
		}
		return
	})
}

func sendConfirmationRestorePassword(c *gin.Context, d *restorePasswordData) {
	scd := shared.SecretConfigData
	// TODO: если нет файлов сертификатов, используйте http an7
	confirmationLinkBase := shared.SitesProtocol() + "//" + scd.SiteRoot + shared.SitesPort() + "/changepasswordform"
	parameters := url.Values{"email": {d.Email}, "confirmationkey": {d.Confirmationkey}}
	u, err := url.Parse(confirmationLinkBase)
	apperror.GracefullyExitAppIf(err, "Невозможно разобрать базовый URL для ссылки подтверждения")
	u.RawQuery = parameters.Encode()
	confirmationLink := u.String()
	body := fmt.Sprintf(
		"Здравствуйте, чтобы восстановить пароль, пожалуйста, перейдите по ссылке восстановления: <a href=%[1]s>%[1]s</a>",
		confirmationLink,
	)

	err = SendEmail(
		d.Email,
		"Восстановите пароль семантического словаря!",
		body)

	if err != nil {
		// Мы предполагаем, что неспособность отправить электронное письмо может быть вызвана
		// временными проблемами в сети
		apperror.Panic500AndLogAttackIf(err, c, "Не удалось отправить подтверждение по электронной почте")
	}

	noteRestorePasswordConfirmationEMailSentWithDb(d)
	return
}

func noteRestorePasswordConfirmationEMailSentWithDb(d *restorePasswordData) {
	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		sddb.CheckDbAlive()
		_, err1 = trans.Tx.NamedExec(
			`update registrationattempt set rastatus='e-mail sent' WHERE
            registrationemail = :email and confirmationkey = :confirmationkey;`,
			d)
		return
	})
	sddb.FatalDatabaseErrorIf(err, "Ошибка, помнящая, что электронная почта была отправлена, ошибка заключается в следующем %#v", err)
	return
}
