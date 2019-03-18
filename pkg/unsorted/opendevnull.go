// Package unsorted is for new things which are, well, unsorted
package unsorted

import (
	"log"
	"os"
)

// OpenDevNullForWrite opens /dev/null for write
func OpenDevNullForWrite() *os.File {
	f, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open DevNull")
	}
	return f
}
