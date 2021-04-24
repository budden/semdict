package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// parameters
type profileEditParamsType struct {
	Sduserid int32
}

// data for the form obtained from the DB
type profileEditDataType struct {
	ID         int64
	Slug       string
	Commentary string
	Ownerid    *int64
}

// profileHTMLTemplateParamsType are params for profile.t.html
type profileEditHTMLTemplateParamsType struct {
	D  []*profileEditDataType
	PD *profileDataType
}

func readProfileEditDataFromDb(p *profileEditParamsType) (d []*profileEditDataType) {

	reply, err1 := sddb.NamedReadQuery(
		`
select id,
       slug,
       commentary,
       ownerid
from tlanguage;
`, &p)
	apperror.Panic500AndErrorIf(err1, "Failed to extract data, sorry")
	defer sddb.CloseRows(reply)()
	for reply.Next() {
		r := &profileEditDataType{}
		err1 = reply.StructScan(r)
		sddb.FatalDatabaseErrorIf(err1, "Error obtaining a theme: %#v", err1)
		d = append(d, r)
	}
	return
}

func ProfileEditPageHandler(c *gin.Context) {
	sduserID := user.GetSDUserIdOrZero(c)
	if sduserID > 0 {
		pd := readProfileDataFromDb(&profileParamsType{Sduserid: sduserID})
		d := readProfileEditDataFromDb(&profileEditParamsType{Sduserid: sduserID})
		c.HTML(http.StatusOK, "profile-edit.t.html", profileEditHTMLTemplateParamsType{
			D:  d,
			PD: pd,
		})
		return
	}
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Register or login."})
}
