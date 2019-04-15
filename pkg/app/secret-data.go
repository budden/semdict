package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/budden/semdict/pkg/shared"
	// "github.com/flynn/json5"
)

func saveSecretDataConfigTToFile(scd *shared.SecretConfigDataT, filename *string) (err error) {
	var text []byte
	text, err = json.MarshalIndent(scd, "", " ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(*filename, text, 0600)
	return
}

// DefaultConfigFileName is a default config file name
const DefaultConfigFileName = "semdict.config.json"

// ConfigFileName is a secret data configuration file name
var ConfigFileName *string

// TemplateBaseDir is a directory in which template and static directory is located
// FIXME rename. It is not "assets", because static files are not assets, only templates are.
// but some "root directory".
var TemplateBaseDir *string

// SaveSecretConfigDataExample is called from the test suite.
// As a side effect, semdict.config.json.example is created
func SaveSecretConfigDataExample(fileName *string) (scd *shared.SecretConfigDataT, err error) {
	scd = &shared.SecretConfigDataT{
		Comment:                     shared.SecretConfigDataTComment,
		SiteRoot:                    "localhost",
		UnderAProxy:                 0,
		ServerPort:                  "8085",
		SMTPServer:                  "smtp.example.net",
		SMTPUser:                    "Кирилл",
		SMTPPassword:                "bla-bla-bla",
		SenderEMail:                 "den@example.net",
		PostgresqlServerURL:         "postgresql://localhost:5432",
		TLSCertFile:                 "example.pem",
		TLSKeyFile:                  "example.key",
		UserAlwaysLoggedIn:          0,
		HideGinStartupDebugMessages: 1,
		GinDebugMode:                0}
	err = saveSecretDataConfigTToFile(scd, fileName)
	return
}

// LoadSecretConfigData reads the config file and inititalizes a SecretConfigData global
func LoadSecretConfigData(configFileName *string) (err error) {
	shared.SecretConfigData = &shared.SecretConfigDataT{}
	scd := shared.SecretConfigData
	fn := *configFileName
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
	err = dec.Decode(scd)
	if err != nil {
		fmt.Printf("Error reading config file %s: %#v\n", fn, err)
		return
	}
	return
}

// ValidateConfiguration validates a "secret config data"
func ValidateConfiguration() (err error) {
	scd := shared.SecretConfigData
	cert := scd.TLSCertFile
	key := scd.TLSKeyFile
	switch scd.UnderAProxy {
	case 0:
		{
			if cert == "" && key == "" {
				// ok
			} else if cert != "" && key != "" {
				// both must exist
				var probe bool
				probe, err = shared.IsFileExist(cert)
				if err != nil {
					err = errors.Wrapf(err, "Failed to stat a TLS cert file")
				} else if !probe {
					err = errors.Errorf("TLSCertFile «%s» not found", cert)
				}
				probe, err = shared.IsFileExist(key)
				if err != nil {
					err = errors.Wrapf(err, "Failed to stat a TLS key file")
				} else if !probe {
					err = errors.Errorf("TLSKeyFile «%s» not found", key)
				}
			} else {
				err = errors.New("In a standalone mode, you must supply either both TLSCertFile and TLSKeyFile, or none of them")
			}
		}
	case 1:
		{
			if cert != "" || key != "" {
				err = errors.New("Under a proxy, don't supply TLSCertFile and TLSKeyFile")
			}
		}
	default:
		{
			err = errors.New("UnderAProxy must be 0 (standalone mode) or 1 (behind a proxy)")
		}
	}
	return
}
