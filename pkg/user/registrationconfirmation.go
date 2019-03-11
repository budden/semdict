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
	confirmationID := c.Param("confirmationid")
	err := processRegistrationConfirmationWithDb(confirmationID)
	message := "Registration confirmed. Now you can proceed to the <a href=/>Login page</a>"
	status := http.StatusOK
	if err != nil {
		status = http.StatusInternalServerError
		message = err.Error()
	}
	c.HTML(status,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
}

func processRegistrationConfirmationWithDb(confirmationID string) (err error) {

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
		`select process_registrationconfirmation(:confirmationid)`,
		confirmationID)
	if err == nil {
		err = tx.Commit()
	}
	//if err != nil {
	//	err = handleRegistrationAttemptInsertError(err)
	//}
	return
}
