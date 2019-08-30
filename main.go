package main

import (
	"log"
	"net"

	"github.com/fasmide/onionpass/ssh"
)

func main() {
	sshConfig, err := ssh.DefaultConfig()
	if err != nil {
		log.Fatalf("cannot get default ssh config: %s", err)
	}

	listener, err := net.ListenTCP("tcp", &net.TCPAddr{})
	if err != nil {
		log.Fatalf("cannot listen for ssh connections: %s", err)
	}

	sshServer := ssh.Server{Config: sshConfig}
	sshServer.Listen(listener)
}
