package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/budden/semdict/pkg/shared"
	// "github.com/flynn/json5"
)

func saveSecretDataConfigTToFile(sds *shared.SecretConfigDataT, filename string) (err error) {
	var text []byte
	text, err = json.MarshalIndent(sds, "", " ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(filename, text, 0600)
	return
}

// ConfigFileName is a secret data configuration file name
const ConfigFileName = "secret-data.config.json"

// SaveSecretConfigDataExample is called from the test suite.
// As a side effect, secret-data.config.json.example is created
func SaveSecretConfigDataExample(fileName string) (sds *shared.SecretConfigDataT, err error) {
	sds = &shared.SecretConfigDataT{
		Comment: "Example config file. Copy this one to the secret-data.config.json and edit. TLSCertFile and TLSKeyFile" +
			" are file names of files in PEM format",
		SiteRoot:            "localhost",
		WebServerPort:       "8085",
		SenderEMail:         "den@example.net",
		RecieverEMail:       "world@example.net",
		SMTPServer:          "smtp.example.net",
		SMTPUser:            "Кирилл",
		SMTPPassword:        "bla-bla-bla",
		TLSCertFile:         "example.pem",
		TLSKeyFile:          "example.key",
		PostgresqlServerURL: "postgresql://localhost:5432"}
	err = saveSecretDataConfigTToFile(sds, fileName)
	return
}

// LoadSecretConfigData reads the config file and inititalizes a SecretConfigData global
func LoadSecretConfigData(configFileName string) (err error) {
	sds := &shared.SecretConfigData
	fn := configFileName
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		fmt.Printf("No config file %s found. Create one by copying from %s.example\n",
			fn, fn)
		os.Exit(shared.ExitCodeNoConfigFileFound)
	}
	var bytes []byte
	bytes, err = ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("Unable to read config %s\n", fn)
		return
	}
	dec := json.NewDecoder(strings.NewReader(string(bytes)))
	dec.DisallowUnknownFields()
	err = dec.Decode(sds)
	if err != nil {
		fmt.Printf("Error reading config file %s: %#v\n", fn, err)
		return
	}
	return
}
