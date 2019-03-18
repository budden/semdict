package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/budden/semdict/pkg/database"
	"github.com/budden/semdict/pkg/shutdown"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/user"

	"github.com/budden/semdict/pkg/query"
	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/netutil"
)

func homePageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "general.html", shared.GeneralTemplateParams{Message: "Hello from gin"})
}

const connectionLimit = 500

func handleDirStrippingPrefix(dir string, handlerFunc http.HandlerFunc) {
	http.Handle(dir,
		http.StripPrefix(dir,
			http.HandlerFunc(handlerFunc)))
}

// ThisHTTPServer is a main http server
var ThisHTTPServer *http.Server

// https://golang.hotexamples.com/examples/golang.org.x.net.netutil/-/LimitListener/golang-limitlistener-function-examples.html
// https://habr.com/ru/post/197468/
func playWithServer() {
	port := ":" + shared.WebServerPort
	log.Printf("Starting server on %s - kill app to stop\n", port)

	engine := gin.New()

	engine.Use(gin.Logger(), user.SetUserStatus(), apperror.HandlePanicInRequestHandler())

	engine.LoadHTMLGlob("templates/*")
	engine.GET("/", homePageHandler)
	engine.GET("/searchform", query.SearchFormPageHandler)
	engine.GET("/searchresult", query.SearchResultPageHandler)
	// FIXME - change a way of addressing articles to be adequate!
	engine.GET("/articleview/:word", query.ArticleViewDirHandler)
	engine.GET("/articleedit/:word", query.ArticleEditDirHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginformsubmit", user.PerformLogin)
	engine.GET("/logout", user.Logout)

	engine.POST("/articlepost", query.ArticlePostDataPageHandler)

	// "/articlepost/"

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
	log.Print(ThisHTTPServer.Serve(limitListener))

	closer1 := func() { ThisHTTPServer.Shutdown(context.TODO()) }
	closer := func() { go closer1() }
	shutdown.Actions = append(shutdown.Actions, closer)
}

func actualFatalDatabaseErrorHandler(err error, c *database.ConnectionType, format string, args ...interface{}) {
	database.SetConnectionDead(c)
	log.Printf("Fatal error: "+format, args...)
	debug.PrintStack()
	shutdown.InitiateGracefulShutdown()
	apperror.Panic500If(apperror.ErrDummy, "Internal error")
}
