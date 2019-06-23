package query

import (
	"net/http"
	"strconv"

	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

type senseProposalAcceptOrRejectSubmitDataType struct {
	Proposalid     int64 // must be here
	Acceptorreject int64 // 1 = accept, 2 = reject
	Ownerid        int32
}

// SenseProposalAcceptOrRejectSubmitPostHandler posts an sense data
func SenseProposalAcceptOrRejectSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	paorsd := &senseProposalAcceptOrRejectSubmitDataType{}
	senseProposalAcceptOrRejectSubmitExtractDataFromRequest(c, paorsd)
	senseProposalAcceptOrRejectSubmitSanitizeData(paorsd)
	commonId := senseProposalAcceptOrRejectSubmitWriteToDb(paorsd)
	// https://github.com/gin-gonic/gin/issues/444
	if commonId == -1 {
		c.Redirect(http.StatusFound, "/")
	} else {
		c.Redirect(http.StatusFound,
			"/sensebyidview/"+strconv.FormatInt(commonId, 10))
	}
}

func senseProposalAcceptOrRejectSubmitSanitizeData(paorsd *senseProposalAcceptOrRejectSubmitDataType) {
	if 1 > paorsd.Acceptorreject || paorsd.Acceptorreject > 2 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Wrong value for «Acceptorreject»")
	}
}

func senseProposalAcceptOrRejectSubmitExtractDataFromRequest(
	c *gin.Context, paorsd *senseProposalAcceptOrRejectSubmitDataType) {
	apperror.Panic500AndErrorIf(apperror.ErrDummy, "FIXME fix here around")
	paorsd.Proposalid = extractIdFromRequest(c, "proposalid")
	paorsd.Acceptorreject = extractIdFromRequest(c, "acceptorreject")
	paorsd.Ownerid = user.GetSDUserIdOrZero(c)
}

// commonId means that the record was deleted
func senseProposalAcceptOrRejectSubmitWriteToDb(paorsd *senseProposalAcceptOrRejectSubmitDataType) (commonId int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavepersonalsense(:proposalid, :acceptorreject)`, paorsd)
	apperror.Panic500AndErrorIf(err1, "Failed to update a sense")
	dataFound := false
	for res.Next() {
		err1 = res.Scan(&commonId)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No common id from server")
	}
	return
}

/* Example of nested records in the template:

package main

import (
	"html/template"
	"log"
	"os"
)

func main() {
	type z struct{ Msg string; Child *z }
	v := z{Msg: "hi", Child: &z{Msg: "wow"}}
	master := "Greeting: {{ .Msg}}, {{ .Child.Msg}}"
	masterTmpl, err := template.New("master").Parse(master)
	if err != nil {
		log.Fatal(err)
	}
	if err := masterTmpl.Execute(os.Stdout, v); err != nil {
		log.Fatal(err)
	}
}

*/
