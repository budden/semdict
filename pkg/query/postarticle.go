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
	ID         int32 // originId
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
		"/senseview/"+strconv.Itoa(int(pad.ID)))
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

func extractIdFromRequest(c *gin.Context) (id int32) {
	idAsString := c.PostForm("senseid")
	if idAsString == "" {
		idAsString = c.Param("senseid")
	}
	if idAsString != "" {
		padID, err := strconv.Atoi(idAsString)
		apperror.Panic500AndErrorIf(err, "Wrong sense ID")
		id = int32(padID)
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No sense ID given")
	}
	return
}

func extractDataFromRequest(c *gin.Context, pad *articlePostDataType) {
	pad.ID = extractIdFromRequest(c)
	pad.Phrase = c.PostForm("phrase")
	pad.Word = c.PostForm("word")
	pad.Ownerid = user.GetSDUserIdOrZero(c)
}

func writeToDb(pad *articlePostDataType) {
	if pad.ID != 0 {
		res, err1 := sddb.NamedExec(
			`select fnsavepersonalsense(:ownerid, :id, :phrase, :word, false)`, pad)
		_ = res
		/* res, err1 := sddb.NamedExec(
		`update tsense set phrase = :phrase, word = :word where id=:id`, pad) */
		apperror.Panic500AndErrorIf(err1, "Failed to update an article")
		// count, err2 := res.RowsAffected()
		// sddb.FatalDatabaseErrorIf(err2, "Unable to check if the record was updated")
		count := 1
		if count == 0 {
			apperror.Panic500AndErrorIf(apperror.ErrDummy, "Sense with id = %v not found", pad.ID)
		}
	} else {
		reply, err := sddb.NamedUpdateQuery(
			`insert into tsense (dialectid, phrase, word) values (1, :phrase, :word) returning id`,
			pad)
		apperror.Panic500AndErrorIf(err, "Failed to insert an article")
		dataFound := false
		for reply.Next() {
			dataFound = true
			err1 := reply.Scan(&pad.ID)
			sddb.FatalDatabaseErrorIf(err1, "Error obtaining id of a new article, err = %#v", err1)
		}
		if !dataFound {
			sddb.FatalDatabaseErrorIf(apperror.ErrDummy, "Id of a new sense is not returned")
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
