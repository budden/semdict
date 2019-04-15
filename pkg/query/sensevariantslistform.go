package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры, к-рые нужны для выполнения запроса
type senseVariantsListQueryParams struct {
	Sduserid int64 // 0 для незарег. польз.
	Senseid  int64
}

type senseVariantsListQueryHeader struct {
	Senseid      int64
	Languageid   int64
	Languageslug string
}

type senseVariantsListQueryRecord struct {
	VariantId   int64
	Phrase      string
	Word        string
	OwnerId     int64
	MySense     bool
	CommonSense bool
}

// SenseVariantsListFormRouteHandler ...
func SenseVariantsListFormRouteHandler(c *gin.Context) {
	var svlqp *senseVariantsListQueryParams
	var records []*senseVariantsListQueryRecord
	svlqp, records = senseVariantsListInner(c)

	// Параметры шаблона
	type senseVariantsListFormTemplateParamsType struct {
		P              *senseVariantsListQueryParams
		Header         *senseVariantsListQueryHeader
		Records        []*senseVariantsListQueryRecord
		IsLoggedIn     bool
		LoggedInUserId int64
	}

	c.HTML(http.StatusOK,
		"sensevariantslistform.t.html",
		senseVariantsListFormTemplateParamsType{P: svlqp,
			Records:        records,
			IsLoggedIn:     user.IsLoggedIn(c),
			LoggedInUserId: int64(user.GetSDUserIdOrZero(c))})
}

func senseVariantsListInner(c *gin.Context) (svlqp *senseVariantsListQueryParams, records []*senseVariantsListQueryRecord) {
	svlqp = &senseVariantsListQueryParams{}

	svlqp.Senseid = extractIdFromRequest(c)
	svlqp.Sduserid = int64(user.GetSDUserIdOrZero(c))

	records = readSenseVariantsListQueryFromDb(svlqp)
	return
}

func readSenseVariantsListQueryFromDb(svlqp *senseVariantsListQueryParams) (
	records []*senseVariantsListQueryRecord) {
	reply, err1 := sddb.NamedReadQuery(`select vari.id as variantid
	,vari.phrase, vari.word, vari.ownerid
	,false as commonsense 
	,case when ownerid = :sduserid then true else false end as mysense
	from tsense vari where originid = :senseid
	union all select s.id as variantid
	,s.phrase, s.word, cast(0 as bigint) as ownerid
	,true as commonsense
	,false as mysense
	from tsense s where id = :senseid
	order by commonsense desc, mysense desc
	`, svlqp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)
	records = make([]*senseVariantsListQueryRecord, 0)
	var last int
	for last = 0; reply.Next(); last++ {
		wsqr := &senseVariantsListQueryRecord{}
		err1 = reply.StructScan(wsqr)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining variants of sense: %#v", err1)
		records = append(records, wsqr)
	}
	return
}
