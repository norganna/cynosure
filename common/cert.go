package common

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"net"
	"time"
)

// CreateCACert creates a certificate authority.
func CreateCACert(cn string) (cert string, key string, err error) {
	return createCert("CA", nil, nil, cn, nil, 15*365)
}

// CreateServerCert creates a server cert from a certificate authority.
func CreateServerCert(ca *tls.Certificate, cn string, hosts []string, days int) (cert string, key string, err error) {
	if !ca.Leaf.IsCA {
		return "", "", ErrorMsg("supplied authority is not a CA")
	}
	return createCert("Server", ca.Leaf, ca.PrivateKey, cn, hosts, days)
}

// CreateClientCert creates a client authentication cert from a certificate authority.
func CreateClientCert(ca *tls.Certificate, cn string, days int) (cert string, key string, err error) {
	if !ca.Leaf.IsCA {
		return "", "", ErrorMsg("supplied authority is not a CA")
	}
	return createCert("Client", ca.Leaf, ca.PrivateKey, cn, nil, days)
}

func createCert(variant string, parent *x509.Certificate, pKey crypto.PrivateKey, cn string, hosts []string, days int) (cert string, key string, err error) {
	prv, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		return "", "", Error(err, "failed to generate key")
	}

	buf := make([]byte, 32)
	_, err = rand.Read(buf)
	if err != nil {
		return "", "", Error(err, "failed to generate serial")
	}
	serial := big.NewInt(0).SetBytes(buf)

	if cn == "" {
		cn = variant
	}

	template := x509.Certificate{
		SerialNumber: serial,

		Subject: pkix.Name{
			Organization: []string{"Cynosure"},
			CommonName:   cn,
		},

		NotBefore: time.Now().Add(-time.Hour),
		NotAfter:  time.Now().Add(time.Hour * 24 * time.Duration(days)),

		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,

		BasicConstraintsValid: true,
	}

	signer := pKey

	switch variant {
	case "CA":
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		parent = &template
		signer = prv
	case "Server":
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
		for _, host := range hosts {
			if ip := net.ParseIP(host); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)
			} else {
				template.DNSNames = append(template.DNSNames, host)
			}
		}
	case "Client":
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, publicKey(prv), signer)
	if err != nil {
		return "", "", Error(err, "failed to create certificate")
	}

	cert = base64.RawStdEncoding.EncodeToString(derBytes)
	data := bytesForKey(prv)
	key = base64.RawStdEncoding.EncodeToString(data)
	return cert, key, nil
}

// TLSCertKey returns the cert and key derBytes from the certificate.
func TLSCertKey(crt *tls.Certificate) (cert, key string) {
	cert = base64.RawStdEncoding.EncodeToString(crt.Certificate[0])
	data := bytesForKey(crt.PrivateKey)
	key = base64.RawStdEncoding.EncodeToString(data)
	return cert, key
}

func publicKey(prv interface{}) interface{} {
	switch k := prv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func bytesForKey(prv interface{}) []byte {
	switch k := prv.(type) {
	case *rsa.PrivateKey:
		return x509.MarshalPKCS1PrivateKey(k)
	case *ecdsa.PrivateKey:
		b, _ := x509.MarshalECPrivateKey(k)
		return b
	default:
		return nil
	}
}

func isServerCert(crt *tls.Certificate) bool {
	return certHasUsage(crt, x509.ExtKeyUsageServerAuth)
}

func isClientCert(crt *tls.Certificate) bool {
	return certHasUsage(crt, x509.ExtKeyUsageClientAuth)
}

func certHasUsage(crt *tls.Certificate, has x509.ExtKeyUsage) bool {
	if crt == nil || crt.Leaf == nil {
		return false
	}

	for _, usage := range crt.Leaf.ExtKeyUsage {
		if usage == has {
			return true
		}
	}
	return false
}
