// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: requests that already arrived over HTTPS are not redirected.
//
// Purpose: on :80 LuxRail answers every acme or internal domain with a 308
// to https://. A Cloudflare tunnel (or any proxy that terminates TLS) that
// sends its traffic to Caddy over plain HTTP therefore loops: the visitor
// asked for https, the proxy forwards http, Caddy redirects to https, and
// the proxy forwards http again. Such proxies say what the visitor used in
// X-Forwarded-Proto, so a request marked "https" is served on :80 instead
// of redirected. A request marked "http", or not marked, is still
// redirected as before.
// Scope: the plain-HTTP server of the ingress Caddy config, for domains
// that are redirected (TLS mode acme or internal). Private installs serve
// everything on :80 already and are unchanged.
// Operational role: renderCaddy adds these routes ahead of the redirect.
//
// The header is the client's claim. Believing it can only make Caddy serve
// that client over plain HTTP, which it already chose; it reaches no other
// visitor and skips no authentication.
package netw

// forwardedHTTPSRoute serves domain with handler h when the request says it
// was HTTPS before a proxy forwarded it over HTTP.
//
// Parameters: domain is the route's host name; h is the same reverse proxy
// handler the HTTPS server uses for it. Returns the Caddy route.
func forwardedHTTPSRoute(domain string, h obj) obj {
	return obj{
		"match":    []obj{{"host": []string{domain}, "header": obj{"X-Forwarded-Proto": []string{"https"}}}},
		"handle":   []obj{h},
		"terminal": true,
	}
}
