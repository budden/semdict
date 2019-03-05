package main

import (
	"golang.org/x/net/netutil"
	"net"
	"fmt"; "log"; "time"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

const connectionLimit = 500

// https://golang.hotexamples.com/examples/golang.org.x.net.netutil/-/LimitListener/golang-limitlistener-function-examples.html
// https://habr.com/ru/post/197468/
func playWithServer() {
	http.HandleFunc("/", handler)
 s := &http.Server{
  Addr:           ":8085",
  Handler:        nil,
  ReadTimeout:    10 * time.Second, 
  WriteTimeout:   10 * time.Second,
  MaxHeaderBytes: 1 << 20 }

 listener, err := net.Listen("tcp",s.Addr)
 if err != nil { log.Fatalln(err) }
 limitListener := netutil.LimitListener(listener,connectionLimit)
 log.Fatal(s.Serve(limitListener)) }
