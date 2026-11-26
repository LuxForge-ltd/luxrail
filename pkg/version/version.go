// SPDX-License-Identifier: MIT

// Package version holds build metadata set with -ldflags at release time:
//
//	-X github.com/luxforge-ltd/luxrail/pkg/version.Version=v0.1.0
//	-X github.com/luxforge-ltd/luxrail/pkg/version.Commit=<sha>
//	-X github.com/luxforge-ltd/luxrail/pkg/version.Date=<rfc3339>
package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)
