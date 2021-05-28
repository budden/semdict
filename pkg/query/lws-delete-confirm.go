package query

import (
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type lwsDeleteConfirmParamsType = senseViewParamsType

// LwsDeleteConfirmHTMLTemplateParamsType являются параметрами для sensedeleteconfirm.t.html
type LwsDeleteConfirmHTMLTemplateParamsType = SenseViewHTMLTemplateParamsType

func LwsDeleteConfirmRequestHandler(c *gin.Context) {

	lnep := &lwsEditParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	lnep.Senseid = extractIdFromRequest(c, "senseid")
	lnep.Languageid = extractIdFromRequest(c, "languageid")
	lnep.Lwsid = extractIdFromRequest(c, "lwsid")

	lned := readLwsEditDataFromDb(lnep)

	phraseHTML := template.HTML(lned.Phrase)

	c.HTML(http.StatusOK,
		"lwsdeleteconfirm.t.html",
		lwsNewEditHTMLTemplateParamsType{Lep: lnep, Led: lned, Phrase: phraseHTML})

}
