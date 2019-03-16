package user

import (
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/budden/a/pkg/database"

	"github.com/budden/a/pkg/apperror"
	"github.com/budden/a/pkg/shared"
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
	appErr := doRegistrationFormSubmit(&rd)
	if appErr == nil {
		c.HTML(http.StatusOK,
			"general.html",
			shared.GeneralTemplateParams{
				Message: "Check your E-Mail for a confirmation code, which will be valid for 10 minutes"})
	} else {
		c.HTML(http.StatusOK,
			"general.html",
			shared.GeneralTemplateParams{Message: appErr.Message})
	}
}

func doRegistrationFormSubmit(rd *RegistrationData) (apperr *apperror.AppErr) {
	apperr = processRegistrationFormSubmitWithDb(rd)
	if apperr == nil {
		// sendConfirmationEmail only produces 500 in case of failure
		sendConfirmationEmail(rd)
	}
	return apperr
}

func sendConfirmationEmail(rd *RegistrationData) {
	confirmationLinkBase := "localhost:" + shared.WebServerPort + "/registrationconfirmation"
	parameters := url.Values{"nickname": {rd.Nickname}, "confirmationkey": {rd.ConfirmationKey}}
	u, err := url.Parse(confirmationLinkBase)
	apperror.GracefullyExitAppIf(err, "Unable to parse base URL for a confirmation link")
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

	if err != nil {
		// We assume that failure to send an E-mail can be due to temporary
		// network issues
		apperror.Panic500If(err, "Failed to send a confirmation E-mail")
	}

	noteRegistrationConfirmationEMailSentWithDb(rd)
	return
}

func noteRegistrationConfirmationEMailSentWithDb(rd *RegistrationData) {
	err := WithSDUsersDbTransaction(func(trans *database.TransactionType) (err1 error) {
		database.CheckDbAlive(trans.Conn)
		_, err1 = trans.Tx.NamedExec(
			`select note_registrationconfirmation_email_sent(:nickname, :confirmationkey)`,
			rd)
		return
	})
	database.FatalDatabaseErrorIf(err, database.SDUsersDb, "Error remembering that E-Mail was sent, error is %#v", err)
	return
}

var mapViolatedConstraintNameToMessage = map[string]string{
	"i_registrationattempt__confirmationkey":   "You're lucky to hit a very seldom random number clash. Please retry a registration",
	"i_registrationattempt__registrationemail": "Someone is already trying to register with the same E-mail",
	"i_registrationattempt__nickname":          "Someone is already trying to register with the same Nickname",
	"i_sduser_registrationemail":               "There is already a user with the same E-mail",
	"i_sduser_nickname":                        "There is already a user with the same nickname"}

func deleteExpiredRegistrationAttempts(trans *database.TransactionType) error {
	conn := trans.Conn
	tx := trans.Tx
	database.CheckDbAlive(conn)
	_, err1 := tx.Exec("select delete_expired_registrationattempts()")
	// it's not a fatal error (rare case!)
	apperror.Panic500If(err1,
		"Failed to register. Please try again later or contact us for assistance")
	database.CheckDbAlive(conn)
	err1 = tx.Commit()
	database.FatalDatabaseErrorIf(err1, conn,
		"Failed to commit after delete_expired_registrationattempts, error = %#v",
		err1)
	return nil
}

// processRegistrationFormSubmitWithDb inserts a registration attempt into sdusers_db
// If some "normal" error happens like non-unique nickname, it is returned in dberror.
func processRegistrationFormSubmitWithDb(rd *RegistrationData) *apperror.AppErr {

	db := database.SDUsersDb
	err := WithSDUsersDbTransaction(deleteExpiredRegistrationAttempts)
	database.FatalDatabaseErrorIf(err,
		db,
		"Failed around delete_expired_registrationattempts, %#v",
		err)

	err = WithSDUsersDbTransaction(func(trans *database.TransactionType) (err error) {
		rd.Calculatedhash, rd.Calculatedsalt = HashAndSaltPassword(rd.Password)
		rd.ConfirmationKey = GenNonce(20)
		database.CheckDbAlive(trans.Conn)
		_, err = trans.Tx.NamedExec(
			`select add_registrationattempt(:nickname, :calculatedhash, :calculatedsalt, :registrationemail, :confirmationkey)`,
			rd)
		if err == nil {
			database.CheckDbAlive(trans.Conn)
			err = trans.Tx.Commit()
		}
		return
	})
	return handleRegistrationAttemptInsertError(err)
}

func handleRegistrationAttemptInsertError(err error) *apperror.AppErr {
	//xt := reflect.TypeOf(err1).Kind()
	if e, ok := err.(*pq.Error); ok {
		if e.Code == PostgresqlErrorCodeUniqueViolation {
			message, found := mapViolatedConstraintNameToMessage[e.Constraint]
			if found {
				return apperror.NewAppErrf(message)
			}
		}
	}
	database.FatalDatabaseErrorIf(err, database.SDUsersDb, "Unexpected error in the registrationformsubmit, %#v\n", err)
	return nil
}
