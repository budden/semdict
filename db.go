package main

import (
	"fmt";	"os"; "time"
//	neturl "net/url"
//	"reflect"
//	"regexp"
//	"strings"
//	"time"
 "database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq")

func playWithDb() {
 url := "postgresql://localhost:5432"

	db, err, dbCloser := openDb(url)	
	if err != nil { 
		logicalPanic(fmt.Sprintf("Unable to connect to Postgresql, error is %#v", err)) }
	defer dbCloser()

 justAQuery := func(query string) {
  rows, err := db.Query(query)
  if err != nil { logicalPanic(fmt.Sprintf("Query error, query: «%s», error: %#v",query, err))}
  rows.Close()
 }

 justAQuery(`drop table if exists budden_a`)
 justAQuery(`create table budden_a (name text)`)
 justAQuery(`create unique index budden_a_name on budden_a (name)`)
 
	m := map[string]interface{}{"name" : `",sql 'inject?`}
	for i := 0; i < 2; i++ {
		var res sql.Result
		res, err1 := db.NamedExec(`insert into budden_a values (:name)`,
			m)
	//xt := reflect.TypeOf(err1).Kind()
		if err1 != nil {
			switch e := interface{}(err1).(type)	{
			case *pq.Error:
				if e.Code == "23505" {
					fmt.Printf("Duplicate key in %s",e.Constraint);
				} else {
					fmt.Printf("Error inserting: %#v\n",err1) }
			default: 
				fmt.Printf("Error insertiing: %#v\n",err1) }
		} else {
			fmt.Printf("Inserted %#v\n",res)	}}
	
	genExpiryDate(db)
 }

 func openDb(url string) (db *sqlx.DB, err error, closer func()) {
		db, err = sqlx.Open("postgres", url)
		closer = func() {
			err := db.Close()
			if err != nil {
				fmt.Printf("Error closing db: %v\n",err)
				// Not exiting because this function is called from the defer
			}}
		return }

	func genExpiryDate(db *sqlx.DB) {
		res1, err2 := db.Query(`select current_timestamp + interval '1' day`)
		if err2 != nil {
			fmt.Printf("Wow!"); os.Exit(1)	}
		if !res1.Next() {
			fmt.Printf("No rows here. Why?"); os.Exit(1) }
		var magic time.Time
		res1.Scan(&magic)
		fmt.Printf("Expiry at %s\n", magic.Format("2006-01-02 15:04 -0700"))	}
