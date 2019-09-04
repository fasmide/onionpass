package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/fasmide/onionpass/ssh"
	"golang.org/x/net/proxy"
)

var sshPort = flag.Int("sshport", 0, "ssh listen port")
var httpPort = flag.Int("httpport", 0, "http listen port")

func main() {
	flag.Parse()

	// lets do a webserver which redirects anyone to the project
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/fasmide/onionpass", http.StatusFound)
	})

	go func() {
		l := fmt.Sprintf(":%d", *httpPort)
		log.Print("http listening on ", l)
		log.Printf("http server failed: %s", http.ListenAndServe(l, nil))
		// we dont care if this http server fails
	}()

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *sshPort})
	if err != nil {
		log.Fatalf("cannot listen for ssh connections: %s", err)
	}

	sshServer := ssh.Server{Config: sshConfig, Dial: proxy.Dial}
	sshServer.Listen(listener)
}
