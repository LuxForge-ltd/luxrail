// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: the address used in links LuxRail hands out.
//
// Purpose: "linkUrl": "https://luxrail.example.com" makes CLI login,
// invite, password reset, email verification and notification links start
// with that address, for example a tunnel name that works from anywhere,
// while the console keeps answering on its own name and IP. Unlike
// consoleUrl it does not change the console's name in Caddy.
// Scope: config.json only; 5050d reads it at start.
// Operational role: File.Validate calls ValidateLinkURL, and Core.LinkURL
// prefers it over the console's address when building links.
package config

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateLinkURL reports whether v is empty or an absolute https:// URL
// with a host and no user name, password, query or fragment.
//
// Parameters: v is config.json's linkUrl. Returns nil, or an error that
// says what is wrong as a sentence.
func ValidateLinkURL(v string) error {
	if v == "" {
		return nil
	}
	u, err := url.Parse(v)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		strings.Trim(u.Path, "/") != "" {
		return fmt.Errorf("the link address %q must be https:// and a host, such as https://luxrail.example.com", v)
	}
	return nil
}

// LinkBase returns the address links start with: linkUrl when set,
// otherwise "" so the caller falls back to the console's address.
func (f *File) LinkBase() string { return strings.TrimRight(f.LinkURL, "/") }
