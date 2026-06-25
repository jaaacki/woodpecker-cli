package safety

import (
	"bytes"
	"testing"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func TestGateCanWrite(t *testing.T) {
	g := NewGate(true, "")
	if !g.CanWrite() {
		t.Fatal("write gate should allow writes")
	}
}

func TestGateCannotWrite(t *testing.T) {
	g := NewGate(false, "")
	if g.CanWrite() {
		t.Fatal("write gate should block writes")
	}
}

func TestCheckWriteAllows(t *testing.T) {
	g := NewGate(true, "")
	ctx := output.Context{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if !g.CheckWrite(ctx) {
		t.Fatal("expected CheckWrite to allow when write is true")
	}
}

func TestCheckConfirmAllows(t *testing.T) {
	g := NewGate(true, "target")
	ctx := output.Context{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if !g.CheckConfirm(ctx, "target") {
		t.Fatal("expected CheckConfirm to allow when confirm matches target")
	}
}

// TestOutputExitSafetyValue ensures the safety exit code is exported consistently.
func TestOutputExitSafetyValue(t *testing.T) {
	if output.ExitSafety != 6 {
		t.Fatalf("expected safety exit code 6, got %d", output.ExitSafety)
	}
}
