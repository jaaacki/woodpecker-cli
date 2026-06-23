package safety

import (
	"fmt"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// Gate checks whether a mutating or destructive operation is allowed.
type Gate struct {
	Write   bool
	Confirm string
}

// NewGate builds a gate from persistent flags.
func NewGate(write bool, confirm string) Gate {
	return Gate{Write: write, Confirm: confirm}
}

// CanWrite returns true if the operation is permitted to mutate state.
func (g Gate) CanWrite() bool {
	return g.Write
}

// CheckWrite fails with a JSON-aware safety error unless --write was passed.
func (g Gate) CheckWrite(ctx output.Context) bool {
	if g.Write {
		return true
	}
	ctx.Error("This command may change remote state. Re-run with --write to proceed.", output.ExitSafety)
	return false
}

// CheckConfirm fails unless --write was passed and --confirm matches the target.
func (g Gate) CheckConfirm(ctx output.Context, target string) bool {
	if !g.Write {
		ctx.Error("This command may change remote state. Re-run with --write to proceed.", output.ExitSafety)
		return false
	}
	if g.Confirm != target {
		ctx.Error(fmt.Sprintf("Confirmation mismatch: expected --confirm %q", target), output.ExitSafety)
		return false
	}
	return true
}

// RequireWrite is a convenience helper for write-gated commands.
// It returns true only when write is enabled; otherwise it exits.
func (g Gate) RequireWrite(ctx output.Context) bool {
	return g.CheckWrite(ctx)
}
