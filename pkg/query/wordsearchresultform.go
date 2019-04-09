package query

// Прорисовывает грид

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// параметры из запроса
type wordSearchResultFormParams struct {
	Wordpattern string
}

// данные формы. Лишняя сущность, т.к. нам нужны только параметры
type wordSearchResultFormDataType struct {
	P *wordSearchResultFormParams
}

// Параметры шаблона. Опять же, в данном случае - лишняя сущность.
type wordSearchResultFormTemplateParamsType struct {
	D *wordSearchResultFormDataType
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchresult". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var wsrfp wordSearchResultFormParams
	wsrfp.Wordpattern = c.Query("wordpattern")
	if wsrfp.Wordpattern == "" {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Empty search pattern")
	}

	c.HTML(http.StatusOK,
		"wordsearchresultform.html",
		wordSearchResultFormTemplateParamsType{D: &wordSearchResultFormDataType{P: &wsrfp}})
}

func readWordSearchResultFromDb(frp *wordSearchResultFormParams) (fd *wordSearchResultFormDataType) {
	fd = &wordSearchResultFormDataType{}
	return
}
