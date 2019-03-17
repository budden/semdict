package app

import (
	"github.com/budden/semdict/pkg/database"
)

func init() {
	database.FatalDatabaseErrorHandler = actualFatalDatabaseErrorHandler
}
