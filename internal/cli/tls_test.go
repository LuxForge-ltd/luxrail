// SPDX-License-Identifier: MIT

package cli

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newTLSHarness serves the fake API over HTTPS with httptest's own CA, which
// no system trusts, and writes that CA to ca.crt.
func newTLSHarness(t *testing.T) (*harness, string) {
	t.Helper()
	h := &harness{t: t, f: &fake{}, dir: t.TempDir()}
	h.srv = httptest.NewTLSServer(h.f.handler(t))
	t.Cleanup(h.srv.Close)
	t.Setenv("FIFTY_CONFIG", filepath.Join(h.dir, "config.json"))
	t.Setenv("FIFTY_URL", "")
	t.Setenv("FIFTY_TOKEN", "")
	t.Setenv("FIFTY_CA_FILE", "")
	t.Setenv("LUXRAIL_CA_FILE", "")
	ca := filepath.Join(h.dir, "ca.crt")
	writePEM(t, ca, "CERTIFICATE", h.srv.Certificate().Raw)
	return h, ca
}

func writePEM(t *testing.T, path, typ string, der ...[]byte) {
	t.Helper()
	var b []byte
	for _, d := range der {
		b = append(b, pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: d})...)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoginWithCAFile(t *testing.T) {
	h, ca := newTLSHarness(t)
	code, _, errs := h.run("luxrail", "login", h.srv.URL, "--token", tok)
	if code == 0 || !strings.Contains(errs, "--ca-file") {
		t.Fatalf("untrusted CA should fail with a hint: %d %s", code, errs)
	}
	if out := h.must("login", h.srv.URL, "--token", tok, "--ca-file", ca); !strings.Contains(out, "ada@example.test") {
		t.Fatal(out)
	}
	// The CA is kept with the login, so later commands verify against it.
	if out := h.must("whoami"); !strings.Contains(out, "ada@example.test") {
		t.Fatal(out)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cfg.CAs[h.srv.URL], "BEGIN CERTIFICATE") {
		t.Errorf("CA not stored: %q", cfg.CAs)
	}
	// Logging in again without --ca-file keeps using the stored CA.
	h.must("login", h.srv.URL, "--token", tok)
	h.must("logout")
	if cfg, _ = LoadConfig(); len(cfg.CAs) != 0 {
		t.Errorf("logout kept the CA: %q", cfg.CAs)
	}
}

func TestCAFileFromEnvironment(t *testing.T) {
	h, ca := newTLSHarness(t)
	t.Setenv("LUXRAIL_URL", h.srv.URL)
	t.Setenv("LUXRAIL_TOKEN", tok)
	if code, _, errs := h.run("luxrail", "whoami"); code == 0 || !strings.Contains(errs, "does not trust") {
		t.Fatalf("whoami without the CA: %d %s", code, errs)
	}
	t.Setenv("LUXRAIL_CA_FILE", ca)
	if out := h.must("whoami"); !strings.Contains(out, "ada@example.test") {
		t.Fatal(out)
	}
	var v map[string]string
	if err := json.Unmarshal([]byte(h.must("version", "--json")), &v); err != nil {
		t.Fatal(err)
	}
}

func TestReadCAFile(t *testing.T) {
	dir := t.TempDir()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	cert := func(isCA bool) []byte {
		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"},
			NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour),
			IsCA: isCA, BasicConstraintsValid: true,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
		if err != nil {
			t.Fatal(err)
		}
		return der
	}
	keyDER, _ := x509.MarshalECPrivateKey(key)

	good := filepath.Join(dir, "ca.crt")
	writePEM(t, good, "CERTIFICATE", cert(true))
	if b, err := readCAFile(good); err != nil || !strings.Contains(string(b), "BEGIN CERTIFICATE") {
		t.Errorf("CA refused: %v", err)
	}

	cases := map[string]struct {
		typ  string
		der  []byte
		want string
	}{
		"leaf.crt": {"CERTIFICATE", cert(false), "not a CA"},
		"key.pem":  {"EC PRIVATE KEY", keyDER, "private key"},
		"junk.crt": {"NOTHING", []byte("x"), "no PEM certificate"},
	}
	for name, c := range cases {
		p := filepath.Join(dir, name)
		writePEM(t, p, c.typ, c.der)
		if _, err := readCAFile(p); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: got %v, want %q", name, err, c.want)
		}
	}
	if _, err := readCAFile(filepath.Join(dir, "missing.crt")); err == nil {
		t.Error("missing file accepted")
	}
}
