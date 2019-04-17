package query

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/microcosm-cc/bluemonday"

	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/gin-gonic/gin"
)

type articlePostDataType struct {
	Proposalid int64 // must be here
	Commonid   int64 // can be 0 if no origin (adding proposal)
	Languageid int32
	Phrase     string
	Word       string
	Deleted    bool
	Ownerid    int32
}

// SenseEditFormSubmitPostHandler posts an sense data
func SenseEditFormSubmitPostHandler(c *gin.Context) {
	user.EnsureLoggedIn(c)
	pad := &articlePostDataType{}
	extractDataFromRequest(c, pad)
	sanitizeData(pad)
	writeToDb(pad)
	// promote the user to Sd Db. If we crash here, user will be able to login,
	// (and unable to register again), but wil be missing from the main content db

	// https://github.com/gin-gonic/gin/issues/444
	c.Redirect(http.StatusFound,
		"/sensebyidview/"+strconv.Itoa(int(pad.Proposalid)))
	//// https://stackoverflow.com/a/43429641/9469533
	//url.PathEscape(pad.Word))
}

func sanitizeData(pad *articlePostDataType) {
	// example just from the title page of https://github.com/microcosm-cc/bluemonday
	p := bluemonday.UGCPolicy()
	pad.Phrase = p.Sanitize(pad.Phrase)
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L} ]+$`, []byte(pad.Word))
	if (err != nil) || !matched {
		// https://www.linux.org.ru/forum/development/14877320
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Word can only contain letters, digits and spaces")
	}
}

func extractIdFromRequest(c *gin.Context, paramName string) (id int64) {
	idAsString := c.PostForm(paramName)
	if idAsString == "" {
		idAsString = c.Param(paramName)
	}
	if idAsString != "" {
		padID, err := strconv.ParseInt(idAsString, 10, 64)
		apperror.Panic500AndErrorIf(err, "Wrong "+paramName)
		id = padID
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No "+paramName+" given")
	}
	return
}

func extractDataFromRequest(c *gin.Context, pad *articlePostDataType) {
	pad.Proposalid = extractIdFromRequest(c, "proposalid")
	pad.Commonid = extractIdFromRequest(c, "commonid")
	pad.Phrase = c.PostForm("phrase")
	pad.Word = c.PostForm("word")
	pad.Ownerid = user.GetSDUserIdOrZero(c)
}

func writeToDb(pad *articlePostDataType) {
	res, err1 := sddb.NamedExec(
		`select fnsavepersonalsense(:ownerid, :commonid, :proposalid, :phrase, :word, false)`, pad)
	_ = res
	/* res, err1 := sddb.NamedExec(
	`update tsense set phrase = :phrase, word = :word where id=:id`, pad) */
	apperror.Panic500AndErrorIf(err1, "Failed to update an article")
	// count, err2 := res.RowsAffected()
	// sddb.FatalDatabaseErrorIf(err2, "Unable to check if the record was updated")
	count := 1
	if count == 0 {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sense with id = %v not found", pad.Proposalid)
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
