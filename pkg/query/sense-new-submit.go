package query

import (
	"net/http"
	"strconv"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type senseAddParamsType struct {
	Sduserid int64
	OWord    string
}

type senseNewSubmitDataType = senseEditSubmitDataType

func sanitizeNewSenseData(pad *senseNewSubmitDataType) {
	sanitizeData(pad)
}

func extractDataFromNewSubmitRequest(c *gin.Context, pad *senseEditSubmitDataType) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	pad.OWord = c.PostForm("oword")
	pad.Theme = c.PostForm("theme")
	pad.Phrase = c.PostForm("phrase")
}

func SenseNewSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &senseNewSubmitDataType{}
	extractDataFromNewSubmitRequest(c, pad)
	sanitizeNewSenseData(pad)
	newSenseId := makeNewSenseidInDb(pad)
	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/wordsearchresultform?dummyid=0&senseid="+strconv.FormatInt(newSenseId, 10))
}

func makeNewSenseidInDb(sap *senseNewSubmitDataType) (id int64) {
	reply, err1 := sddb.NamedUpdateQuery(
		`insert into tsense (ownerid, oword, theme, phrase) 
			values (:sduserid, :oword, :theme, :phrase) 
			returning id`, &sap)
	apperror.Panic500AndErrorIf(err1, "Не удалось вставить смысл, извините")
	defer sddb.CloseRows(reply)()
	var dataFound bool
	for reply.Next() {
		err1 = reply.Scan(&id)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Вставка не вернула запись")
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибочное получение свежего смысла: %#v", err1)
	return
}
