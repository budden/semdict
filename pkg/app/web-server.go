package app

import (
	"log"
	"net"
	"net/http"
	"time"

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

// https://golang.hotexamples.com/examples/golang.org.x.net.netutil/-/LimitListener/golang-limitlistener-function-examples.html
// https://habr.com/ru/post/197468/
func playWithServer() {
	port := ":8085"
	log.Printf("Starting server on %s - kill app to stop\n", port)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", homePageHandler)
	router.GET("/searchform", query.SearchFormPageHandler)
	router.GET("/searchresult", query.SearchResultPageHandler)
	router.GET("/articleview/:articleslug", query.ArticleViewDirHandler)
	router.GET("/articleedit/:articleslug", query.ArticleEditDirHandler)

	router.GET("/registrationform", user.RegistrationFormPageHandler)
	router.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	router.GET("/registrationconfirmation/:confirmationid", user.RegistrationConfirmationPageHandler)

	// "/articlepost/"

	// https://habr.com/ru/post/197468/
	s := &http.Server{
		Addr:           port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Fatalln(err)
	}
	limitListener := netutil.LimitListener(listener, connectionLimit)
	log.Fatal(s.Serve(limitListener))
}
