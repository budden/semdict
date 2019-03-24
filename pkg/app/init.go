package app

import (
	"github.com/budden/semdict/pkg/sddb"
)

func init() {
	sddb.FatalDatabaseErrorHandler = actualFatalDatabaseErrorHandler
}
