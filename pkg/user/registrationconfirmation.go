package user

import (
	"fmt"
	"net/http"

	"github.com/budden/a/pkg/apperror"
	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/shared"
	"github.com/gin-gonic/gin"
)

// RegistrationConfirmationPageHandler processes a registration confirmation
func RegistrationConfirmationPageHandler(c *gin.Context) {
	query := c.Request.URL.Query()
	message := "Registration confirmed. Now you can proceed to the <a href=/>Login page</a>"
	status := http.StatusOK
	confirmationkeys, ok1 := query["confirmationkey"]
	nicknames, ok2 := query["nickname"]

	if !ok1 || !ok2 ||
		len(confirmationkeys) == 0 ||
		len(nicknames) == 0 {
		apperror.Panic500If(apperror.ErrDummy, "Bad registration confirmation URL")
	}
	var rd RegistrationData
	rd.Nickname = nicknames[0]
	rd.ConfirmationKey = confirmationkeys[0]
	processRegistrationConfirmationWithSDUsersDbStage1(&rd)
	c.HTML(status,
		"general.html",
		shared.GeneralTemplateParams{Message: message})
	// processRegistrationConfirmationWithSDDb(&rd)
}

func processRegistrationConfirmationWithSDUsersDbStage1(rd *RegistrationData) {
	err := WithSDUsersDbTransaction(func(trans *database.TransactionType) (err1 error) {
		database.CheckDbAlive(trans.Conn)
		var reply *sqlx.Rows
		reply, err1 = trans.Tx.NamedQuery(
			`select * from process_registrationconfirmation(:confirmationkey, :nickname)`,
			rd)
		apperror.Panic500If(err1, "Failed to confirm registration, sorry")
		for reply.Next() {
			err1 = reply.Scan(&rd.UserID)
			fmt.Printf("UserID = %v\n", rd.UserID)
			database.FatalDatabaseErrorIf(err1, database.SDUsersDb, "Error obtaining id of a new user, err = %#v", err1)
		}
		// hence err1 == nil
		return
	})
	// if we have error here, it is an error in commit, so is fatal
	database.FatalDatabaseErrorIf(err, database.SDUsersDb, "Failed around registrationconfirmation, error is %#v", err)
	return
}
