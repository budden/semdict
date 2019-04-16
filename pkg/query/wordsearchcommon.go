package query

// Общая часть для /wordsearchquery и /wordsearchresultform

import (
	"database/sql"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры из URL
type wordSearchQueryParams struct {
	Dummyid     int32 // не имеет значения
	Wordpattern string
	Sduserid    int32 // 0 для незарег. польз.
	Offset      int32
	Limit       int32 // 0 - значит «без ограничения»
}

type wordSearchQueryRecord struct {
	Personalid   sql.NullInt64 // Id of the data visible to some user only
	Commonid     sql.NullInt64 // Id of the corresponding common data
	Languageid   int32
	Languageslug string
	Phrase       string
	Word         string
	// Proposalid       sql.NullInt64 // is non-null when this record is a proposal.
	Countofproposals int32
	Commonorproposal string
	Whos             string
	Kindofchange     string
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
	frp.Sduserid = user.GetSDUserIdOrZero(c)

	fd = readWordSearchQueryFromDb(frp)
	return
}

func readWordSearchQueryFromDb(frp *wordSearchQueryParams) (
	fd []*wordSearchQueryRecord) {
	var queryText string
	queryText = `select ps.r_originid as id, ts.languageid, ts.languageslug, ts.word, ts.phrase, 
		ps.r_proposalid as proposalid,
		ps.r_countofproposals as countofproposals,
		(explainSenseStatusVsProposals(ts.id, ts.originid, :sduserid, ts.ownerid, ts.deleted)).*
		from fnpersonalsenses(:sduserid) ps 
		left join vsense ts on coalesce(ps.r_proposalid, ps.r_originid) = ts.id
		order by word, languageslug, id offset :offset limit :limit`
	reply, err1 := sddb.NamedReadQuery(
		queryText, frp)
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
