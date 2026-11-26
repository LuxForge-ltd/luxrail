// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: validation of the LAN console address in config.json.
//
// Purpose: "consoleLan": "10.10.180.5:8443" serves the console on that LAN
// address over HTTPS (Caddy's internal CA), so it can be reached by IP
// without a public name. Caddy's whole config is rendered from state, so
// unlike a route added by hand through Caddy's admin API it survives every
// rebuild.
// Scope: config.json only; 5050d reads it at start.
// Operational role: File.Validate calls ValidateConsoleLAN, so a bad value
// stops 5050d with a sentence instead of producing a Caddy config that
// cannot load.
package config

import (
	"fmt"
	"net"
	"strconv"
)

// ValidateConsoleLAN reports whether v is empty or an IPv4 address with a
// port that Caddy's own listeners do not use.
//
// Parameters: v is config.json's consoleLan. Returns nil, or an error that
// says what is wrong as a sentence.
func ValidateConsoleLAN(v string) error {
	if v == "" {
		return nil
	}
	host, port, err := net.SplitHostPort(v)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || ip.To4() == nil || ip.IsUnspecified() || ip.IsLoopback() {
		return fmt.Errorf("the LAN console address %q must be one IPv4 address of this host with a port, such as 10.0.0.5:8443", v)
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return fmt.Errorf("the LAN console address %q has no valid port", v)
	}
	switch n {
	case 80, 443, 5050, 5051:
		return fmt.Errorf("the LAN console address %q uses port %d, which LuxRail already uses; pick another, such as 8443", v, n)
	}
	return nil
}
