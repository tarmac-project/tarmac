/*
Package tlsconfig is a helper package used to create TLS configuration that adheres to best practices.

This package aims to write less repetitive code when creating TLS configurations. Opening certs, bundling certificate
authorities, configuring ciphers, etc. Just use this package to save yourself some headaches.

	// Create an instance of config
	cfg := tlsconfig.New()

	// Load Certificates
	err := cfg.CertsFromFile(cert, key)
	if err != nil {
		// do something
	}

	// Disable Host Validation
	cfg.IgnoreHostValidation()

	// Use the Config
	h := &http.Server{TLSConfig: cfg.Generate()}
*/
package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config is used to create an instance of the configuration helper. It holds the basic TLS configuration for
// repeated generation.
type Config struct {
	config tls.Config
}

// New will create a new config instance with basic TLS best practices pre-defined.
func New() *Config {
	c := Config{
		config: tls.Config{
			// Restrict to use TLS 1.2 as the minimum TLS versions
			MinVersion: tls.VersionTLS12,
			// Restrict Cipher Suites
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_AES_128_GCM_SHA256,
				tls.TLS_AES_256_GCM_SHA384,
			},
		},
	}
	return &c
}

// CertsFromFile will read the certificate and key file and create an X509 KeyPair loaded as
// Certificates. The files must contain PEM encoded data. The certificate file may contain
// intermediate certificates following the leaf certificate to form a certificate chain.
func (c *Config) CertsFromFile(cert, key string) error {
	if cert == "" || key == "" {
		return fmt.Errorf("cert and key cannot be empty")
	}

	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return fmt.Errorf("unable to load keypair - %s", err)
	}

	c.config.Certificates = append(c.config.Certificates, pair)
	return nil
}

// CAFromFile will read the PEM encoded certificate authority file and register the
// certificate as an authority for Client Authentication. This function is for m-TLS
// configuration at the server level. By default, this function sets Client
// Authentication to Require and Verify the Certificate.
func (c *Config) CAFromFile(ca string) error {
	c.config.ClientAuth = tls.RequireAndVerifyClientCert

	if ca == "" {
		return fmt.Errorf("ca cannot be empty")
	}

	b, err := os.ReadFile(ca)
	if err != nil {
		return fmt.Errorf("unable to read ca file - %s", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(b) {
		return fmt.Errorf("unable to load ca certificate - %s", err)
	}

	c.config.ClientCAs = pool
	return nil
}

// IgnoreClientCert will set client certificate authentication to verify the certificate
// only if provided. Otherwise, if no certificate is provided, the client will still be allowed.
func (c *Config) IgnoreClientCert() {
	c.config.ClientAuth = tls.VerifyClientCertIfGiven
}

// IgnoreHostValidation will turn off the hostname validation of certificates. This
// setting is dangerous and should only be used in testing.
func (c *Config) IgnoreHostValidation() {
	c.config.InsecureSkipVerify = true
}

// Generate will create a TLS configuration type based on the defaults and settings called.
// Users can run this multiple times to produce the same configuration.
func (c *Config) Generate() *tls.Config {
	return c.config.Clone()
}
