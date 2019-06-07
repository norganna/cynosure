// Package port is a network port checker.
//
// Accepts no config.
//
// The "wait" parameter is of the form:
//   [HOST]:PORT[/PROTO]
//
// HOST is either an IPv4, IPv6 address or resolvable host name (default = 127.0.0.1).
// PORT is a port number.
// PROTO is either "TCP" or "UDP" (default = TCP).
//
// Examples:
//   :3306
//   google.com:80
//   172.17.1.123:53/UDP
package port

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "port"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(_ common.StringMap) (deps.Broker, error) {
	b := &broker{}
	return b, nil
}

type broker struct{}

func (b *broker) Dep(wait string) (deps.Depender, error) {
	proto := "tcp"
	if strings.HasSuffix(wait, "/TCP") {
		wait = wait[:len(wait)-4]
	} else if strings.HasSuffix(wait, "/UDP") {
		proto = "udp"
		wait = wait[:len(wait)-4]
	}

	host, port, err := net.SplitHostPort(wait)
	if err != nil {
		return nil, common.Error(err, "failed to parse host/port %s", wait)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return &dep{
		host:  host,
		port:  port,
		proto: proto,
	}, nil
}

type dep struct {
	host  string
	port  string
	proto string
}

func (d *dep) Check() (msg string, ok bool) {
	msg = fmt.Sprintf("%s %s:%s/%s", Kind, d.host, d.port, strings.ToUpper(d.proto))

	conn, err := net.DialTimeout(d.proto, fmt.Sprintf("%s:%s", d.host, d.port), time.Second)
	defer func() {
		_ = conn.Close()
	}()

	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		return msg + " timeout", false
	}

	if err != nil {
		return msg + " " + err.Error(), false
	}

	return msg + " open", true
}
