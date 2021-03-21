package query

// Общая часть для /wordsearchquery и /wordsearchresultform

import (
	"database/sql"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

/* Мы добились, чтобы url запроса
заполнял форму поиска и форму результата одинаково.
Теперь надо встроить начинку формы поиска в форму результата
(перетащить в wordsearchform-inner.t.html побольше начинки)

ещё один вопрос - куда деть нужный javascript.  */

// параметры из URL
type wordSearchQueryParams struct {
	Dummyid     int32 // не имеет значения
	Wordpattern string
	Languageid  int64 // 0 значит «все»
	// Эти поля не вводятся пользователем
	Sduserid int32 // 0 для незарег. польз.
	Offset   int32
	Limit    int32 // 0 - значит «без ограничения»
}

type wordSearchQueryRecord struct {
	Id int64
	// Sdusernickname sql.NullString
	Oword   string
	Theme   string
	Phrase  string
	Ownerid int64
	Lwsjson sql.NullString
	// Proposalid       sql.NullInt64 // is non-null when this record is a proposal.

}

func wordSearchCommonPart(c *gin.Context) (wsqp *wordSearchQueryParams, fd []*wordSearchQueryRecord) {
	wsqp = getWordSearchQueryParamsFromRequest(c)

	if wsqp.Wordpattern == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Empty search pattern")
	}

	wsqp.Offset = int32(GetZeroOrOneNonNegativeIntFormValue(c, "offset"))
	wsqp.Limit = int32(GetZeroOrOneNonNegativeIntFormValue(c, "limit"))
	LimitLimit(&wsqp.Limit)
	wsqp.Sduserid = user.GetSDUserIdOrZero(c)

	fd = readWordSearchQueryFromDb(wsqp)
	return
}

// select * from tsense where to_tsvector(phrase)||to_tsvector(word) @@ 'go';
// select row_to_json(x) from (select * from tsense) x;
// https://eax.me/postgresql-full-text-search/

func readWordSearchQueryFromDb(wsqp *wordSearchQueryParams) (
	fd []*wordSearchQueryRecord) {
	var queryText string
	queryText = `select tsense.*, 
   (select jsonb_agg(row_to_json(detail)) 
    from (select tlws.* from tlws where senseid=tsense.id order by word) as detail)
			as lwsjson from tsense	order by oword, theme, id offset :offset limit :limit`
	reply, err1 := sddb.NamedReadQuery(queryText, wsqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	fd = make([]*wordSearchQueryRecord, wsqp.Limit)
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
