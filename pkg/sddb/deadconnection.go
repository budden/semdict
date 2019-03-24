package sddb

import (
	"log"
	"runtime/debug"

	"github.com/budden/semdict/pkg/apperror"
)

// SetConnectionDead is used to announce that db connection is broken. We do that when something
// wrong happened like error in the rollback.
func SetConnectionDead(conn *ConnectionType) {
	conn.IsDead = true
}

// IsConnectionDead should be called before every use of the db connection to check if it is not dead
// If you find it dead, return status 500
func IsConnectionDead(conn *ConnectionType) bool {
	return conn.IsDead
}

// FatalDatabaseErrorHandlerType is a type for FatalDatabaseErrorHandler
type FatalDatabaseErrorHandlerType func(err error, conn *ConnectionType, format string, args ...interface{})

func initialFatalDatabaseErrorHandler(err error, conn *ConnectionType, format string, args ...interface{}) {
	if err != nil {
		log.Printf("Early call to FatalDatabaseErrorIf: "+format, args...)
		debug.PrintStack()
		log.Fatal(err)
	}
}

// FatalDatabaseErrorHandler is used by FatalDatabaseError function
var FatalDatabaseErrorHandler FatalDatabaseErrorHandlerType = initialFatalDatabaseErrorHandler

// FatalDatabaseErrorIf function is called from request handler to inform that we can't any more work
// with this database and have to shut down. If this happens, we first declare that database as dead.
// Next, we initiate a "graceful shutdown". Last, we arrange to return status 500 to the client.
// But this module knows nothing about http, so we only have a stub here. Actual handler is stored
// in the FatalDatabaseErrorHandler variable and is set up later to app.actualFatalDatabaseErrorHandler
// FatalDataBaseErrorIf is believed to be thread safe
func FatalDatabaseErrorIf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	c := sdUsersDb
	FatalDatabaseErrorHandler(err, c, format, args...)
	return
}

// CheckDbAlive is to be called in page handlers before every db interaction
func CheckDbAlive() {
	c := sdUsersDb
	if IsConnectionDead(c) {
		apperror.Panic500If(apperror.ErrDummy, "Internal error")
	}
}
