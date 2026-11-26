// SPDX-License-Identifier: MIT

package cli

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableVT turns on escape-sequence processing for the console on out, so a
// remote shell's colours, cursor moves and full-screen programs draw properly,
// and returns a func that puts the previous mode back.
func enableVT(out *os.File) func() {
	if out == nil {
		return func() {}
	}
	h := windows.Handle(out.Fd())
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return func() {}
	}
	want := mode | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING | windows.DISABLE_NEWLINE_AUTO_RETURN
	if want == mode || windows.SetConsoleMode(h, want) != nil {
		return func() {}
	}
	return func() { _ = windows.SetConsoleMode(h, mode) }
}
