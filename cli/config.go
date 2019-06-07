package cli

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/norganna/cynosure/common"
)

func init() {
	registerHandler("config", func(config *common.Config, args []string) {
		host := common.HostIP()
		if len(args) > 0 {
			command := args[0]
			switch command {
			case "local":
				host = "127.0.0.1"
			case "host":
				if len(args) > 1 {
					host = args[1]
				}
			case "client-cert", "server-cert":
				var crt *tls.Certificate
				var err error

				var certType string
				if command == "client-cert" {
					crt, err = config.ClientCert(365)
					certType = "Client"
				} else {
					crt, err = config.ServerCert(365)
					certType = "Server"
				}
				if err != nil {
					config.Log().Fatalf("Failed to create %s certificate", strings.ToLower(certType))
				}

				dumpCert(config, crt, certType)
				return
			}
		}
		fmt.Println(createClientConfig(config, host))
	})
}

func createServerConfig(folder string) string {
	cert, key, err := common.CreateCACert("ca.cynosure")
	if err != nil {
		log.Fatalf("Failed to create CA cert")
	}

	config := &common.Config{
		Server: ":8055",
		Names: []string{
			"127.0.0.1",
			"::1",
			common.HostIP(),
			"localhost",
			"api.cynosure",
		},
		Root: path.Join(folder, "root"),
		Authority: &common.ConfigCertificate{
			Cert: string(cert),
			Key:  string(key),
		},
		Brokers: map[string]*common.ConfigBroker{
			"cyno-local": {
				Kind: "cyno",
				Config: common.ConfigBrokerConfigs{
					Default: common.StringMap{
						"type": "self",
					},
				},
			},
			"kube-in-cluster": {
				Kind: "kube",
				Config: common.ConfigBrokerConfigs{
					Default: common.StringMap{
						"type": "in-cluster",
					},
				},
			},
			"port": {
				Kind: "port",
			},
			"http": {
				Kind: "http",
			},
			"wait": {
				Kind: "wait",
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config", err)
	}

	return string(data)
}

func createClientConfig(config *common.Config, host string) string {
	sHost, sPort, err := net.SplitHostPort(config.Server)
	if err != nil {
		config.Log().Fatal("Unable to determine Server host/port from ", config.Server)
	}

	if host != "" {
		sHost = host
	} else {
		if common.IsPrivateHost(sHost) {
			sHost = common.HostIP()
		}
	}

	crt, err := config.ClientCert(365)
	if err != nil {
		config.Log().Fatal("Unable to create client certificate: ", err)
	}

	cert, key := common.TLSCertKey(crt)
	client := &common.Config{
		Server: net.JoinHostPort(sHost, sPort),
		Certificate: &common.ConfigCertificate{
			CA:   config.Authority.Cert,
			Cert: cert,
			Key:  key,
		},
	}

	data, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal config", err)
	}

	return string(data)
}

func dumpCert(config *common.Config, crt *tls.Certificate, certType string) {
	cert, key := common.TLSCertKey(crt)

	certData, keyData, caData := decodeCerts(cert, config, key)

	fmt.Printf("%s certificate:\n", certType)
	fmt.Println(string(common.Pemify("CERTIFICATE", certData)))

	fmt.Println("Root CA:")
	fmt.Println(string(common.Pemify("CERTIFICATE", caData)))

	fmt.Printf("%s private key:\n", certType)
	fmt.Println(string(common.Pemify("EC PRIVATE KEY", keyData)))
}

func decodeCerts(cert string, config *common.Config, key string) (certData []byte, keyData []byte, caData []byte) {
	certData, err := base64.RawStdEncoding.DecodeString(cert)
	if err != nil {
		config.Log().Fatal("Failed to decode cert")
	}
	keyData, err = base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		config.Log().Fatal("Failed to decode key")
	}
	caData, err = base64.RawStdEncoding.DecodeString(config.Authority.Cert)
	if err != nil {
		config.Log().Fatal("Failed to decode ca")
	}
	return certData, keyData, caData
}
