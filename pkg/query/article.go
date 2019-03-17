package query

import (
	"fmt"
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/database"

	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

type articleViewDirHandlerParams struct {
	LanguageSlug string
	DialectSlug  string
	Word         string
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

	db := database.SDUsersDb
	reply, err1 := db.Db.NamedQuery(
		`select phrase from tsense 
			inner join tdialect on tsense.dialectid = tdialect.id 
			inner join tlanguage on tdialect.languageid = tlanguage.id
			where 
			tlanguage.slug = :languageslug 
			and tdialect.slug = :dialectslug
			and word = :word
			limit 1`, &avdhp)
	apperror.Panic500If(err1, "Failed to extract an article, sorry")
	dataFound := false
	var phrase string
	for reply.Next() {
		err1 = reply.Scan(&phrase)
		database.FatalDatabaseErrorIf(err1, database.SDUsersDb, "Error obtaining phrase of sense", err1)
		dataFound = true
	}
	if dataFound {
		c.HTML(http.StatusOK,
			"general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Article page for «%s»:\n%s", avdhp.Word, phrase)})
	} else {
		c.HTML(http.StatusBadRequest,
			"general.html",
			shared.GeneralTemplateParams{Message: fmt.Sprintf("Sorry, no article (yet?) for «%s»", avdhp.Word)})
	}
}

// ArticleEditDirHandler is a handler to open edit page
func ArticleEditDirHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
}
