// SPDX-License-Identifier: AGPL-3.0-only

// LuxRail: a service's website link.
//
// Purpose: a service can be marked as a web frontend and given the https URL
// people reach it at, which may be an address the install itself does not
// serve (for example a Cloudflare tunnel in front of an internal domain). The
// console shows the link on the service card and opens it in a new tab.
// Operational role: the two fields live in the service config but change
// immediately, like the icon, because they never affect a deployment.
package core

import (
	"context"
	"database/sql"
	"net/url"
	"strings"

	"github.com/luxforge-ltd/luxrail/internal/cp/core/fail"
)

// maxWebsiteLen bounds the stored URL.
const maxWebsiteLen = 2048

// checkWebsite validates a website link: empty, or an absolute https URL with
// a host and no user name or password. It returns the trimmed URL.
func checkWebsite(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return "", nil
	}
	u, err := url.Parse(v)
	if err != nil || len(v) > maxWebsiteLen || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Opaque != "" {
		return "", fail.Validation("website", "The website must be an https:// address, such as https://app.example.com.")
	}
	return v, nil
}

// websiteValues works out the stored pair from the current values and a
// request: a link needs the web frontend flag, and clearing the flag clears
// the link.
func websiteValues(curFrontend bool, curSite string, frontend *bool, site *string) (bool, string, error) {
	f, w := curFrontend, curSite
	if frontend != nil {
		f = *frontend
	}
	if site != nil {
		v, err := checkWebsite(*site)
		if err != nil {
			return false, "", err
		}
		w = v
	}
	if !f {
		if site != nil && w != "" {
			return false, "", fail.Validation("website", "Mark the service as a web frontend to add a website.")
		}
		w = ""
	}
	return f, w, nil
}

// setWebsite stores the web frontend flag and website of a service instance.
func setWebsite(ctx context.Context, tx *sql.Tx, in instance, frontend *bool, site *string) error {
	f, w, err := websiteValues(in.Config.WebFrontend, in.Config.Website, frontend, site)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE service_instances SET config = json_set(config, '$.webFrontend', json(?), '$.website', ?) WHERE service_id = ? AND environment_id = ?`,
		map[bool]string{true: "true", false: "false"}[f], w, in.ServiceID, in.EnvID)
	return err
}

// refuseWebsitePatch keeps the website fields out of staged config patches,
// so every change to them goes through checkWebsite.
func refuseWebsitePatch(patch map[string]any) error {
	for _, k := range []string{"webFrontend", "website"} {
		if _, ok := patch[k]; ok {
			return fail.Validation(k, "Set the website in the service's settings; it is not part of a staged change.")
		}
	}
	return nil
}
