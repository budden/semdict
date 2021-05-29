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
	// пример только с титульного листа https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Word = p.Sanitize(pad.Word)
	pad.Commentary = p.Sanitize(pad.Commentary)
	matched, err := regexp.Match(`^[0-9a-zA-Zа-яА-ЯёЁ\p{L}\- ]+$`, []byte(pad.Word))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Слово может содержать только русские или латинские буквы, цифры, знак '-' и пробелы")
	}
}

func LwsNewSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &lwsNewSubmitDataType{}
	extractLwsDataFromRequest(c, pad, true)
	sanitizeNewLwsData(pad)
	_ = makeNewLwsidInDb(pad)
	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/wordsearchresultform?dummyid=0&senseid="+strconv.FormatInt(pad.Senseid, 10))
}

func makeNewLwsidInDb(sap *lwsNewSubmitDataType) (id int64) {
	reply, err1 := sddb.NamedUpdateQuery(
		`insert into tlws (languageid, senseid, word, commentary, ownerid) 
			values (:languageid, :senseid, :word, :commentary, :sduserid) 
			returning id`, &sap)
	apperror.Panic500AndErrorIf(err1, "Не удалось вставить lws, извините")
	defer sddb.CloseRows(reply)()
	var dataFound bool
	for reply.Next() {
		err1 = reply.Scan(&id)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Вставка не вернула запись")
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении идентификатора свежего lws: %#v", err1)
	return
}
