package app

import (
	"context"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"time"

	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shutdown"
	"github.com/budden/semdict/pkg/unsorted"
	"github.com/coreos/go-systemd/daemon"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/query"
	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/netutil"
)

func homePageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Welcome to semantic dictionary"})
}

const connectionLimit = 500

func handleDirStrippingPrefix(dir string, handlerFunc http.HandlerFunc) {
	http.Handle(dir,
		http.StripPrefix(dir,
			http.HandlerFunc(handlerFunc)))
}

// ThisHTTPServer is a main http server
var ThisHTTPServer *http.Server

// https://golang.hotexamples.com/examples/golang.org.x.net.netutil/-/LimitListener/golang-limitlistener-function-examples.t.html
// https://habr.com/ru/post/197468/
func playWithServer() {
	port := ":" + shared.SecretConfigData.ServerPort
	log.Printf("Starting server on %s - kill app to stop\n", port)

	// https://stackoverflow.com/a/52830435/9469533
	// FIXME conditionalize
	if shared.SecretConfigData.GinDebugMode != 0 {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	//This will disable hot template reloading, so we'll try to disable any messaging for a whil

	engine := initRouter()

	// https://habr.com/ru/post/197468/
	ThisHTTPServer := &http.Server{
		Addr:           port,
		Handler:        engine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20}

	listener, err := net.Listen("tcp", ThisHTTPServer.Addr)

	if err != nil {
		log.Fatalln(err)
	}

	limitListener := netutil.LimitListener(listener, connectionLimit)

	scd := shared.SecretConfigData

	daemon.SdNotify(false, "READY=1")

	if scd.TLSCertFile == "" && scd.TLSKeyFile == "" {
		log.Print(ThisHTTPServer.Serve(limitListener))
	} else {
		log.Print(ThisHTTPServer.ServeTLS(limitListener, scd.TLSCertFile, scd.TLSKeyFile))
	}

	closer1 := func() { ThisHTTPServer.Shutdown(context.TODO()) }
	closer := func() { go closer1() }
	shutdown.Actions = append(shutdown.Actions, closer)
}

func actualFatalDatabaseErrorHandler(err error, c *sddb.ConnectionType, format string, args ...interface{}) {
	sddb.SetConnectionDead(c)
	log.Printf("Fatal error: "+format, args...)
	debug.PrintStack()
	shutdown.InitiateGracefulShutdown()
	apperror.Panic500If(apperror.ErrDummy, "Internal error")
}

func initRouter() *gin.Engine {

	if shared.SecretConfigData.HideGinStartupDebugMessages != 0 {
		oldStderr := os.Stderr
		devNull := unsorted.OpenDevNullForWrite()
		os.Stderr = devNull
		defer func() { os.Stderr = oldStderr; devNull.Close() }()
	}

	engine := gin.New()

	engine.Use(gin.Logger(), user.SetUserStatus(), apperror.HandlePanicInRequestHandler())

	setupTemplates(engine)
	engine.GET("/", homePageHandler)
	engine.GET("/menu", menuPageHandler)
	engine.GET("/wordsearchform", query.WordSearchFormRouteHandler)
	engine.GET("/wordsearchresultform", query.WordSearchResultRouteHandler)
	engine.GET("/wordsearchquery", query.WordSearchQueryRouteHandler)
	engine.GET("/sensebyidview/:senseid", query.SenseByIdViewDirHandler)
	engine.GET("/sensebyoriginidview/:originid", query.SenseByOriginIdViewDirHandler)
	engine.GET("/sensebyoriginidedit/:originid", query.SenseByOriginIdEditDirHandler)
	engine.POST("/senseproposaladdform", query.SenseProposalAddFormPageHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginformsubmit", user.LoginFormSubmitPostHandler) // FIXME rename handler
	engine.GET("/logout", user.Logout)
	engine.Static("/static", *TemplateBaseDir+"static")

	engine.POST("/senseeditformsubmit/:proposalid", query.SenseEditFormSubmitPostHandler)
	engine.GET("/senseproposalslistform/:originid", query.SenseProposalsListFormRouteHandler)

	//engine.GET("/captcha/:imagefilename", ReverseProxy)
	return engine
}

func castAsHTML(s string) template.HTML {
	return template.HTML(s)
}

func setupTemplates(engine *gin.Engine) {
	funcMap := template.FuncMap{
		"castAsHTML": castAsHTML}
	engine.SetFuncMap(funcMap)
	templatesGlob := *TemplateBaseDir + "templates/*.t.html"
	engine.LoadHTMLGlob(templatesGlob)
}

// ReverseProxy https://stackoverflow.com/a/39009974/9469533
func ReverseProxy(c *gin.Context) {
	target := "localhost:8666"
	director := func(req *http.Request) {
		r := c.Request
		//req = r
		req.URL.Scheme = "http"
		req.URL.Host = target
		req.Header["my-header"] = []string{r.Header.Get("my-header")}
		// Golang camelcases headers
		delete(req.Header, "My-Header")
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}
