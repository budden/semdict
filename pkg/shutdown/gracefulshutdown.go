package shutdown

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/budden/semdict/pkg/shared"
)

// https://android.wekeepcoding.com/article/10973673/How+to+delete+a+file+using+golang+on+program+exit%3F

// Sigs - канал, принимающий сигналы os
var Sigs *chan os.Signal

// Таймаут используется для обработки "льготного отключения". Когда он пройдёт,
// os.Exit(ExitCodeGracefulShutdownTimeout) вызовет нулевое значение, означающее "тайм-аут отсутствует"
//
var Timeout = time.Second * 1

// ArrayOfParameterlessFunctions является типом для переменной Actions
type ArrayOfParameterlessFunctions = []func()

// Actions - see RunSignalListener
var Actions = ArrayOfParameterlessFunctions{}

// RunSignalListener создаёт слушателя для перехвата SIGINT и SIGTERM.
// Когда поступает сигнал, последовательно выполняются функции из Actions.
// Затраченное время контролируется параметром Timeout
func RunSignalListener() {
	sigs := make(chan os.Signal, 1)
	Sigs = &sigs
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go signalListener()
}

func signalListener() {
	<-*Sigs
	log.Println("Signal!")
	// есть два таймаута, и оба приводят к завершению работы приложения.
	// Первый запускается до начала действий по очистке
	// и вызывает выход с ExitCodeGracefulShutdownTimeout
	// Второй запускается после всех действий по очистке. 
	// Мы могли бы использовать waitgroup, но здесь в этом нет необходимости,
	// так как оба охранника вызывают exit (надеюсь, exit безопасен для потока).
	if Timeout != 0 {
		go timeoutGuard(shared.ExitCodeGracefulShutdownTimeout)
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

// InitiateGracefulShutdown запускает процесс выключения и возвращается без задержки
// Предназначены для вызова пользовательским кодом (включая обработчики запросов),
// в отличие от выключения по сигналу, инициируемого из внешнего мира
func InitiateGracefulShutdown() {
	as := artificialSignal{}
	*Sigs <- &as
}
