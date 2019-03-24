// Package sddb contains things for db connection
package sddb

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	//	neturl "net/url"
	//	"reflect"
	//	"regexp"
	//	"strings"
	//	"time"
	"database/sql"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/shutdown"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ConnectionType encapsulates connection with
// some other minor things like a broken state of a connection pool.
// There should only one ConnectionType instance for any DB instance
type ConnectionType struct {
	Db *sqlx.DB

	// We hold the mutex for all writes to the database.
	// This way we control the number of concurrent transactions in the database.
	// For now we only have one instance of the service, so there will be no more than
	// one concurrent write to the database. If we have several instances, there will be
	// no more concurrent writes than there are service instances.
	Mutex *sync.Mutex

	IsDead bool
}

// TransactionType is a transaction and reference to
// ConnectionType, which we need due to inability to get db from
// transaction.
type TransactionType struct {
	Conn *ConnectionType
	Tx   *sqlx.Tx
}

// sdUsersDb contains, after the call to OpenSdUsersDb, a connection to sdusers_db
var sdUsersDb *ConnectionType

// OpenSdUsersDb opens sdusers_db
func OpenSdUsersDb() {
	if sdUsersDb != nil {
		log.Fatal("An attempt to re-open SDUsers database")
	}
	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"
	sdUsersDb = OpenDb(url, "sdusers_db", true)
	return
}

// PlayWithDb is used to manually test db functionality
func PlayWithDb() {
	justAQuery := func(query string) {
		rows, err := sdUsersDb.Db.Query(query)
		apperror.LogicalPanicIf(err, "Query error, query: «%s», error: %#v", query, err)
		rows.Close()
	}

	justAQuery(`drop table if exists budden_a`)
	justAQuery(`create table budden_a (name text)`)
	justAQuery(`create unique index budden_a_name on budden_a (name)`)

	m := map[string]interface{}{"name": `",sql 'inject?`}
	for i := 0; i < 2; i++ {
		var res sql.Result
		res, err1 := sdUsersDb.Db.NamedExec(`insert into budden_a values (:name)`,
			m)
		//xt := reflect.TypeOf(err1).Kind()
		if err1 != nil {
			switch e := interface{}(err1).(type) {
			case *pq.Error:
				if e.Code == "23505" {
					fmt.Printf("Duplicate key in %s", e.Constraint)
				} else {
					fmt.Printf("Error inserting: %#v\n", err1)
				}
			default:
				fmt.Printf("Error insertiing: %#v\n", err1)
			}
		} else {
			fmt.Printf("Inserted %#v\n", res)
		}
	}
	genExpiryDate(sdUsersDb.Db)
}

const maxOpenConns = 4
const maxIdleConns = 4
const connMaxLifetime = 10 * time.Second

// CommitIfActive commits a transaction if it is still active.
func CommitIfActive(trans *TransactionType) (err error) {
	err = trans.Tx.Commit()
	if err == sql.ErrTxDone {
		err = nil
	}
	return
}

// RollbackIfActive rolls back transaction if it is still active.
// Defer this one if you're opening any transaction
// If failed to rollback, will shutdown the application gracefully
// If already panicking with a non-application error (something unexpected happened),
// will continue to do the same, just logging the failure to rollback.
func RollbackIfActive(trans *TransactionType) {
	err := trans.Tx.Rollback()
	if err == nil || err == sql.ErrTxDone {
		return
	}
	preExistingPanic := recover()
	if preExistingPanic == nil {
		FatalDatabaseErrorIf(apperror.ErrDummy, "Failed to rollback transaction")
	} else if ae, ok := preExistingPanic.(apperror.Exception500); ok {
		FatalDatabaseErrorIf(apperror.ErrDummy,
			"Failed to rollback transaction with Exception500 pending: %#v",
			ae)
	} else {
		debug.PrintStack()
		log.Fatalf("Failed to rollback transaction while panicking with %#v", preExistingPanic)
	}
}

// OpenDb obtains a connection to db. Connections are pooled, beware.
// logFriendlyName is for the case when url contains passwords
func OpenDb(url, logFriendlyName string, withMutex bool) *ConnectionType {
	var err error
	var db *sqlx.DB
	db, err = sqlx.Open("postgres", url)
	apperror.ExitAppIf(err, 6, "Failed to open «%s» database", logFriendlyName)
	err = db.Ping()
	apperror.ExitAppIf(err, 7, "Failed to ping «%s» database", logFriendlyName)
	// http://go-database-sql.org/connection-pool.html
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	closer1 := func() {
		err := db.Close()
		if err != nil {
			// we don't know if stdout is writable, but we're in goroutine already
			log.Printf("Error closing database «%s»: %#v\n", logFriendlyName, err)
		} else {
			log.Printf("Closed database «%s»\n", logFriendlyName)
		}
	}
	closer := func() { go closer1() }
	var mtx *sync.Mutex
	if withMutex {
		var mtx2 sync.Mutex
		mtx = &mtx2
	}
	shutdown.Actions = append(shutdown.Actions, closer)
	result := ConnectionType{Db: db, Mutex: mtx}
	return &result
}

func genExpiryDate(db *sqlx.DB) {
	res1, err2 := db.Query(`select current_timestamp + interval '10' minutes`)
	if err2 != nil {
		fmt.Printf("Wow!")
		os.Exit(1)
	}
	if !res1.Next() {
		fmt.Printf("No rows here. Why?")
		os.Exit(1)
	}
	var magic time.Time
	res1.Scan(&magic)
	fmt.Printf("Expiry at %s\n", magic.Format("2006-01-02 15:04 -0700"))
}

// WithTransaction opens a transaction in the sdusers_db, then runs body
// Then, if there is no error, and transaction is still active, commit transaction and returns commit's error
// If there was an error or panic while executing body, tries to rollback the tran transaction,
// see RollbackIfActive
func WithTransaction(body func(tx *TransactionType) (err error)) (err error) {

	conn := sdUsersDb
	CheckDbAlive()

	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}

	var tx *sqlx.Tx
	CheckDbAlive()
	tx, err = conn.Db.Beginx()
	trans := TransactionType{Conn: conn, Tx: tx}
	FatalDatabaseErrorIf(err, "Unable to start transaction")
	defer func() { RollbackIfActive(&trans) }()
	CheckDbAlive()
	_, err = tx.Exec(`set transaction isolation level repeatable read`)
	FatalDatabaseErrorIf(err, "Unable to set transaction isolation level")
	CheckDbAlive()
	err = body(&trans)
	if err == nil {
		CheckDbAlive()
		err = CommitIfActive(&trans)
	}
	return
}

// NamedUpdateQuery is a query in the sdUsersDb which updates the db. So we hold our mutex
// to ensure serialization of all writes in scope of instances.
func NamedUpdateQuery(sql string, params interface{}) (res *sqlx.Rows, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	CheckDbAlive()
	res, err = conn.Db.NamedQuery(sql, params)
	return
}

// NamedExec is like sqlx.NamedExec and also holds the mutex. Use it whenever the query executed
// can update the db
func NamedExec(sql string, params interface{}) (res sql.Result, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	CheckDbAlive()
	res, err = conn.Db.NamedExec(sql, params)
	return
}

// NamedReadQuery is for queries which are read-only. We introduce one as we encapsulate
// sql(x) connection of sdusers_db in the sddb module.
func NamedReadQuery(sql string, params interface{}) (res *sqlx.Rows, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	res, err = conn.Db.NamedQuery(sql, params)
	return
}
