// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the LAN console server in the ingress Caddy config.
//
// Purpose: serves the console on a LAN address of the control host over
// HTTPS, with a certificate for that IP from Caddy's internal CA. It
// replaces a hand-made listener: LuxRail renders Caddy's whole config from
// state on every change, so anything added through the admin API is lost,
// while this server is part of the rendered config.
// Scope: the control host's ingress, only when config.json sets consoleLan.
// Operational role: renderCaddy adds the server and the IP's certificate
// policy; the agent leaves it out while the address is not on the host.
package netw

import (
	"fmt"
	"net"

	"github.com/luxforge-ltd/luxrail/pkg/proto"
)

// consoleLANSubjects returns the certificate subject (the IP) of the LAN
// console, or nothing when there is none or its address is unusable.
//
// Parameters: lan is the ingress's LAN console setting, possibly nil.
func consoleLANSubjects(lan *proto.ConsoleLAN) []string {
	if lan == nil {
		return nil
	}
	host, _, err := net.SplitHostPort(lan.Listen)
	if err != nil || net.ParseIP(host) == nil {
		return nil
	}
	return []string{host}
}

// consoleLANServer renders the LAN console's HTTPS server, a reverse proxy
// to 5050d that sets the visitor's address the way the console relay does.
//
// Parameters: lan is the setting (nil renders nothing); delaySec keeps
// websockets open across reloads; logs is the access log setting shared
// with the other servers. Returns the server, or an error that says why it
// was left out.
func consoleLANServer(lan *proto.ConsoleLAN, delaySec int, logs obj) (obj, error) {
	if lan == nil {
		return nil, nil
	}
	if len(consoleLANSubjects(lan)) == 0 {
		return nil, fmt.Errorf("the LAN console address %q is not IP:port, so the console is not served on the LAN", lan.Listen)
	}
	if _, _, err := net.SplitHostPort(lan.Target); err != nil {
		return nil, fmt.Errorf("the LAN console target %q is not host:port, so the console is not served on the LAN", lan.Target)
	}
	return obj{
		"listen": []string{lan.Listen},
		"routes": []obj{{"handle": []obj{{
			"handler":            "reverse_proxy",
			"upstreams":          []obj{{"dial": lan.Target}},
			"stream_close_delay": fmt.Sprintf("%ds", delaySec),
			"headers": obj{"request": obj{"set": obj{
				"X-Forwarded-For": []string{"{http.request.remote.host}"},
				"X-Real-Ip":       []string{"{http.request.remote.host}"},
			}}},
		}}, "terminal": true}},
		"tls_connection_policies": []obj{{}},
		"automatic_https":         obj{"disable": true},
		"logs":                    logs,
	}, nil
}
