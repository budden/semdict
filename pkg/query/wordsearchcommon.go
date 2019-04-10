package query

// Общая часть для /wordsearchquery и /wordsearchresultform

import (
	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

// параметры из URL
type wordSearchQueryParams struct {
	Dummyid     int32 // не имеет значения
	Wordpattern string
	Offset      int32
	Limit       int32 // 0 - значит «без ограничения»
}

type wordSearchQueryRecord struct {
	Id           int32
	Languageid   int32
	Languageslug string
	Phrase       string
	Word         string
}

func wordSearchCommonPart(c *gin.Context) (frp *wordSearchQueryParams, fd []*wordSearchQueryRecord) {
	frp = &wordSearchQueryParams{}

	frp.Wordpattern = c.Query("wordpattern")
	if frp.Wordpattern == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Empty search pattern")
	}

	frp.Offset = int32(GetZeroOrOneNonNegativeIntFormValue(c, "offset"))
	frp.Limit = int32(GetZeroOrOneNonNegativeIntFormValue(c, "limit"))
	LimitLimit(&frp.Limit)

	// Прочитать данные из базы данных. Если нет данных, паниковать
	fd = readWordSearchQueryFromDb(frp)
	return
}

func readWordSearchQueryFromDb(frp *wordSearchQueryParams) (fd []*wordSearchQueryRecord) {
	reply, err1 := sddb.NamedReadQuery(
		`select id, languageid, languageslug, word, phrase from vsense 
			where word like :wordpattern
			order by word, languageslug, id offset :offset limit :limit`, frp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	fd = make([]*wordSearchQueryRecord, frp.Limit)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &wordSearchQueryRecord{}
		err1 = reply.StructScan(wsqr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
		fd[last] = wsqr
	}
	fd = fd[:last]
	return
}
