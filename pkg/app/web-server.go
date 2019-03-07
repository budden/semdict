package app

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/budden/a/pkg/query"
	"github.com/budden/a/pkg/shared"
	"golang.org/x/net/netutil"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	if tail := r.URL.Path[1:]; tail != "" {
		http.Error(w, "Sorry 2, page not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Welcome to semantic dictionary!")
}

const connectionLimit = 500

// https://golang.hotexamples.com/examples/golang.org.x.net.netutil/-/LimitListener/golang-limitlistener-function-examples.html
// https://habr.com/ru/post/197468/
func playWithServer() {
	port := ":8085"
	log.Printf("Starting server on %s - kill app to stop\n", port)

	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/searchform", query.SearchFormPageHandler)
	http.HandleFunc("/searchresult", query.SearchFormPageHandler)
	// /articleview/SLUG
	http.HandleFunc(shared.ArticleViewDirPath, query.ArticleViewDirHandler)
	// /articleedit/SLUG
	http.HandleFunc(shared.ArticleEditDirPath, query.ArticleEditDirHandler)

	s := &http.Server{
		Addr:           port,
		Handler:        nil,
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
