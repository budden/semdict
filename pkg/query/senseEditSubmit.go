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

type senseEditSubmitDataType struct {
	Proposalid       int64 // must be here
	Commonid         int64 // can be 0 if no origin (adding proposal)
	Languageid       int32
	Proposalstatus   string
	Phrase           string
	Word             string
	Phantom          bool // Does it make sense?
	Deletionproposed bool
	Ownerid          int32
}

// SenseEditSubmitPostHandler posts an sense data
func SenseEditSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &senseEditSubmitDataType{}
	extractDataFromRequest(c, pad)
	sanitizeData(pad)
	newProposalId := writeToDb(pad)
	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/sensebyidview/"+strconv.FormatInt(newProposalId, 10))
}

func sanitizeData(pad *senseEditSubmitDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Proposalstatus = p.Sanitize(pad.Proposalstatus)
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(pad.Word))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Word can only contain letters, digits and spaces")
	}
}

func extractDataFromRequest(c *gin.Context, pad *senseEditSubmitDataType) {
	pad.Proposalid = extractIdFromRequest(c, "proposalid")
	pad.Commonid = extractIdFromRequest(c, "commonid")
	pad.Proposalstatus = c.PostForm("proposalstatus")
	pad.Phrase = c.PostForm("phrase")
	pad.Word = c.PostForm("word")
	pad.Deletionproposed = extractCheckBoxFromRequest(c, "deletionproposed")
	pad.Ownerid = user.GetSDUserIdOrZero(c)
}

func writeToDb(pad *senseEditSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavepersonalsense(:ownerid, :commonid, :proposalid, :proposalstatus, :phrase, :word, :deletionproposed)`, pad)
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
