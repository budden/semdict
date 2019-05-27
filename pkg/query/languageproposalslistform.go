package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры, к-рые нужны для выполнения запроса
type languageProposalsListQueryParams struct {
	Sduserid int64 // 0 для незарег. польз.
	Commonid int64
}

type languageProposalsListQueryHeader struct {
	Commonid     int64
	Languageid   int64
	Languageslug string
}

type languageProposalsListQueryRecord struct {
	Commonid         int64
	Proposalid       int64
	Senseid          int64
	Proposalstatus   string
	Phrase           string
	Word             string
	Deleted          bool
	OwnerId          int64
	Sdusernickname   string
	Languageslug     string
	Commonorproposal string
	Whos             string
	Kindofchange     string
	Iscommon         bool
	Ismine           bool
}

// Параметры шаблона
type languageProposalsListFormTemplateParamsType struct {
	P              *languageProposalsListQueryParams
	Header         *languageProposalsListQueryHeader
	Records        []*languageProposalsListQueryRecord
	IsLoggedIn     bool
	LoggedInUserId int64
}

// LanguageProposalsListFormRouteHandler ...
func LanguageProposalsListFormRouteHandler(c *gin.Context) {
	sduserid := int64(user.GetSDUserIdOrZero(c))
	commonid := extractIdFromRequest(c, "commonid")
	svlqp := &languageProposalsListQueryParams{Sduserid: sduserid, Commonid: commonid}

	var records []*languageProposalsListQueryRecord
	records = readLanguageProposalsListQueryFromDb(svlqp)

	svp := &senseViewParamsType{Sduserid: sduserid, Senseid: commonid}
	dataFound, header1 := readSenseFromDb(svp)
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No common sense found - unable to get proposals")
	}
	header := &languageProposalsListQueryHeader{Commonid: commonid, Languageslug: header1.Languageslug}

	c.HTML(http.StatusOK,
		"senseproposalslistform.t.html",
		languageProposalsListFormTemplateParamsType{P: svlqp,
			Header:         header,
			Records:        records,
			IsLoggedIn:     user.IsLoggedIn(c),
			LoggedInUserId: int64(user.GetSDUserIdOrZero(c))})
}

// reads both common sense and proposals
func readLanguageProposalsListQueryFromDb(svlqp *languageProposalsListQueryParams) (
	records []*languageProposalsListQueryRecord) {
	reply, err1 := sddb.NamedReadQuery("select * from fnlanguageproposals(:sduserid, :commonid)", svlqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	records = make([]*languageProposalsListQueryRecord, 0)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &languageProposalsListQueryRecord{}
		err1 = reply.StructScan(wsqr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining proposals of sense: %#v", err1)
		records = append(records, wsqr)
	}
	return
}
