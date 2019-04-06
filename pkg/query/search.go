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
// должна из этого генерироваться. Для переходимости метданные мы можем сделать
// константой, переменной или функцией, хотя переходить к определению будет и неудобно.
type frmDataType struct {
	Frmid       int32 // для формы поиска - всегда 1
	Wordpattern string
}

// Параметры, которые могут участвовать в маршруте. Например, невидимые поля не могут.
// А может быть могут быть и параметры такие, как
// В десктопе аналогом может быть params из runWrappedSprav
// Интересный вопрос - можно ли передать wrapper через URL? Вообще говоря, wrapper - слишком могущественная вещь
// для URL-ов.
type frmRouteParams struct {
	Wordpattern string
}

// Это нужно для статической типизации параметров шаблона? Или вообще неyужно?
type frmTemplateParamsType struct {
	Ad *frmDataType
}

// FrmRouteHandler нужен для случая, когда форма поиска заполняет через URL...
// По идее, это - runWrappedSprav - его частный случай
func FrmRouteHandler(c *gin.Context) {
	var avdhp frmRouteParams

	// Извлечь параметры из запроса
	// avdhp.Id = extractIdFromRequest(c)

	// Прочитать данные из базы данных. Если нет данных, паниковать
	ad := readFrmFromDb(&avdhp)

	// Здесь мы генерируем интерфейс, заполненный данными (или содержащий функции AJAX для динамического заполнения)
	// и отправляем клиенту
	c.HTML(http.StatusOK,
		// возможно, тут нужна развязка в зависимости от того, открываем ли мы на чтение или на ред-е - разные шаблоны
		"senseedit.html",
		frmTemplateParamsType{Ad: ad})
}

func readFrmFromDb(avdhp *frmRouteParams) (ad *frmDataType) {
	reply, err1 := sddb.NamedReadQuery(
		`select 1 as frmid, :wordPattern as wordPattern`, &avdhp)
	apperror.Panic500AndErrorIf(err1, "Db query failed")
	ad = &frmDataType{}
	dataFound := false
	for reply.Next() {
		err1 = reply.StructScan(ad)
		dataFound = true
	}
	if !dataFound {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No data found")
	}
	sddb.FatalDatabaseErrorIf(err1, "Error obtaining data of sense: %#v", err1)
	return
}
