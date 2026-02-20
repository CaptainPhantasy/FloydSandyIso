// Package terminal provides an embedded PTY terminal component for the
// Floyd TUI. It uses creack/pty to spawn a real shell and
// charmbracelet/x/vt to parse and render the output inside an
// ultraviolet screen region.
package terminal

import (
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

// Session wraps an OS-level PTY and the shell process running inside
// it. Callers read output via the io.Reader interface and write input
// via Write.
type Session struct {
	ptyFile *os.File
	cmd     *exec.Cmd
	mu      sync.Mutex
	closed  bool
}

// NewSession spawns the user's shell (from $SHELL, defaulting to
// /bin/sh) inside a PTY with the given initial size.
func NewSession(rows, cols int, cwd string) (*Session, error) {
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/sh"
	}

	c := exec.Command(shellPath)
	c.Env = os.Environ()
	c.Env = append(c.Env, "TERM=xterm-256color")
	if cwd != "" {
		c.Dir = cwd
	}

	f, err := pty.StartWithSize(c, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return nil, err
	}

	return &Session{
		ptyFile: f,
		cmd:     c,
	}, nil
}

// Read reads output from the PTY. This is called by the VT emulator
// goroutine to feed terminal data into the parser.
func (s *Session) Read(p []byte) (int, error) {
	return s.ptyFile.Read(p)
}

// Write sends input to the PTY (user keystrokes).
func (s *Session) Write(p []byte) (int, error) {
	return s.ptyFile.Write(p)
}

// Resize informs the PTY of a new window size so the shell reflows
// correctly.
func (s *Session) Resize(rows, cols int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	_ = pty.Setsize(s.ptyFile, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Close tears down the session — kills the process and closes the PTY
// file descriptor.
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	// Kill the process first, then close the PTY.
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	err := s.ptyFile.Close()
	_ = s.cmd.Wait()
	return err
}

// Closed reports whether the session has been torn down.
func (s *Session) Closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

var _ io.ReadWriter = (*Session)(nil)
