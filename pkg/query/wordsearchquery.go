package query

// Возвращает выборку подходящих слов в формате JSON

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WordSearchQueryRouteHandler - обработчик для "/wordsearchquery".
func WordSearchQueryRouteHandler(c *gin.Context) {
	_, fd := wordSearchCommonPart(c)
	// Выдать как JSON
	c.JSON(http.StatusOK, fd)
}
