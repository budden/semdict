package query

import (
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type lwsDeleteConfirmParamsType = senseViewParamsType

// LwsDeleteConfirmHTMLTemplateParamsType являются параметрами для sensedeleteconfirm.t.html
type LwsDeleteConfirmHTMLTemplateParamsType = SenseViewHTMLTemplateParamsType

func LwsDeleteConfirmRequestHandler(c *gin.Context) {
	svp := &lwsDeleteConfirmParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}

	paramValue := extractIdFromRequest(c, "lwsid")
	svp.Senseid = paramValue
	dataFound, lwsDataForEdit := readLwsFromDb(svp)

	if dataFound {
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"lwsdeleteconfirm.t.html",
			LwsDeleteConfirmHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Извините, нет смысла (пока?) с id = «%d»", svp.Senseid)
	}
}
