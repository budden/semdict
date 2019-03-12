package user

import (
	"fmt"
	"net/http"

	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/shared"
	"github.com/budden/a/pkg/unsorted"
	"github.com/gin-gonic/gin"
)

// RegistrationConfirmationPageHandler processes a registration confirmation
func RegistrationConfirmationPageHandler(c *gin.Context) {
	query := c.Request.URL.Query()
	message := "Registration confirmed. Now you can proceed to the <a href=/>Login page</a>"
	status := http.StatusOK
	confirmationids, ok1 := query["confirmationid"]
	nicknames, ok2 := query["nickname"]

	if !ok1 || !ok2 ||
		len(confirmationids) == 0 ||
		len(nicknames) == 0 {
		status = http.StatusInternalServerError
		message = "Bad registration confirmation URL"
	} else {
		var rd RegistrationData
		rd.Nickname = nicknames[0]
		rd.ConfirmationID = confirmationids[0]
		err := processRegistrationConfirmationWithDb(&rd)
		if err != nil {
			status = http.StatusInternalServerError
			message = err.Error()
		}
	}
	c.HTML(status,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
}

func processRegistrationConfirmationWithDb(rd *RegistrationData) (err error) {

	writeSDUsersMutex.Lock()
	defer writeSDUsersMutex.Unlock()
	db, dbCloser := openSDUsersDb()
	defer dbCloser()

	var tx *sqlx.Tx
	tx, err = db.Beginx()
	if err != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to start transaction, error is %#v", err))
	}
	defer func() { database.RollbackIfActive(tx) }()

	tx.MustExec(`set transaction isolation level repeatable read`)

	_, err = tx.NamedExec(
		`select process_registrationconfirmation(:confirmationid, :nickname)`,
		rd)
	if err == nil {
		err = tx.Commit()
	}
	//if err != nil {
	//	err = handleRegistrationAttemptInsertError(err)
	//}
	return
}
