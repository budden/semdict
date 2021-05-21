// Пакет unsorted предназначен для новых вещей, которые, в общем, не отсортированы
package unsorted

import (
	"log"
	"os"
)

// OpenDevNullForWrite открывает /dev/null для записи
func OpenDevNullForWrite() *os.File {
	f, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Не удалось открыть DevNull")
	}
	return f
}
