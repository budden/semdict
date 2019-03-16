package apperror

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/budden/a/pkg/shared"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// AppErr is an application level error which
// should not crash the application
type AppErr struct {
	Message string
}

// NewAppErrf returns a new AppErr with a message
func NewAppErrf(format string, args ...interface{}) *AppErr {
	message := fmt.Sprintf(format, args...)
	result := AppErr{Message: message}
	return &result
}

func (be *AppErr) Error() string {
	return fmt.Sprintf("AppErr: %s", be.Message)
}

// Exception500 means that something relatively bad happened,
// but our program is still operational
type Exception500 struct {
	Message string
}

// HandlePanicInRequestHandler returns a middleware
// that, for our known "good" panics recovers and
// writes a 500, otherwise it prints the panic and exits the app.
func HandlePanicInRequestHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch et := err.(type) {
				case *Exception500:
					{
						c.HTML(http.StatusInternalServerError,
							"general.html",
							shared.GeneralTemplateParams{Message: et.Message})
						return
					}
				default:
					{
						// this will exit app if no other too smart middleware
						// would recover
						panic(err)
					}
				}
			}
		}()
		c.Next()
	}
}

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

// ExitAppIf is called when something really bad happened
func ExitAppIf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Printf(format, args...)
		debug.PrintStack()
		log.Fatal(err)
	}
}

// Panic500If should be called inside an http request handler, cancel handling, unwind the stack
// and return 500 status with the formatted message
func Panic500If(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		data := Exception500{Message: msg}
		panic(&data)
	}
}

// Panic200 should be called inside an http request handler and will cause the
// default template to be built with the message
/* func Panic200(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	data := Exception200{Message: msg}
	panic(&data)
} */

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

// ErrDummy can be used as a first argument to DoSomethingIf(err,...) if there is
// no real error at hand
var ErrDummy = errors.New("Dummy error")
