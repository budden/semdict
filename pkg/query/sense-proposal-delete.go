package query

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type senseProposalDeleteParamsType struct {
	Sduserid   int64
	Proposalid int64
}

// SenseProposalAddFormPageHandler handles POST senseproposaladdform
func SenseProposalDeleteRequestHandler(c *gin.Context) {
	// FIXME handle empty drafts, like calling this page many times and never calling post.
	// Like have timeout for a draft, or a draft status, or even not add into the db until the
	// first save
	user.EnsureLoggedIn(c)
	svp := &senseProposalDeleteParamsType{
		Sduserid:   int64(user.GetSDUserIdOrZero(c)),
		Proposalid: extractIdFromRequest(c, "proposalid")}
	deleteProposalFromDb(svp)
	c.HTML(http.StatusOK,
		"general.t.html",
		shared.GeneralTemplateParams{Message: "Proposal deleted successfully"})
}

func deleteProposalFromDb(spdp *senseProposalDeleteParamsType) {
	reply, err1 := sddb.NamedUpdateQuery(
		`delete from tsense where id = :proposalid and ownerid = :sduserid returning id`, &spdp)
	apperror.Panic500AndErrorIf(err1, "Failed to delete a proposal, sorry")
	var dataFound bool
	for reply.Next() {
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Failed to delete a proposal (maybe it is not yours, not a proposal, or does not exist")
	}
	return
}
