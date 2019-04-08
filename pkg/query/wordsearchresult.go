package query

// Здесь будет грид заложён

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Здесь перечислены все данные.
// На самом деле это должна быть модель (метаданные) - аналог deftbl из fb2/dict-ripol.lisp или
// definterface из wrapper.lisp. А структура
// должна из этого генерироваться. Для переходимости метданные мы можем сделать
// константой, переменной или функцией, переходить к определению с помощью go to symbol in workspace
// apperror.Panic500If (почему-то без префикса пакета, ну и ладно - там есть выбор одноимённых)
type wordSearchResultDataType struct {
	Resultid    int32 // для формы поиска - всегда 1
	Wordpattern string
}

// Параметры, которые могут участвовать в маршруте. Например, невидимые поля не могут.
// А может быть могут быть и параметры такие, как
// В десктопе аналогом может быть params из runWrappedSprav
// Интересный вопрос - можно ли передать wrapper через URL? Вообще говоря, wrapper - слишком могущественная вещь
// для URL-ов.
type wordSearchResultRouteParams struct {
	Wordpattern string
}

// Это нужно для статической типизации параметров шаблона? Или вообще неyужно?
type wordSearchResultTemplateParamsType struct {
	Fd *wordSearchResultDataType
}

// WordSearchResultRouteHandler - обработчик для "/wordsearchform". Поддерживается случай, когда форма поиска
// заполняет через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchResultRouteHandler(c *gin.Context) {
	var frp wordSearchResultRouteParams

	// Извлечь параметры из запроса
	// frp.Id = extractIdFromRequest(c)

	// Прочитать данные из базы данных. Если нет данных, паниковать
	fd := readWordSearchResultFromDb(&frp)

	// Здесь мы генерируем интерфейс, заполненный данными (или содержащий функции AJAX для динамического заполнения)
	// и отправляем клиенту
	c.HTML(http.StatusOK,
		// возможно, тут нужна развязка в зависимости от того, открываем ли мы на чтение или на ред-е - разные шаблоны
		"wordsearchresult.html",
		wordSearchResultTemplateParamsType{Fd: fd})
}

func readWordSearchResultFromDb(frp *wordSearchResultRouteParams) (fd *wordSearchResultDataType) {
	// FIXME скопируй всё это из wordsearchform.html, если первая порция должна впечатываться сразу
	fd = &wordSearchResultDataType{}
	return
}
