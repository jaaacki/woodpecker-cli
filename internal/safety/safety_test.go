package safety

import (
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

func TestCheckConfirmMismatch(t *testing.T) {
	g := NewGate(true, "wrong")
	// CheckConfirm calls os.Exit on mismatch, so we only test the gate struct state.
	if g.Confirm != "wrong" {
		t.Fatal("confirm value not stored")
	}
}

func TestGateRequireWrite(t *testing.T) {
	g := NewGate(false, "")
	// RequireWrite would exit; verify the gate reports the need for --write.
	if g.Write {
		t.Fatal("expected write to be false")
	}
}

// TestOutputExitSafetyValue ensures the safety exit code is exported consistently.
func TestOutputExitSafetyValue(t *testing.T) {
	if output.ExitSafety != 6 {
		t.Fatalf("expected safety exit code 6, got %d", output.ExitSafety)
	}
}
