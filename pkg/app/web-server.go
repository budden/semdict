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
	engine.GET("/articleview/:articleslug", query.ArticleViewDirHandler)
	engine.GET("/articleedit/:articleslug", query.ArticleEditDirHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	/* // Group user related routes together
	userRoutes := router.Group("/u")
	{
		// Handle the GET requests at /u/login
		// Show the login page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/login", ensureNotLoggedIn(), showLoginPage)

		// Handle POST requests at /u/login
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/login", ensureNotLoggedIn(), performLogin)

		// Handle GET requests at /u/logout
		// Ensure that the user is logged in by using the middleware
		userRoutes.GET("/logout", ensureLoggedIn(), logout)

		// Handle the GET requests at /u/register
		// Show the registration page
		// Ensure that the user is not logged in by using the middleware
		userRoutes.GET("/register", ensureNotLoggedIn(), showRegistrationPage)

		// Handle POST requests at /u/register
		// Ensure that the user is not logged in by using the middleware
		userRoutes.POST("/register", ensureNotLoggedIn(), register)
	}*/

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
