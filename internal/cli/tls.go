// SPDX-License-Identifier: MIT

package cli

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// readCAFile reads a PEM file of CA certificates, for an install whose
// console certificate comes from its own CA (Caddy's internal CA on a LAN).
// It returns only the certificates, re-encoded, so a key that sits in the same
// file is never copied into the config.
func readCAFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the CA file: %w", err)
	}
	return parseCAs(b, path)
}

func parseCAs(b []byte, name string) ([]byte, error) {
	var out []byte
	for {
		var blk *pem.Block
		blk, b = pem.Decode(b)
		if blk == nil {
			break
		}
		if strings.Contains(blk.Type, "PRIVATE KEY") {
			return nil, fmt.Errorf("%s holds a private key. Pass the CA certificate only", name)
		}
		if blk.Type != "CERTIFICATE" {
			continue
		}
		c, err := x509.ParseCertificate(blk.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s has a certificate that could not be read: %w", name, err)
		}
		if !c.IsCA {
			return nil, fmt.Errorf("%s holds %q, which is not a CA certificate. Pass the CA that signed the console's certificate", name, c.Subject.CommonName)
		}
		out = append(out, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: blk.Bytes})...)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s holds no PEM certificate", name)
	}
	return out, nil
}

// TrustCA makes the client accept certificates signed by the CAs in pemData,
// as well as the system's own roots. Verification stays on.
func (c *Client) TrustCA(pemData []byte) error {
	if len(pemData) == 0 {
		return nil
	}
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(pemData) {
		return errors.New("the stored CA certificate could not be read. Run `luxrail login` again with --ca-file")
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	c.HTTP = &http.Client{Transport: tr}
	return nil
}

// caHint adds what to do when the console's certificate comes from a CA this
// machine does not trust.
func caHint(base string, err error) error {
	var ua x509.UnknownAuthorityError
	if errors.As(err, &ua) {
		return fmt.Errorf("could not reach %s: its certificate is signed by a CA this computer does not trust. If the install uses its own CA, run `luxrail login %s --ca-file CA.crt`", base, base)
	}
	return nil
}
