package ssh

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

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

func (f *forwardMsg) To() string {
	return fmt.Sprintf("%s:%d", f.Addr, f.Rport)
}

type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// Server represents a listening ssh server
type Server struct {
	// Config is the ssh serverconfig
	Config *ssh.ServerConfig

	// Dialer provides means to dial forwards
	Dialer Dialer

	listener net.Listener
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

		forwardInfo := forwardMsg{}
		err := ssh.Unmarshal(newChannel.ExtraData(), &forwardInfo)
		if err != nil {
			logger.Printf("unable to unmarshal forward information: %s", err)
			continue
		}

		logger.Printf("accepting forward to %s:%d", forwardInfo.Addr, forwardInfo.Rport)

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logger.Print("Could not accept channel: ", err)
		}

		go ssh.DiscardRequests(requests)
		go s.connectForward(channel, forwardInfo)
	}

	logger.Print("client went away ", conn.RemoteAddr())
}

func (s *Server) connectForward(c ssh.Channel, forwardInfo forwardMsg) {
	ctx, cancelTimeout := context.WithTimeout(context.Background(), 25*time.Second)
	conn, err := s.Dialer.DialContext(ctx, "tcp", forwardInfo.To())

	cancelTimeout()
	if err != nil {
		logger.Printf("unable to dial %s: %s", forwardInfo.To(), err)
		c.Stderr().Write([]byte(fmt.Sprintf("unable to dial %s: %s", forwardInfo.To(), err)))
		c.Close()
	}

	// pass traffic
	go io.Copy(conn, c)
	go io.Copy(c, conn)
}
