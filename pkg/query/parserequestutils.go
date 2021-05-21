package query

import (
	"strconv"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// GetZeroOrOneNonNegativeIntFormValue извлекает (уникальное) целочисленное значение по ключу
// из запроса (URL или форма)
func GetZeroOrOneNonNegativeIntFormValue(c *gin.Context, key string) (
	value int64) {
	values, ok1 := c.GetQueryArray(key)
	if !ok1 || len(values) == 0 {
		return
	}
	if len(values) > 1 {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Параметр запроса «%s» дублируется", key)
	}
	valueS := values[0]
	if valueS == "" {
		return
	}
	var err error
	value, err = strconv.ParseInt(valueS, 10, 64)
	apperror.Panic500AndLogAttackIf(err, c, "Нецелое значение «%s»", key)
	if value < 0 {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Отрицательное значение параметра «%s»", key)
	}
	return
}

// FIXME искать магическую константу 5000 в шаблонах.
const MaxDataSetRecordCountLimit = 5000

// Если Limit равен 0, установите его на какое-то разумное максимальное значение
func LimitLimit(limit *int32) {
	if *limit == 0 {
		*limit = MaxDataSetRecordCountLimit
	}
}

func extractCheckBoxFromRequest(c *gin.Context, paramName string) (res bool) {
	txt := c.PostForm(paramName)
	if txt == "" {
		txt, _ = c.GetQuery(paramName)
	}
	if txt == "on" {
		res = true
	} else if txt == "off" || txt == "" {
		res = false
	} else {
		apperror.Panic500If(apperror.ErrDummy, "Плохое значение параметра запроса «%s»", txt)
	}
	return
}

func extractIdFromRequest(c *gin.Context, paramName string) (id int64) {
	idAsString := c.PostForm(paramName)
	if idAsString == "" {
		idAsString = c.Param(paramName)
	}
	if idAsString != "" {
		padID, err := strconv.ParseInt(idAsString, 10, 64)
		apperror.Panic500AndErrorIf(err, "Wrong "+paramName)
		id = padID
	} else {
		apperror.Panic500AndErrorIf(apperror.ErrDummy, "No "+paramName+" given")
	}
	return
}

func extractStringFromRequest(c *gin.Context, paramName string) (res string, found bool) {
	res, found = c.GetPostForm(paramName)
	if !found {
		res, found = c.Params.Get(paramName)
	}
	return
}
