package main

import (
	"strings"; "fmt"; "os"
	"encoding/json"
	"io/ioutil"
	// "github.com/flynn/json5"
)

type SecretConfigDataStruct struct {
	Comment       string
	RecieverEMail string
	SMTPServer    string
	SMTPUser      string
	SMTPPassword  string
	SenderEMail   string }

var SecretConfigData SecretConfigDataStruct

func (sds *SecretConfigDataStruct) SaveToFile(filename string) (err error) {
	var text []byte
	text, err = json.MarshalIndent(sds,""," ")
	if err != nil { return	}	
	err = ioutil.WriteFile(filename, text, 0600)
	return }

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
	if err != nil {	panic(err)	}}

func loadSecretConfigData() (err error) {
	sds := &SecretConfigData
	fn := ConfigFileName
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		fmt.Printf("No config file %s found. Create one by copying from %s.example\n",
			fn, fn)
		return	}
	var bytes []byte
	bytes, err = ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("Unable to read config %s\n", fn)
		return	}
	dec := json.NewDecoder(strings.NewReader(string(bytes)))
	dec.DisallowUnknownFields() 
	err = dec.Decode(sds)
	if err != nil {
		fmt.Printf("Error reading config file %s: %#v\n", fn, err)
		return	}
	fmt.Printf("playWithSecretConfigData returned %#v\n", sds)
	return }