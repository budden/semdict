package query

// Прорисовывает грид

import (
	"net/http"
	"net/url"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// параметры из запроса
type wordSearchResultFormParams struct {
	Wordpattern string
}

// Параметры шаблона
type wordSearchResultFormTemplateParamsType struct {
	P                  *wordSearchResultFormParams
	Wordsearchqueryurl string
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchresult". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var wsrfp wordSearchResultFormParams
	wsrfp.Wordpattern = c.Query("wordpattern")
	wp := wsrfp.Wordpattern
	if wp == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Empty search pattern")
	}

	queryURL := "/wordsearchquery" + "?wordpattern=" + url.QueryEscape(wp)

	c.HTML(http.StatusOK,
		"wordsearchresultform.html",
		wordSearchResultFormTemplateParamsType{P: &wsrfp, Wordsearchqueryurl: queryURL})
}
