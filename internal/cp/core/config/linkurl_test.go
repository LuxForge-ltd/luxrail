// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for the link address setting.
package config

import "testing"

func TestValidateLinkURL(t *testing.T) {
	for _, ok := range []string{"", "https://luxrail.example.com", "https://luxrail.example.com/", "https://luxrail.example.com:8443"} {
		if err := ValidateLinkURL(ok); err != nil {
			t.Errorf("ValidateLinkURL(%q): %v", ok, err)
		}
	}
	for _, bad := range []string{"http://luxrail.example.com", "luxrail.example.com", "https://", "https://u:p@luxrail.example.com", "https://luxrail.example.com/console", "https://luxrail.example.com/?a=1"} {
		if err := ValidateLinkURL(bad); err == nil {
			t.Errorf("ValidateLinkURL(%q) accepted", bad)
		}
	}
	f := File{LinkURL: "https://luxrail.example.com/"}
	if got := f.LinkBase(); got != "https://luxrail.example.com" {
		t.Errorf("LinkBase() = %q", got)
	}
}
