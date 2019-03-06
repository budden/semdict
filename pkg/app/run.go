package app

// To run this one you need that current user is able to connect
// to pgsql via postgres://localhost:5432
// This is achieved like this (unchecked)
/*
- run psql via `su postgres`
- create user budden with superuser login
- create database budden
*/

import (
	"log"
)

// Run is an application entry point
func Run(commandLineArgs []string) {
	log.Println("Nothing to run yet")
	return
}
