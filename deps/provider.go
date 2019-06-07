package deps

import "github.com/norganna/cynosure/common"

type providerManager struct {
	kinds map[string]Provider
}

var pm = &providerManager{
	kinds: map[string]Provider{},
}

// RegisterProvider allows a plugin provider to register itself.
func RegisterProvider(name string, provider Provider) {
	// No mutex required as all these will be called at init() phase.
	pm.kinds[name] = provider
}

// Provider will return a broker for a given configuration.
type Provider func(config common.StringMap) (Broker, error)
