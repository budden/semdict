package query

import (
	"net/http"
	"strings"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type SenseNewEditHTMLTemplateParamsType = SenseViewHTMLTemplateParamsType

func SenseNewEditRequestHandler(c *gin.Context) {
	svp := &senseViewParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	sdfe := &senseDataForEditType{}
	oword := c.Query("oword")
	sdfe.OWord = convertWordpatternToNewWork(oword)
	sdfe.Allth = AllKnownThemes()
	c.HTML(http.StatusOK,
		"sensenewedit.t.html",
		SenseNewEditHTMLTemplateParamsType{Svp: svp, Sdfe: sdfe})
}

type ThemeRecord struct {
	Theme string
}

func AllKnownThemes() (records []*ThemeRecord) {
	var queryText string
	queryText = `select distinct theme from tsense order by theme`
	// https://stackoverflow.com/questions/56178312/run-a-sql-query-without-parameters
	reply, err1 := sddb.ReadQuery(queryText)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)()
	records = make([]*ThemeRecord, 0)
	var last int
	for last = 0; reply.Next(); last++ {
		thr := &ThemeRecord{}
		err1 = reply.StructScan(thr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining a theme: %#v", err1)
		records = append(records, thr)
	}
	return
}

func convertWordpatternToNewWork(pattern string) string {
	return strings.Replace(pattern, "%", "", -1)
}
