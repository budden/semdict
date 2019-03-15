// Package database contains things for db connection
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	//	neturl "net/url"
	//	"reflect"
	//	"regexp"
	//	"strings"
	//	"time"
	"database/sql"

	"github.com/budden/a/pkg/apperror"
	"github.com/budden/a/pkg/gracefulshutdown"
	"github.com/budden/a/pkg/shared"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// SDUsersDb contains, after the call to OpenSDUsersDb, a connection to sdusers_db
var SDUsersDb *sqlx.DB

// OpenSDUsersDb opens sdusers_db
func OpenSDUsersDb() (db *sqlx.DB) {
	url := shared.SecretConfigData.PostgresqlServerURL + "/sdusers_db"
	SDUsersDb = OpenDb(url, "sdusers_db")
	return
}

// PlayWithDb is used to manually test db functionality
func PlayWithDb() {
	justAQuery := func(query string) {
		rows, err := SDUsersDb.Query(query)
		apperror.LogicalPanicIf(err, "Query error, query: «%s», error: %#v", query, err)
		rows.Close()
	}

	justAQuery(`drop table if exists budden_a`)
	justAQuery(`create table budden_a (name text)`)
	justAQuery(`create unique index budden_a_name on budden_a (name)`)

	m := map[string]interface{}{"name": `",sql 'inject?`}
	for i := 0; i < 2; i++ {
		var res sql.Result
		res, err1 := SDUsersDb.NamedExec(`insert into budden_a values (:name)`,
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
	genExpiryDate(SDUsersDb)
}

const maxOpenConns = 4
const maxIdleConns = 4
const connMaxLifetime = 10 * time.Second

// CommitIfActive commits a transaction if it is still active.
func CommitIfActive(tx *sqlx.Tx) (err error) {
	err = tx.Commit()
	if err == sql.ErrTxDone {
		err = nil
	}
	return
}

// RollbackIfActive rolls back transaction if it is still active.
// Defer this one if you're opening any transaction
// If failed to rollback, will panic. If already panicking, would ignore
// rollback error silently.
func RollbackIfActive(tx *sqlx.Tx) {
	err := tx.Rollback()
	if err == nil || err == sql.ErrTxDone {
		return
	}
	preExistingPanic := recover()
	if preExistingPanic == nil {
		apperror.LogicalPanicIf(err, "Failed to rollback transaction")
	}
	log.Printf("Failed to rollback transaction while panicking. Err is %#v", err)
	apperror.LogicalPanicIf(preExistingPanic, "Failed to rollback tranaction while panicking")
}

// OpenDb obtains a connection to db. Connections are pooled, beware.
// logFriendlyName is for the case when url contains passwords
func OpenDb(url, logFriendlyName string) (db *sqlx.DB) {
	var err error
	db, err = sqlx.Open("postgres", url)
	apperror.GlobalPanicIf(err, "Failed to open «%s» database", logFriendlyName)
	err = db.Ping()
	apperror.GlobalPanicIf(err, "Failed to ping «%s» database", logFriendlyName)
	// http://go-database-sql.org/connection-pool.html
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	closer := func() {
		// FIXME db.Close can take an indefinite time. We should run closer in the goroutine and
		// wait for it in the graceful shutdown cleanup job.
		// have gracefulshutdown.Timeout
		err := db.Close()
		if err != nil {
			// we don't know if stdout is writable, but we're in goroutine already
			log.Printf("Error closing database «%s»: %#v\n", logFriendlyName, err)
		} else {
			log.Printf("Gracefully shut down database «%s»\n", logFriendlyName)
		}
	}
	gracefulshutdown.Actions = append(gracefulshutdown.Actions, closer)
	return
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
