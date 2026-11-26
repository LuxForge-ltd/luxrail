// SPDX-License-Identifier: MIT

// Command LuxRail is the LuxRail command line client. Installed as `railway`
// it also accepts Railway's verbs and flags.
package main

import (
	"os"

	"github.com/luxforge-ltd/luxrail/internal/cli"
)

func main() { os.Exit(cli.Main()) }
