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
	Showdeleted bool
	// Эти поля не вводятся пользователем
	Sduserid int32 // 0 для незарег. польз.
	Offset   int32
	Limit    int32 // 0 - значит «без ограничения»
}

type wordSearchQueryRecord struct {
	Commonid       int64
	Proposalid     int64
	Senseid        int64
	Proposalstatus string
	Languageid     int64
	Languageslug   string
	Sdusernickname sql.NullString
	Phantom        bool
	Phrase         string
	Word           string
	// Proposalid       sql.NullInt64 // is non-null when this record is a proposal.
	Countofproposals int32
	Commonorproposal string
	Whos             string
	Kindofchange     string
	Deletionproposed bool
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

func readWordSearchQueryFromDb(wsqp *wordSearchQueryParams) (
	fd []*wordSearchQueryRecord) {
	var queryText string
	queryText = `select ts.commonid,	ts.proposalid, ts.senseid
	 ,ts.proposalstatus
		,ts.languageid, ts.languageslug, ts.word, ts.phrase
		,ps.r_countofproposals as countofproposals
		,ts.sdusernickname
		,ts.phantom
		,ts.deletionproposed
		,(explainSenseEssenseVsProposals(:sduserid, ts.commonid, ts.proposalid, ts.ownerid, ts.phantom, ts.deletionproposed)).*
		from fnpersonalsenses(:sduserid,:showdeleted) ps 
		left join vsense_wide ts on coalesce(nullif(ps.r_proposalid,0), ps.r_commonid) = ts.id
		order by word, languageslug, senseid offset :offset limit :limit`
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
