// Package testcerts enables users to create temporary x509 Certificates for testing.
//
// There are many Certificate generation tools out there, but most focus on being a CLI tool. This package is focused
// on providing helper functions for creating Certificates. These helper functions can be used as part of your unit
// and integration tests as per the example below.
//
//  func TestSomething(t *testing.T) {
//    err := testcerts.GenerateCertsToFile("/tmp/cert", "/tmp/key")
//    if err != nil {
//      // do stuff
//    }
//
//    _ = something.Run("/tmp/cert", "/tmp/key")
//    // do more testing
//  }
//
// The goal of this package, is to make testing TLS based services easier. Without having to leave the comfort of your
// editor, or place test certificates in your repo.
package testcerts

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

// GenerateCerts will create a temporary x509 Certificate and Key.
//	cert, key, err := GenerateCerts()
//	if err != nil {
//		// do stuff
//	}
func GenerateCerts() ([]byte, []byte, error) {
	// Create certs and return as []byte
	c, k, err := genCerts()
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(c), pem.EncodeToMemory(k), nil
}

// GenerateCertsToFile will create a temporary x509 Certificate and Key. Writing them to the file provided.
//  err := GenerateCertsToFile("/path/to/cert", "/path/to/key")
//  if err != nil {
//    // do stuff
//  }
//
func GenerateCertsToFile(certFile, keyFile string) error {
	// Create Certs
	c, k, err := genCerts()
	if err != nil {
		return err
	}

	// Write to Certificate File
	cfh, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("unable to create certificate file - %s", err)
	}
	defer cfh.Close()
	err = pem.Encode(cfh, c)
	if err != nil {
		return fmt.Errorf("unable to create certificate file - %s", err)
	}

	// Write to Key File
	kfh, err := os.Create(keyFile)
	if err != nil {
		return fmt.Errorf("unable to create certificate file - %s", err)
	}
	defer kfh.Close()
	err = pem.Encode(kfh, k)
	if err != nil {
		return fmt.Errorf("unable to create certificate file - %s", err)
	}

	return nil
}

// genCerts will perform the task of creating a temporary Certificate and Key.
func genCerts() (*pem.Block, *pem.Block, error) {
	// Create a Certificate Authority Cert
	ca := &x509.Certificate{
		Subject: pkix.Name{
			Organization: []string{"Never Use this Certificate in Production Inc."},
		},
		SerialNumber:          big.NewInt(42),
		NotAfter:              time.Now().Add(2 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Create a Private Key
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not generate rsa key - %s", err)
	}

	// Use CA Cert to sign a CSR and create a Public Cert
	csr := &key.PublicKey
	cert, err := x509.CreateCertificate(rand.Reader, ca, ca, csr, key)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not generate certificate - %s", err)
	}

	// Convert keys into pem.Block
	c := &pem.Block{Type: "CERTIFICATE", Bytes: cert}
	k := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return c, k, nil
}
