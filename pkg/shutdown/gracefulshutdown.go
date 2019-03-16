package shutdown

import (
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
	log.Println("Signal!")
	// there are two timeout guards, and both cause the app to exit.
	// First one is started before starting cleanup actions and causes
	// exit with ExitCodeGracefulShutdownTimeout
	// Second one is started after all cleanup actions. We could use waitgroup,
	// but no need here as both guards call exit (hopefully exit is thread safe)
	if Timeout != 0 {
		go timeoutGuard(ExitCodeGracefulShutdownTimeout)
	}
	cleanupAllTheThings()
	time.Sleep(Timeout)
	timeoutGuard(1)
}

func timeoutGuard(exitCode int) {
	time.Sleep(Timeout)
	os.Exit(exitCode)
}

func cleanupAllTheThings() {
	for _, fn := range Actions {
		fn()
	}
}

// Implement os.Signal
type artificialSignal struct {
	Mark int
}

func (as *artificialSignal) String() string { return "" }
func (as *artificialSignal) Signal()        { return }

// InitiateGracefulShutdown starts the shutdown process and returns without delay
// Intended to be called by the user's code (including request handlers), in contrast with
// signal-based shutdown initiated from the outer world
func InitiateGracefulShutdown() {
	as := artificialSignal{}
	*Sigs <- &as
}
