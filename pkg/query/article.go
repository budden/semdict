package query

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/budden/a/pkg/shared"
)

// ArticleViewDirHandler is a handler for an article view directory
func ArticleViewDirHandler(w http.ResponseWriter, r *http.Request) {
	// note https://github.com/golang/go/issues/24366#issuecomment-372764978
	SLUG := r.URL.Path
	data := shared.GeneralTemplateParams{
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
