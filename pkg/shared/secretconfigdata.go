package shared

// SecretConfigDataT specifies the fields of semdict.config.json
// That file contains the data which is secret and site-specific so it can't be stored to git
type SecretConfigDataT struct {
	Comment             []string
	SiteRoot            string
	UnderAProxy         int8 // 0 means false, 1 means true
	ServerPort          string
	SMTPServer          string
	SMTPUser            string
	SMTPPassword        string
	SenderEMail         string
	PostgresqlServerURL string
	TLSCertFile         string
	TLSKeyFile          string
	// If set to non-zero, acts as if a user with this id is always logged in,
	// useful for debugging of user-based routes
	UserAlwaysLoggedIn int
	// Some gin messages are annoying, set this switch to 1 to hush them
	HideGinStartupDebugMessages int
	// Set GinDebugMode to 1 to enable gin debug mode
	GinDebugMode int
}

// SecretConfigDataTComment is actually a documentation for SecretConfigData, which is placed to a config sample file
var SecretConfigDataTComment = []string{"Example config file. Copy this one to the semdict.config.json and edit.",
	"UnderAProxy is an integer value with legal values of 0 (false) and 1 (true)",
	"Set UnderAProxy to 0 if gin is used as a web server (standalone mode)",
	"UnderAProxy to 1 when semdict is started as a service behind a TLS-enabled reverse proxy (service mode)",
	"ServerPort is included into the registration E-mails only is UnderAProxy == 1.",
	"TLSCertFile and TLSKeyFile (PEM format) can only be used in a standalone mode to enable https",
	"Pass empty strings to use plain http",
	"If an SMTPServer is set to an empty string, emails are printed to stdout instead of actually being sent"}

// SecretConfigData is an in-memory copy of a semdict.config.json configuration file
var SecretConfigData *SecretConfigDataT

// SitesProtocol returns "http:" or "https:"
func SitesProtocol() string {
	scd := SecretConfigData
	if scd.UnderAProxy == 1 || scd.TLSKeyFile != "" {
		return "https:"
	}
	return "http:"
}

// SitesPort returns "port:" if there is a non-standard port.
// Under a proxy, returns nothing
func SitesPort() string {
	scd := SecretConfigData
	if scd.UnderAProxy == 1 {
		return ""
	}
	return ":" + scd.ServerPort
}
