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
	LanguageSlug string
	DialectSlug  string
	Word         string
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
	avdhp.LanguageSlug = "en"
	avdhp.DialectSlug = "-"
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
	db := sddb.SDUsersDb
	reply, err1 := db.Db.NamedQuery(
		`select 
			s.id as senseid
			,l.slug as languageslug
			,d.slug as dialectslug 
			,phrase
			,word 
			from tsense as s
			inner join tdialect as d on s.dialectid = d.id 
			inner join tlanguage as l on d.languageid = l.id
			where 
			l.slug = :languageslug 
			and d.slug = :dialectslug
			and word = :word
			limit 1`, &avdhp)
	apperror.Panic500AndErrorIf(err1, "Failed to extract an article, sorry")
	ad = &articleDataForEditType{}
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	sddb.FatalDatabaseErrorIf(err1, sddb.SDUsersDb, "Error obtaining data of sense: %#v", err1)
	return
}

// ArticleEditDirHandler is a handler to open edit page
func ArticleEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	var avdhp articleViewDirHandlerParams
	avdhp.LanguageSlug = "en"
	avdhp.DialectSlug = "-"
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
