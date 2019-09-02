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

// Server represents a listening ssh server
type Server struct {
	// Config is the ssh serverconfig
	Config *ssh.ServerConfig

	// Dial provides means to dial forwards
	Dial func(ctx context.Context, network, address string) (net.Conn, error)

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
	// auth timeout
	// only give people 10 seconds to ssh handshake and authenticate themselves
	timer := time.AfterFunc(10*time.Second, func() {
		c.Close()
	})

	// ssh handshake and auth
	conn, chans, reqs, err := ssh.NewServerConn(c, s.Config)
	if err != nil {
		logger.Print("failed to handshake: ", err)
		return
	}

	timer.Stop()

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
	for channelRequest := range chans {

		if channelRequest.ChannelType() != "direct-tcpip" {
			channelRequest.Reject(ssh.Prohibited, fmt.Sprintf("no %s allowed, only direct-tcpip", channelRequest.ChannelType()))
			continue
		}

		forwardInfo := forwardMsg{}
		err := ssh.Unmarshal(channelRequest.ExtraData(), &forwardInfo)
		if err != nil {
			logger.Printf("unable to unmarshal forward information: %s", err)
			channelRequest.Reject(ssh.UnknownChannelType, "failed to parse forward information")
			continue
		}

		if !forwardInfo.IsOnion() {
			channelRequest.Reject(ssh.Prohibited, "onionpass only passes traffic to .onion services")
			continue
		}

		// it seems channel.Stderr() is somehow broken between golang and the regular ssh client
		// so to give a meaning full error response to users . we connect to the forward endpoint
		// before accepting the channel - if this connection fails we reject the channel request
		// with a meaningfull message
		ctx, cancelTimeout := context.WithTimeout(context.Background(), 25*time.Second)
		forwardConnection, err := s.Dial(ctx, "tcp", forwardInfo.To())
		cancelTimeout()
		if err != nil {
			logger.Printf("unable to dial %s: %s", forwardInfo.To(), err)
			channelRequest.Reject(ssh.ConnectionFailed, fmt.Sprintf("failed to dial %s: %s", forwardInfo.To(), err))
			continue
		}

		// Accept channel from ssh client
		logger.Printf("accepting forward to %s:%d", forwardInfo.Addr, forwardInfo.Rport)
		channel, requests, err := channelRequest.Accept()
		if err != nil {
			logger.Print("could not accept forward channel: ", err)
			continue
		}

		go ssh.DiscardRequests(requests)

		// pass traffic in both directions - close channel when io.Copy returns
		go func() {
			io.Copy(forwardConnection, channel)
			channel.Close()
		}()
		go func() {
			io.Copy(channel, forwardConnection)
			channel.Close()
		}()
	}

	logger.Print("client went away ", conn.RemoteAddr())
}
