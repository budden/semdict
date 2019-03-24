package user

import (
	"regexp"
	"sync"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/jmoiron/sqlx"
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

// WithTransaction opens a transaction in the sdusers_db, then runs body
// Then, if there is no error, and transaction is still active, commit transaction and returns commit's error
// If there was an error or panic while executing body, tries to rollback the tran transaction,
// see sddb.RollbackIfActive
func WithTransaction(body func(tx *sddb.TransactionType) (err error)) (err error) {

	conn := sddb.SDUsersDb
	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}

	var tx *sqlx.Tx
	sddb.CheckDbAlive(conn)
	tx, err = conn.Db.Beginx()
	trans := sddb.TransactionType{Conn: conn, Tx: tx}
	sddb.FatalDatabaseErrorIf(err, conn, "Unable to start transaction")
	defer func() { sddb.RollbackIfActive(&trans) }()
	sddb.CheckDbAlive(conn)
	_, err = tx.Exec(`set transaction isolation level repeatable read`)
	sddb.FatalDatabaseErrorIf(err, conn, "Unable to set transaction isolation level")
	sddb.CheckDbAlive(conn)
	err = body(&trans)
	if err == nil {
		sddb.CheckDbAlive(conn)
		err = sddb.CommitIfActive(&trans)
	}
	return
}

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
