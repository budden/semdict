package query

import (
	"fmt"
	"html"
	"net/http"

	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
)

// ArticleViewDirHandler ...
func ArticleViewDirHandler(c *gin.Context) {
	articleslug := c.Param("articleslug")
	if articleslug == "" {
		c.HTML(http.StatusNotFound, "", nil)
		return
	}
	c.HTML(http.StatusOK,
		"general.html",
		shared.GeneralTemplateParams{Message: fmt.Sprintf("Article page: %s", html.EscapeString(articleslug))})
}

// ArticleEditDirHandler is a handler to open edit page
func ArticleEditDirHandler(c *gin.Context) {
}
