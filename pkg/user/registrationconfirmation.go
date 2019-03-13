package user

import (
	"net/http"

	"github.com/jmoiron/sqlx"

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
		status = http.StatusInternalServerError
		message = "Bad registration confirmation URL"
	} else {
		var rd RegistrationData
		rd.Nickname = nicknames[0]
		rd.ConfirmationKey = confirmationkeys[0]
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
	err = WithSDUsersDbTransaction(func(tx *sqlx.Tx) (err error) {
		_, err = tx.NamedExec(
			`select process_registrationconfirmation(:confirmationkey, :nickname)`,
			rd)
		return
	})
	return
}
