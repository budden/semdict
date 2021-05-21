package shared

// SecretConfigDataT определяет поля semdict.config.json
// Этот файл содержит данные, которые являются секретными и специфичными для конкретного сайта, поэтому они не могут быть сохранены в git
type SecretConfigDataT struct {
	Comment             []string
	SiteRoot            string
	UnderAProxy         int8 // 0 означает ложь, 1 - истину
	ServerPort          string
	SMTPServer          string
	SMTPUser            string
	SMTPPassword        string
	SenderEMail         string
	PostgresqlServerURL string
	TLSCertFile         string
	TLSKeyFile          string
	// Если установлено ненулевое значение, действует так, будто пользователь с этим идентификатором всегда входит в систему,
	// что полезно для отладки маршрутов, основанных на пользователях.
	UserAlwaysLoggedIn int
	// Некоторые сообщения gin раздражают, установите этот переключатель на 1, чтобы заглушить их.
	HideGinStartupDebugMessages int
	// Установите GinDebugMode в 1, чтобы включить режим отладки gin
	GinDebugMode int
}

// SecretConfigDataTComment - это фактически документация для SecretConfigData, которая помещается в файл образца конфигурации
var SecretConfigDataTComment = []string{"Пример конфигурационного файла. Скопируйте его в файл semdict.config.json и отредактируйте.",
	"UnderAProxy - целочисленное значение с допустимыми значениями 0 (false) и 1 (true)",
	"Установите UnderAProxy на 0, если gin используется в качестве веб-сервера (автономный режим)",
	"UnderAProxy - 1, когда semdict запускается как служба за обратным прокси-сервером с поддержкой TLS (режим службы).",
	"ServerPort включается в регистрацию E-mails только в том случае, если UnderAProxy == 1.",
	"TLSCertFile и TLSKeyFile (формат PEM) можно использовать только в автономном режиме для включения https",
	"Передавайте пустые строки для использования обычного http",
	"Если SMTPServer установлен на пустую строку, электронные письма выводятся в stdout, а не отправляются."}

// SecretConfigData - это копия в памяти конфигурационного файла semdict.config.json
var SecretConfigData *SecretConfigDataT

// SitesProtocol возвращает "http:". или "https:".
func SitesProtocol() string {
	scd := SecretConfigData
	if scd.UnderAProxy == 1 || scd.TLSKeyFile != "" {
		return "https:"
	}
	return "http:"
}

// SitesPort возвращает "порт:". если имеется нестандартный порт.
// По доверенности, ничего не возвращает
func SitesPort() string {
	scd := SecretConfigData
	if scd.UnderAProxy == 1 {
		return ""
	}
	return ":" + scd.ServerPort
}
