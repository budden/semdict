package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type lwsDeleteParamsType struct {
	Sduserid int64
	Lwsid    int64
	Action   string
}

// LwsDeleteRequestHandler = POST lwsdelete
func LwsDeleteRequestHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	svp := &lwsDeleteParamsType{
		Sduserid: int64(user.GetSDUserIdOrZero(c)),
		Lwsid:    extractIdFromRequest(c, "lwsid"),
		Action:   c.PostForm("action"),
	}

	if svp.Action == "delete" {
		if svp.Sduserid != 1 /*tsar*/ {
			c.HTML(http.StatusOK,
				"general.t.html",
				shared.GeneralTemplateParams{Message: "Попросите об этом царя, пожалуйста, передав ему url формы подтверждения удаления"})
			return
		}
		deleteLwsFromDb(svp)
		c.HTML(http.StatusOK,
			"general.t.html",
			shared.GeneralTemplateParams{Message: "Перевод удалён успешно"})
	} else if svp.Action == "cancel" {
		c.HTML(http.StatusFound,
			"general.t.html",
			shared.GeneralTemplateParams{Message: "Вы отказались удалить перевод"})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Неизвестное действие в форме")
	}

}

func deleteLwsFromDb(spdp *lwsDeleteParamsType) {
	reply, err1 := sddb.NamedUpdateQuery(
		`delete from tlws where id = :lwsid returning id`, &spdp)
	apperror.Panic500AndErrorIf(err1, "Не удалось удалить перевод, извините")
	defer sddb.CloseRows(reply)()
	var dataFound bool
	for reply.Next() {
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Не удалось удалить перевод (возможно, он не ваш или не существует)")
	}
	return
}
