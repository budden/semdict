package shared

// SecretConfigDataT specifies the fields of secret-data.config.json
// That file contains the data which is secret and site-specific so it can't be stored to git
type SecretConfigDataT struct {
	Comment             string
	RecieverEMail       string
	SMTPServer          string
	SMTPUser            string
	SMTPPassword        string
	SenderEMail         string
	PostgresqlServerURL string
}

// SecretConfigData is an in-memory copy of a secret-data.config.json configuration file
var SecretConfigData SecretConfigDataT
