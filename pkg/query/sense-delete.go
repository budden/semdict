package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type senseDeleteParamsType struct {
	Sduserid int64
	Senseid  int64
	Action   string
}

// SenseDeleteRequestHandler = POST sensedelete
func SenseDeleteRequestHandler(c *gin.Context) {
	// FIXME обрабатывать пустые черновики, например, много раз вызывать эту страницу и ни разу не вызвать пост.
	// Например, таймаут для черновика, или статус черновика, или даже не добавлять в базу данных
	// до первого сохранения.
	user.EnsureLoggedIn(c)
	svp := &senseDeleteParamsType{
		Sduserid: int64(user.GetSDUserIdOrZero(c)),
		Senseid:  extractIdFromRequest(c, "senseid"),
		Action:   c.PostForm("action"),
	}

	if svp.Action == "delete" {
		if svp.Sduserid != 1 /*tsar*/ {
			c.HTML(http.StatusOK,
				"general.t.html",
				shared.GeneralTemplateParams{Message: "Попросите об этом царя, пожалуйста, передав ему url формы подтверждения удаления"})
			return
		}
		deleteSenseFromDb(svp)
		c.HTML(http.StatusOK,
			"general.t.html",
			shared.GeneralTemplateParams{Message: "Смысл удалён успешно"})
	} else if svp.Action == "cancel" {
		c.HTML(http.StatusFound,
			"general.t.html",
			shared.GeneralTemplateParams{Message: "Вы отказались удалить смысл"})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Неизвестное действие в форме")
	}

}

func deleteSenseFromDb(spdp *senseDeleteParamsType) {
	reply, err1 := sddb.NamedUpdateQuery(
		`delete from tsense where id = :senseid returning id`, &spdp)
	apperror.Panic500AndErrorIf(err1, "Не удалось удалить смысл, извините")
	defer sddb.CloseRows(reply)()
	var dataFound bool
	for reply.Next() {
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Не удалось удалить смысл (возможно, он не ваш или не существует)")
	}
	return
}
