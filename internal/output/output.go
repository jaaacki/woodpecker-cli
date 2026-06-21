package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
)

// Exit codes mirror the PDR. The full table includes safety (6), but this
// read-only phase only needs 0-5.
const (
	ExitSuccess   = 0
	ExitRuntime   = 1
	ExitUsage     = 2
	ExitConfig    = 3
	ExitAuth      = 4
	ExitAPI       = 5
	ExitSafety    = 6
	ExitWriteGate = 6 // alias for clarity
	ExitCancel    = 1 // user cancelled confirmation
)

// Context carries output preferences down a command invocation.
type Context struct {
	JSON bool
	Raw  bool
	Out  io.Writer
	Err  io.Writer
}

// NewContext returns a context writing to stdout/stderr.
func NewContext() Context {
	return Context{Out: os.Stdout, Err: os.Stderr}
}

// NewJSONContext returns a context with JSON enabled.
func NewJSONContext() Context {
	return Context{JSON: true, Out: os.Stdout, Err: os.Stderr}
}

// Println writes a line of text output when not in JSON/raw mode.
func (c Context) Println(a ...any) {
	if c.JSON || c.Raw {
		return
	}
	fmt.Fprintln(c.Out, a...)
}

// Printf writes formatted text output when not in JSON/raw mode.
func (c Context) Printf(format string, a ...any) {
	if c.JSON || c.Raw {
		return
	}
	fmt.Fprintf(c.Out, format, a...)
}

// PrintTable renders tab-separated rows as a table when not in JSON/raw mode.
func (c Context) PrintTable(headers []string, rows [][]string) {
	if c.JSON || c.Raw {
		return
	}
	w := tabwriter.NewWriter(c.Out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

// Data wraps a Go value in the JSON success envelope and writes it.
func (c Context) Data(v any) {
	if c.Raw {
		// Raw output should have already been streamed by the caller.
		return
	}
	if !c.JSON {
		return
	}
	enc := json.NewEncoder(c.Out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"ok":   true,
		"data": v,
	})
}

// RawBytes writes raw bytes directly when --raw is set.
func (c Context) RawBytes(b []byte) {
	if !c.Raw {
		return
	}
	if _, err := c.Out.Write(b); err == nil && len(b) > 0 && b[len(b)-1] != '\n' {
		if syncer, ok := c.Out.(interface{ Sync() error }); ok {
			_ = syncer.Sync()
		}
	}
}

// JSONError writes a structured error envelope and exits with the supplied code.
func (c Context) JSONError(kind, message string, code int) {
	enc := json.NewEncoder(c.Err)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(map[string]any{
		"ok": false,
		"error": map[string]any{
			"kind":    kind,
			"message": message,
		},
	})
	os.Exit(code)
}

// Error prints a human or JSON error and exits.
func (c Context) Error(message string, code int) {
	if c.JSON {
		kind := errorKind(code)
		c.JSONError(kind, message, code)
		return
	}
	fmt.Fprintln(c.Err, "Error:", message)
	os.Exit(code)
}

// Errorf formats and prints an error message, then exits.
func (c Context) Errorf(format string, a ...any) {
	c.Error(fmt.Sprintf(format, a...), ExitRuntime)
}

// errorKind maps an exit code to a stable JSON error kind.
func errorKind(code int) string {
	switch code {
	case ExitUsage:
		return "usage_error"
	case ExitConfig:
		return "config_error"
	case ExitAuth:
		return "auth_error"
	case ExitAPI:
		return "api_error"
	case ExitSafety:
		return "safety_error"
	default:
		return "runtime_error"
	}
}

// Fatal is a convenience alias for Error with a runtime exit code.
func (c Context) Fatal(message string) {
	c.Error(message, ExitRuntime)
}

// JSONString returns a compact JSON representation of v.
func JSONString(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\":\"marshal failed: %s\"}", err)
	}
	return string(b)
}

// Marshal writes a pretty-printed JSON value to out.
func Marshal(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
