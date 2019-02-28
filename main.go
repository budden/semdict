package main

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
	_ "github.com/lib/pq"
)

func main() {
	log.Println("hello")
	fmt.Println("hi")
	url := "http://localhost:5432"
	db, err := sqlx.Open("postgres", url)
	if err != nil { 
		log.Printf("Unable to connect to Postgresql, error is %#x", err)
		os.Exit(1) 
		}
	err = db.Close()
	if err != nil {
		log.Printf("Error closing db")
		os.Exit(1)
	}
	return
}