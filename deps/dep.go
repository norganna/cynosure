package deps

import (
	"fmt"
	"strings"

	"github.com/norganna/cynosure/common"
)

// Standard error messages.
var (
	ErrIncorrectType = common.ErrorMsg("cannot add dependency (incorrect type)")
	ErrUnknownConfig = common.ErrorMsg("unrecognised config value")
)

// Depender is checkable and returns a message and whether the dependency is met.
type Depender interface {
	Check() (string, bool)
}

// DepList contains a set of dependencies and checks for each, which if any are successful meet the requirement.
type DepList map[string][]Depender

// Add adds a new dependency to the named requirement.
func (d DepList) Add(name string, dep Depender) {
	d[name] = append(d[name], dep)
}

// Check checks the list of dependencies and works out if they are met or not.
func (d DepList) Check() (messages []string, ok bool) {
	if d == nil {
		return nil, true
	}

	all := true
	for name, deps := range d {
		var msg []string
		var m string
		var ok bool

		for _, dep := range deps {
			m, ok = dep.Check()
			if ok {
				msg = []string{m}
				break
			}

			msg = append(msg, m)
		}

		if !ok {
			m = fmt.Sprintf("[FAIL] %s: %s", name, strings.Join(msg, "; "))
			all = false
		} else {
			m = fmt.Sprintf("[ OK ] %s: %s", name, msg[0])
		}
		messages = append(messages, m)
	}
	return messages, all
}
