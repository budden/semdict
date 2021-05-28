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
	Sduserid   int64 // извлечённые из сессии
	Lwsid      int64 // должны быть для сохранения и удаления
	Languageid int64
	Word       string
	Senseid    int64
	Commentary string
	Action     string
}

// LwsEditSubmitPostHandler размещает данные lws
func LwsEditSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &lwsEditSubmitDataType{}
	extractLwsDataFromRequest(c, pad, false)
	if pad.Action == "save" {
		sanitizeLwsEditData(pad)
		writeLwsToDb(pad)
		// https://github.com/gin-gonic/gin/issues/444
		c.Redirect(http.StatusFound,
			"/sensebyidview/"+strconv.FormatInt(pad.Senseid, 10))
	} else if pad.Action == "delete" {
		c.Redirect(http.StatusFound,
			"/lwsdeleteconfirm/"+strconv.FormatInt(pad.Senseid, 10)+"/"+
				strconv.FormatInt(pad.Languageid, 10)+"/"+
				strconv.FormatInt(pad.Lwsid, 10))
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Неизвестное действие в форме")
	}

}

func sanitizeLwsEditData(pad *lwsEditSubmitDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Commentary = p.Sanitize(pad.Commentary)
	matched, err := regexp.Match(`^[0-9a-zA-Zа-яА-ЯёЁ\p{L} ]+$`, []byte(pad.Word))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Слово может содержать только русские и латинские буквы, цифры и пробелы")
	}
}

func extractLwsDataFromRequest(c *gin.Context, pad *lwsEditSubmitDataType, forInsert bool) {
	pad.Sduserid = int64(user.GetSDUserIdOrZero(c))
	if forInsert {
		pad.Lwsid = 0
	} else {
		pad.Lwsid = extractIdFromRequest(c, "lwsid")
	}
	pad.Languageid = extractIdFromRequest(c, "languageid")
	pad.Word = c.PostForm("word")
	pad.Senseid = extractIdFromRequest(c, "senseid")
	pad.Commentary = c.PostForm("Commentary")
	pad.Action = c.PostForm("action")
}

func writeLwsToDb(pad *lwsEditSubmitDataType) (newProposalid int64) {
	res, err1 := sddb.NamedUpdateQuery(
		`select fnsavelws(:sduserid, :lwsid, :languageid, :word, :senseid, :commentary, 'save')`, pad)
	apperror.Panic500AndErrorIf(err1, "Не удалось обновить lws")
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
