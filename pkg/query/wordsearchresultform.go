package query

import (
	"net/http"
	"net/url"

	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// Параметры шаблона
type wordSearchResultFormTemplateParamsType struct {
	P                     *wordSearchQueryParams
	Wordpatternurlencoded string
	Records               []*wordSearchQueryRecord
	IsLoggedIn            bool
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchresult". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var frp *wordSearchQueryParams
	var fd []*wordSearchQueryRecord
	frp, fd = wordSearchCommonPart(c)

	wpu := url.QueryEscape(frp.Wordpattern)

	c.HTML(http.StatusOK,
		"wordsearchresultform.html",
		wordSearchResultFormTemplateParamsType{P: frp,
			Wordpatternurlencoded: wpu,
			Records:               fd,
			IsLoggedIn:            user.IsLoggedIn(c)})
}
