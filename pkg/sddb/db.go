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

	// If we serialize writes to this db, we
	// hold the mutex while being in the transaction (FIXME - do that only in write transaction)
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

// SDUsersDb contains, after the call to OpenSDUsersDb, a connection to sdusers_db
var SDUsersDb *ConnectionType

// OpenSDUsersDb opens sdusers_db
func OpenSDUsersDb() {
	if SDUsersDb != nil {
		log.Fatal("An attempt to re-open SDUsers database")
	}
	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"
	SDUsersDb = OpenDb(url, "sdusers_db", true)
	return
}

// PlayWithDb is used to manually test db functionality
func PlayWithDb() {
	justAQuery := func(query string) {
		rows, err := SDUsersDb.Db.Query(query)
		apperror.LogicalPanicIf(err, "Query error, query: «%s», error: %#v", query, err)
		rows.Close()
	}

	justAQuery(`drop table if exists budden_a`)
	justAQuery(`create table budden_a (name text)`)
	justAQuery(`create unique index budden_a_name on budden_a (name)`)

	m := map[string]interface{}{"name": `",sql 'inject?`}
	for i := 0; i < 2; i++ {
		var res sql.Result
		res, err1 := SDUsersDb.Db.NamedExec(`insert into budden_a values (:name)`,
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
	genExpiryDate(SDUsersDb.Db)
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
		FatalDatabaseErrorIf(apperror.ErrDummy, trans.Conn, "Failed to rollback transaction")
	} else if ae, ok := preExistingPanic.(apperror.Exception500); ok {
		FatalDatabaseErrorIf(apperror.ErrDummy,
			trans.Conn,
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
