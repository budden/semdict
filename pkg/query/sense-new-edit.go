package query

import (
	"net/http"
	"strings"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type SenseNewEditHTMLTemplateParamsType = SenseViewHTMLTemplateParamsType

func SenseNewEditRequestHandler(c *gin.Context) {
	svp := &senseViewParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	sdfe := &senseDataForEditType{}
	oword := c.PostForm("oword")
	sdfe.OWord = convertWordpatternToNewWork(oword)
	c.HTML(http.StatusOK,
		"sensenewedit.t.html",
		SenseNewEditHTMLTemplateParamsType{Svp: svp, Sdfe: sdfe})
}

func convertWordpatternToNewWork(pattern string) string {
	return strings.Replace(pattern, "%", "", -1)
}
