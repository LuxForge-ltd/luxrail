// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the console on the bare IP of a private console name.
//
// Purpose: an install reached by a name such as console.10-200-0-5.sslip.io
// is usually also opened as https://10.200.0.5. Caddy's config is rendered
// from state on every change, so a route added by hand for the IP is lost on
// the next reboot, agent restart or domain change; this route is rendered
// with the rest, with a certificate for the IP from Caddy's internal CA.
// Scope: console names that can only point at a private address
// (routes.PrivateIPName); public names get no IP route.
// Operational role: ingressFor adds the route next to the console's own.
package engine

import (
	"github.com/luxforge-ltd/luxrail/internal/cp/orch/routes"
	"github.com/luxforge-ltd/luxrail/pkg/proto"
)

// consoleIPRoute returns the console route for the private IP written in
// the console's name, or nil when the name is not a private wildcard name.
//
// Parameters: domain is the console's host name; port is 5050d's port.
func consoleIPRoute(domain string, port int) *proto.Route {
	ip := routes.PrivateIPName(domain)
	if ip == "" {
		return nil
	}
	return &proto.Route{Domain: ip, UpstreamName: consoleUpstream, Port: port, TLS: "internal"}
}
