package server

// The following imports add the dependency providers.
//
// If you wish to compile in additional providers, include them here:
import (
	_ "github.com/norganna/cynosure/deps/always" // Plugin.
	_ "github.com/norganna/cynosure/deps/consul" // Plugin.
	_ "github.com/norganna/cynosure/deps/cyno"   // Plugin.
	_ "github.com/norganna/cynosure/deps/etcd"   // Plugin.
	_ "github.com/norganna/cynosure/deps/http"   // Plugin.
	_ "github.com/norganna/cynosure/deps/kube"   // Plugin.
	_ "github.com/norganna/cynosure/deps/port"   // Plugin.
	_ "github.com/norganna/cynosure/deps/wait"   // Plugin.
)
