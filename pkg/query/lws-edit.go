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

// параметры
type lwsEditParamsType struct {
	Lwsid      int64 // 0 при создании
	Sduserid   int64 // взято из сеанса, 0 при отсутствии входа в систему
	Senseid    int64
	Languageid int64
}

// данные для формы, полученные из БД
type lwsEditDataType struct {
	Word          string
	Commentary    template.HTML
	Languageslug  string
	OWord         string
	Theme         string
	Phrase        template.HTML
	OwnerId       sql.NullInt64 // обладатель смысла
	Ownernickname string        // обладатель смысла (прямого или подразумеваемого)
}

// SenseViewHTMLTemplateParamsType являются параметрами для senseview.t.html
type lwsNewEditHTMLTemplateParamsType struct {
	Lep    *lwsEditParamsType
	Led    *lwsEditDataType
	Phrase template.HTML
}

// читать смысл, см. представление vsense и senseViewParamsType для объяснения
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
	apperror.Panic500AndErrorIf(err1, "Не удалось извлечь данные, извините")
	defer sddb.CloseRows(reply)()
	lned = &lwsEditDataType{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(lned)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении данных: %#v", err1)
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Данные не найдены")
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
