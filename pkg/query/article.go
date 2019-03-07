package query

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/budden/a/pkg/shared"
)

// ArticleViewDirHandler is a handler for an article view directory
func ArticleViewDirHandler(w http.ResponseWriter, r *http.Request) {
	// note https://github.com/golang/go/issues/24366#issuecomment-372764978
	// Its a bad idea to make this duplicating. Another options are StripPrefix,
	// changing a signature of handler to accept a prefix,
	SLUG := strings.TrimPrefix(r.URL.Path, shared.ArticleViewDirPath)
	data := GeneralTemplateParams{
		Message: SLUG}
	fileName := "templates/general.html"
	tmpl, err := template.ParseFiles(fileName)
	if err != nil {
		log.Printf("Error parsing %s: %#v\n", fileName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// ArticleEditDirHandler is a handler to open edit page
func ArticleEditDirHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ArticleEditDirHandler stub")
}
