package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// Params for SenseProposalAcceptOrRejectDirHandler
type senseProposalAcceptOrRejectParamsType struct {
	Sduserid   int64
	Proposalid int64
}

// SenseViewHTMLTemplateParamsType are params for senseview.t.html

type senseProposalAcceptOrRejectTemplateParamsInnerType struct {
	Proposalid       int64
	Commonid         int64
	Phantom          bool // FIXME fill it!
	Deletionproposed bool
	Phraseold        string
	Phrasechanged    bool
	Phrasenew        string
	Wordold          string
	Wordchanged      bool
	Wordnew          string
	// FIXME that is insufficient!
}

type senseProposalAcceptOrRejectTemplateParamsType struct {
	Data *senseProposalAcceptOrRejectTemplateParamsInnerType
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
	spaorhtpi := senseProposalAcceptOrRejectCalculateTemplateParams(spaorp, records)
	spaorhtp := &senseProposalAcceptOrRejectTemplateParamsType{Data: spaorhtpi}
	c.HTML(http.StatusOK,
		"senseproposalacceptorrejectform.t.html",
		spaorhtp)
	/*
		phraseHTML := template.HTML(senseDataForEdit.Phrase)
		c.HTML(http.StatusOK,
			"senseview.t.html",
			SenseViewHTMLTemplateParamsType{Svp: svp, Sdfe: senseDataForEdit, Phrase: phraseHTML})
	*/
}

func senseProposalAcceptOrRejectCalculateTemplateParams(spaorp *senseProposalAcceptOrRejectParamsType,
	records []*senseAndProposalsListQueryRecord) (spaorhtpi *senseProposalAcceptOrRejectTemplateParamsInnerType) {
	spaorhtpi = &senseProposalAcceptOrRejectTemplateParamsInnerType{Proposalid: spaorp.Proposalid}
	n := len(records)
	if n == 0 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No proposal with id = %d", spaorp.Proposalid)
	}
	p := records[0]
	var o *senseAndProposalsListQueryRecord
	if n == 2 {
		o = records[1]
		spaorhtpi.Commonid = o.Commonid
		spaorhtpi.Proposalid = p.Proposalid
		spaorhtpi.Deletionproposed = p.Deletionproposed
		spaorhtpi.Phrasenew = p.Phrase
		spaorhtpi.Phraseold = o.Phrase
		spaorhtpi.Phrasechanged = p.Phrase != o.Phrase
		spaorhtpi.Wordnew = p.Word
		spaorhtpi.Wordold = o.Word
		spaorhtpi.Wordchanged = p.Word != o.Word
		// Теперь надо в вызывающей ф-ии заполнить шаблон, который нужно сделать.
	} else if n == 1 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Proposal with no original sense (is it an addition?) - unable to handle (yet)")
	}
	//if p.Kindofchange
	return
}
