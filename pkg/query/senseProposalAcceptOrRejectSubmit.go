package query

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/microcosm-cc/bluemonday"

	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

type senseProposalAcceptOrRejectSubmitDataType struct {
	Proposalid       int64 // must be here
	Commonid         int64 // can be 0 if no origin (adding senseProposal)
	Languageid       int32
	Proposalstatus   string
	Phrase           string
	Word             string
	Phantom          bool // Does it make sense?
	Deletionproposed bool // Not used! FIXME
	Ownerid          int32
}

// SenseProposalAcceptOrRejectSubmitPostHandler posts an sense data
func SenseProposalAcceptOrRejectSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	paorsd := &senseProposalAcceptOrRejectSubmitDataType{}
	senseProposalAcceptOrRejectSubmitExtractDataFromRequest(c, paorsd)
	senseProposalAcceptOrRejectSubmitSanitizeData(paorsd)
	newProposalId := senseProposalAcceptOrRejectSubmitWriteToDb(paorsd)
	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/sensebyidview/"+strconv.FormatInt(newProposalId, 10))
}

func senseProposalAcceptOrRejectSubmitSanitizeData(paorsd *senseProposalAcceptOrRejectSubmitDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	paorsd.Proposalstatus = p.Sanitize(paorsd.Proposalstatus)
	paorsd.Phrase = p.Sanitize(paorsd.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(paorsd.Word))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Word can only contain letters, digits and spaces")
	}
}

// убедиться, что не сломалось. Закоммитить. Продолжать реализацию слияния смыслов
//  сделать удаление и добавление смысла. Сразу историю?

func senseProposalAcceptOrRejectSubmitExtractDataFromRequest(
	c *gin.Context, paorsd *senseProposalAcceptOrRejectSubmitDataType) {
	apperror.Panic500AndErrorIf(apperror.ErrDummy, "FIXME fix here around")
	paorsd.Proposalid = extractIdFromRequest(c, "proposalid")
	paorsd.Commonid = extractIdFromRequest(c, "commonid")
	paorsd.Proposalstatus = c.PostForm("proposalstatus")
	paorsd.Phrase = c.PostForm("phrase")
	paorsd.Word = c.PostForm("word")
	paorsd.Ownerid = user.GetSDUserIdOrZero(c)
}

func senseProposalAcceptOrRejectSubmitWriteToDb(paorsd *senseProposalAcceptOrRejectSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavepersonalsense(:ownerid, :commonid, :proposalid, :proposalstatus, :phrase, :word, false)`, paorsd)
	apperror.Panic500AndErrorIf(err1, "Failed to update a sense")
	dataFound := false
	for res.Next() {
		err1 = res.Scan(&newProposalid)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No proposal id from server")
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
