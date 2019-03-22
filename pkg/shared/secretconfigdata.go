package shared

// SecretConfigDataT specifies the fields of semdict.config.json
// That file contains the data which is secret and site-specific so it can't be stored to git
type SecretConfigDataT struct {
	Comment             []string
	SiteRoot            string
	WebServerPort       string
	RecieverEMail       string
	SMTPServer          string
	SMTPUser            string
	SMTPPassword        string
	SenderEMail         string
	PostgresqlServerURL string
	TLSCertFile         string
	TLSKeyFile          string
}

// SecretConfigData is an in-memory copy of a semdict.config.json configuration file
var SecretConfigData *SecretConfigDataT

// SitesProtocol returns "http:" if there are no TLS sertificates
// and "https:" if there are
func SitesProtocol() string {
	if SecretConfigData.TLSKeyFile == "" {
		return "http:"
	}
	return "https:"
}
