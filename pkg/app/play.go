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
	"fmt"

	"github.com/budden/semdict/pkg/database"
	"github.com/budden/semdict/pkg/shutdown"
	"github.com/budden/semdict/pkg/user"
)

// Play runs a set of exercises/demos
func Play(commandLineArgs []string) {
	shutdown.RunSignalListener()
	/// Uncomment next line to create secret-data.config.json.example
	//saveSecretConfigDataExample()

	loadSecretConfigData()
	database.OpenSDUsersDb()
	/* playWithPanic()
	playWithNonce(16)
	playWithSaltAndHash()
	user.PlayWithEmail() */
	playWithServer()
}

func playWithNonce(length uint8) {
	fmt.Println("FIXME: test that those numbers are sufficiently random!")
	for i := 0; i < 5; i++ {
		str := user.GenNonce(length)
		fmt.Println("Nonce1:", str)
	}
}

func playWithPanic() {
	unwind := func() {
		if r := recover(); r != nil {
			fmt.Printf("recover %#v\n", r)
			//panic(r)
		}
	}
	defer unwind()
	panic("It's a panic")
}
