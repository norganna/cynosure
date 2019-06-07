// Package cyno is a provider that checks running instances in a cynosure server.
//
// Accepts config:
//    type: "self" or "config"
//    file: PATH
//
// If type is "self", will use internal methods to check running instances.
//
// If type is "config", will use a RPC using the client config file at PATH.
// You can generate a remote config file using `cynosure config` on the destination server.
package cyno

import (
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "cyno"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(config common.StringMap) (br deps.Broker, err error) {
	b := &broker{}
	t := config["type"]
	switch t {
	case "self":
		b.internal = true
	case "config":
		b.config, err = common.LoadConfig(common.Logger(), config["file"])
		if err != nil {
			return nil, err
		}
	default:
		return nil, common.ErrorMsg(`unknown type %s`, t)
	}

	return b, nil
}

type broker struct {
	internal bool
	config   *common.Config
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
	// TODO implement checks once the process manager is completed.
	panic("implement me")
}
