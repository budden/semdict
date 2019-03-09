// Package query is for search. I wanted to call
// it 'search', but lint complains.
package query

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/budden/a/pkg/shared"
)

// SearchFormPageHandler is a handler for a search page
func SearchFormPageHandler(w http.ResponseWriter, r *http.Request) {
	data := shared.GeneralTemplateParams{
		Message: "Search form"}
	fileName := "templates/general.html"
	tmpl, err := template.ParseFiles(fileName)
	if err != nil {
		log.Printf("Error parsing %s: %#v\n", fileName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// SearchResultPageHandler is a handler for search results page
func SearchResultPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to semantic dictionary!")
}
