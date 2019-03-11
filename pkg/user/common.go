package user

import (
	"fmt"
	"sync"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/shared"
	"github.com/budden/a/pkg/unsorted"
	"github.com/jmoiron/sqlx"
)

// Mutex we lock for any writes to sdusers_db to minimize parallelism at the db level
var writeSDUsersMutex sync.Mutex

// PostgresqlErrorCodeUniqueViolation is a unique_violation,
// https://postgrespro.ru/docs/postgrespro/9.5/errcodes-appendix
const PostgresqlErrorCodeUniqueViolation = "23505"

// PostgresqlErrorCodeNoData = no_data warning
const PostgresqlErrorCodeNoData = "02000"

func openSDUsersDb() (db *sqlx.DB, dbCloser func()) {
	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"
	var err error
	db, dbCloser, err = database.OpenDb(url)
	if err != nil {
		unsorted.LogicalPanic(fmt.Sprintf("Unable to connect to Postgresql, error is %#v", err))
	}
	return
}
