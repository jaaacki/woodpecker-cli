package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// namedArg requires exactly one positional argument and names the placeholder in
// the error so the message is actionable ("wpci account add requires <alias>")
// instead of cobra's generic "accepts 1 arg(s), received 0".
func namedArg(placeholder string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("%s requires <%s>", cmd.CommandPath(), placeholder)
		}
		return nil
	}
}

// NewAccountCommand returns `wpci account ...`.
func NewAccountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage Woodpecker accounts",
		Long:  "Add, list, show, remove, and test saved Woodpecker accounts and their tokens.",
	}

	cmd.AddCommand(
		accountAddCommand(),
		accountListCommand(),
		accountShowCommand(),
		accountRemoveCommand(),
		accountTestCommand(),
		accountTokenCommand(),
	)
	return cmd
}

func accountTokenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage account bearer tokens",
		Long:  "Set, show, and remove bearer tokens stored separately from account configuration.",
	}
	cmd.AddCommand(
		accountTokenSetCommand(),
		accountTokenStatusCommand(),
		accountTokenRemoveCommand(),
	)
	return cmd
}

func accountAddCommand() *cobra.Command {
	var (
		server     string
		apiBase    string
		tlsSkip    bool
		timeout    int
		token      string
		tokenStdin bool
	)

	cmd := &cobra.Command{
		Use:   "add <alias>",
		Short: "Add or update an account",
		Args:  namedArg("alias"),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]
			ctx := NewContextFromCmd(cmd)
			if server == "" {
				ctx.Error("--server is required", output.ExitUsage)
				return nil
			}
			acct := config.Account{
				Alias:          alias,
				Server:         server,
				APIBase:        apiBase,
				Auth:           config.DefaultAuth,
				TLSSkipVerify:  tlsSkip,
				TimeoutSeconds: timeout,
			}
			if err := config.EnsureConfigDirs(); err != nil {
				ctx.Error("creating config directories: "+err.Error(), output.ExitConfig)
				return nil
			}
			if err := acct.Save(); err != nil {
				ctx.Error("saving account: "+err.Error(), output.ExitConfig)
				return nil
			}

			if tokenStdin {
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					ctx.Error("reading token from stdin: "+err.Error(), output.ExitConfig)
					return nil
				}
				token = strings.TrimSpace(line)
			}
			if token != "" {
				if err := auth.NewToken(alias).Save(token); err != nil {
					ctx.Error("saving token: "+err.Error(), output.ExitConfig)
					return nil
				}
			}

			ctx.Println("Account", alias, "saved.")
			return nil
		},
		SilenceUsage: true,
	}

	cmd.Flags().StringVar(&server, "server", "", "Woodpecker server URL (required)")
	cmd.Flags().StringVar(&apiBase, "api-base", "/api", "API base path")
	cmd.Flags().BoolVar(&tlsSkip, "tls-skip-verify", false, "Skip TLS verification")
	cmd.Flags().IntVar(&timeout, "timeout", config.DefaultTimeoutSeconds, "HTTP timeout in seconds")
	cmd.Flags().StringVar(&token, "token", "", "Bearer token")
	cmd.Flags().BoolVar(&tokenStdin, "token-stdin", false, "Read token from stdin")
	_ = cmd.MarkFlagRequired("server")
	return cmd
}

func accountListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List configured accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			aliases, err := config.ListAccounts()
			if err != nil {
				ctx.Error("listing accounts: "+err.Error(), output.ExitConfig)
				return nil
			}
			if ctx.JSON {
				ctx.Data(aliases)
				return nil
			}
			if len(aliases) == 0 {
				ctx.Println("No accounts configured.")
				return nil
			}
			for _, alias := range aliases {
				acct, err := config.LoadAccount(alias)
				if err != nil {
					continue
				}
				tok := auth.NewToken(alias)
				status := tok.Status()
				ctx.Printf("%s\t%s\t%s\n", alias, acct.Server, status)
			}
			return nil
		},
		SilenceUsage: true,
	}
}

func accountShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <alias>",
		Short: "Show account configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			acct, err := config.LoadAccount(args[0])
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			if ctx.JSON {
				ctx.Data(acct.SanitizeForDisplay())
				return nil
			}
			ctx.Println(output.JSONString(acct.SanitizeForDisplay()))
			return nil
		},
		SilenceUsage: true,
	}
}

func accountRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <alias>",
		Short: "Remove an account and its token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := args[0]
			if err := config.RemoveAccount(alias); err != nil && !os.IsNotExist(err) {
				ctx.Error("removing account: "+err.Error(), output.ExitConfig)
				return nil
			}
			_ = auth.NewToken(alias).Remove()
			ctx.Println("Account", alias, "removed.")
			return nil
		},
		SilenceUsage: true,
	}
}

func accountTestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "test <alias>",
		Short: "Test account token against the server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := args[0]
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var user api.User
			if err := c.GetJSON(c.URL("user"), &user); err != nil {
				code := client.ExitForError(err)
				ctx.Error(err.Error(), code)
				return nil
			}
			version := probeVersion(c)
			if ctx.JSON {
				ctx.Data(map[string]any{
					"alias":   alias,
					"server":  c.Account.Server,
					"user":    user,
					"version": version,
				})
				return nil
			}
			if version.Available {
				ctx.Println("Account", alias, "OK -", version.Value.Source, version.Value.Version)
			} else {
				ctx.Println("Account", alias, "OK - version unavailable ("+version.Note+")")
			}
			return nil
		},
		SilenceUsage: true,
	}
}

func accountTokenSetCommand() *cobra.Command {
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "set <alias> [token]",
		Short: "Set the bearer token for an account",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := args[0]
			var value string
			if len(args) == 2 {
				value = args[1]
			} else if fromStdin {
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					ctx.Error("reading token: "+err.Error(), output.ExitConfig)
					return nil
				}
				value = strings.TrimSpace(line)
			} else {
				ctx.Error("provide token as argument or --stdin", output.ExitUsage)
				return nil
			}
			if value == "" {
				ctx.Error("token cannot be empty", output.ExitUsage)
				return nil
			}
			if _, err := config.LoadAccount(alias); err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			if err := auth.NewToken(alias).Save(value); err != nil {
				ctx.Error("saving token: "+err.Error(), output.ExitConfig)
				return nil
			}
			ctx.Println("Token saved for", alias)
			return nil
		},
		SilenceUsage: true,
	}
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read token from stdin")
	return cmd
}

func accountTokenStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status <alias>",
		Short: "Show token status for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := args[0]
			status := auth.NewToken(alias).Status()
			if ctx.JSON {
				ctx.Data(map[string]string{"alias": alias, "status": status})
				return nil
			}
			ctx.Println(alias, status)
			return nil
		},
		SilenceUsage: true,
	}
}

func accountTokenRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rm <alias>",
		Short: "Remove the bearer token for an account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := args[0]
			if err := auth.NewToken(alias).Remove(); err != nil && !os.IsNotExist(err) {
				ctx.Error("removing token: "+err.Error(), output.ExitConfig)
				return nil
			}
			ctx.Println("Token removed for", alias)
			return nil
		},
		SilenceUsage: true,
	}
}
