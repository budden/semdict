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
	Sduserid int64
	Senseid  int64 // must be here
	OWord    string
	Theme    string
	Phrase   string
	Ownerid  int32
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
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(pad.OWord))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "English word can only contain latin letters, digits and spaces")
	}
}

func extractDataFromRequest(c *gin.Context, pad *senseEditSubmitDataType) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	pad.Senseid = extractIdFromRequest(c, "senseid")
	pad.Phrase = c.PostForm("phrase")
	pad.OWord = c.PostForm("oword")
	pad.Ownerid = int32(extractIdFromRequest(c, "ownerid"))
}

func writeToDb(pad *senseEditSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavesense(:sduserid, :senseid, :oword, :theme, :phrase, :ownerid)`, pad)
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
