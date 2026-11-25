// SPDX-License-Identifier: AGPL-3.0-only

package engine

import "testing"

func TestReleaseVersion(t *testing.T) {
	cases := map[string]bool{
		"v0.1.9":                   true,
		"0.1.9":                    true,
		"v0.1.9-lf.1":              true,
		"v1.20.300-lf.42":          true,
		"v0.1.9-rc.1":              false,
		"v0.1.9-lf":                false,
		"v0.1.9-lf.x":              false,
		"v0.1.10-snapshot.408e8d2": false,
		"v0.1.9-lf.1-dirty":        false,
		"dev":                      false,
		"v0.1.9-dev":               false,
		"v0.1":                     false,
		"v0.1.9.4":                 false,
		"":                         false,
	}
	for v, want := range cases {
		if got := releaseVersion(v); got != want {
			t.Errorf("releaseVersion(%q) = %v, want %v", v, got, want)
		}
	}
}
