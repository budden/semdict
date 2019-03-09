// Package query is for search. I wanted to call
// it 'search', but lint complains.
package query

import (
	"net/http"

	"github.com/budden/a/pkg/shared"
	"github.com/gin-gonic/gin"
)

// SearchFormPageHandler ...
func SearchFormPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK,
		"general.html",
		shared.GeneralTemplateParams{Message: "Search Form"})
}

// SearchResultPageHandler is a handler for search results page
func SearchResultPageHandler(c *gin.Context) {
	c.HTML(http.StatusInternalServerError, "", nil)
}
