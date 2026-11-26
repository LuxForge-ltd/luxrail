// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/luxforge-ltd/luxrail/pkg/api"
)

// termSize is the size of the terminal on out, or 80x24 when it has none.
func termSize(out *os.File) (cols, rows int) {
	if out != nil {
		if w, h, err := term.GetSize(int(out.Fd())); err == nil && w > 0 && h > 0 {
			return w, h
		}
	}
	return 80, 24
}

// rawTerminal puts the terminal on in into raw mode, so keys such as Tab,
// arrows and Ctrl+C go to the remote shell, and lets out show its escape
// sequences (Windows consoles need that turned on). It works on Windows as
// well as macOS and Linux; the returned func puts both back.
func rawTerminal(in, out *os.File) (func(), error) {
	st, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return nil, err
	}
	undoVT := enableVT(out)
	return func() {
		undoVT()
		_ = term.Restore(int(in.Fd()), st)
	}, nil
}

// watchResize sends the terminal's new size to the remote shell when it
// changes. Polling works the same on every OS (Windows has no SIGWINCH).
func watchResize(ctx context.Context, ws *websocket.Conn, out *os.File, cols, rows int) {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		c, r := termSize(out)
		if c == cols && r == rows {
			continue
		}
		cols, rows = c, r
		b, _ := json.Marshal(api.ShellControl{Resize: &[2]int{cols, rows}})
		if ws.Write(ctx, websocket.MessageText, b) != nil {
			return
		}
	}
}

func (a *App) cmdSSH() *cobra.Command {
	var replica, session string
	c := &cobra.Command{
		Use:   "ssh [-- CMD]",
		Short: "Shell inside a running replica (WebSocket exec)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			t, err := a.resolve(ctx, true)
			if err != nil {
				return err
			}
			tty := a.interactive() && len(args) == 0
			inF, _ := a.In.(*os.File)
			outF, _ := a.Out.(*os.File)
			cols, rows := 80, 24
			if tty {
				cols, rows = termSize(outF)
			}
			ws, err := t.C.Dial(ctx, svcPath(t)+"/shell"+q("replica", replica, "cols", strconv.Itoa(cols), "rows", strconv.Itoa(rows), "session", session))
			if err != nil {
				return err
			}
			defer ws.CloseNow()
			if tty {
				if restore, err := rawTerminal(inF, outF); err == nil {
					defer restore()
				}
				go watchResize(ctx, ws, outF, cols, rows)
			}
			if len(args) > 0 {
				line := strings.Join(args, " ") + "\nexit $?\n"
				if err := ws.Write(ctx, websocket.MessageBinary, []byte(line)); err != nil {
					return err
				}
			}
			go func() {
				buf := make([]byte, 32*1024)
				for {
					n, err := a.In.Read(buf)
					if n > 0 {
						if ws.Write(ctx, websocket.MessageBinary, buf[:n]) != nil {
							return
						}
					}
					if err != nil {
						return
					}
				}
			}()
			for {
				typ, data, err := ws.Read(ctx)
				if err != nil {
					if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
						return nil
					}
					if ctx.Err() != nil {
						return nil
					}
					return errors.New("the shell connection closed")
				}
				if typ == websocket.MessageBinary {
					_, _ = a.Out.Write(data)
					continue
				}
				var ctl api.ShellControl
				if json.Unmarshal(data, &ctl) == nil && ctl.Exit != nil {
					if *ctl.Exit != 0 {
						return &exitError{code: *ctl.Exit}
					}
					return nil
				}
			}
		},
	}
	c.Flags().StringVar(&replica, "replica", "", "replica id (default: the first running one)")
	c.Flags().StringVar(&session, "session", "", "session name to reattach (kept for 10 minutes)")
	return c
}
