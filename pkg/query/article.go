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
type senseViewDirHandlerParams struct {
	Id int32 // it is actually a number!
}

// FIXME shall we create a record for each query?
type senseDataForEditType struct {
	Senseid      int32
	Languageslug string
	Dialectslug  string
	Phrase       string
	Word         string
}

type senseEditTemplateParams struct {
	Ad *senseDataForEditType
}

// SenseViewDirHandler ...
func SenseViewDirHandler(c *gin.Context) {
	var avdhp senseViewDirHandlerParams

	avdhp.Id = extractIdFromRequest(c)
	dataFound, ad := readArticleFromDb(&avdhp)

	if dataFound {
		c.HTML(http.StatusOK,
			"senseview.html",
			shared.SenseViewParams{Id: ad.Senseid, Word: ad.Word, Phrase: template.HTML(ad.Phrase)})
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sorry, no sense (yet?) with id = «%d»", avdhp.Id)
	}
}

func readArticleFromDb(avdhp *senseViewDirHandlerParams) (dataFound bool, ad *senseDataForEditType) {
	reply, err1 := sddb.NamedReadQuery(
		`select 
			s.id as senseid
			,get_language_slug(l.id) as languageslug
			,phrase
			,word 
			from tsense as s
			inner join tlanguage as l on s.languageid = l.id
			where s.id = :id
			limit 1`, &avdhp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract an article, sorry")
	ad = &senseDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}

// SenseEditDirHandler is a handler to open edit page
func SenseEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	var avdhp senseViewDirHandlerParams

	avdhp.Id = extractIdFromRequest(c)
	dataFound, ad := readArticleFromDb(&avdhp)

	if !dataFound {
		c.HTML(http.StatusBadRequest,
			"general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no sense (yet?) for «%d»", avdhp.Id)})
		return
	}

	aetp := &senseEditTemplateParams{Ad: ad}
	c.HTML(http.StatusOK,
		"senseedit.html",
		aetp)
}
