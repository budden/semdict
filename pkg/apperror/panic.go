package apperror

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/budden/semdict/pkg/shutdown"

	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// AppErr is an application level error which
// should not crash the application. It implements
// error interface and should be used as a return value.
// For panic, use Exception500
type AppErr struct {
	Message string
}

// Exception500 means that something relatively bad happened,
// and we want to return 500 error code.
// But the issue is local for current event handler, and
// our program is still operational. Use Exception500 as an argument to
// panic. For error return value, consider AppErr
type Exception500 struct {
	// Message is sent to the client
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

// HandlePanicInRequestHandlerMiddleware returns a middleware
// that, for our known "good" panics recovers and
// writes a 500, otherwise it prints the panic and exits the app.
func HandlePanicInRequestHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				switch et := err.(type) {
				case *Exception500:
					{
						c.HTML(http.StatusInternalServerError,
							"general.t.html",
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

// GracefullyExitAppIf can be used if error is considered not so horrible
// and we can afford to shutdown the server gracefully. Error is not printed
func GracefullyExitAppIf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Printf(format, args...)
		debug.PrintStack()
		shutdown.InitiateGracefulShutdown()
	}
}

// ExitAppIf closes the app abruptly
func ExitAppIf(err error, exitCode int, format string, args ...interface{}) {
	if err != nil {
		log.Printf(format, args...)
		debug.PrintStack()
		os.Exit(exitCode)
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

// Panic500AndErrorIf is like Panic500, but logs error. We need to log error to know what is wrong and
// for fail2ban
// FIXME TODO - fail2ban and error logging are two different things, write two functions!
func Panic500AndErrorIf(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		data := Exception500{Message: msg}
		log.Printf("Panic500AndErrorIf: error is %#v, message for the user is «%s», stack follows\n", err, msg)
		debug.PrintStack()
		panic(&data)
	}
}

// Panic500AndLogAttackIf is used when the error can be a part of attack. We're going to put it to
// the log (maybe a separated log) in the form understandible by fail2ban
func Panic500AndLogAttackIf(err error, c *gin.Context, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		log.Printf("Panic500AndLogAttaciIf: message is «%s»\n", msg)
		LogAttack(c, err)
		data := Exception500{Message: msg}
		panic(&data)
	}
}

// LogAttack is recording a request which is a suspected attack (for fail2ban)
func LogAttack(c *gin.Context, err error) {
	log.Printf("Error (may be an attack). IP address is %s\n, Error is «%#v»\n", c.ClientIP(), err)
	// FIXME integrate with fail2ban
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

// ErrDummy can be used as a first argument to DoSomethingIf(err,...) if there is
// no real error at hand
var ErrDummy = errors.New("Dummy error")
