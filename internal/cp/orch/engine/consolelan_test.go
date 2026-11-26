// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests that only the control host serves the LAN console.
package engine

import (
	"testing"

	"github.com/luxforge-ltd/luxrail/internal/cp/orch/hub"
)

func TestConsoleLANOnlyOnTheControlHost(t *testing.T) {
	o := &Orchestrator{o: Options{ConsoleLAN: "192.0.2.5:8443", LocalURL: "http://127.0.0.1:5050"}}
	lan := o.consoleLAN(hub.HostRow{MeshIndex: 1})
	if lan == nil || lan.Listen != "192.0.2.5:8443" || lan.Target != "127.0.0.1:5050" {
		t.Fatalf("control host: %+v", lan)
	}
	if lan := o.consoleLAN(hub.HostRow{MeshIndex: 2}); lan != nil {
		t.Errorf("another host: %+v", lan)
	}
	o.o.ConsoleLAN = ""
	if lan := o.consoleLAN(hub.HostRow{MeshIndex: 1}); lan != nil {
		t.Errorf("not configured: %+v", lan)
	}
}
