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
)

// Execute runs the root command.
func Execute() error {
	if dir := os.Getenv("WPCI_CONFIG_DIR"); dir != "" {
		_ = config.SetConfigDir(dir)
	}
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

Quick start:
  wpci account add              Add a Woodpecker account
  wpci <alias> doctor --json    Validate account and token
  wpci <alias> repo ls          List repositories
  wpci <alias> pipeline last   Show the latest pipeline`,
		Version: commands.CompiledVersion(),
		// ArbitraryArgs (instead of cobra's default legacyArgs) lets an
		// unrecognised first argument reach RunE so we can produce a helpful
		// "unknown account / did you mean" message instead of cobra's bare
		// `unknown command "x"`. Known subcommands still route normally.
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				_ = cmd.Help()
				return nil
			}
			return unknownCommandError(cmd, args[0])
		},
		SilenceUsage: true,
	}

	root.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output JSON")
	root.PersistentFlags().BoolVar(&rawFlag, "raw", false, "Output raw upstream response")
	root.PersistentFlags().BoolVar(&writeFlag, "write", false, "Enable write operations")
	root.PersistentFlags().StringVar(&confirmFlag, "confirm", "", "Confirm destructive operation target")

	root.AddCommand(commands.NewAccountCommand())
	root.AddCommand(commands.NewVersionCommand())
	return root
}

// addAliasCommands attaches a command for every configured account — both those
// added with `wpci account add` and any defined only via WPCI_<ALIAS>_SERVER env
// (zero-config accounts). This makes `wpci <alias> repo ls` work with Cobra.
func addAliasCommands(root *cobra.Command) {
	registered := map[string]bool{}
	stored, _ := config.ListAccounts()
	for _, alias := range stored {
		acct, err := config.LoadAccount(alias)
		if err != nil {
			continue
		}
		registerAliasCommand(root, alias, fmt.Sprintf("Account %s (%s)", alias, acct.Server))
		registered[alias] = true
	}
	for _, alias := range config.EnvAccounts() {
		// The env suffix maps non-alphanumerics to '_', so an alias like
		// "home-syno" is discovered from WPCI_HOME_SYNO_SERVER as "home_syno".
		// Register both the underscore form and the hyphen form so the user can
		// type either. ResolveAccount keys env vars off the typed name's suffix,
		// so both forms resolve to the same account.
		forms := []string{alias, strings.ReplaceAll(alias, "_", "-")}
		for _, f := range forms {
			if registered[f] {
				continue
			}
			registerAliasCommand(root, f, fmt.Sprintf("Account %s (env)", f))
			registered[f] = true
		}
	}
}

// registerAliasCommand builds and attaches the scoped command for one account
// alias. Account/token are resolved at call time (stored account or env).
func registerAliasCommand(root *cobra.Command, alias, short string) {
	aliasCmd := &cobra.Command{
		Use:   alias,
		Short: short,
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

// ParseAlias extracts the account alias from the command path. Cobra makes the
// alias the first command segment after root.
func ParseAlias(cmd *cobra.Command) string {
	parts := strings.Split(cmd.CommandPath(), " ")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// unknownCommandError explains an unrecognised first argument. It distinguishes
// "no accounts configured yet" (point the user at `account add`) from "accounts
// exist but this name is unknown" (list configured aliases, suggest typos), and
// always offers cobra's closest-match suggestions (e.g. "acount" -> "account").
func unknownCommandError(root *cobra.Command, name string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "unknown command or account %q for \"wpci\"", name)

	if sugs := root.SuggestionsFor(name); len(sugs) > 0 {
		b.WriteString("\n\nDid you mean this?\n")
		for _, s := range sugs {
			fmt.Fprintf(&b, "\t%s\n", s)
		}
	}

	accounts, _ := config.ListAccounts()
	envAccounts := config.EnvAccounts()
	if len(accounts) == 0 && len(envAccounts) == 0 {
		b.WriteString("\nNo Woodpecker accounts configured. Add one:\n")
		b.WriteString("\twpci account add <alias> --server <url> --token-stdin\n")
		b.WriteString("\nor set WPCI_<ALIAS>_SERVER and WPCI_<ALIAS>_TOKEN in your shell.\n")
	} else {
		if len(accounts) > 0 {
			b.WriteString("\nConfigured accounts:\n")
			for _, a := range accounts {
				fmt.Fprintf(&b, "\t%s\n", a)
			}
		}
		if len(envAccounts) > 0 {
			b.WriteString("\nEnv-only accounts:\n")
			for _, a := range envAccounts {
				fmt.Fprintf(&b, "\t%s\n", a)
			}
		}
		b.WriteString("\nAdd another with: wpci account add <alias> --server <url>\n")
	}
	return fmt.Errorf("%s", strings.TrimRight(b.String(), "\n"))
}
