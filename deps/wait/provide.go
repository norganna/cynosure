// Package wait simply returns true after a given interval has passed.
//
// Accepts no config.
//
// The "wait" parameter is of the form:
//   MILLISECONDS
//
// MILLISECONDS is just a number of milliseconds to wait before returning true.
package wait

import (
	"fmt"
	"strconv"
	"time"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "wait"

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
	ms, err := strconv.ParseInt(wait, 10, 64)
	if err != nil {
		return nil, common.Error(err, "failed to parse interval %s", wait)
	}

	return &dep{
		ms:      ms,
		expires: time.Now().Add(time.Duration(ms) * time.Millisecond),
	}, nil
}

type dep struct {
	ms      int64
	expires time.Time
}

func (d *dep) Check() (msg string, ok bool) {
	msg = fmt.Sprintf("%s %dms", Kind, d.ms)

	if time.Now().Before(d.expires) {
		return msg + " waiting", false
	}

	return msg + " complete", true
}
