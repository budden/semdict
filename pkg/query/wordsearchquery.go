package query

// Возвращает выборку подходящих слов в формате JSON

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/gin-gonic/gin"
)

type wordSearchQueryParams struct {
	Dummyid     int32 // не имеет значения
	Wordpattern string
	Offset      int32
	Limit       int32
}

type wordSearchQueryRecord struct {
	Id         int32
	Languageid int32
	Phrase     string
	Word       string
}

// WordSearchQueryRouteHandler - обработчик для "/wordsearchquery".
func WordSearchQueryRouteHandler(c *gin.Context) {
	var frp wordSearchQueryParams

	frp.Wordpattern = c.Query("wordpattern")
	frp.Offset = int32(GetZeroOrOneNonNegativeIntFormValue(c, "offset"))
	frp.Limit = int32(GetZeroOrOneNonNegativeIntFormValue(c, "limit"))
	LimitLimit(&frp.Limit)

	// Извлечь параметры из запроса
	// frp.Id = extractIdFromRequest(c)

	// Прочитать данные из базы данных. Если нет данных, паниковать
	fd := readWordSearchQueryFromDb(&frp)
	// Выдать как JSON
	c.JSON(http.StatusOK, fd)
}

func readWordSearchQueryFromDb(frp *wordSearchQueryParams) (fd []*wordSearchQueryRecord) {
	reply, err1 := sddb.NamedReadQuery(
		`select id, languageid, word, phrase from tsense 
			where word like :wordpattern
			order by word, languageid, id offset :offset limit :limit`, frp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	fd = make([]*wordSearchQueryRecord, frp.Limit)
	var last int
	for last = 0; reply.Next(); last++ {
		fd[last] = &wordSearchQueryRecord{}
		err1 = reply.StructScan(fd[last])
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	}
	fd = fd[:last]
	return
}
