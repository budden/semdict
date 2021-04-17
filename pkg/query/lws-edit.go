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
type lwsEditParamsType struct {
	Lwsid      int64 // 0 when creating
	Sduserid   int64 // taken from sesion, 0 when not logged in
	Senseid    int64
	Languageid int64
}

// data for the form obtained from the DB
type lwsEditDataType struct {
	Word          string
	Commentary    template.HTML
	Languageslug  string
	OWord         string
	Theme         string
	Phrase        template.HTML
	OwnerId       sql.NullInt64 // owner of a sense
	Ownernickname string        // owner of a sense (direct or implied)
}

// SenseViewHTMLTemplateParamsType are params for senseview.t.html
type lwsNewEditHTMLTemplateParamsType struct {
	Lep    *lwsEditParamsType
	Led    *lwsEditDataType
	Phrase template.HTML
}

// read the sense, see the vsense view and senseViewParamsType for the explanation
func readLwsNewEditDataFromDb(lnep *lwsEditParamsType) (lned *lwsEditDataType) {
	reply, err1 := sddb.NamedReadQuery(
		`select 
 		coalesce(tlws.word,'') as word,
	 	coalesce(tlws.commentary,'') as commentary,
			tsense.oword, tsense.phrase, tsense.phrase, coalesce(sense_owner.nickname,cast('' as varchar(128))) as ownernickname,
			coalesce(tlanguage.slug,'') as languageslug
		 from tsense 
			left join tlws on tlws.id = :lwsid and tlws.senseid = :senseid
			left join sduser as sense_owner on tsense.ownerid = sense_owner.id
			left join tlanguage on tlanguage.id = :languageid
			where tsense.id = :senseid`, &lnep)
	apperror.Panic500AndErrorIf(err1, "Failed to extract data, sorry")
	defer sddb.CloseRows(reply)()
	lned = &lwsEditDataType{}
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

	lnep := &lwsEditParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	lnep.Senseid = extractIdFromRequest(c, "senseid")
	lnep.Languageid = extractIdFromRequest(c, "languageid")

	lned := readLwsNewEditDataFromDb(lnep)

	phrase := template.HTML(lned.Phrase)

	c.HTML(http.StatusOK,
		"lws-new-edit.t.html",
		lwsNewEditHTMLTemplateParamsType{Lep: lnep, Led: lned, Phrase: phrase})
}

func LwsEditGetHandler(c *gin.Context) {

	lnep := &lwsEditParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	lnep.Senseid = extractIdFromRequest(c, "senseid")
	lnep.Languageid = extractIdFromRequest(c, "languageid")
	lnep.Lwsid = extractIdFromRequest(c, "lwsid")

	lned := readLwsNewEditDataFromDb(lnep)

	c.HTML(http.StatusOK,
		"lws-edit.t.html",
		lwsNewEditHTMLTemplateParamsType{Lep: lnep, Led: lned})
}
