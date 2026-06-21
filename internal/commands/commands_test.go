package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewContextFromCmd(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.PersistentFlags().Bool("json", false, "")
	cmd.PersistentFlags().Bool("raw", false, "")
	_ = cmd.ParseFlags([]string{"--json"})

	ctx := NewContextFromCmd(cmd)
	if !ctx.JSON {
		t.Fatal("expected json flag to be true")
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	// Verify it can be invoked without arguments.
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatal(err)
	}
}

// TestWriter is a minimal io.Writer for future command tests.
type TestWriter struct {
	bytes []byte
}

func (w *TestWriter) Write(p []byte) (int, error) {
	w.bytes = append(w.bytes, p...)
	return len(p), nil
}
