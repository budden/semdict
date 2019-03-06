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

	"github.com/budden/a/pkg/my"
)

const actuallySendEmailP = false

// Play runs a set of exercises/demos
func Play(commandLineArgs []string) {
	playWithDb()
	playWithPanic()
	playWithNonce(16)
	playWithHashAndSalt()
	/// Uncomment next line to create secret-data.config.json.example
	//saveSecretConfigDataExample()
	loadSecretConfigData()
	if actuallySendEmailP {
		my.PlayWithEmail()
	} else {
		fmt.Println("Bypassing sending E-mail due to ActuallySendEmailP == false")
	}
	playWithServer()
}

func playWithNonce(length uint8) {
	fmt.Println("FIXME: test that those numbers are sufficiently random!")
	for i := 0; i < 5; i++ {
		str := my.GenNonce(length)
		fmt.Println("Nonce1:", str)
	}
}

func playWithHashAndSalt() {
	for i := 0; i < 2; i++ {
		password := "kvack"
		hash, salt := my.HashAndSaltPassword(password)
		fmt.Printf("playWithHashAndSalt: hash=%s, salt=%s\n", hash, salt)
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
