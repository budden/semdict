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
	"fmt"
	"log"
	"os"
//	neturl "net/url"
//	"reflect"
//	"regexp"
//	"strings"
//	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func main() {
	log.Println("hello")
	fmt.Println("hi")
	url := "postgresql://localhost:5432"
	db, err := sqlx.Open("postgres", url)
	if err != nil { 
		log.Printf("Unable to connect to Postgresql, error is %#v", err)
		os.Exit(1) 
		}
	m := map[string]interface{}{"id" : "5"}
	res, err1 := db.NamedExec(`insert into zz1 (id) values (:id) returning id`,
		m)
	//xt := reflect.TypeOf(err1).Kind()
	if err1 != nil {
		switch e := interface{}(err1).(type)	{
		case *pq.Error:
			if e.Code == "23505" {
				fmt.Printf("Duplicate key in %s",e.Constraint)
			} else {
				fmt.Printf("Error inserting: %#v\n",err1)
			}
		default: 
			fmt.Printf("Error insertiing: %#v\n",err1)
	
		}
	} else {
		fmt.Printf("Inserted %#v\n",res)
	}
	err = db.Close()
	if err != nil {
		log.Printf("Error closing db")
		os.Exit(1)
	}
	return
}