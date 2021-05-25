package sddb

import (
	"log"
	"runtime/debug"

	"github.com/budden/semdict/pkg/apperror"
)

// SetConnectionDead используется для объявления о том, что соединение с db разорвано.
// Мы делаем это, когда произошло что-то неправильное, например, ошибка при откате.
func SetConnectionDead(conn *ConnectionType) {
	conn.IsDead = true
}

// IsConnectionDead следует вызывать перед каждым использованием соединения с бд, чтобы проверить, не умерло ли оно
// Если вы обнаружите, что он мёртв, верните статус 500
func IsConnectionDead(conn *ConnectionType) bool {
	return conn.IsDead
}

// FatalDatabaseErrorHandlerType - это тип для FatalDatabaseErrorHandler
type FatalDatabaseErrorHandlerType func(err error, conn *ConnectionType, format string, args ...interface{})

func initialFatalDatabaseErrorHandler(err error, conn *ConnectionType, format string, args ...interface{}) {
	if err != nil {
		log.Printf("Ранний вызов FatalDatabaseErrorIf: "+format, args...)
		debug.PrintStack()
		log.Fatal(err)
	}
}

// FatalDatabaseErrorHandler используется функцией FatalDatabaseError
var FatalDatabaseErrorHandler FatalDatabaseErrorHandlerType = initialFatalDatabaseErrorHandler

// Функция FatalDatabaseErrorIf вызывается из обработчика запроса, чтобы сообщить, что мы больше не можем работать с этой базой данных и должны закрыться.
// Если это произойдет, мы сначала объявим эту базу данных мёртвой.
// Далее мы инициируем "льготное отключение". Наконец, мы организуем возврат статуса 500 клиенту.
// Но этот модуль ничего не знает об http, поэтому здесь у нас только заглушка.
// Фактический обработчик хранится в переменной FatalDatabaseErrorHandler и устанавливается позже в app.actualFatalDatabaseErrorHandler
// FatalDataBaseErrorIf считается потокобезопасным
func FatalDatabaseErrorIf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	c := sdUsersDb
	FatalDatabaseErrorHandler(err, c, format, args...)
	return
}

// CheckDbAlive должен вызываться в обработчиках страниц перед каждым взаимодействием с БД
func CheckDbAlive() {
	c := sdUsersDb
	if IsConnectionDead(c) {
		apperror.Panic500If(apperror.ErrDummy, "Внутренняя ошибка")
	}
}
