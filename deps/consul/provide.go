// Package consul is a provider that can find things in consul.
//
// Accepts config:
//    address: ADDRESS
//
// Pass the ADDRESS to the consul you wish to interrogate.
package consul

import (
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "consul"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(config common.StringMap) (deps.Broker, error) {
	b := &broker{}
	addr := config["address"]
	if addr == "" {
		return nil, common.ErrorMsg("must supply a consul server api endpoint address")
	}

	return b, nil
}

type broker struct {
	address string
}

func (b *broker) Dep(wait string) (deps.Depender, error) {
	return &dep{
		b:    b,
		wait: wait,
	}, nil
}

type dep struct {
	b    *broker
	wait string
}

func (d *dep) Check() (msg string, ok bool) {
	panic("not implemented")
}
