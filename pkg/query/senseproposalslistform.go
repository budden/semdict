package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры, к-рые нужны для выполнения запроса
type senseAndProposalsListQueryParams struct {
	Sduserid int64 // 0 для незарег. польз.
	Commonid int64
}

type senseAndProposalsListQueryHeader struct {
	Commonid     int64
	Languageid   int64
	Languageslug string
}

type senseAndProposalsListQueryRecord struct {
	Commonid       int64
	Proposalid     int64
	Senseid        int64
	Phrase         string
	Word           string
	Deleted        bool
	OwnerId        int64
	Sdusernickname string
	Languageslug   string
	Whos           string
	Kindofchange   string
	Iscommon       bool
	Ismine         bool
}

// SenseAndProposalsListFormRouteHandler ...
func SenseAndProposalsListFormRouteHandler(c *gin.Context) {
	var svlqp *senseAndProposalsListQueryParams
	var records []*senseAndProposalsListQueryRecord
	svlqp, records = senseAndProposalsListInner(c)

	// Параметры шаблона
	type senseAndProposalsListFormTemplateParamsType struct {
		P              *senseAndProposalsListQueryParams
		Header         *senseAndProposalsListQueryHeader
		Records        []*senseAndProposalsListQueryRecord
		IsLoggedIn     bool
		LoggedInUserId int64
	}

	c.HTML(http.StatusOK,
		"senseproposalslistform.t.html",
		senseAndProposalsListFormTemplateParamsType{P: svlqp,
			Records:        records,
			IsLoggedIn:     user.IsLoggedIn(c),
			LoggedInUserId: int64(user.GetSDUserIdOrZero(c))})
}

func senseAndProposalsListInner(c *gin.Context) (svlqp *senseAndProposalsListQueryParams, records []*senseAndProposalsListQueryRecord) {
	svlqp = &senseAndProposalsListQueryParams{}

	svlqp.Commonid = extractIdFromRequest(c, "commonid")
	svlqp.Sduserid = int64(user.GetSDUserIdOrZero(c))

	records = readCommonSenseAndProposalsListQueryFromDb(svlqp)
	return
}

// reads both common sense and proposals
func readCommonSenseAndProposalsListQueryFromDb(svlqp *senseAndProposalsListQueryParams) (
	records []*senseAndProposalsListQueryRecord) {
	reply, err1 := sddb.NamedReadQuery("select fncommonsenseandproposals(:sduserid, :commonid)", svlqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	records = make([]*senseAndProposalsListQueryRecord, 0)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &senseAndProposalsListQueryRecord{}
		err1 = reply.StructScan(wsqr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining proposals of sense: %#v", err1)
		records = append(records, wsqr)
	}
	return
}
