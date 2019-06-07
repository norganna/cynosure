package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/yosuke-furukawa/json5/encoding/json5"
	"google.golang.org/grpc/grpclog"
)

// ConfigCertificate contains the certificate details for the config file.
type ConfigCertificate struct {
	CA   string `json:"ca,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
}

// ConfigBrokerConfigs contains the broker configuration for the config file.
type ConfigBrokerConfigs struct {
	Default    StringMap            `json:"default,omitempty"`
	Namespaced map[string]StringMap `json:"namespaced,omitempty"`
}

// ConfigBroker specifies the broker kind for the config file.
type ConfigBroker struct {
	Kind   string              `json:"kind,omitempty"`
	Config ConfigBrokerConfigs `json:"config,omitempty"`
}

// Config contains the config file details.
type Config struct {
	Server      string                   `json:"server,omitempty"`
	Names       []string                 `json:"names,omitempty"`
	Root        string                   `json:"root,omitempty"`
	Authority   *ConfigCertificate       `json:"authority,omitempty"`
	Certificate *ConfigCertificate       `json:"certificate,omitempty"`
	Brokers     map[string]*ConfigBroker `json:"brokers,omitempty"`

	log       grpclog.LoggerV2
	auth      *tls.Certificate
	crtPool   *x509.CertPool
	crtServer *tls.Certificate
	crtClient *tls.Certificate
}

// Log returns the logger that things running from this config should use to output info.
func (c *Config) Log() grpclog.LoggerV2 {
	return c.log
}

// CertPool returns a pool of certificates that we will trust for connections.
func (c *Config) CertPool() (pool *x509.CertPool, err error) {
	if c.crtPool != nil {
		return c.crtPool, nil
	}

	c.crtPool = x509.NewCertPool()

	var ca *x509.Certificate
	var pemData []byte
	if c.Certificate != nil && c.Certificate.CA != "" {
		pemData = Pemify("CERTIFICATE", []byte(c.Certificate.CA))
	} else if c.Authority != nil && c.Authority.CA != "" {
		pemData = Pemify("CERTIFICATE", []byte(c.Authority.CA))
	}

	if len(pemData) > 0 {
		pb, _ := pem.Decode(pemData)
		ca, err = x509.ParseCertificate(pb.Bytes)
		if err != nil {
			return nil, Error(err, "failed to parse supplied CA")
		}
	} else if c.auth.Leaf != nil {
		ca = c.auth.Leaf
	} else {
		return nil, ErrorMsg("cannot find a valid CA certificate to use")
	}

	c.crtPool.AddCert(ca)
	return c.crtPool, nil
}

// ServerCert returns a valid server certificate (or generates one if possible and none available).
func (c *Config) ServerCert(days int) (crt *tls.Certificate, err error) {
	if c.crtServer != nil {
		return c.crtServer, nil
	}

	crt, err = c.getCertificate(true, "api.cynosure", days)
	if err != nil {
		return nil, err
	}

	c.crtServer = crt
	return crt, nil
}

// ClientCert returns a valid client certificate (or generates one if possible and none available).
func (c *Config) ClientCert(days int) (crt *tls.Certificate, err error) {
	if c.crtClient != nil {
		return c.crtClient, nil
	}

	crt, err = c.getCertificate(false, "client.cynosure", days)
	if err != nil {
		return nil, err
	}

	c.crtClient = crt
	return crt, nil
}

func (c *Config) getCertificate(server bool, cn string, days int) (crt *tls.Certificate, err error) {
	if c.Certificate != nil && c.Certificate.Key != "" {
		crt, err = LoadCertificate(c.Certificate.Cert, c.Certificate.Key)
		if err != nil && c.auth == nil {
			return nil, Error(err, "unable to load supplied certificate")
		}
		if server && isServerCert(crt) || !server && isClientCert(crt) {
			return crt, nil
		}
	}

	if c.auth == nil {
		return nil, nil
	}

	var cert, key string
	if server {
		cert, key, err = CreateServerCert(c.auth, cn, c.Names, days)
	} else {
		cert, key, err = CreateClientCert(c.auth, cn, days)
	}

	if err != nil {
		return nil, Error(err, "failed to create new certificate")
	}

	crt, err = LoadCertificate(cert, key)
	if err != nil {
		return nil, Error(err, "failed to load new certificate")
	}

	return crt, nil
}

// Pemify returns a PEM formatted version of the derBytes provided.
func Pemify(variant string, data []byte) []byte {
	d := base64.StdEncoding.EncodeToString(data)
	return []byte(fmt.Sprintf(
		"-----BEGIN %s-----\n%s\n-----END %s-----\n",
		variant,
		strings.Join(chunk(d, 64), "\n"),
		variant,
	))
}

func chunk(in string, size int) []string {
	var chunks []string
	runes := []rune(in)

	n := len(runes)
	if n == 0 {
		return []string{in}
	}

	for i := 0; i < n; i += size {
		nn := i + size
		if nn > n {
			nn = n
		}
		chunks = append(chunks, string(runes[i:nn]))
	}
	return chunks
}

// LoadConfig reads the passed in config filename, parses it and returns the config object.
func LoadConfig(log grpclog.LoggerV2, filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, Error(err, "could not read config file %s", filename)
	}

	config := &Config{log: log}
	err = json5.Unmarshal(data, config)
	if err != nil {
		return nil, Error(err, "failed to parse config")
	}

	a := config.Authority
	if a != nil {
		config.auth, err = LoadCertificate(a.Cert, a.Key)
		if err != nil {
			return nil, Error(err, "could not load authority certificate")
		}

		if !config.auth.Leaf.IsCA {
			return nil, ErrorMsg("supplied authority is not a CA")
		}
	}

	err = setupRoot(config.Root)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// LoadCertificate will load a TLS certificate from the provided base 64 encoded derBytes.
func LoadCertificate(certEnc, keyEnc string) (crt *tls.Certificate, err error) {
	var data []byte

	data, err = base64.RawStdEncoding.DecodeString(certEnc)
	if err != nil {
		return nil, err
	}
	cert := Pemify("CERTIFICATE", data)

	data, err = base64.RawStdEncoding.DecodeString(keyEnc)
	if err != nil {
		return nil, err
	}
	key := Pemify("EC PRIVATE KEY", data)

	var xc tls.Certificate
	xc, err = tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, Error(err, "failed to load X509 key pair")
	}

	xc.Leaf, err = x509.ParseCertificate(xc.Certificate[0])
	if err != nil {
		return nil, Error(err, "failed to parse leaf certificate")
	}

	return &xc, nil
}

func setupRoot(root string) (err error) {
	if runtime.GOOS == "darwin" {
		err = setupDarwinMount(root)
		if err != nil {
			return err
		}
	}

	for _, sub := range []string{"data", "images", "instances"} {
		err = os.MkdirAll(path.Join(root, sub), 0755)
		if err != nil {
			return Error(err, "failed to create root subdirectory %s", sub)
		}
	}

	return nil
}

func setupDarwinMount(root string) (err error) {
	// On OS X, we need a root image that is HFS+J to be able to hard link directories.
	// This is to overcome the fact that OSX does not support bind mounting volumes.
	image := root + ".sparsebundle"

	if !DirExists(root) {
		err = os.Mkdir(root, 0755)
		if err != nil {
			return Error(err, "failed to create root directory %s", root)
		}
	}

	if !DirExists(image) {
		err = RunCmd(
			"hdiutil",
			"create",
			"-fs", "HFS+J",
			"-size", "500g",
			"-type", "SPARSEBUNDLE",
			"-volname", "cynosure-root",
			image,
		)
		if err != nil {
			return Error(err, "failed to create root sparsebundle")
		}
	}

	err = RunCmd(
		"hdiutil",
		"attach",
		"-noverify",
		"-nobrowse",
		"-noautoopen",
		"-mountpoint", root,
		image,
	)
	if err != nil {
		return Error(err, "failed to mount root sparsebundle")
	}

	return nil
}
