package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// параметры
type profileParamsType struct {
	Sduserid int32
}

// Данные для формы, полученной из БД
type profileDataType struct {
	Nickname           string
	Email              string
	LanguageId         *int64
	Languageslug       string
	Languagecommentary string
}

// profileHTMLTemplateParamsType являются параметрами для profile.t.html
type profileHTMLTemplateParamsType struct {
	Pdt *profileDataType
}

func readProfileDataFromDb(pp *profileParamsType) (pd *profileDataType) {
	reply, err1 := sddb.NamedReadQuery(
		`
select
    sduser.nickname                   as nickname,
    sduser.registrationemail          as email,
	tlanguage.id 					  as languageid,
    coalesce(tlanguage.slug,'')       as languageslug,
    coalesce(tlanguage.commentary,'') as languagecommentary
from sduser
         left join sduser_profile as profile on sduser.id = profile.id
         left join tlanguage on tlanguage.id = profile.favorite_tlanguageid
where sduser.id = :sduserid;
`, &pp)
	apperror.Panic500AndErrorIf(err1, "Не удалось извлечь данные, извините")
	defer sddb.CloseRows(reply)()
	pd = &profileDataType{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(pd)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении данных: %#v", err1)
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Данные не найдены")
	}
	return
}

func ProfilePageHandler(c *gin.Context) {
	sduserID := user.GetSDUserIdOrZero(c)
	if sduserID > 0 {
		pd := readProfileDataFromDb(&profileParamsType{Sduserid: sduserID})
		c.HTML(http.StatusOK, "profile.t.html", profileHTMLTemplateParamsType{
			Pdt: pd,
		})
		return
	}
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Регистрация или логин."})
}
