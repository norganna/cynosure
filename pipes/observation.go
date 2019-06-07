package pipes

import (
	"sync"

	"github.com/norganna/cynosure/proto/cynosure"
)

type observation struct {
	sync.RWMutex

	live     bool
	ready    bool
	watches  map[string]*cynosure.Watch
	observed map[string]string
}

// NewObservation creates a new Observer.
func NewObservation() Observer {
	return &observation{
		watches:  map[string]*cynosure.Watch{},
		observed: map[string]string{},
	}
}

func (o *observation) AddWatch(name string, watch *cynosure.Watch) bool {
	o.Lock()
	defer o.Unlock()

	if o.live {
		// Once live, can't add more watches.
		return false
	}

	o.watches[name] = watch
	return true
}

func (o *observation) Clear() {
	o.Lock()
	defer o.Unlock()

	o.ready = false
	o.observed = map[string]string{}
}

func (o *observation) Observation(watch *cynosure.Watch, name, line string) {
	o.Lock()
	defer o.Unlock()

	o.observed[name] = line
	switch watch.State {
	case cynosure.Watch_MakeReady:
		o.ready = true
	case cynosure.Watch_NotReady:
		o.ready = false
	}
}

func (o *observation) Observed() map[string]string {
	o.RLock()
	defer o.RUnlock()

	observed := map[string]string{}
	for k, v := range o.observed {
		observed[k] = v
	}
	return observed
}

func (o *observation) Ready() bool {
	return o.ready
}

func (o *observation) Watches() map[string]*cynosure.Watch {
	if !o.live {
		o.RLock()
		defer o.RUnlock()

		// The first time watches are asked for, make us live.
		o.live = true
	}

	return o.watches
}
