package app

import (
	"fmt"
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
		// we can't drop db because server holds the connections.
		t.Run("teardownDatabase", teardownDatabase)
	}()

	time.Sleep(1 * time.Second)

	setupServer()
	defer func() {
		teardownServer(t)
		time.Sleep(1 * time.Second)
	}()

	// FIXME there must be a better way to wait for server to start
	time.Sleep(1 * time.Second)

	if !assert.Truef(t,
		reportIfErr(setupClient()),
		"setupClient failed") {
		return
	}

	/* defer func() {
		time.Sleep(1 * time.Second)
		assert.Truef(t,
			reportIfErr(teardownClient()),
			"teardownClient failed")
	}() */

	fmt.Println("Ughhhhhhh!")
	// t.Run("testDataImportCSVAlternativeDelimiter", testDataImportCSVAlternativeDelimiter)
}

// Run runs an app
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
	apiURL := "http://localhost:8081/api/import/csv"

	fd := formDataType{
		"importCSVTableName":      strings.NewReader(tableName),
		"importCSVFieldDelimiter": strings.NewReader(fieldDelimiter),
		"importCSVFile":           mustOpen(fileName),
	}

	var req *http.Request
	req, err = preparePostRequest(apiURL, fd)
	// Now that you have a form, you can submit it to your handler.
	if err != nil {
		return
	}

	// Submit the request
	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		return
	}

	// Check the response
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
