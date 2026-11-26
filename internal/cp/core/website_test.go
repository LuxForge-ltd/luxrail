// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: tests for a service's website link.
package core

import "testing"

func TestCheckWebsite(t *testing.T) {
	good := map[string]string{
		"":                                 "",
		" https://app.example.com ":        "https://app.example.com",
		"https://app.example.com/path?q=1": "https://app.example.com/path?q=1",
		"https://app.example.com:8443/":    "https://app.example.com:8443/",
	}
	for in, want := range good {
		got, err := checkWebsite(in)
		if err != nil || got != want {
			t.Errorf("checkWebsite(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"http://app.example.com", "app.example.com", "javascript:alert(1)", "https://", "https://user:pw@app.example.com", "ftp://x.example.com", "https:app.example.com"} {
		if _, err := checkWebsite(in); err == nil {
			t.Errorf("checkWebsite(%q) accepted", in)
		}
	}
}

func TestWebsiteValues(t *testing.T) {
	yes, no := true, false
	site, empty := "https://app.example.com", ""
	cases := []struct {
		name      string
		curF      bool
		curW      string
		f         *bool
		w         *string
		wantF     bool
		wantW     string
		wantError bool
	}{
		{"mark and link", false, "", &yes, &site, true, site, false},
		{"link needs the flag", false, "", nil, &site, false, "", true},
		{"clearing the flag clears the link", true, site, &no, nil, false, "", false},
		{"link kept when only the flag is sent", true, site, &yes, nil, true, site, false},
		{"clear the link", true, site, nil, &empty, true, "", false},
	}
	for _, tc := range cases {
		f, w, err := websiteValues(tc.curF, tc.curW, tc.f, tc.w)
		if (err != nil) != tc.wantError || f != tc.wantF || w != tc.wantW {
			t.Errorf("%s: got %v, %q, %v", tc.name, f, w, err)
		}
	}
	if err := refuseWebsitePatch(map[string]any{"website": site}); err == nil {
		t.Error("a staged patch may not set the website")
	}
}
