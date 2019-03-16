package database

import (
	"sync"

	"github.com/budden/a/pkg/apperror"
	"github.com/jmoiron/sqlx"
)

// DeadConnections is only published for introspection. Don't write here!
var DeadConnections map[*sqlx.DB]bool
var deadConnectionsMutex sync.Mutex

// SetConnectionDead is used to announce that db connection is broken. We do that when something
// wrong happened like error in the rollback.
func SetConnectionDead(db *sqlx.DB) {
	deadConnectionsMutex.Lock()
	defer deadConnectionsMutex.Unlock()
	DeadConnections[db] = true
}

// IsConnectionDead should be called before every use of the db connection to check if it is not dead
// If you find it dead, return status 500
func IsConnectionDead(db *sqlx.DB) bool {
	deadConnectionsMutex.Lock()
	defer deadConnectionsMutex.Unlock()
	value, ok := DeadConnections[db]
	return ok && value
}

// FatalDatabaseErrorHandlerType is a type for FatalDatabaseErrorHandler
type FatalDatabaseErrorHandlerType func(err error, db *sqlx.DB, format string, args ...interface{})

func initialFatalDatabaseErrorHandler(err error, db *sqlx.DB, format string, args ...interface{}) {
	apperror.ExitAppIf(err, "Early call to FatalDatabaseErrorIf: "+format, args...)
}

// FatalDatabaseErrorHandler is used by FatalDatabaseError function
var FatalDatabaseErrorHandler FatalDatabaseErrorHandlerType = initialFatalDatabaseErrorHandler

// FatalDatabaseError function is called from request handler to inform that we can't any more work
// with this database and have to shut down. If this happens, we first declare that database as dead.
// Next, we initiate a "graceful shutdown". Last, we arrange to return status 500 to the client.
// But this module knows nothing about http, so we only have a stub here. Actual handler is stored
// in the FatalDatabaseErrorHandler variable and is set up later
func FatalDatabaseError(err error, db *sqlx.DB, format string, args ...interface{}) {
	if err == nil {
		return
	}
	FatalDatabaseErrorHandler(err, db, format, args...)
	return
}

// CheckDbAlive is to be called in page handlers before every db interaction
func CheckDbAlive(db *sqlx.DB) {
	if IsConnectionDead(db) {
		apperror.Panic500If(apperror.ErrDummy, "Internal error")
	}
}
