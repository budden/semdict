package query

import (
	"net/http"
	"strconv"

	"github.com/microcosm-cc/bluemonday"

	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/database"
	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
)

type postArticleDataType struct {
	ID           int64
	LanguageSlug string // unprocessed
	DialectSlug  string // unprocessed
	Phrase       string
	Word         string
}

// ArticlePostDataPageHandler posts an article data
func ArticlePostDataPageHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	var pad postArticleDataType
	extractDataFromRequest(c, &pad)
	sanitizeData(&pad)
	writeToDb(&pad)
	// promote the user to Sd Db. If we crash here, user will be able to login,
	// (and unable to register again), but wil be missing from the main content db
	c.HTML(http.StatusMovedPermanently,
		"general.html",
		shared.GeneralTemplateParams{Message: "Registration confirmed. Now you can proceed to the <a href=/>Login page</a>"})
}

func sanitizeData(pad *postArticleDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	// todo: match word with this: /^[a-zA-Z ]+$/\p{L}
}

func extractDataFromRequest(c *gin.Context, pad *postArticleDataType) {
	query := c.Request.URL.Query()
	phrases, ok1 := query["phrase"]
	words, ok2 := query["word"]
	ids, _ := query["id"]

	if !ok1 || !ok2 ||
		len(phrases) == 0 ||
		len(words) == 0 {
		apperror.Panic500If(apperror.ErrDummy, "Bad query")
	}
	pad.Phrase = phrases[0]
	pad.Word = words[0]
	if len(ids) > 0 {
		idAsString := ids[0]
		padID, err := strconv.Atoi(idAsString)
		apperror.Panic500If(err, "Wrong article ID")
		pad.ID = int64(padID)
	}
}

func writeToDb(pad *postArticleDataType) {
	db := database.SDUsersDb
	database.CheckDbAlive(db)
	if pad.ID != 0 {
		res, err1 := db.Db.NamedExec(
			`update tsense set phrase = :phrase, word = :word where	dialectid = 1 and id=:id`, pad)
		apperror.Panic500If(err1, "Failed to update an article")
		count, err2 := res.RowsAffected()
		database.FatalDatabaseErrorIf(err2, db, "Unable to check if the record was updated")
		if count == 0 {
			apperror.Panic500If(apperror.ErrDummy, "Article with id = %v not found", pad.ID)
		}
	} else {
		reply, err := db.Db.NamedQuery(
			`insert into tsense (dialectid, phrase, word) values (1, :phrase, :word) returning id`, pad)
		apperror.Panic500If(err, "Failed to insert an article")
		dataFound := false
		for reply.Next() {
			dataFound = true
			err1 := reply.Scan(&pad.ID)
			database.FatalDatabaseErrorIf(err1, database.SDUsersDb, "Error obtaining id of a new article, err = %#v", err1)
		}
		if !dataFound {
			database.FatalDatabaseErrorIf(apperror.ErrDummy, database.SDUsersDb, "Id of a new article is not returned")
		}
	}
}
