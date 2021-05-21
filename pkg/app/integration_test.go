package app

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/shutdown"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	t.Run("setupDatabase", setupDatabase)
	if t.Failed() {
		return
	}
	defer func() {
		time.Sleep(1 * time.Second)
		// мы не можем бросить db, поскольку сервер удерживает соединения.
		t.Run("teardownDatabase", teardownDatabase)
	}()

	time.Sleep(1 * time.Second)

	setupServer()
	defer func() {
		teardownServer(t)
		time.Sleep(1 * time.Second)
	}()

	// FIXME должен быть лучший способ дождаться запуска сервера
	time.Sleep(1 * time.Second)

	if !assert.Truef(t,
		reportIfErr(setupClient()),
		"setupClient failed") {
		return
	}

	t.Run("getHomePage", getHomePage)

	// t.Run("testDataImportCSVAlternativeDelimiter", testDataImportCSVAlternativeDelimiter)
}

func getHomePage(t *testing.T) {
	// https://stackoverflow.com/a/38807963/9469533
	url := serviceURL + "/"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Не удалось получить %s, ошибка %#v", url, err)
		t.Fail()
		return
	}
	defer resp.Body.Close()
	responseData, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		log.Printf("Не удалось прочитать ответ от %s, ошибка %#v", url, err)
		t.Fail()
		return
	}

	responseString := string(responseData)

	if strings.Index(responseString, "Добро пожаловать в семантический словарь") < 0 {
		t.Fail()
	}
}

// Пуск запускает приложение
func runForTesting() {
	tbd := "../../"
	TemplateBaseDir = &tbd
	setSecretConfigDataForIntegrationTest()
	shutdown.RunSignalListener()
	sddb.OpenSdUsersDb(serverDatabase)
	playWithServer()
}

func setSecretConfigDataForIntegrationTest() {
	postgresqlServerURL := "postgresql://" + serverHost + ":" + serverPort
	shared.SecretConfigData = &shared.SecretConfigDataT{
		SiteRoot:            "localhost",
		UnderAProxy:         0,
		ServerPort:          "8085",
		SenderEMail:         "budden@example.net",
		SMTPServer:          "",
		SMTPUser:            "ignored",
		SMTPPassword:        "ignored",
		TLSCertFile:         "",
		TLSKeyFile:          "",
		PostgresqlServerURL: postgresqlServerURL}
}

func setupServer() {
	go runForTesting()
}

func teardownServer(t *testing.T) {
	err := sddb.CloseSdUsersDb()
	if err != nil {
		log.Println(err)
		t.Fail()
	}
}

func dataImportCSV(tableName, fieldDelimiter, fileName string) (err error) {
	var client *http.Client
	client = &http.Client{Timeout: time.Second * 10}
	apiURL := "http://localhost:????/api/import/csv"

	fd := formDataType{
		"importCSVTableName":      strings.NewReader(tableName),
		"importCSVFieldDelimiter": strings.NewReader(fieldDelimiter),
		"importCSVFile":           mustOpen(fileName),
	}

	var req *http.Request
	req, err = preparePostRequest(apiURL, fd)
	// Теперь, когда у вас есть форма, вы можете отправить её в обработчик.
	if err != nil {
		return
	}

	// Отправить заявку
	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		return
	}

	// Проверьте ответ
	if res.StatusCode != http.StatusOK {
		err = decodeErrorFromHTTPResponsesBody(res)
	}
	return
}

func testDataImportCSVAlternativeDelimiter(t *testing.T) {
	if !assert.True(t,
		reportIfErr(dataImportCSV("from_csv_alternative_delimiter",
			";",
			"../../data/import_csv/alternative-delimiter.csv"))) {
		return
	}
	if !assert.True(t,
		reportIfErr(
			errIfQueryResultMismatch(t,
				"select id, line from from_csv_alternative_delimiter order by id",
				`{"columns":["id","line"],"rows":[["1","line"]]}`))) {
		return
	}
}
