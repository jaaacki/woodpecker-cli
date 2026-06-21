package safety

import (
	"fmt"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// Gate checks whether a mutating or destructive operation is allowed.
// For the read-only milestone every read-only command bypasses this.
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
func (g Gate) CheckWrite(ctx output.Context) {
	if g.Write {
		return
	}
	ctx.Error("This command may change remote state. Re-run with --write to proceed.", output.ExitSafety)
}

// CheckConfirm fails unless --write was passed and --confirm matches the target.
func (g Gate) CheckConfirm(ctx output.Context, target string) {
	if !g.Write {
		ctx.Error("This command may change remote state. Re-run with --write to proceed.", output.ExitSafety)
	}
	if g.Confirm != target {
		ctx.Error(fmt.Sprintf("Confirmation mismatch: expected --confirm %q", target), output.ExitSafety)
	}
}

// RequireWrite is a convenience helper for write-gated commands in this phase.
// It returns true only when write is enabled; otherwise it exits.
func (g Gate) RequireWrite(ctx output.Context) bool {
	if !g.Write {
		ctx.Error("This command writes to the API. Pass --write to enable.", output.ExitSafety)
	}
	return true
}
