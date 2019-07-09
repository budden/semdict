package query

import (
	"strconv"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// GetZeroOrOneNonNegativeIntFormValue extracts an (unique) integer value by the key
// from the request query (URL or form)
func GetZeroOrOneNonNegativeIntFormValue(c *gin.Context, key string) (
	value int64) {
	values, ok1 := c.GetQueryArray(key)
	if !ok1 || len(values) == 0 {
		return
	}
	if len(values) > 1 {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Query parameter «%s» is duplicated", key)
	}
	valueS := values[0]
	if valueS == "" {
		return
	}
	var err error
	value, err = strconv.ParseInt(valueS, 10, 64)
	apperror.Panic500AndLogAttackIf(err, c, "Non-integer value of «%s»", key)
	if value < 0 {
		apperror.Panic500AndLogAttackIf(apperror.ErrDummy, c, "Negative value of parameter «%s»", key)
	}
	return
}

// FIXME искать магическую константу 5000 в шаблонах.
const MaxDataSetRecordCountLimit = 5000

// If Limit is 0, set it to some reasonable maximum value
func LimitLimit(limit *int32) {
	if *limit == 0 {
		*limit = MaxDataSetRecordCountLimit
	}
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
