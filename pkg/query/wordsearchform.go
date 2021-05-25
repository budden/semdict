package query

// Здесь мы пытаемся реализовать пример формы поиска как канонический пример формы (аналог редактора из conbred)

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"

	"github.com/gin-gonic/gin"
)

// Здесь перечислены все данные.
// На самом деле это должна быть модель (метаданные) - аналог deftbl из fb2/dict-ripol.lisp или
// definterface из wrapper.lisp. А структура
// должна из этого генерироваться. Для переносимости мы можем сделать метданные
// константой, переменной или функцией, переходить к определению с помощью go to symbol in workspace
// apperror.Panic500If (почему-то без префикса пакета, ну и ладно - там есть выбор одноимённых)
type wordSearchFormDataType struct {
	Dummyid     int32 // для формы поиска - всегда 1
	Wordpattern string
}

type wordSearchFormTemplateParamsType struct {
	Wsfd *wordSearchFormDataType
	Wsqp *wordSearchQueryParams
}

// WordSearchFormRouteHandler - обработчик для "/wordsearchform". Поддерживается случай, когда форма поиска
// заполняется через URL... По идее, это - runWrappedSprav - его частный случай
func WordSearchFormRouteHandler(c *gin.Context) {
	wsqp := getWordSearchQueryParamsFromRequest(c)

	// Прочитать данные из базы данных. Если нет данных, паниковать
	wsfd := readWordSearchFormFromDb(wsqp)

	// Здесь мы генерируем интерфейс, заполненный данными (или содержащий функции AJAX для динамического заполнения)
	// и отправляем клиенту
	c.HTML(http.StatusOK,
		// возможно, тут нужна развязка в зависимости от того, открываем ли мы на чтение или на ред-е - разные шаблоны
		"wordsearchform.t.html",
		wordSearchFormTemplateParamsType{Wsfd: wsfd, Wsqp: wsqp})
}

func getWordSearchQueryParamsFromRequest(c *gin.Context) (wsqp *wordSearchQueryParams) {
	wsqp = new(wordSearchQueryParams)
	wsqp.Wordpattern, _ = c.GetQuery("wordpattern")
	return
}

func readWordSearchFormFromDb(frp *wordSearchQueryParams) (fd *wordSearchFormDataType) {
	reply, err1 := sddb.NamedReadQuery(
		`select 1 as dummyid, cast(:wordpattern as text) as wordpattern`, frp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	defer sddb.CloseRows(reply)()
	fd = &wordSearchFormDataType{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(fd)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "Данные не найдены")
	}
	sddb.FatalDatabaseErrorIf(err1, "Ошибка при получении данных о смысле: %#v", err1)
	return
}
