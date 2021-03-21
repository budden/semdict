package query

import (
	"net/http"
	"net/url"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

/// Поиск смысла

// Параметры шаблона
type wordSearchResultFormTemplateParamsType struct {
	Wsqp                  *wordSearchQueryParams
	Wordpatternurlencoded string
	Records               []*wordSearchQueryRecord
	IsLoggedIn            bool
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchresult". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var wsqp *wordSearchQueryParams
	var wsfd []*wordSearchQueryRecord
	wsqp, wsfd = wordSearchCommonPart(c)

	wpu := url.QueryEscape(wsqp.Wordpattern)

	c.HTML(http.StatusOK,
		"wordsearchresultform.t.html",
		wordSearchResultFormTemplateParamsType{Wsqp: wsqp,
			Wordpatternurlencoded: wpu,
			Records:               wsfd,
			IsLoggedIn:            user.IsLoggedIn(c)})
}
