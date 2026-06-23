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

func TestCheckWriteDenies(t *testing.T) {
	g := NewGate(false, "")
	var buf bytes.Buffer
	ctx := output.Context{Out: &buf, Err: &buf}
	if g.CheckWrite(ctx) {
		t.Fatal("expected CheckWrite to deny when write is false")
	}
	if !bytes.Contains(buf.Bytes(), []byte("--write")) {
		t.Fatalf("expected safety message to mention --write, got: %s", buf.String())
	}
}

func TestCheckConfirmMatches(t *testing.T) {
	g := NewGate(true, "target")
	ctx := output.Context{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}
	if !g.CheckConfirm(ctx, "target") {
		t.Fatal("expected CheckConfirm to pass when confirm matches target")
	}
}

func TestCheckConfirmRequiresWrite(t *testing.T) {
	g := NewGate(false, "target")
	var buf bytes.Buffer
	ctx := output.Context{Out: &buf, Err: &buf}
	if g.CheckConfirm(ctx, "target") {
		t.Fatal("expected CheckConfirm to fail when write is false")
	}
}

func TestCheckConfirmMismatch(t *testing.T) {
	g := NewGate(true, "wrong")
	var buf bytes.Buffer
	ctx := output.Context{Out: &buf, Err: &buf}
	if g.CheckConfirm(ctx, "target") {
		t.Fatal("expected CheckConfirm to fail on mismatch")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Confirmation mismatch")) {
		t.Fatalf("expected mismatch message, got: %s", buf.String())
	}
}

func TestRequireWrite(t *testing.T) {
	g := NewGate(false, "")
	var buf bytes.Buffer
	ctx := output.Context{Out: &buf, Err: &buf}
	if g.RequireWrite(ctx) {
		t.Fatal("expected RequireWrite to return false when write is false")
	}
}

// TestOutputExitSafetyValue ensures the safety exit code is exported consistently.
func TestOutputExitSafetyValue(t *testing.T) {
	if output.ExitSafety != 6 {
		t.Fatalf("expected safety exit code 6, got %d", output.ExitSafety)
	}
}
