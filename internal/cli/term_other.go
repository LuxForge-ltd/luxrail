// SPDX-License-Identifier: MIT

//go:build !windows

package cli

import "os"

// enableVT is a no-op outside Windows: terminals there already process
// escape sequences.
func enableVT(*os.File) func() { return func() {} }
