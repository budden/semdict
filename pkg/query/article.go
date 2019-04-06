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
type articleViewDirHandlerParams struct {
	Word string
}

// FIXME shall we create a record for each query?
type articleDataForEditType struct {
	Senseid      int32
	Languageslug string
	Dialectslug  string
	Phrase       string
	Word         string
}

type articleEditTemplateParams struct {
	Ad *articleDataForEditType
}

// ArticleViewDirHandler ...
func ArticleViewDirHandler(c *gin.Context) {
	var avdhp articleViewDirHandlerParams
	avdhp.Word = c.Param("word")

	if avdhp.Word == "" {
		c.HTML(http.StatusNotFound, "", nil)
		return
	}

	dataFound, ad := readArticleFromDb(&avdhp)

	if dataFound {
		c.HTML(http.StatusOK,
			"articleview.html",
			shared.ArticleViewParams{Word: ad.Word, Phrase: template.HTML(ad.Phrase)})
	} else {
		c.HTML(http.StatusBadRequest,
			"general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no article (yet?) for «%s»", avdhp.Word)})
	}
}

func readArticleFromDb(avdhp *articleViewDirHandlerParams) (dataFound bool, ad *articleDataForEditType) {
	reply, err1 := sddb.NamedReadQuery(
		`select 
			s.id as senseid
			,get_language_slug(l.id) as languageslug
			,phrase
			,word 
			from tsense as s
			inner join tlanguage as l on s.languageid = l.id
			where word = :word
			limit 1`, &avdhp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract an article, sorry")
	ad = &articleDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}

// ArticleEditDirHandler is a handler to open edit page
func ArticleEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	var avdhp articleViewDirHandlerParams
	avdhp.Word = c.Param("word")

	if avdhp.Word == "" {
		apperror.Panic500If(apperror.ErrDummy, "No article for empty word")
		return
	}

	dataFound, ad := readArticleFromDb(&avdhp)

	if !dataFound {
		c.HTML(http.StatusBadRequest,
			"general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no article (yet?) for «%s»", avdhp.Word)})
		return
	}

	aetp := &articleEditTemplateParams{Ad: ad}
	c.HTML(http.StatusOK,
		"articleedit.html",
		aetp)
}
