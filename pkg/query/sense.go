package query

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// Params to show a sense.
type senseViewParamsType struct {
	Sduserid int64
	Senseid  int64
}

// senseDataForEditType is also used for a view.
type senseDataForEditType struct {
	Senseid        int64
	OWord          string
	Theme          string
	Phrase         string
	Sdusernickname sql.NullString // owner (direct or implied)
}

type senseEditTemplateParams struct {
	Ad *senseDataForEditType
}

// SenseViewHTMLTemplateParamsType are params for senseview.t.html
type SenseViewHTMLTemplateParamsType struct {
	Svp    *senseViewParamsType
	Sdfe   *senseDataForEditType
	Phrase template.HTML
}

func SenseByIdViewDirHandler(c *gin.Context) {
	svp := &senseViewParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}

	paramValue := extractIdFromRequest(c, "senseid")
	svp.Senseid = paramValue
	dataFound, senseDataForEdit := readSenseFromDb(svp)

	if dataFound {
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"senseview.t.html",
			SenseViewHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no sense (yet?) with id = «%d»", svp.Senseid)
	}
}

// read the sense, see the vsense view and senseViewParamsType for the explanation
func readSenseFromDb(svp *senseViewParamsType) (dataFound bool, ad *senseDataForEditType) {
	reply, err1 := sddb.NamedReadQuery(
		`select * from tsense where id = :senseid`, &svp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract a sense, sorry")
	ad = &senseDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}

//  SenseEditDirHandler serves /senseedit/:senseid
func SenseEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	Senseid := extractIdFromRequest(c, "commonid")
	svp := &senseViewParamsType{
		Sduserid: int64(user.GetSDUserIdOrZero(c)),
		Senseid:  Senseid}

	var dataFound bool
	var ad *senseDataForEditType
	dataFound, ad = readSenseFromDb(svp)

	if !dataFound {
		c.HTML(http.StatusBadRequest,
			"general.t.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no sense (yet?) for «%d»", svp.Senseid)})
		return
	}

	aetp := &senseEditTemplateParams{Ad: ad}
	c.HTML(http.StatusOK,
		"senseedit.t.html",
		aetp)
}
