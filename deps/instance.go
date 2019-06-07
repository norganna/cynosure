package deps

import "github.com/norganna/cynosure/common"

type instanceManager struct {
	// instances[identity][namespace]
	instances map[string]map[string]Broker
}

var im = &instanceManager{
	instances: map[string]map[string]Broker{},
}

// NewInstance creates an instance with the given identity/namespace.
func NewInstance(identity, namespace, kind string, config common.StringMap) error {
	provider := pm.kinds[kind]
	if provider == nil {
		return common.ErrorMsg("failed to find provider of kind %s", kind)
	}
	broker, err := provider(config)
	if err != nil {
		return err
	}

	if _, ok := im.instances[identity]; !ok {
		im.instances[identity] = map[string]Broker{}
	}
	im.instances[identity][namespace] = broker

	return nil
}

// Instance returns the instance with the given identity/namespace.
func Instance(identity, namespace string) Broker {
	ii := im.instances[identity]
	if b, ok := ii[namespace]; ok {
		return b
	}
	return ii[""]
}
