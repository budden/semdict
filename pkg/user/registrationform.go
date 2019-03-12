package user

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/database"
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
	ConfirmationID    string
}

// RegistrationFormSubmitPostHandler processes a registrationformsubmit form post request
func RegistrationFormSubmitPostHandler(c *gin.Context) {
	var rd RegistrationData
	rd.Nickname = c.PostForm("nickname")
	rd.Password = c.PostForm("password")
	rd.Registrationemail = c.PostForm("registrationemail")
	err := doRegistrationFormSubmit(&rd)
	message := "Check your E-Mail for a confirmation code, which will be valid for 10 minutes"
	if err != nil {
		message = err.Error()
	}
	c.HTML(http.StatusOK,
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
	parameters := url.Values{"nickname": {rd.Nickname}, "confirmationid": {rd.ConfirmationID}}
	u, err1 := url.Parse(confirmationLinkBase)
	if err1 != nil {
		unsorted.LogicalPanic("Very bad: unable to parse base URL for a confirmation link")
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

	return
}

var mapViolatedConstraintNameToMessage = map[string]string{
	"i_registrationattempt__confirmationid":    "You're lucky to hit a very seldom random number clash. Please retry a registration",
	"i_registrationattempt__registrationemail": "Someone is already trying to register with the same E-mail",
	"i_registrationattempt__nickname":          "Someone is already trying to register with the same Nickname",
	"i_sduser_registrationemail":               "There is already a user with the same E-mail",
	"i_sduser_nickname":                        "There is already a user with the same nickname"}

// processRegistrationFormSubmitWithDb inserts a registration attempt into sdusers_db
// If some "normal" error happens, it is returned in err. err.String() can be
// used to present an error to the user. In case of unexpected error, LogicalPanic is invoked
func processRegistrationFormSubmitWithDb(rd *RegistrationData) (err error) {

	writeSDUsersMutex.Lock()
	defer writeSDUsersMutex.Unlock()

	db, dbCloser := openSDUsersDb()
	defer dbCloser()

	db.MustExec("select delete_expired_registrationattempts()")

	var tx *sqlx.Tx
	tx, err = db.Beginx()
	if err != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to start transaction, error is %#v", err))
	}
	defer func() { database.RollbackIfActive(tx) }()

	tx.MustExec(`set transaction isolation level repeatable read`)

	rd.Calculatedhash, rd.Calculatedsalt = HashAndSaltPassword(rd.Password)
	rd.ConfirmationID = GenNonce(20)
	_, err = tx.NamedExec(
		`select process_registrationformsubmit(:nickname, :calculatedhash, :calculatedsalt, :registrationemail, :confirmationid)`,
		rd)
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		err = handleRegistrationAttemptInsertError(err)
	}
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
	unsorted.LogicalPanic(fmt.Sprintf("Unexpected error in the registrationformsubmit: %#v\n", err))
	return err
}
