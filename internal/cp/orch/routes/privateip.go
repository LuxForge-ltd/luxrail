// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: names that can only ever point at a private address.
//
// Purpose: wildcard DNS names such as console.10-200-0-5.sslip.io resolve
// to the IP written in them. When that IP is private, no public CA can
// reach the host, so asking Let's Encrypt for it fails on every retry and
// leaves the name with no certificate at all. Such names, and bare IPs,
// get Caddy's internal CA instead (TLSMode).
// Scope: sslip.io and nip.io names with a dotted or dashed IPv4 address.
// Operational role: TLSMode calls PrivateIPName; the engine also serves the
// console on the bare IP of a private console name.
package routes

import (
	"net/netip"
	"strings"
)

// wildcardDNS are the services that answer a name with the IP written in it.
var wildcardDNS = []string{".sslip.io", ".nip.io"}

// PrivateIPName returns the IPv4 address written in a wildcard DNS name
// when that address is private, loopback or link-local, and "" otherwise.
//
// Parameters: domain is a lower-case host name without a trailing dot.
// Returns the address, such as "10.200.0.5" for
// "console.10-200-0-5.sslip.io", or "".
func PrivateIPName(domain string) string {
	ip, ok := wildcardIP(domain)
	if !ok || !(ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast()) {
		return ""
	}
	return ip.String()
}

// wildcardIP finds the IPv4 address in a wildcard DNS name: the last dashed
// label ("10-200-0-5") or the last four labels ("10.200.0.5") before the
// service's suffix.
func wildcardIP(domain string) (netip.Addr, bool) {
	d := strings.ToLower(strings.TrimSuffix(domain, "."))
	for _, suf := range wildcardDNS {
		rest, ok := strings.CutSuffix(d, suf)
		if !ok {
			continue
		}
		labels := strings.Split(rest, ".")
		if ip, err := netip.ParseAddr(strings.ReplaceAll(labels[len(labels)-1], "-", ".")); err == nil && ip.Is4() {
			return ip, true
		}
		if len(labels) >= 4 {
			if ip, err := netip.ParseAddr(strings.Join(labels[len(labels)-4:], ".")); err == nil && ip.Is4() {
				return ip, true
			}
		}
	}
	return netip.Addr{}, false
}
