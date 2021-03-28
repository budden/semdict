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

type lwsEditSubmitDataType struct {
	Sduserid int64
	Lwsid    int64 // must be here
	OWord    string
	Theme    string
	Phrase   string
	Ownerid  int32
	Action   string
}

// LwsEditSubmitPostHandler posts an lws data
func LwsEditSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &lwsEditSubmitDataType{}
	extractLwsDataFromRequest(c, pad)
	if pad.Action == "save" {
		sanitizeLwsEditData(pad)
		newLwsId := writeLwsToDb(pad)
		// https://github.com/gin-gonic/gin/issues/444
		c.Redirect(http.StatusFound,
			"/sensebyidview/"+strconv.FormatInt(newLwsId, 10))
	} else if pad.Action == "delete" {
		c.Redirect(http.StatusFound,
			"/sensedeleteconfirm/"+strconv.FormatInt(pad.Lwsid, 10))
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Unknown action in the form")
	}

}

func sanitizeLwsEditData(pad *lwsEditSubmitDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(pad.OWord))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "English word can only contain latin letters, digits and spaces")
	}
}

func extractLwsDataFromRequest(c *gin.Context, pad *lwsEditSubmitDataType) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	pad.Lwsid = extractIdFromRequest(c, "lwsid")
	pad.Phrase = c.PostForm("phrase")
	pad.OWord = c.PostForm("oword")
	pad.Theme = c.PostForm("theme")
	pad.Ownerid = int32(extractIdFromRequest(c, "ownerid"))
	pad.Action = c.PostForm("action")
}

func writeLwsToDb(pad *lwsEditSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavelws(:sduserid, :senseid, :oword, :theme, :phrase, :ownerid)`, pad)
	apperror.Panic500AndErrorIf(err1, "Failed to update a lws")
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
