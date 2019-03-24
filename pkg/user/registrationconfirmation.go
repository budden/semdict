package user

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/jmoiron/sqlx"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

// RegistrationConfirmationPageHandler processes a registration confirmation
func RegistrationConfirmationPageHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	var rd RegistrationData
	// fill nickname and confirmationkey
	extractNicknameAndConfirmationKeyFromRequest(c, &rd)
	// update sdusers_db, fill userid
	processRegistrationConfirmationWithSDUsersDbStage1(&rd)
	// promote the user to Sd Db. If we crash here, user will be able to login,
	// (and unable to register again), but wil be missing from the main content db
	c.HTML(http.StatusMovedPermanently,
		"registrationconfirmation.html",
		gin.H{})
}

func extractNicknameAndConfirmationKeyFromRequest(c *gin.Context, rd *RegistrationData) {
	query := c.Request.URL.Query()
	nicknames, ok1 := query["nickname"]
	confirmationkeys, ok2 := query["confirmationkey"]

	if !ok1 || !ok2 ||
		len(nicknames) == 0 ||
		len(confirmationkeys) == 0 {
		apperror.Panic500If(apperror.ErrDummy, "Bad registration confirmation URL")
	}
	rd.Nickname = nicknames[0]
	rd.ConfirmationKey = confirmationkeys[0]
}

func processRegistrationConfirmationWithSDUsersDbStage1(rd *RegistrationData) {
	err := WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		var reply *sqlx.Rows
		reply, err1 = trans.Tx.NamedQuery(
			`select * from process_registrationconfirmation(:confirmationkey, :nickname)`,
			rd)
		apperror.Panic500If(err1, "Failed to confirm registration, sorry")
		for reply.Next() {
			err1 = reply.Scan(&rd.UserID)
			//fmt.Printf("UserID = %v\n", rd.UserID)
			sddb.FatalDatabaseErrorIf(err1, sddb.SDUsersDb, "Error obtaining id of a new user, err = %#v", err1)
		}
		// hence err1 == nil
		return
	})
	// if we have error here, it is an error in commit, so is fatal
	sddb.FatalDatabaseErrorIf(err, sddb.SDUsersDb, "Failed around registrationconfirmation, error is %#v", err)
	return
}
