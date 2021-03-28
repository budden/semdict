package query

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/microcosm-cc/bluemonday"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type lwsAddParamsType struct {
	Sduserid int64
	OWord    string
}

type lwsNewSubmitDataType = lwsEditSubmitDataType

func sanitizeNewLwsData(pad *lwsNewSubmitDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Zа-яА-ЯёЁ\p{L} ]+$`, []byte(pad.OWord))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Word can only contain Russian or latin letters, digits and spaces")
	}
}

func extractDataFromLwsNewSubmitRequest(c *gin.Context, pad *lwsEditSubmitDataType) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	pad.OWord = c.PostForm("oword")
	pad.Theme = c.PostForm("theme")
	pad.Phrase = c.PostForm("phrase")
}

func LwsNewSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &lwsNewSubmitDataType{}
	extractDataFromLwsNewSubmitRequest(c, pad)
	sanitizeNewLwsData(pad)
	newLwsId := makeNewLwsidInDb(pad)
	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/lwsbyidview/"+strconv.FormatInt(newLwsId, 10))
}

func makeNewLwsidInDb(sap *lwsNewSubmitDataType) (id int64) {
	reply, err1 := sddb.NamedUpdateQuery(
		`insert into tlws (languageid, senseid, word, commentary) 
			values (:languageid, :senseid, :word, :commentary) 
			returning id`, &sap)
	apperror.Panic500AndErrorIf(err1, "Failed to insert a lws, sorry")
	var dataFound bool
	for reply.Next() {
		err1 = reply.Scan(&id)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Insert didn't return a record")
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining id of a fresh lws: %#v", err1)
	return
}
