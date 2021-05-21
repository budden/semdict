package user

import (
	"regexp"
	"sync"
)

// Мы блокируем мьютекс для любых записей в sdusers_db, чтобы минимизировать параллелизм на уровне БД
// TODO создать фн WithMutex и использовать его для всех запросов на запись
// Кроме того, подумайте о блокировках на уровне базы данных (помещать каждое обновление в транзакцию?).
var writeSDUsersMutex sync.Mutex

// PostgresqlErrorCodeUniqueViolation is a unique_violation,
// https://postgrespro.ru/docs/postgrespro/9.5/errcodes-appendix
const PostgresqlErrorCodeUniqueViolation = "23505"

// PostgresqlErrorCodeNoData = no_data warning
const PostgresqlErrorCodeNoData = "02000"

func isNicknameInValidFormat(nickname string) bool {
	matched, err := regexp.Match(`^[0-9a-zA-Z\p{L}]+$`, []byte(nickname))
	return err == nil && matched
}

// пароль должен быть длиной от 8 до 25 символов и не должен содержать пробелов

func isEmailInValidFormat(email string) bool {
	//https://socketloop.com/tutorials/golang-validate-email-address-with-regular-expression
	matched, err := regexp.Match(
		`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`,
		[]byte(email))
	return err == nil && matched
}
