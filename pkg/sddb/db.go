// Пакет sddb содержит элементы для подключения к базе данных
package sddb

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	//	neturl "net/url"
	//	"reflect"
	//	"regexp"
	//	"strings"
	//	"time"
	"database/sql"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/shutdown"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

// ConnectionType инкапсулирует соединение с некоторыми другими незначительными вещами, 
// такими как нерабочее состояние пула соединений.
// Для любого экземпляра БД должен существовать только один экземпляр ConnectionType
type ConnectionType struct {
	Db *sqlx.DB

	// Мы держим мьютекс для всех записей в базу данных.
	// Таким образом мы контролируем количество одновременных транзакций в базе данных.
	// Пока у нас только один экземпляр службы, 
	// поэтому одновременных записей в базу данных будет не более одной.
	// Если у нас несколько экземпляров, то одновременных записей будет не больше, чем экземпляров сервиса.
	Mutex *sync.Mutex

	IsDead bool
}

// TransactionType - это транзакция и ссылка на ConnectionType, 
// которая нам нужна из-за невозможности получить db из транзакции.
type TransactionType struct {
	Conn *ConnectionType
	Tx   *sqlx.Tx
}

// sdUsersDb содержит, после вызова OpenSdUsersDb, соединение с sdusers_db
var sdUsersDb *ConnectionType

// OpenSdUsersDb открывает заданное имя базы данных
func OpenSdUsersDb(dbName string) {
	if sdUsersDb != nil {
		log.Fatal("Попытка повторного открытия базы данных SDUsers")
	}
	url := shared.SecretConfigData.PostgresqlServerURL + "/" + dbName
	sdUsersDb = OpenDb(url, dbName, true)
	return
}

// CloseSdUsersDb закрывает базу данных
func CloseSdUsersDb() error {
	return sdUsersDb.Db.Close()
}

// PlayWithDb используется для ручного тестирования функциональности БД
func PlayWithDb() {
	justAQuery := func(query string) {
		rows, err := sdUsersDb.Db.Query(query)
		apperror.LogicalPanicIf(err, "Query error, query: «%s», error: %#v", query, err)
		rows.Close()
	}

	justAQuery(`drop table if exists budden_a`)
	justAQuery(`create table budden_a (name text)`)
	justAQuery(`create unique index budden_a_name on budden_a (name)`)

	m := map[string]interface{}{"name": `",sql 'inject?`}
	for i := 0; i < 2; i++ {
		var res sql.Result
		res, err1 := sdUsersDb.Db.NamedExec(`insert into budden_a values (:name)`,
			m)
		//xt := reflect.TypeOf(err1).Kind()
		if err1 != nil {
			/* switch e := interface{}(err1).(type) {
			case *pgx.Error:
				if e.Code == "23505" {
					fmt.Printf("Duplicate key in %s", e.Constraint)
				} else {
					fmt.Printf("Error inserting: %#v\n", err1)
				}
			default:
				fmt.Printf("Error insertiing: %#v\n", err1)
			} */
			fmt.Printf("не удалось вставить: %#v\n", err1)
		} else {
			fmt.Printf("Вставлено %#v\n", res)
		}
	}
	genExpiryDate(sdUsersDb.Db)
}

const maxOpenConns = 4
const maxIdleConns = 4
const connMaxLifetime = 10 * time.Second

// CommitIfActive фиксирует транзакцию, если она всё ещё активна.
func CommitIfActive(trans *TransactionType) (err error) {
	err = trans.Tx.Commit()
	if err == sql.ErrTxDone {
		err = nil
	}
	return
}

// RollbackIfActive откатывает транзакцию, если она всё ещё активна.
// Отложите этот вопрос, если вы открываете какую-либо сделку
// Если откат не удался, приложение будет завершено изящно
// Если уже возникла паника из-за ошибки, не связанной с приложением (произошло что-то неожиданное), 
// будет продолжать делать то же самое, просто регистрируя сбой для отката.
func RollbackIfActive(trans *TransactionType) {
	err := trans.Tx.Rollback()
	if err == nil || err == sql.ErrTxDone {
		return
	}
	preExistingPanic := recover()
	if preExistingPanic == nil {
		FatalDatabaseErrorIf(apperror.ErrDummy, "Не удалось выполнить откат транзакции")
	} else if ae, ok := preExistingPanic.(apperror.Exception500); ok {
		FatalDatabaseErrorIf(apperror.ErrDummy,
			"Не удалось откатить транзакцию с ожиданием Exception500: %#v",
			ae)
	} else {
		debug.PrintStack()
		log.Fatalf("Не удалось откатить транзакцию во время паники с %#v", preExistingPanic)
	}
}

