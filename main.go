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
	"fmt" )

func main() {
	playWithDb()
	playWithPanic()
	playWithNonce(16)
	playWithHashAndSalt()
	return }


	func playWithPanic() {
		unwind := func() {
			if r := recover(); r != nil {
				fmt.Printf("recover %#v\n",r)
				//panic(r)
			}
		}
		defer unwind()
		panic("It's a panic")	
	}


