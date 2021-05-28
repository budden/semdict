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

// AppErr - это ошибка на уровне приложения, которая
// не должна приводить к краху приложения. Она реализует
// интерфейс ошибок и должна использоваться в качестве возвращаемого значения.
// Для паники используйте Exception500
type AppErr struct {
	Message string
}

// Exception500 означает, что произошло что-то относительно плохое,
// и мы хотим вернуть код ошибки 500.
// Но проблема локальна для текущего обработчика события, и
// наша программа продолжает работать. Используйте Exception500 в качестве аргумента для
// panic. Для возврата значения ошибки рассмотрим AppErr
type Exception500 struct {
	// Сообщение отправляется клиенту
	Message string
}

// NewAppErrf возвращает новый AppErr с сообщением
func NewAppErrf(format string, args ...interface{}) *AppErr {
	message := fmt.Sprintf(format, args...)
	result := AppErr{Message: message}
	return &result
}

func (be *AppErr) Error() string {
	return fmt.Sprintf("AppErr: %s", be.Message)
}

// HandlePanicInRequestHandlerMiddleware возвращает промежуточное ПО
// что для нашего известного "блага" паника восстанавливается и
// пишет 500, в противном случае печатает сообщение о панике и выходит из приложения.
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
						// это приведет к завершению приложения, если нет другого слишком умного промежуточного ПО
						// для восстановления
						panic(err)
					}
				}
			}
		}()
		c.Next()
	}
}

// ErrorWithContents это обёртка для любого значения.
// Может использоваться для преобразования паники в ошибку
type ErrorWithContents struct {
	Message  string
	Contents interface{}
}

func (i ErrorWithContents) Error() string {
	return fmt.Sprintf("«ErrorWithContents: «%s», %v»", i.Message, i.Contents)
}

func coerceToError(x interface{}) (e error) {
	if x == nil {
		log.Fatal("Попытка принудительного приведения nil к ошибке")
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

// GracefullyExitAppIf можно использовать, если ошибка считается не столь ужасной
// и мы можем позволить себе изящно выключить сервер. Ошибка не печатается
func GracefullyExitAppIf(err error, format string, args ...interface{}) {
	if err != nil {
		log.Printf(format, args...)
		debug.PrintStack()
		shutdown.InitiateGracefulShutdown()
	}
}

// ExitAppIf резко закрывает приложение
func ExitAppIf(err error, exitCode int, format string, args ...interface{}) {
	if err != nil {
		log.Printf(format, args...)
		debug.PrintStack()
		os.Exit(exitCode)
	}
}

// Panic500If следует вызывать внутри обработчика http-запросов.
// Она прерывает обработку, разматывает стек
// и возвращает статус 500 с отформатированным сообщением
func Panic500If(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		data := Exception500{Message: msg}
		panic(&data)
	}
}

// Panic500AndErrorIf похож на Panic500, но регистрирует ошибки. Мы должны регистрировать ошибки, чтобы знать, что не так и
// для fail2ban
// FIXME TODO - fail2ban и регистрация ошибок - это две разные вещи, напишите две функции!
func Panic500AndErrorIf(err error, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		data := Exception500{Message: msg}
		log.Printf("Panic500AndErrorIf: ошибка - %#v, сообщение для пользователя - '%s', стек следующий\n", err, msg)
		debug.PrintStack()
		panic(&data)
	}
}

// Panic500AndLogAttackIf используется, когда ошибка может быть частью атаки. Мы собираемся поставить его на
// журнал (может быть разделённый журнал) в форме, понятной fail2ban
func Panic500AndLogAttackIf(err error, c *gin.Context, format string, args ...interface{}) {
	if err != nil {
		msg := fmt.Sprintf(format, args...)
		log.Printf("Panic500AndLogAttackIf: сообщение «%s», ошибка %#v\n", msg, err)
		LogAttack(c, err)
		data := Exception500{Message: msg}
		panic(&data)
	}
}

// LogAttack регистрирует запрос, который является подозреваемой атакой (для fail2ban)
func LogAttack(c *gin.Context, err error) {
	log.Printf("Ошибка (может быть атакой). IP-адрес %s\n, Ошибка «%#v»\n", c.ClientIP(), err)
	// FIXME интеграция с fail2ban
}

// Panic200 должен быть вызван внутри обработчика http запроса и вызовет
// шаблон по умолчанию, который будет построен вместе с сообщением
/* func Panic200(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	data := Exception200{Message: msg}
	panic(&data)
} */

// LogicalPanicIf должен выполняться в обработчике веб-запросов и разворачивать только текущую горутину
func LogicalPanicIf(subject interface{}, format string, args ...interface{}) {
	if subject != nil {
		err := errors.WithMessagef(coerceToError(subject), format, args...)
		panic(err)
	}
}

// ErrDummy можно использовать в качестве первого аргумента DoSomethingIf(err,...), если
// реальной ошибки нет
var ErrDummy = errors.New("Фиктивная ошибка")
