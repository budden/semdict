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
	Sduserid int64
	Senseid  int64
}

// SenseDeleteRequestHandler = POST sensedelete
func SenseDeleteRequestHandler(c *gin.Context) {
	// FIXME handle empty drafts, like calling this page many times and never calling post.
	// Like have timeout for a draft, or a draft status, or even not add into the db until the
	// first save
	user.EnsureLoggedIn(c)
	svp := &senseProposalDeleteParamsType{
		Sduserid: int64(user.GetSDUserIdOrZero(c)),
		Senseid:  extractIdFromRequest(c, "senseid")}
	deleteSenseFromDb(svp)
	c.HTML(http.StatusOK,
		"general.t.html",
		shared.GeneralTemplateParams{Message: "Proposal deleted successfully"})
}

func deleteSenseFromDb(spdp *senseProposalDeleteParamsType) {
	reply, err1 := sddb.NamedUpdateQuery(
		`delete from tsense where id = :senseid returning id`, &spdp)
	apperror.Panic500AndErrorIf(err1, "Failed to delete a sense, sorry")
	var dataFound bool
	for reply.Next() {
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Failed to delete a sense (maybe it is not yours, or does not exist)")
	}
	return
}
