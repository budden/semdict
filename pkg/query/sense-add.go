package query

import (
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

// SenseAddPageHandler handles POST senseadd
func SenseAddPageHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
}

func convertWordpatternToNewWork(pattern string) string {
	return strings.Replace(pattern, "%", "", -1)
}

func makeNewSenseidInDb(sap *senseAddParamsType) (id int64) {
	reply, err1 := sddb.NamedUpdateQuery(
		`insert into tsense (ownerid, word, proposalstatus, languageid, phrase) 
			values (:sduserid, :word, 'draft', 1/*language engligh*/, '') 
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
