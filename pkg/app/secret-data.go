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

// DefaultConfigFileName это имя файла конфигурации по умолчанию
const DefaultConfigFileName = "semdict.config.json"

// ConfigFileName имя файла конфигурации секретных данных
var ConfigFileName *string

// TemplateBaseDir каталог, в котором находится каталог шаблонов и статических файлов
// FIXME Переименование. Это не "активы", потому что статические файлы не являются активами, ими являются только шаблоны.
// но некий "корневой каталог".
var TemplateBaseDir *string

// SaveSecretConfigDataExample вызывается из набора тестов.
// В качестве побочного эффекта создается файл semdict.config.json.example
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

// LoadSecretConfigData считывает файл конфигурации и инициализирует глобальный файл SecretConfigData
func LoadSecretConfigData(configFileName *string) (err error) {
	shared.SecretConfigData = &shared.SecretConfigDataT{}
	scd := shared.SecretConfigData
	fn := *configFileName
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		fmt.Printf("Не найден файл конфигурации %s. Создайте его, скопировав из %s.example\n",
			fn, fn)
		os.Exit(shared.ExitCodeNoConfigFileFound)
	}
	var bytes []byte
	bytes, err = ioutil.ReadFile(fn)
	if err != nil {
		fmt.Printf("Невозможно прочитать конфигурацию %s\n", fn)
		return
	}
	dec := json.NewDecoder(strings.NewReader(string(bytes)))
	dec.DisallowUnknownFields()
	err = dec.Decode(scd)
	if err != nil {
		fmt.Printf("Ошибка чтения файла конфигурации %s: %#v\n", fn, err)
		return
	}
	return
}

// ValidateConfiguration проверяет "секретные данные конфигурации".
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
					err = errors.Wrapf(err, "Не удалось установить файл сертификата TLS")
				} else if !probe {
					err = errors.Errorf("TLSCertFile «%s» не найден", cert)
				}
				probe, err = shared.IsFileExist(key)
				if err != nil {
					err = errors.Wrapf(err, "Не удалось установить файл ключа TLS")
				} else if !probe {
					err = errors.Errorf("TLSKeyFile «%s» не найден", key)
				}
			} else {
				err = errors.New("В автономном режиме вы должны предоставить либо оба файла TLSCertFile и TLSKeyFile, либо ни один из них.")
			}
		}
	case 1:
		{
			if cert != "" || key != "" {
				err = errors.New("В прокси-сервере не поставляйте TLSCertFile и TLSKeyFile")
			}
		}
	default:
		{
			err = errors.New("UnderAProxy должен быть равен 0 (автономный режим) или 1 (за прокси-сервером).")
		}
	}
	return
}
