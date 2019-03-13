package unsorted

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
)

// ErrorWithContents is a wrapper for any value.
// Can be used to convert panic to error
type ErrorWithContents struct {
	Message  string
	Contents interface{}
}

func (i ErrorWithContents) Error() string {
	return fmt.Sprintf("«ErrorWithContents: «%s», %v»", i.Message, i.Contents)
}

func coerceToError(x interface{}) (e error) {
	if x == nil {
		log.Fatal("Attempting to coerce nil to error")
	}
	switch xt := x.(type) {
	case error:
		e = xt
	default:
		ewc := ErrorWithContents{Message: "<>", Contents: xt}
		e = ewc
	}
	return
}

// LogicalPanicIf should run in a web query handler and unwind current goroutine only
func LogicalPanicIf(subject interface{}, format string, args ...interface{}) {
	if subject != nil {
		err := errors.WithMessagef(coerceToError(subject), format, args...)
		panic(err)
	}
}

// GlobalPanicIf is intended to run in a goroutine other than
// web query handler and crash the application
func GlobalPanicIf(subject interface{}, format string, args ...interface{}) {
	if subject != nil {
		err := errors.WithMessagef(coerceToError(subject), format, args...)
		panic(err)
	}
}
