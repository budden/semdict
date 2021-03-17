package query

import (
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type senseDeleteConfirmParamsType = senseViewParamsType

// SenseDeleteConfirmHTMLTemplateParamsType are params for sensedeleteconfirm.t.html
type SenseDeleteConfirmHTMLTemplateParamsType = SenseViewHTMLTemplateParamsType

func SenseDeleteConfirmRequestHandler(c *gin.Context) {
	svp := &senseDeleteConfirmParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}

	paramValue := extractIdFromRequest(c, "senseid")
	svp.Senseid = paramValue
	dataFound, senseDataForEdit := readSenseFromDb(svp)

	if dataFound {
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"sensedeleteconfirm.t.html",
			SenseDeleteConfirmHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no sense (yet?) with id = «%d»", svp.Senseid)
	}
}
