package main

import (
	"flag"
	"log"
	"net"

	"github.com/fasmide/onionpass/ssh"
)

var sshPort = flag.Int("sshport", 0, "ssh listen port")

func main() {
	flag.Parse()

	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{Port: *sshPort})
	if err != nil {
		log.Fatalf("cannot listen for ssh connections: %s", err)
	}

	sshServer := ssh.Server{Config: sshConfig}
	sshServer.Listen(listener)
}
