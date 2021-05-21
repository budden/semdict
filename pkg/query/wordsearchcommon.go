package query

// Общая часть для /wordsearchquery и /wordsearchresultform

import (
	"database/sql"
	"encoding/json"
	"html/template"

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

// данные, общие для формы
type wordSearchMasterRecord struct {
	FavoriteLanguageId   int64
	FavoriteLanguageSlug string
}

// рекорд по найденным ощущениям
type wordSearchQueryRecord struct {
	Senseid int64
	// Sdusernickname sql.NullString
	Oword                          string
	Theme                          string
	Phrase                         string
	Ownerid                        int64
	Lwsjson                        sql.NullString
	LwsArray                       []TlwsRecordForWordSearch
	HasFavoriteLanguageTranslation int64
	// Proposalid       sql.NullInt64 // не является нулевым, если эта запись является предложением.

}

func wordSearchCommonPart(c *gin.Context) (wsqp *wordSearchQueryParams,
	wsmr *wordSearchMasterRecord,
	fd []*wordSearchQueryRecord) {
	wsqp = getWordSearchQueryParamsFromRequest(c)

	if wsqp.Wordpattern == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Пустой шаблон поиска")
	}

	wsqp.Offset = int32(GetZeroOrOneNonNegativeIntFormValue(c, "offset"))
	wsqp.Limit = int32(GetZeroOrOneNonNegativeIntFormValue(c, "limit"))
	LimitLimit(&wsqp.Limit)
	wsqp.Sduserid = user.GetSDUserIdOrZero(c)

	wsmr, fd = readWordSearchQueryFromDb(wsqp)
	return
}

type TlwsRecordForWordSearch = struct {
	Id           int64
	Word         string
	Commentary   template.HTML
	OwnerId      int64
	SenseId      int64
	LanguageId   int64
	Languageslug string
	Canedit      int
}

// select * from tsense where to_tsvector(phrase)||to_tsvector(word) @@ 'go';
// select row_to_json(x) from (select * from tsense) x;
// https://eax.me/postgresql-full-text-search/

func readWordSearchQueryFromDb(wsqp *wordSearchQueryParams) (
	wsmr *wordSearchMasterRecord, fd []*wordSearchQueryRecord) {
	wsmr = readWordSearchMasterRecordFromDb(wsqp)
	fd = readWordSearchSensesFromDb(wsqp)
	return
}

func readWordSearchMasterRecordFromDb(wsqp *wordSearchQueryParams) (
	wmsr *wordSearchMasterRecord) {
	queryText := "select * from fnwordsearchmasterrecord(:sduserid)"
	reply, err1 := sddb.NamedReadQuery(queryText, wsqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)()
	wmsr = &wordSearchMasterRecord{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(wmsr)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Данные не найдены")
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении данных из профиля пользователя: %#v", err1)
	return
}

func readWordSearchSensesFromDb(wsqp *wordSearchQueryParams) (fd []*wordSearchQueryRecord) {
	queryText := "select * from fnwordsearch(:sduserid,:wordpattern,:offset,:limit)"
	reply, err1 := sddb.NamedReadQuery(queryText, wsqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)()
	fd = make([]*wordSearchQueryRecord, wsqp.Limit)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &wordSearchQueryRecord{}
		err1 = reply.StructScan(wsqr)
		tlws := make([]TlwsRecordForWordSearch, 0)
		if wsqr.Lwsjson.Valid {
			json.Unmarshal([]byte(wsqr.Lwsjson.String), &tlws)
		}
		wsqr.LwsArray = tlws
		sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении данных о смысле: %#v", err1)
		fd[last] = wsqr
	}
	fd = fd[:last]
	return
}
