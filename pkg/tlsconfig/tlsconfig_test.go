package tlsconfig

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/madflojo/testcerts"
)

type TestCase struct {
	name          string
	ca            string
	caPass        bool
	cert          string
	certsPass     bool
	generateCerts bool
	key           string
}

func TestTLSConfig(t *testing.T) {
	var tc []TestCase

	tc = append(tc, TestCase{
		name:          "Happy Path",
		ca:            "/tmp/cert",
		caPass:        true,
		cert:          "/tmp/cert",
		certsPass:     true,
		generateCerts: true,
		key:           "/tmp/key",
	})

	tc = append(tc, TestCase{
		name:          "Empty CA",
		ca:            "",
		caPass:        false,
		cert:          "/tmp/cert",
		certsPass:     true,
		generateCerts: true,
		key:           "/tmp/key",
	})

	tc = append(tc, TestCase{
		name:          "Empty Cert Values",
		ca:            "",
		caPass:        false,
		cert:          "",
		certsPass:     false,
		generateCerts: false,
		key:           "",
	})

	tc = append(tc, TestCase{
		name:          "No CA",
		ca:            "/tmp/nope",
		caPass:        false,
		cert:          "/tmp/cert",
		certsPass:     true,
		generateCerts: true,
		key:           "/tmp/key",
	})

	tc = append(tc, TestCase{
		name:          "No Certs",
		ca:            "/tmp/cert",
		caPass:        false,
		cert:          "/tmp/cert",
		certsPass:     false,
		generateCerts: false,
		key:           "/tmp/key",
	})

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			b := New()

			t.Run("Base Config", func(t *testing.T) {
				bc := b.Generate()

				if bc.MinVersion != tls.VersionTLS12 {
					t.Errorf("Unexpected TLS version set as minimum")
				}

				if bc.InsecureSkipVerify {
					t.Errorf("Unexpected value for Host verification")
				}
			})

			t.Run("WithCerts", func(t *testing.T) {
				// Generate Certs
				if c.generateCerts {
					err := testcerts.GenerateCertsToFile(c.cert, c.key)
					if err != nil {
						t.Fatalf("Could not generate certs for test - %s", err)
					}
					defer os.Remove(c.cert)
					defer os.Remove(c.key)
				}

				err := b.CertsFromFile(c.cert, c.key)
				if err != nil && c.certsPass {
					t.Fatalf("Unable to generate with certs - %s", err)
				}
				if !c.certsPass && err == nil {
					t.Fatalf("Unexpected success Loading certs from a file")
				}

				t.Run("WithCA", func(t *testing.T) {
					err := b.CAFromFile(c.ca)
					if err != nil && c.caPass {
						t.Fatalf("Unable to generate with CA - %s", err)
					}
					if !c.caPass && err == nil {
						t.Fatalf("Unexpected success Loading ca from file")
					}

					bc := b.Generate()
					if bc.ClientAuth != tls.RequireAndVerifyClientCert {
						t.Errorf("Unexpected default client auth value")
					}
				})
			})

			t.Run("IgnoreHostVerification", func(t *testing.T) {
				b.IgnoreHostValidation()
				bc := b.Generate()
				if !bc.InsecureSkipVerify {
					t.Errorf("Unexpected value for Host verification")
				}
			})

			t.Run("IgnoreClientCert", func(t *testing.T) {
				b.IgnoreClientCert()
				bc := b.Generate()
				if bc.ClientAuth != tls.VerifyClientCertIfGiven {
					t.Errorf("Unexpected value for Client Auth")
				}
			})
		})
	}
}
