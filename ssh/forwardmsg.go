package ssh

import (
	"fmt"
	"strings"
)

// forwardMsg request - See RFC4254 7.2 TCP/IP Forwarding Channels
// https://tools.ietf.org/html/rfc4254#page-18
type forwardMsg struct {
	Addr           string
	Rport          uint32
	OriginatorAddr string
	OriginatorPort uint32
}

func (f *forwardMsg) To() string {
	return fmt.Sprintf("%s:%d", f.Addr, f.Rport)
}

func (f *forwardMsg) IsOnion() bool {
	return strings.HasSuffix(f.Addr, ".onion")
}
