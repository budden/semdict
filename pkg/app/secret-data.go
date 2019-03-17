package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/budden/semdict/pkg/apperror"
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

// for development
func saveSecretConfigDataExample() {
	sds := shared.SecretConfigDataT{
		Comment:             "Example config file. Copy this one to the secret-data.config.json and edit",
		SenderEMail:         "den@example.net",
		RecieverEMail:       "world@example.net",
		SMTPServer:          "smtp.example.net",
		SMTPUser:            "Кирилл",
		SMTPPassword:        "bla-bla-bla",
		PostgresqlServerURL: "postgresql://localhost:5432"}
	err := saveSecretDataConfigTToFile(&sds, ConfigFileName+".example")
	apperror.ExitAppIf(err, 5, "Failed to save secret config data example")
}

// loadSecretConfigData reads the config file and inititalizes a SecretConfigData global
func loadSecretConfigData() (err error) {
	sds := &shared.SecretConfigData
	fn := ConfigFileName
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		fmt.Printf("No config file %s found. Create one by copying from %s.example\n",
			fn, fn)
		return
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
	fmt.Printf("playWithSecretConfigData returned %#v\n", sds)
	return
}
