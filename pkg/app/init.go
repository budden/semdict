package app

import (
	"github.com/budden/a/pkg/database"
)

func init() {
	database.FatalDatabaseErrorHandler = actualFatalDatabaseErrorHandler
}
