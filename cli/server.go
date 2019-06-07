package cli

import (
	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/server"
)

func init() {
	registerHandler("server", func(config *common.Config, args []string) {
		server.Run(config, args)
	})
}
