package user

import (
	"regexp"
	"sync"
)

// Mutex we lock for any writes to sdusers_db to minimize parallelism at the db level
// TODO create an WithMutex fn and use it for all writing queries
// Also, think about db-level locks (put every update into transaction?)
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

// password must be from 8 to 25 characters long, must not contain spaces

func isEmailInValidFormat(email string) bool {
	//https://socketloop.com/tutorials/golang-validate-email-address-with-regular-expression
	matched, err := regexp.Match(
		`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`,
		[]byte(email))
	return err == nil && matched
}
