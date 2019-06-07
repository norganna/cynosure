package cli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/norganna/cynosure/common"
	"google.golang.org/grpc/grpclog"
)

type handlerFunc func(config *common.Config, args []string)

func registerHandler(name string, handler handlerFunc) {
	handlers[name] = handler
}

var (
	log grpclog.LoggerV2

	configFile = flag.String("config", "", "The configuration file")

	handlers = map[string]handlerFunc{}
)

// Boot is the entry point for the cli.
func Boot() {
	var err error

	log = grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)
	common.SetLogger(log)

	flag.CommandLine.Init(flag.CommandLine.Name(), flag.ContinueOnError)
	flag.Parse()
	cfgFile := *configFile

	if cfgFile == "" {
		cfgPath := path.Join(os.Getenv("HOME"), ".cyno")
		cfgFile = path.Join(cfgPath, "config")

		if !common.DirExists(cfgPath) {
			err = os.MkdirAll(cfgPath, 0755)
			if err != nil {
				log.Fatalf("Failed creating new config dir %s: %s", cfgPath, err)
			}
		}

		if !common.FileExists(cfgFile) {
			log.Infof("Creating new config file %s", cfgFile)
			configText := createServerConfig(cfgPath)
			err = ioutil.WriteFile(cfgFile, []byte(configText), 0640)
			if err != nil {
				log.Fatalf("Failed writing new config %s: %s", cfgFile, err)
			}
		}
	}

	if !common.FileExists(cfgFile) {
		log.Fatalf("The specified config file does not exist: %s", cfgFile)
	}

	config, err := common.LoadConfig(log, cfgFile)
	if err != nil {
		log.Fatalf("Failed to load config file %s: %s", cfgFile, err)
	}

	params := flag.Args()
	var handler handlerFunc

	if len(params) > 0 {
		if h, ok := handlers[params[0]]; ok {
			handler = h
			params = params[1:]
		}
	}

	if handler == nil {
		fmt.Println("Usage:")
		fmt.Println("  cynosure [--config=CONFIG] COMMAND ARGS")
		fmt.Println("Where command is one of:")
		fmt.Println("  server     Start a cynosure server on this computer")
		fmt.Println("  config     Output a new CLIENT config to stdout")
		return
	}

	handler(config, params)
}
