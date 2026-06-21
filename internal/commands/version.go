package commands

import (
	"github.com/spf13/cobra"
)

// versionString is set by the root package via linker or constant.
var versionString = "0.1.0"

// NewVersionCommand returns `wpci version`.
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			if ctx.JSON {
				ctx.Data(map[string]string{"version": versionString})
				return nil
			}
			ctx.Println("wpci version", versionString)
			return nil
		},
		SilenceUsage: true,
	}
}

// SetVersion allows the root package to override the compiled version string.
func SetVersion(v string) {
	versionString = v
}
