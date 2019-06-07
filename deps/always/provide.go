// Package always is an example provider that you can copy to create a new provider.
//
// Accepts config:
//    state: "true" or "false"
//
// If state is true the broker will always check OK.
package always

import (
	"fmt"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "always"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(config common.StringMap) (deps.Broker, error) {
	b := &broker{}
	state := config["state"]
	switch state {
	case "true":
		b.state = true
	case "false":
		b.state = false
	default:
		return nil, common.ErrorMsg(`unknown state %s`, state)
	}

	return b, nil
}

type broker struct {
	state bool
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
	msg = fmt.Sprintf("%s %t", Kind, d.b.state)
	ok = d.b.state
	return msg, ok
}
