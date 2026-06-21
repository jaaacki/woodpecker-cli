package commands

import (
	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// NewContextFromCmd extracts output flags from a cobra command.
func NewContextFromCmd(cmd *cobra.Command) output.Context {
	ctx := output.NewContext()
	if v, err := cmd.Flags().GetBool("json"); err == nil {
		ctx.JSON = v
	}
	if v, err := cmd.Flags().GetBool("raw"); err == nil {
		ctx.Raw = v
	}
	return ctx
}
