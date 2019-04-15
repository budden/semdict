package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры, к-рые нужны для выполнения запроса
type senseProposalsListQueryParams struct {
	Sduserid int64 // 0 для незарег. польз.
	Senseid  int64
}

type senseProposalsListQueryHeader struct {
	Senseid      int64
	Languageid   int64
	Languageslug string
}

type senseProposalsListQueryRecord struct {
	ProposalId  int64
	Phrase      string
	Word        string
	OwnerId     int64
	MySense     bool
	CommonSense bool
}

// SenseProposalsListFormRouteHandler ...
func SenseProposalsListFormRouteHandler(c *gin.Context) {
	var svlqp *senseProposalsListQueryParams
	var records []*senseProposalsListQueryRecord
	svlqp, records = senseProposalsListInner(c)

	// Параметры шаблона
	type senseProposalsListFormTemplateParamsType struct {
		P              *senseProposalsListQueryParams
		Header         *senseProposalsListQueryHeader
		Records        []*senseProposalsListQueryRecord
		IsLoggedIn     bool
		LoggedInUserId int64
	}

	c.HTML(http.StatusOK,
		"senseproposalslistform.t.html",
		senseProposalsListFormTemplateParamsType{P: svlqp,
			Records:        records,
			IsLoggedIn:     user.IsLoggedIn(c),
			LoggedInUserId: int64(user.GetSDUserIdOrZero(c))})
}

func senseProposalsListInner(c *gin.Context) (svlqp *senseProposalsListQueryParams, records []*senseProposalsListQueryRecord) {
	svlqp = &senseProposalsListQueryParams{}

	svlqp.Senseid = extractIdFromRequest(c)
	svlqp.Sduserid = int64(user.GetSDUserIdOrZero(c))

	records = readSenseProposalsListQueryFromDb(svlqp)
	return
}

func readSenseProposalsListQueryFromDb(svlqp *senseProposalsListQueryParams) (
	records []*senseProposalsListQueryRecord) {
	reply, err1 := sddb.NamedReadQuery(`select vari.id as proposalid
	,vari.phrase, vari.word, vari.ownerid
	,false as commonsense 
	,case when ownerid = :sduserid then true else false end as mysense
	from tsense vari where originid = :senseid
	union all select s.id as proposalid
	,s.phrase, s.word, cast(0 as bigint) as ownerid
	,true as commonsense
	,false as mysense
	from tsense s where id = :senseid
	order by commonsense desc, mysense desc
	`, svlqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	records = make([]*senseProposalsListQueryRecord, 0)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &senseProposalsListQueryRecord{}
		err1 = reply.StructScan(wsqr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining proposals of sense: %#v", err1)
		records = append(records, wsqr)
	}
	return
}
