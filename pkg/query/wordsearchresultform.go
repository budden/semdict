package query

// Прорисовывает грид

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// параметры из запроса
type wordSearchResultFormParams struct {
	Wordpatternurlencoded string
}

// данные формы. Лишняя сущность, т.к. нам нужны только параметры
type wordSearchResultFormDataType struct {
	p *wordSearchResultFormParams
}

// Параметры шаблона. Опять же, в данном случае - лишняя сущность.
type wordSearchResultFormTemplateParamsType struct {
	d *wordSearchResultFormDataType
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchresult". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var wsrfp wordSearchResultFormParams
	wsrfp.Wordpatternurlencoded = c.Query("wordpattern")
	if wsrfp.Wordpatternurlencoded == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Empty search pattern")
	}

	c.HTML(http.StatusOK,
		"wordsearchresultform.html",
		wordSearchResultFormTemplateParamsType{d: &wordSearchResultFormDataType{p: &wsrfp}})
}

func readWordSearchResultFromDb(frp *wordSearchResultFormParams) (fd *wordSearchResultFormDataType) {
	fd = &wordSearchResultFormDataType{}
	return
}
