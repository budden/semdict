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
	Senseid  int64 // должен быть здесь
	OWord    string
	Theme    string
	Phrase   string
	Ownerid  int32
	Action   string
}

// SenseEditSubmitPostHandler размещает данные о смысле
func SenseEditSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &senseEditSubmitDataType{}
	extractDataFromRequest(c, pad)
	if pad.Action == "save" {
		sanitizeData(pad)
		newSenseId := writeToDb(pad)
		// https://github.com/gin-gonic/gin/issues/444
		c.Redirect(http.StatusFound,
			"/sensebyidview/"+strconv.FormatInt(newSenseId, 10))
	} else if pad.Action == "delete" {
		c.Redirect(http.StatusFound,
			"/sensedeleteconfirm/"+strconv.FormatInt(pad.Senseid, 10))
		// spdp := &senseDeleteParamsType{Sduserid: pad.Sduserid, Senseid: pad.Senseid}
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Неизвестное действие в форме")
	}

}

func sanitizeData(pad *senseEditSubmitDataType) {
	// пример только с титульного листа https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(pad.OWord))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Английское слово может содержать только латинские буквы, цифры и пробелы")
	}
}

func extractDataFromRequest(c *gin.Context, pad *senseEditSubmitDataType) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	pad.Senseid = extractIdFromRequest(c, "senseid")
	pad.Phrase = c.PostForm("phrase")
	pad.OWord = c.PostForm("oword")
	pad.Theme = c.PostForm("theme")
	pad.Ownerid = int32(extractIdFromRequest(c, "ownerid"))
	pad.Action = c.PostForm("action")
}

func writeToDb(pad *senseEditSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavesense(:sduserid, :senseid, :oword, :theme, :phrase, :ownerid)`, pad)
	apperror.Panic500AndErrorIf(err1, "Не удалось обновить смысл")
	defer sddb.CloseRows(res)()
	dataFound := false
	for res.Next() {
		err1 = res.Scan(&newProposalid)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Нет идентификатора предложения с сервера")
	}
	return
}
