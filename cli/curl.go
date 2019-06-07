package cli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/norganna/cynosure/common"
)

func init() {
	registerHandler("curl", func(config *common.Config, args []string) {
		crt, err := config.ClientCert(1)
		if err != nil {
			config.Log().Fatal("Failed getting client cert: ", err)
		}

		dir, err := ioutil.TempDir("", "cynocurl")
		if err != nil {
			config.Log().Fatal("Failed to create temp folder: ", err)
		}

		cert, key := common.TLSCertKey(crt)
		certData, keyData, caData := decodeCerts(cert, config, key)

		certPem := common.Pemify("CERTIFICATE", certData)
		keyPem := common.Pemify("EC PRIVATE KEY", keyData)
		caPem := common.Pemify("CERTIFICATE", caData)

		caFile := filepath.Join(dir, "ca.pem")
		err = ioutil.WriteFile(caFile, caPem, 0640)
		if err != nil {
			config.Log().Fatal("Failed to create ca pem file: ", err)
		}

		crtFile := filepath.Join(dir, "crt.pem")
		err = ioutil.WriteFile(crtFile, append(certPem, keyPem...), 0640)
		if err != nil {
			config.Log().Fatal("Failed to create crt pem file: ", err)
		}

		cArgs := append([]string{"curl", "--cacert", caFile, "--cert", crtFile}, args...)
		err = syscall.Exec("/usr/bin/curl", cArgs, os.Environ())
		if err != nil {
			config.Log().Fatal("Failed to exec curl command: ", err)
		}
	})
}
