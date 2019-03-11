package user

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/shared"
	"github.com/budden/a/pkg/unsorted"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

// Mutex we lock for any writes to sdusers_db to minimize parallelism at the db level
var writeSDUsersMutex sync.Mutex

// PlayWithEmail sends an email

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
}

// Filled by init
var mapViolatedConstraintNameToMessage = map[string]string{
	"i_registrationattempt__registrationemail": "Someone is already trying to register with the same E-mail",
	"i_registrationattempt__nickname":          "Someone is already trying to register with the same Nickname",
	"i_sduser_registrationemail":               "There is already a user with the same E-mail",
	"i_sduser_nickname":                        "There is already a user with the same nickname"}

// processRegistrationWithDb inserts a registration attempt into sdusers_db
// If some "normal" error happens, it is returned in err. err.String() can be
// used to present an error to the user. In case of unexpected error, LogicalPanic is invoked
func processRegistrationWithDb(rd *RegistrationData) (err error) {

	writeSDUsersMutex.Lock()
	defer writeSDUsersMutex.Unlock()

	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"

	db, dbCloser, err1 := database.OpenDb(url)
	if err1 != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to connect to Postgresql, error is %#v", err1))
	}
	defer dbCloser()
	var tx *sqlx.Tx
	db.MustExec("select delete_expired_registrationattempts()")
	tx, err = db.Beginx()
	defer func() { database.RollbackIfActive(tx) }()
	if err != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to start transaction, error is %#v", err))
	}
	rd.Calculatedhash, rd.Calculatedsalt = HashAndSaltPassword(rd.Password)
	tx.MustExec(`set transaction isolation level repeatable read`)
	_, err = tx.NamedExec(
		`select process_registrationformsubmit(:nickname, :calculatedhash, :calculatedsalt, :registrationemail)`,
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
		if e.Code == "23505" {
			message, found := mapViolatedConstraintNameToMessage[e.Constraint]
			if found {
				err = errors.New(message)
				return err
			}
		}
	}
	unsorted.LogicalPanic(fmt.Sprintf("Error in the registrationformsubmit: %#v\n", err))
	return err
}

// RegistrationFormSubmitPostHandler processes a registrationformsubmit form post request
func RegistrationFormSubmitPostHandler(c *gin.Context) {
	var rd RegistrationData
	rd.Nickname = c.PostForm("nickname")
	rd.Password = c.PostForm("password")
	rd.Registrationemail = c.PostForm("registrationemail")
	err := processRegistrationWithDb(&rd)
	message := "Check your E-Mail for a confirmation code, which will be valid for 10 minutes"
	if err != nil {
		message = err.Error()
	}
	c.HTML(http.StatusOK,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
}
