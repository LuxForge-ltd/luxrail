// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for 5050d make-private.
package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/luxforge-ltd/luxrail/internal/cp/core/config"
)

// initInstall makes a normal install in a temporary directory, with the
// sslip.io names a default install gets, and returns its config.json path.
func initInstall(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o700); err != nil {
		t.Fatal(err)
	}
	// A sparse ballast of the full size, so init does not write 1 GB.
	b, err := os.Create(filepath.Join(data, "ballast"))
	if err != nil {
		t.Fatal(err)
	}
	if err := b.Truncate(ballastSize); err != nil {
		t.Fatal(err)
	}
	_ = b.Close()
	cfg := filepath.Join(dir, "config.json")
	err = cmdInit(context.Background(), []string{"--config", cfg, "--data-dir", data, "--public-ip", "192.0.2.5",
		"--base-domain", "192-0-2-5.sslip.io", "--console-domain", "console-ab12.192-0-2-5.sslip.io"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestMakePrivateNeedsRealDomains(t *testing.T) {
	cfg := initInstall(t)
	before, _ := os.ReadFile(cfg)
	err := cmdMakePrivate(context.Background(), []string{"--config", cfg}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "nothing was changed") {
		t.Fatalf("err = %v", err)
	}
	after, _ := os.ReadFile(cfg)
	if !bytes.Equal(before, after) {
		t.Error("config.json changed although the command was refused")
	}
}

func TestMakePrivateAndUndo(t *testing.T) {
	cfg := initInstall(t)
	ctx := context.Background()
	var out bytes.Buffer
	if err := cmdMakePrivate(ctx, []string{"--config", cfg, "--base-domain", "Apps.Example.com", "--console-domain", "console.example.com"}, &out); err != nil {
		t.Fatal(err)
	}
	f, err := config.Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !f.Private || f.BaseDomain != "apps.example.com" || f.ConsoleDomain != "console.example.com" {
		t.Fatalf("config %+v", f)
	}
	p, err := openPlane(ctx, cfg, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	in, err := p.c.Install(ctx)
	p.Close()
	if err != nil || in.BaseDomain != "apps.example.com" || in.ConsoleURL != "https://console.example.com" || !in.Private {
		t.Fatalf("state %+v %v", in, err)
	}

	out.Reset()
	if err := cmdMakePrivate(ctx, []string{"--config", cfg}, &out); err != nil || !strings.Contains(out.String(), "already private") {
		t.Fatalf("second run: %v %q", err, out.String())
	}
	if err := cmdMakePrivate(ctx, []string{"--config", cfg, "--undo"}, &out); err != nil {
		t.Fatal(err)
	}
	if f, _ := config.Load(cfg); f.Private || f.BaseDomain != "apps.example.com" {
		t.Fatalf("after undo %+v", f)
	}
}
