package query

import (
	"net/http"
	"strings"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type senseAddParamsType struct {
	Sduserid int64
	Word     string
}

// SenseProposalAddFormPageHandler handles POST senseproposaldelete
func SenseProposalAddFormPageHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	svp := &senseAddParamsType{
		Sduserid: int64(user.GetSDUserIdOrZero(c)),
		Word:     convertWordpatternToNewWork(c.PostForm("wordpattern"))}
	ProposalID := makeNewSenseidInDb(svp)
	ad := &senseDataForEditType{}
	ad.Proposalid = ProposalID
	ad.Word = svp.Word
	// FIXME set language and edit it
	aetp := &senseEditTemplateParams{Ad: ad}
	c.HTML(http.StatusOK,
		"senseedit.t.html",
		aetp)
}

func convertWordpatternToNewWork(pattern string) string {
	return strings.Replace(pattern, "%", "", -1)
}

func makeNewSenseidInDb(sap *senseAddParamsType) (id int64) {
	reply, err1 := sddb.NamedUpdateQuery(
		`insert into tsense (ownerid, word, languageid, phrase) 
			values (:sduserid, :word, 1/*language engligh*/, '') 
			returning id`, &sap)
	apperror.Panic500AndErrorIf(err1, "Failed to insert an article, sorry")
	var dataFound bool
	for reply.Next() {
		err1 = reply.Scan(&id)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Insert didn't return a record")
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining id of a fresh sense: %#v", err1)
	return
}
