package ssh

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "[ssh] ", log.Flags())
}

type forwardMsg struct {
	Addr           string
	Rport          uint32
	OriginatorAddr string
	OriginatorPort uint32
}

// Server represents a listening ssh server
type Server struct {
	Config   *ssh.ServerConfig
	listener net.Listener
}

// DefaultConfig generates a default ssh.ServerConfig
func DefaultConfig() (*ssh.ServerConfig, error) {
	config := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			return nil, nil
		},

		// Remove to disable public key auth.
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			return &ssh.Permissions{
				// Record the public key used for authentication.
				Extensions: map[string]string{
					"pubkey-fp": ssh.FingerprintSHA256(pubKey),
				},
			}, nil
		},
	}

	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		return nil, fmt.Errorf("Failed to load private key: %s", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %s", err)
	}

	config.AddHostKey(private)

	return config, nil
}

// Listen will listen and accept ssh connections
func (s *Server) Listen(l net.Listener) {
	s.listener = l
	logger.Print("listening on ", l.Addr())
	for {
		nConn, err := s.listener.Accept()
		if err != nil {
			logger.Print("failed to accept incoming connection: ", err)
		}
		go s.accept(nConn)
	}
}

func (s *Server) accept(c net.Conn) {

	// Before use, a handshake must be performed on the incoming
	// net.Conn.
	conn, chans, reqs, err := ssh.NewServerConn(c, s.Config)
	if err != nil {
		logger.Print("failed to handshake: ", err)
		return
	}

	logger.Printf("accepted session from %s", conn.RemoteAddr())

	// The incoming Request channel must be serviced.
	go func(reqs <-chan *ssh.Request) {
		for req := range reqs {
			if req.Type == "keepalive@openssh.com" {
				req.Reply(true, nil)
				continue
			}
			req.Reply(false, nil)

		}
	}(reqs)

	// Service the incoming Channel channel.
	for newChannel := range chans {

		if newChannel.ChannelType() != "direct-tcpip" {
			newChannel.Reject(ssh.UnknownChannelType, newChannel.ChannelType())
			continue
		}

		forwardInfo := &forwardMsg{}
		err := ssh.Unmarshal(newChannel.ExtraData(), forwardInfo)
		if err != nil {
			logger.Printf("unable to unmarshal forward information: %s", err)
			continue
		}
		logger.Printf("extra data: %+v, unmarshalled: %+v", newChannel.ExtraData(), forwardInfo)

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logger.Print("Could not accept channel: ", err)
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				logger.Printf("channel request: %+v", req)
				req.Reply(req.Type == "shell", nil)
			}
		}(requests)
		channel.Write([]byte("Hjerteligt velkommen\n\r"))
	}

	logger.Print("client went away ", conn.RemoteAddr())
}
