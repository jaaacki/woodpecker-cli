package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/commands"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

var (
	jsonFlag    bool
	rawFlag     bool
	writeFlag   bool
	confirmFlag string
	versionFlag bool
)

// Execute runs the root command.
func Execute() error {
	root := newRootCommand()
	addAliasCommands(root)
	return root.Execute()
}

// NewContext builds the output context from persistent flags.
func NewContext() output.Context {
	return output.Context{
		JSON: jsonFlag,
		Raw:  rawFlag,
		Out:  os.Stdout,
		Err:  os.Stderr,
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "wpci [account-alias] [api-area] [action] [args...]",
		Short: "AI-friendly Woodpecker CI CLI",
		Long: `wpci is a neutral, unofficial Woodpecker CI API client.

Usage:
  wpci account add              Add a Woodpecker account
  wpci <alias> doctor --json    Validate account and token
  wpci <alias> repo ls          List repositories
  wpci <alias> pipeline last   Show the latest pipeline`,
		Version: commands.CompiledVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Println(commands.CompiledVersion())
				return nil
			}
			_ = cmd.Help()
			return nil
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON")
	root.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw upstream response")
	root.PersistentFlags().BoolVar(&writeFlag, "write", false, "Enable write operations")
	root.PersistentFlags().StringVar(&confirmFlag, "confirm", "", "Confirm destructive operation target")
	root.Flags().BoolVar(&versionFlag, "version", false, "Print version")

	root.AddCommand(commands.NewAccountCommand())
	root.AddCommand(commands.NewVersionCommand())
	return root
}

// addAliasCommands attaches a command for every configured account. This makes
// `wpci <alias> repo ls` work naturally with Cobra.
func addAliasCommands(root *cobra.Command) {
	aliases, err := config.ListAccounts()
	if err != nil {
		return
	}
	for _, alias := range aliases {
		acct, err := config.LoadAccount(alias)
		if err != nil {
			continue
		}
		aliasCmd := &cobra.Command{
			Use:   alias,
			Short: fmt.Sprintf("Account %s (%s)", alias, acct.Server),
			Long:  fmt.Sprintf("Commands scoped to the %q Woodpecker account.", alias),
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				return nil
			},
			SilenceUsage: true,
		}
		commands.RegisterAlias(aliasCmd, alias, NewContext)
		root.AddCommand(aliasCmd)
	}
}

// ParseAlias extracts the account alias from the command path. Cobra makes the
// alias the first command segment after root.
func ParseAlias(cmd *cobra.Command) string {
	parts := strings.Split(cmd.CommandPath(), " ")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