// OpenDb получает соединение с db. Соединения объединяются, остерегайтесь.
// logFriendlyName - для случая, когда url содержит пароли
func OpenDb(url, logFriendlyName string, withMutex bool) *ConnectionType {
	var err error
	var db *sqlx.DB
	db, err = sqlx.Open("pgx", url)
	apperror.ExitAppIf(err, 6, "Не удалось открыть «%s» database", logFriendlyName)
	err = db.Ping()
	apperror.ExitAppIf(err, 7, "Не удалось выполнить пинг «%s» database", logFriendlyName)
	// http://go-database-sql.org/connection-pool.t.html
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)
	closer1 := func() {
		err := db.Close()
		if err != nil {
			// мы не знаем, можно ли записывать stdout, но мы уже находимся в goroutine
			log.Printf("Error closing database «%s»: %#v\n", logFriendlyName, err)
		} else {
			log.Printf("Closed database «%s»\n", logFriendlyName)
		}
	}
	closer := func() { go closer1() }
	var mtx *sync.Mutex
	if withMutex {
		var mtx2 sync.Mutex
		mtx = &mtx2
	}
	shutdown.Actions = append(shutdown.Actions, closer)
	result := ConnectionType{Db: db, Mutex: mtx}
	return &result
}

func genExpiryDate(db *sqlx.DB) {
	res1, err2 := db.Query(`select current_timestamp + interval '10' minutes`)
	if err2 != nil {
		fmt.Printf("Вау!")
		os.Exit(1)
	}
	if !res1.Next() {
		fmt.Printf("Здесь нет рядов. почему?")
		os.Exit(1)
	}
	var magic time.Time
	res1.Scan(&magic)
	fmt.Printf("Срок годности по %s\n", magic.Format("2006-01-02 15:04 -0700"))
}

// WithTransaction открывает транзакцию в sdusers_db, затем запускает тело
// Затем, если ошибки нет, и транзакция всё ещё активна, фиксируем транзакцию и возвращаем ошибку фиксации
// Если во время выполнения тела произошла ошибка или паника, пытается откатить транзакцию,
// см. RollbackIfActive
func WithTransaction(body func(tx *TransactionType) (err error)) (err error) {

	conn := sdUsersDb
	CheckDbAlive()

	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}

	var tx *sqlx.Tx
	CheckDbAlive()
	tx, err = conn.Db.Beginx()
	trans := TransactionType{Conn: conn, Tx: tx}
	FatalDatabaseErrorIf(err, "Невозможно начать транзакцию")
	defer func() { RollbackIfActive(&trans) }()
	CheckDbAlive()
	_, err = tx.Exec(`set transaction isolation level repeatable read`)
	FatalDatabaseErrorIf(err, "Невозможно установить уровень изоляции транзакции")
	CheckDbAlive()
	err = body(&trans)
	if err == nil {
		CheckDbAlive()
		err = CommitIfActive(&trans)
	}
	return
}

// NamedUpdateQuery - это запрос в sdUsersDb, который обновляет базу данных. 
// Поэтому мы держим наш мьютекс для обеспечения сериализации всех записей в области видимости экземпляров.
func NamedUpdateQuery(sql string, params interface{}) (res *sqlx.Rows, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	CheckDbAlive()
	res, err = conn.Db.NamedQuery(sql, params)
	return
}

// NamedExec похож на sqlx.NamedExec и также держит мьютекс.
// Используйте его всякий раз, когда выполняемый запрос может обновить базу данных
func NamedExec(sql string, params interface{}) (res sql.Result, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	mutex := conn.Mutex
	if mutex != nil {
		mutex.Lock()
		defer mutex.Unlock()
	}
	CheckDbAlive()
	res, err = conn.Db.NamedExec(sql, params)
	return
}

// NamedReadQuery предназначен для запросов, которые доступны только для чтения.
// Мы вводим его, когда инкапсулируем sql(x) соединение sdusers_db в модуль sddb.
func NamedReadQuery(sql string, params interface{}) (res *sqlx.Rows, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	res, err = conn.Db.NamedQuery(sql, params)
	return
}

// ReadQuery предназначен для запросов, которые доступны только для чтения.
// Мы вводим его, когда инкапсулируем sql(x) соединение sdusers_db в модуль sddb.
func ReadQuery(sql string) (res *sqlx.Rows, err error) {
	conn := sdUsersDb
	CheckDbAlive()
	res, err = conn.Db.NamedQuery(sql, map[string]interface{}{})
	return
}

// CloseRows предназначены для использования с отсрочкой для закрытия строк в случае, если итерация строк идет неправильно
func CloseRows(r *sqlx.Rows) func() {
	return func() {
		err := r.Close()
		FatalDatabaseErrorIf(err, "Не удалось закрыть строки, «%s»")
	}
}

// UncoalesceInt64 converts 0 to an invalid sql.NullInt64
func UncoalesceInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{Int64: 0, Valid: false}
	} else {
		return sql.NullInt64{Int64: i, Valid: true}
	}
}
