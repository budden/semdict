package query

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/microcosm-cc/bluemonday"

	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/database"
	"github.com/gin-gonic/gin"
)

type articlePostDataType struct {
	ID           int32
	LanguageSlug string // unprocessed
	DialectSlug  string // unprocessed
	Phrase       string
	Word         string
}

// ArticleEditFormSubmitPostHandler posts an article data
func ArticleEditFormSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &articlePostDataType{}
	extractDataFromRequest(c, pad)
	sanitizeData(pad)
	writeToDb(pad)
	// promote the user to Sd Db. If we crash here, user will be able to login,
	// (and unable to register again), but wil be missing from the main content db

	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/articleview/"+
			// https://stackoverflow.com/a/43429641/9469533
			url.PathEscape(pad.Word))
}

func sanitizeData(pad *articlePostDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	// TODO: match word with this: /^[a-zA-Z ]+$/\p{L}
}

func extractDataFromRequest(c *gin.Context, pad *articlePostDataType) {
	idAsString := c.PostForm("senseid")
	if idAsString != "" {
		padID, err := strconv.Atoi(idAsString)
		apperror.Panic500If(err, "Wrong article ID")
		pad.ID = int32(padID)
	} else {
		pad.ID = 0
	}
	pad.Phrase = c.PostForm("phrase")
	pad.Word = c.PostForm("word")
}

func writeToDb(pad *articlePostDataType) {
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

/* Example of nested records in the template:

package main

import (
	"html/template"
	"log"
	"os"
)

func main() {
	type z struct{ Msg string; Child *z }
	v := z{Msg: "hi", Child: &z{Msg: "wow"}}
	master := "Greeting: {{ .Msg}}, {{ .Child.Msg}}"
	masterTmpl, err := template.New("master").Parse(master)
	if err != nil {
		log.Fatal(err)
	}
	if err := masterTmpl.Execute(os.Stdout, v); err != nil {
		log.Fatal(err)
	}
}

*/
