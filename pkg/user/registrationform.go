package user

import (
	"errors"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/shared"
	"github.com/budden/a/pkg/unsorted"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// RegistrationFormPageHandler renders a /registrationform page
func RegistrationFormPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK,
		"registrationform.html",
		shared.GeneralTemplateParams{Message: "Search Form"})
}

// RegistrationData is a transient struct containing data obtained from a /registrationformsubmit query
// as well as some of calculated data
type RegistrationData struct {
	Nickname          string
	Password          string
	Registrationemail string
	Calculatedhash    string
	Calculatedsalt    string
	ConfirmationKey   string
}

// RegistrationFormSubmitPostHandler processes a registrationformsubmit form post request
func RegistrationFormSubmitPostHandler(c *gin.Context) {
	var rd RegistrationData
	rd.Nickname = c.PostForm("nickname")
	rd.Password = c.PostForm("password")
	rd.Registrationemail = c.PostForm("registrationemail")
	err := doRegistrationFormSubmit(&rd)
	status := http.StatusOK
	message := "Check your E-Mail for a confirmation code, which will be valid for 10 minutes"
	if err != nil {
		log.Println("Error sending confirmation E-mail: ", err)
		status = http.StatusInternalServerError
		message = "Error sending confirmation E-mail. Retry registration after 10 minutes or ask for assistance"
	}
	c.HTML(status,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
}

func doRegistrationFormSubmit(rd *RegistrationData) (err error) {
	err = processRegistrationFormSubmitWithDb(rd)
	if err != nil {
		return
	}
	err = sendConfirmationEmail(rd)
	return
}

func sendConfirmationEmail(rd *RegistrationData) (err error) {
	confirmationLinkBase := "localhost:" + shared.WebServerPort + "/registrationconfirmation"
	parameters := url.Values{"nickname": {rd.Nickname}, "confirmationkey": {rd.ConfirmationKey}}
	u, err1 := url.Parse(confirmationLinkBase)
	if err1 != nil {
		log.Fatal("Very bad: unable to parse base URL for a confirmation link")
	}
	u.RawQuery = parameters.Encode()
	confirmationLink := u.String()
	body := fmt.Sprintf(
		"Hello, %s!\nTo activate your account, please follow an <a href=%s>activation link</a>",
		// FIXME should Nickname need html escaping?
		html.EscapeString(rd.Nickname),
		confirmationLink)

	err = SendEmail(
		rd.Registrationemail,
		"Welcome to semantic dictionary!",
		body)

	if err == nil {
		err = noteRegistrationConfirmationEMailSentWithDb(rd)
	}

	return
}

func noteRegistrationConfirmationEMailSentWithDb(rd *RegistrationData) (err error) {
	err = WithSDUsersDbTransaction(func(tx *sqlx.Tx) (err error) {
		_, err = tx.NamedExec(
			`select note_registrationconfirmation_email_sent(:nickname, :confirmationkey)`,
			rd)
		return
	})
	if err != nil {
		message := fmt.Sprintf("Error remembering that E-Mail was sent, error is %#v", err)
		err = errors.New(message)
	}
	return
}

var mapViolatedConstraintNameToMessage = map[string]string{
	"i_registrationattempt__confirmationkey":   "You're lucky to hit a very seldom random number clash. Please retry a registration",
	"i_registrationattempt__registrationemail": "Someone is already trying to register with the same E-mail",
	"i_registrationattempt__nickname":          "Someone is already trying to register with the same Nickname",
	"i_sduser_registrationemail":               "There is already a user with the same E-mail",
	"i_sduser_nickname":                        "There is already a user with the same nickname"}

// processRegistrationFormSubmitWithDb inserts a registration attempt into sdusers_db
// If some "normal" error happens, it is returned in err. err.String() can be
// used to present an error to the user. In case of unexpected error, LogicalPanicIf is invoked
func processRegistrationFormSubmitWithDb(rd *RegistrationData) (err error) {

	err = WithSDUsersDbTransaction(func(tx *sqlx.Tx) (err error) {
		tx.MustExec("select delete_expired_registrationattempts()")
		err = tx.Commit()
		return
	})
	if err != nil {
		return
	}
	err = WithSDUsersDbTransaction(func(tx *sqlx.Tx) (err error) {
		rd.Calculatedhash, rd.Calculatedsalt = HashAndSaltPassword(rd.Password)
		rd.ConfirmationKey = GenNonce(20)
		_, err = tx.NamedExec(
			`select add_registrationattempt(:nickname, :calculatedhash, :calculatedsalt, :registrationemail, :confirmationkey)`,
			rd)
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			err = handleRegistrationAttemptInsertError(err)
		}
		return
	})
	return
}

func handleRegistrationAttemptInsertError(err error) error {
	//xt := reflect.TypeOf(err1).Kind()
	switch e := interface{}(err).(type) {
	case *pq.Error:
		if e.Code == PostgresqlErrorCodeUniqueViolation {
			message, found := mapViolatedConstraintNameToMessage[e.Constraint]
			if found {
				err = errors.New(message)
				return err
			}
		}
	}
	unsorted.LogicalPanicIf(err, "Unexpected error in the registrationformsubmit\n")
	return err
}
