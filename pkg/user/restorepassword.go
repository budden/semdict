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

// RestorePasswordFormPageHandler renders a /restorepasswordform page
func RestorePasswordFormPageHandler(c *gin.Context) {
	sduserID := GetSDUserIdOrZero(c)
	if sduserID > 0 {
		c.Redirect(http.StatusFound, "/changepasswordform")
		return
	}
	c.HTML(http.StatusOK, "restorepasswordform.t.html", nil)
}

// RestorePasswordSubmitPageHandler processes a /restorepasswordsubmit form post request
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
			Message: "If you supplied your valid registration E-mail, then we sent you a message with a reset password link"})
}

type restorePasswordData struct {
	Email           string
	Confirmationkey string
}

func doRestorePasswordSubmit(c *gin.Context, d *restorePasswordData) (err error) {
	err = processRestorePasswordSubmitWithDb(d)
	if err == nil {
		// sendConfirmationEmail only produces 500 in case of failure
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
	// TODO: if there are no certificate files, use http an7
	confirmationLinkBase := shared.SitesProtocol() + "//" + scd.SiteRoot + shared.SitesPort() + "/changepasswordform"
	parameters := url.Values{"email": {d.Email}, "confirmationkey": {d.Confirmationkey}}
	u, err := url.Parse(confirmationLinkBase)
	apperror.GracefullyExitAppIf(err, "Unable to parse base URL for a confirmation link")
	u.RawQuery = parameters.Encode()
	confirmationLink := u.String()
	body := fmt.Sprintf(
		"Hello, to restore your password, please follow an restore link: <a href=%[1]s>%[1]s</a>",
		confirmationLink,
	)

	err = SendEmail(
		d.Email,
		"Restore password of semantic dictionary!",
		body)

	if err != nil {
		// We assume that failure to send an E-mail can be due to temporary
		// network issues
		apperror.Panic500AndLogAttackIf(err, c, "Failed to send a confirmation E-mail")
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
	sddb.FatalDatabaseErrorIf(err, "Error remembering that E-Mail was sent, error is %#v", err)
	return
}
