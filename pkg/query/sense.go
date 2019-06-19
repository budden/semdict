package query

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
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

// Params for SenseProposalAcceptOrRejectDirHandler
type senseProposalAcceptOrRejectParamsType struct {
	Sduserid   int64
	Proposalid int64
}

// SenseViewHTMLTemplateParamsType are params for senseview.t.html

type senseProposalAcceptOrRejectHTMLTemplateParamsType struct {
	Proposalid       int64
	Commonid         int64
	Deletionproposed bool
	Phraseold        string
	Phrasechanged    bool
	Phrasenew        string
	Wordold          string
	Wordchanged      bool
	Wordnew          string
	// FIXME that is insufficient!
}

// senseDataForEditType is also used for a view.
type senseDataForEditType struct {
	Commonid         int64
	Proposalid       int64
	Senseid          int64
	Proposalstatus   string
	Phrase           string
	Word             string
	Phantom          bool
	Deletionproposed bool
	Sdusernickname   sql.NullString
	Languageslug     string
	Commonorproposal string
	Whos             string
	Kindofchange     string
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

// SenseByCommonidViewDirHandler ...
func SenseByCommonidViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, "commonid")
}

// SenseByIdViewDirHandler ...
func SenseByIdViewDirHandler(c *gin.Context) {
	senseOrProposalDirHandlerCommon(c, "senseid")
}

// SenseProposalAcceptOrRejectDirHandler ...
func SenseProposalAcceptOrRejectDirHandler(c *gin.Context) {
	spaorp := &senseProposalAcceptOrRejectParamsType{Sduserid: int64(user.GetSDUserIdOrZero(c))}
	spaorp.Proposalid = extractIdFromRequest(c, "proposalid")
	var records []*senseAndProposalsListQueryRecord
	records = readSenseProposalAcceptOrRejectDataFromDb(spaorp)
	if len(records) == 0 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no proposal (yet?) with id = «%d»", spaorp.Proposalid)
	}
	spaorhtp := senseProposalAcceptOrRejectCalculateTemplateParams(spaorp, records)
	c.HTML(http.StatusOK,
		"general.t.html",
		shared.GeneralTemplateParams{Message: "So far so good. TODO Now convert those records to senseProposalAcceptOrRejectHTMLTemplateParamsType and use diff.js to visualize"})
	/*
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"senseview.t.html",
			SenseViewHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	*/
}

func senseProposalAcceptOrRejectCalculateTemplateParams(spaorp *senseProposalAcceptOrRejectParamsType,
	records []*senseAndProposalsListQueryRecord) (spaorhtp *senseProposalAcceptOrRejectHTMLTemplateParamsType) {
	spaorhtp = &senseProposalAcceptOrRejectHTMLTemplateParamsType{Proposalid: spaorp.Proposalid}
	n := len(records)
	if n == 0 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No proposal with id = %d", spaorp.Proposalid)
	}
	p := records[0]
	var o *senseAndProposalsListQueryRecord
	if n == 2 {
		o = records[1]
		spaorhtp.Commonid = o.Commonid
		spaorhtp.Proposalid = p.Proposalid
		spaorhtp.Deletionproposed = p.Deletionproposed
		spaorhtp.Phrasenew = p.Phrase
		spaorhtp.Phraseold = o.Phrase
		spaorhtp.Phrasechanged = p.Phrase != o.Phrase
		spaorhtp.Wordnew = p.Word
		spaorhtp.Wordold = o.Word
		spaorhtp.Wordchanged = p.Word != o.Word
		log.Printf("We got %#v", spaorhtp)
		// Теперь надо в вызывающей ф-ии заполнить шаблон, который нужно сделать.
	} else if n == 1 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Proposal with no original sense (is it an addition?) - unable to handle (yet)")
	}
	//if p.Kindofchange
	return
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
			SenseViewHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
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
	Commonid := extractIdFromRequest(c, "commonid")
	if Proposalid != 0 {
		Commonid = 0
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
