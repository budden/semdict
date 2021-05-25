package app

// Для его запуска необходимо, чтобы текущий пользователь мог подключиться
// в pgsql через postgres://localhost:5432
// Это достигается следующим образом (не отмечено)
/*
- запустите psql через `su postgres`
- создать пользователя budden с логином суперпользователя
- создать базу данных буден
*/

import (
	"fmt"

	"github.com/budden/semdict/pkg/shared"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shutdown"
	"github.com/budden/semdict/pkg/user"
)

// Пуск запускает приложение
func Run(commandLineArgs []string) {
	shutdown.RunSignalListener()
	err := LoadSecretConfigData(ConfigFileName)
	apperror.ExitAppIf(err,
		shared.ExitCodeBadConfigFile,
		"Не удалось загрузить конфигурацию, ошибка «%s»",
		err)
	err = ValidateConfiguration()
	apperror.ExitAppIf(err,
		shared.ExitCodeBadConfigFile,
		"Неверная конфигурация, ошибка «%s»",
		err)
	sddb.OpenSdUsersDb("sduser_db")
	/* playWithPanic()
	playWithNonce(16)
	playWithSaltAndHash() */
	playWithServer()
}

func playWithNonce(length uint8) {
	fmt.Println("FIXME: проверить, что эти числа достаточно случайны!")
	for i := 0; i < 5; i++ {
		str := user.GenNonce(length)
		fmt.Println("Nonce1:", str)
	}
}

func playWithPanic() {
	unwind := func() {
		if r := recover(); r != nil {
			fmt.Printf("recover %#v\n", r)
			//panic(r)
		}
	}
	defer unwind()
	panic("Это паника")
}
