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

// Params to show a sense. First non-zero sense of three sense ids is used.
type senseViewParamsType struct {
	Sduserid   int64
	Commonid   int64 // We want to see proposal for this common id if Sduserid has one, otherwise common sense.
	Proposalid int64 // We want to see this record which must be a proposal
	Senseid    int64 // We want to see sense by id, regardless of it is a common sense or a proposal
}

//fnsenseorproposalforview(p_sduserid bigint, p_id bigint, p_proposalifexists bool)
//returns table (commonid bigint, senseorproposalid bigint, phrase text, word varchar(512), deleted bool, languageslug text)

// senseDataForEditType is also used for a view.
type senseDataForEditType struct {
	Commonid         int64
	Proposalid       int64
	Senseid          int64
	Phrase           string
	Word             string
	Deleted          bool
	Languageslug     string
	Commonorproposal string
	Whos             string
	Kindofchange     string
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

// SenseByCommonidViewDirHandler ...
func SenseByCommonidViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, "commonid")
}

// SenseByIdViewDirHandler ...
func SenseByIdViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, "senseid")
}

func senseOrProposalDirHandlerCommon(c *gin.Context, paramName string) {
	svp := &senseViewParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}

	paramValue := extractIdFromRequest(c, paramName)
	if paramName == "commonid" {
		svp.Commonid = paramValue
	} else if paramName == "proposalid" {
		svp.Proposalid = paramValue
	} else if paramName == "senseid" {
		svp.Senseid = paramValue
	} else {
		apperror.GracefullyExitAppIf(apperror.ErrDummy, "unknown paramName")
	}
	dataFound, senseDataForEdit := readSenseFromDb(svp)

	if dataFound {
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"senseview.t.html",
			SenseViewParams{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no sense (yet?) with id = «%d»", svp.Senseid)
	}
}

// read the sense, see the vsense view and senseViewParamsType for the explanation
func readSenseFromDb(svp *senseViewParamsType) (dataFound bool, ad *senseDataForEditType) {
	reply, err1 := sddb.NamedReadQuery(
		`select * from fnsenseorproposalforview(:sduserid, :commonid, :proposalid, :senseid)`, &svp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract a sense, sorry")
	ad = &senseDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}

//  SenseEditDirHandler serves /senseedit/:commonid/:proposalid
func SenseEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	Proposalid := extractIdFromRequest(c, "proposalid")
	var Commonid int64
	if Proposalid != 0 {
		Commonid = 0
	} else {
		Commonid = extractIdFromRequest(c, "proposalid")
	}
	svp := &senseViewParamsType{
		Sduserid:   int64(user.GetSDUserIdOrZero(c)),
		Commonid:   Commonid,
		Proposalid: Proposalid}

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
