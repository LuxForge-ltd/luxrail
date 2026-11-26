// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the LAN console listener in the control host's desired state.
//
// Purpose: turns config.json's consoleLan into the ingress setting the
// agent renders into Caddy.
// Scope: the control host only, because 5050d listens on its loopback.
// Operational role: called by ingressFor for every ingress host.
package engine

import (
	"net"
	"strconv"

	"github.com/musakkka/openrail/internal/cp/orch/hub"
	"github.com/musakkka/openrail/pkg/proto"
)

// consoleLAN returns the LAN console listener for host, or nil when none is
// configured or host is not the control host.
//
// Parameters: host is the ingress host whose desired state is being built.
func (o *Orchestrator) consoleLAN(host hub.HostRow) *proto.ConsoleLAN {
	if o.o.ConsoleLAN == "" || host.MeshIndex != 1 {
		return nil
	}
	return &proto.ConsoleLAN{
		Listen: o.o.ConsoleLAN,
		Target: net.JoinHostPort("127.0.0.1", strconv.Itoa(consolePort(o.o.LocalURL))),
	}
}
