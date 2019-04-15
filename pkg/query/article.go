package query

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// params for a query for a word
type senseViewParamsType struct {
	Id       int64
	Sduserid int64
	/// if Proposalifexists is true, we try to show current user's proposal, and if theres no one, a common sense
	/// if Proposalifexists is false, we show just exact Id given, regardless of whether it is a common sense or proposal
	Proposalifexists bool
}

//fnsenseorproposalforview(p_sduserid bigint, p_id bigint, p_proposalifexists bool)
//returns table (originid bigint, senseorproposalid bigint, phrase text, word varchar(512), deleted bool, languageslug text)

// senseDataForEditType is also used for a view.
type senseDataForEditType struct {
	Senseorproposalid int64 // it is just the id of the record we see
	Originid          int64 // it is an origin id, 0 for additions and common sense
	Phrase            string
	Word              string
	Deleted           bool
	Languageslug      string
	Commonorproposal  string
	Whos              string
	Kindofchange      string
}

type senseEditTemplateParams struct {
	Ad *senseDataForEditType
}

// SenseViewParams are params for senseview.t.html
type SenseViewParams struct {
	Svp    *senseViewParamsType
	Sdfe   *senseDataForEditType
	Phrase template.HTML
}

// SenseByOriginIdViewDirHandler ...
func SenseByOriginIdViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, true)
}

// SenseByIdViewDirHandler ...
func SenseByIdViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, false)
}

func senseOrProposalDirHandlerCommon(c *gin.Context, proposalIfExists bool) {
	svp := &senseViewParamsType{
		Id:               extractIdFromRequest(c),
		Sduserid:         int64(user.GetSDUserIdOrZero(c)),
		Proposalifexists: proposalIfExists}
	dataFound, senseDataForEdit := readSenseFromDb(svp)

	if dataFound {
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"senseview.t.html",
			SenseViewParams{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no sense (yet?) with id = «%d»", svp.Id)
	}
}

// read the sense appropriate for edit. That is, either mine or a common one.
func readSenseFromDb(svp *senseViewParamsType) (dataFound bool, ad *senseDataForEditType) {
	reply, err1 := sddb.NamedReadQuery(
		`select * from fnsenseorproposalforview(:sduserid, :id, :proposalifexists)`, &svp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract an article, sorry")
	ad = &senseDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}

// SenseByOriginIdEditDirHandler is a handler to open a user's proposal, or an original record if there
// is no user's proposal
func SenseByOriginIdEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	svp := &senseViewParamsType{
		Id:               extractIdFromRequest(c),
		Sduserid:         int64(user.GetSDUserIdOrZero(c)),
		Proposalifexists: true}

	dataFound, ad := readSenseFromDb(svp)

	if !dataFound {
		c.HTML(http.StatusBadRequest,
			"general.t.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no sense (yet?) for «%d»", svp.Id)})
		return
	}

	aetp := &senseEditTemplateParams{Ad: ad}
	c.HTML(http.StatusOK,
		"sensebyoriginidedit.t.html",
		aetp)
}
