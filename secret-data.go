package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	// "github.com/flynn/json5"
)

// SecretConfigDataStruct specifies the contents of secret-data.config.json
// That file contains the data which is secret and site-specific so it can't be stored to git
type SecretConfigDataStruct struct {
	Comment       string
	RecieverEMail string
	SMTPServer    string
	SMTPUser      string
	SMTPPassword  string
	SenderEMail   string
}

// SecretConfigData is a configuration loaded from the secret-data-config.json
var SecretConfigData SecretConfigDataStruct

// SaveToFile is only used to create a sample config file
func (sds *SecretConfigDataStruct) SaveToFile(filename string) (err error) {
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
	sds := SecretConfigDataStruct{
		Comment:       "Example config file. Copy this one to the secret-data.config.json and edit",
		SenderEMail:   "den@example.net",
		RecieverEMail: "world@example.net",
		SMTPServer:    "smtp.example.net",
		SMTPUser:      "Кирилл",
		SMTPPassword:  "bla-bla-bla"}
	err := sds.SaveToFile(ConfigFileName + ".example")
	if err != nil {
		panic(err)
	}
}

func loadSecretConfigData() (err error) {
	sds := &SecretConfigData
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
