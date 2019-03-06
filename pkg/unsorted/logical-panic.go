package unsorted

// Panic which should crash current goroutine only
func LogicalPanic(message string) {
	panic(message)
}
