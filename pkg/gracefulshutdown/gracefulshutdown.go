package gracefulshutdown

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// https://android.wekeepcoding.com/article/10973673/How+to+delete+a+file+using+golang+on+program+exit%3F

// Sigs holds a channel accepting os signals
var Sigs *chan os.Signal

// Timeout is used to handle "graceful shutdown". When it passes,
// os.Exit(ExitCodeGracefulShutdownTimeout) is called
// zero value means "no timeout"
var Timeout = time.Second * 1

// ArrayOfParameterlessFunctions is a type for Actions variable
type ArrayOfParameterlessFunctions = []func()

// Actions - see RunSignalListener
var Actions = ArrayOfParameterlessFunctions{}

// ExitCodeGracefulShutdownTimeout exit code means that graceful shutdown timed out
const ExitCodeGracefulShutdownTimeout = 107

// RunSignalListener creates a listener to catch SIGINT and SIGTERM. When
// signal arrives, functions from Actions are run
// sequentially. Time spent is controlled by the Timeout
func RunSignalListener() {
	sigs := make(chan os.Signal, 1)
	Sigs = &sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go signalListener()
}

func signalListener() {
	<-*Sigs
	fmt.Println("Signal!")
	log.Println("Signal!")
	if Timeout != 0 {
		go timeoutGuard()
	}
	cleanupAllTheThings()
	os.Exit(0)
}

func timeoutGuard() {
	time.Sleep(Timeout)
	os.Exit(ExitCodeGracefulShutdownTimeout)
}

func cleanupAllTheThings() {
	for _, fn := range Actions {
		fn()
	}
}
