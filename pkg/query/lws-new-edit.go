package query

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// parameters
type lwsNewEditParamsType struct {
	Sduserid   int64
	Senseid    int64
	Languageid int64
}

// data for the form obtained from the DB
type lwsNewEditDataType struct {
	Id            int64 // sense id, unnecessary
	Languageslug  string
	OWord         string
	Theme         string
	Phrase        string
	OwnerId       sql.NullInt64 // owner of a sense
	Ownernickname string        // owner of a sense (direct or implied)
}

// SenseViewHTMLTemplateParamsType are params for senseview.t.html
type lwsNewEditHTMLTemplateParamsType struct {
	Lnep   *lwsNewEditParamsType
	Lned   *lwsNewEditDataType
	Phrase template.HTML
}

// read the sense, see the vsense view and senseViewParamsType for the explanation
func readLwsNewEditDataFromDb(lnep *lwsNewEditParamsType) (lned *lwsNewEditDataType) {
	reply, err1 := sddb.NamedReadQuery(
		`select tsense.*, coalesce(sense_owner.nickname,cast('' as varchar(128))) as ownernickname,
			tlanguage.slug as languageslug
		 from tsense left join sduser as sense_owner on tsense.ownerid = sense_owner.id
			left join tlanguage on tlanguage.id = :languageid
			where tsense.id = :senseid`, &lnep)
	apperror.Panic500AndErrorIf(err1, "Failed to extract data, sorry")
	lned = &lwsNewEditDataType{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(lned)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data: %#v", err1)
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Data not found")
	}
	return
}

func LwsNewEditRequestHandler(c *gin.Context) {

	lnep := &lwsNewEditParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	lnep.Senseid = extractIdFromRequest(c, "senseid")
	lnep.Languageid = extractIdFromRequest(c, "languageid")

	lned := readLwsNewEditDataFromDb(lnep)

	phrase := template.HTML(lned.Phrase)

	c.HTML(http.StatusOK,
		"lws-new-edit.t.html",
		lwsNewEditHTMLTemplateParamsType{Lnep: lnep, Lned: lned, Phrase: phrase})
}
