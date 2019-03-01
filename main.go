package main
// To run this one you need that current user is able to connect
// to pgsql via postgres://localhost:5432
// This is achieved like this (unchecked)
/*
- run psql via `su postgres` 
- create user budden with superuser login
- create database budden
*/

import (
	"fmt";	"log";	"os"; "time"
//	neturl "net/url"
//	"reflect"
//	"regexp"
//	"strings"
//	"time"
 "github.com/jmoiron/sqlx"
	"github.com/lib/pq")

func main() {
	url := "postgresql://localhost:5432"
	db, err := sqlx.Open("postgres", url)
	if err != nil { 
		log.Printf("Unable to connect to Postgresql, error is %#v", err)
		os.Exit(1) }
	m := map[string]interface{}{"name" : `",sql 'inject?`}
	res, err1 := db.NamedExec(`insert into users values (:name)`,
		m)
	//xt := reflect.TypeOf(err1).Kind()
	if err1 != nil {
		switch e := interface{}(err1).(type)	{
		case *pq.Error:
			if e.Code == "23505" {
				fmt.Printf("Duplicate key in %s",e.Constraint)
			} else {
				fmt.Printf("Error inserting: %#v\n",err1) }
		default: 
			fmt.Printf("Error insertiing: %#v\n",err1) }
	} else {
		fmt.Printf("Inserted %#v\n",res)	}
	res1, err2 := db.Query(`select current_timestamp + interval '1' day`)
	if err2 != nil {
		fmt.Printf("Wow!"); os.Exit(1)	}
	if !res1.Next() {
		fmt.Printf("No rows here. Why?"); os.Exit(1) }
	var magic time.Time
	res1.Scan(&magic)
	// посмотреть, во что превратится timestamp
	fmt.Printf("Expiry at %#v\n", magic);

	err = db.Close()
	if err != nil {
		log.Printf("Error closing db")
		os.Exit(1) }
	return }