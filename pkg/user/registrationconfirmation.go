package user

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/jmoiron/sqlx"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

// RegistrationConfirmationPageHandler обрабатывает подтверждение регистрации
func RegistrationConfirmationPageHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	var rd RegistrationData
	// введите псевдоним и ключ подтверждения
	extractNicknameAndConfirmationKeyFromRequest(c, &rd)
	// обновить sdusers_db, заполнить userid
	processRegistrationConfirmationWithSDUsersDbStage1(&rd)
	// перевести пользователя в Sd Db. Если здесь произойдет сбой, пользователь сможет войти в систему,
	// (и не сможет зарегистрироваться снова), но будет отсутствовать в основном контенте db
	c.HTML(http.StatusMovedPermanently,
		"registrationconfirmation.t.html",
		gin.H{})
}

func extractNicknameAndConfirmationKeyFromRequest(c *gin.Context, rd *RegistrationData) {
	query := c.Request.URL.Query()
	nicknames, ok1 := query["nickname"]
	confirmationkeys, ok2 := query["confirmationkey"]

	if !ok1 || !ok2 ||
		len(nicknames) == 0 ||
		len(confirmationkeys) == 0 {
		apperror.Panic500If(apperror.ErrDummy, "Плохой URL-адрес подтверждения регистрации")
	}
	rd.Nickname = nicknames[0]
	rd.ConfirmationKey = confirmationkeys[0]
}

func processRegistrationConfirmationWithSDUsersDbStage1(rd *RegistrationData) {
	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		var reply *sqlx.Rows
		reply, err1 = trans.Tx.NamedQuery(
			`select * from process_registrationconfirmation(:confirmationkey, :nickname)`,
			rd)
		apperror.Panic500AndErrorIf(err1, "Не удалось подтвердить регистрацию, извините")
		defer reply.Close()
		for reply.Next() {
			err1 = reply.Scan(&rd.UserID)
			//fmt.Printf("UserID = %v\n", rd.UserID)
			sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении идентификатора нового пользователя, err = %#v", err1)
		}
		// hence err1 == nil
		return
	})
	// если у нас здесь ошибка, то это ошибка в коммите, поэтому она фатальна.
	sddb.FatalDatabaseErrorIf(err, "Не удалось пройти процедуру подтверждения регистрации, ошибка заключается в следующем %#v", err)
	return
}
