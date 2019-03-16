package app

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/budden/a/pkg/database"
	"github.com/budden/a/pkg/gracefulshutdown"
	"github.com/jmoiron/sqlx"

	"github.com/budden/a/pkg/apperror"
	"github.com/budden/a/pkg/user"

	//"github.com/budden/a/pkg/query"
	"github.com/budden/a/pkg/query"
	"github.com/budden/a/pkg/shared"
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
	engine.Use(gin.Logger(), apperror.HandlePanicInRequestHandler() /*, gin.Recovery()*/)
	engine.LoadHTMLGlob("templates/*")
	engine.GET("/", homePageHandler)
	engine.GET("/searchform", query.SearchFormPageHandler)
	engine.GET("/searchresult", query.SearchResultPageHandler)
	engine.GET("/articleview/:articleslug", query.ArticleViewDirHandler)
	engine.GET("/articleedit/:articleslug", query.ArticleEditDirHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

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
	gracefulshutdown.Actions = append(gracefulshutdown.Actions, closer)
}

func fatalDatabaseErrorHandler(err error, db *sqlx.DB, format string, args ...interface{}) {
	database.SetConnectionDead(db)
	gracefulshutdown.InitiateGracefulShutdown()
	apperror.Panic500If(apperror.ErrDummy, "Internal error")
}

// CheckDbAlive is to be called in page handlers before every db interaction
func CheckDbAlive(db *sqlx.DB) {
	if database.IsConnectionDead(db) {
		apperror.Panic500If(apperror.ErrDummy, "Internal error")
	}
}

// with this database and have to shut down. If this happens, we first declare that database as dead.
// Next, we initiate a "graceful shutdown". Last, we arrange to return status 500 to the client.
