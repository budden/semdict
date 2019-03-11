package user

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/shared"
	"github.com/budden/a/pkg/unsorted"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

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

// See init
var messageFormats gin.H

// processRegistrationWithDb inserts a registration attempt into sdusers_db
// If some "normal" error happens, it is returned in err. err.String() can be
// used to present an error to the user. In case of unexpected error, LogicalPanic is invoked
func processRegistrationWithDb(rd *RegistrationData) (err error) {
	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"

	db, dbCloser, err1 := database.OpenDb(url)
	if err1 != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to connect to Postgresql, error is %#v", err1))
	}
	defer dbCloser()
	var tx *sqlx.Tx
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
	//xt := reflect.TypeOf(err1).Kind()
	if err != nil {
		switch e := interface{}(err).(type) {
		case *pq.Error:
			if e.Code == "23505" {
				fmt.Printf("Duplicate key in %s", e.Column)
			} else {
				unsorted.LogicalPanic(fmt.Sprintf("Error inserting: %#v\n", err))
			}
		default:
			unsorted.LogicalPanic(fmt.Sprintf("Error insertiing: %#v\n", err))
		}
	} else {
		fmt.Printf("Inserted %#v\n", 10050)
	}
	return
}

// RegistrationFormSubmitPostHandler processes a registrationformsubmit form post request
func RegistrationFormSubmitPostHandler(c *gin.Context) {
	var rd RegistrationData
	rd.Nickname = c.PostForm("Nickname")
	rd.Password = c.PostForm("Password")
	rd.Registrationemail = c.PostForm("Registrationemail")
	err := processRegistrationWithDb(&rd)
	message := "Check your E-Mail for a confirmation code, which will be valid for 10 minutes"
	if err != nil {
		message = err.Error()
	}
	c.HTML(http.StatusOK,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
}
