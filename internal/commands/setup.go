package commands

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/auth"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/config"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// NewSetupCommand returns `wpci setup`: a one-shot bootstrap that stores a
// Woodpecker account (server + token) and writes a `wpci-<alias>` wrapper script
// onto PATH. The wrapper is a real script, not a shell alias, so it is inherited
// by non-interactive shells and AI agents (which never load .zshrc aliases).
func NewSetupCommand() *cobra.Command {
	var (
		server     string
		token      string
		tokenStdin bool
		alias      string
		binDir     string
		apiBase    string
		tlsSkip    bool
		timeout    int
	)
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure an account and install a wpci-<alias> wrapper",
		Long: `setup stores a Woodpecker account (server + token) and writes a small
wpci-<alias> wrapper script into a directory on PATH so the account is usable
from interactive shells, scripts, and AI agents. Agents do not inherit shell
aliases, so the wrapper is a script (exec wpci <alias>) rather than an alias.

Token via stdin is recommended:

  printf '%s' "$WPCI_TOKEN" | wpci setup --server https://ci.example.com --token-stdin

After setup:

  wpci-<alias> doctor
  wpci-<alias> repo ls`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			if alias == "" {
				alias = "local"
			}
			if server == "" {
				ctx.Error("--server is required", output.ExitUsage)
				return nil
			}

			tok := strings.TrimSpace(token)
			if tokenStdin {
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil && line == "" {
					ctx.Error("reading token from stdin: "+err.Error(), output.ExitConfig)
					return nil
				}
				tok = strings.TrimSpace(line)
			}
			if tok == "" {
				ctx.Error("token is required (--token or --token-stdin)", output.ExitUsage)
				return nil
			}

			if apiBase == "" {
				apiBase = "/api"
			}
			if timeout <= 0 {
				timeout = 30
			}
			if binDir == "" {
				binDir = defaultBinDir()
			}

			if err := config.EnsureConfigDirs(); err != nil {
				ctx.Error("creating config directories: "+err.Error(), output.ExitConfig)
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
			if err := acct.Save(); err != nil {
				ctx.Error("saving account: "+err.Error(), output.ExitConfig)
				return nil
			}
			if err := auth.NewToken(alias).Save(tok); err != nil {
				ctx.Error("saving token: "+err.Error(), output.ExitConfig)
				return nil
			}

			wrapperPath := filepath.Join(binDir, "wpci-"+alias)
			content := "#!/bin/sh\n" +
				"# wpci-" + alias + " - runs `wpci " + alias + "` (account configured via `wpci setup`).\n" +
				"# A real script (not a shell alias) so non-interactive shells and AI agents inherit it.\n" +
				"exec wpci " + alias + " \"$@\"\n"
			if err := os.MkdirAll(binDir, 0o755); err != nil {
				ctx.Error("creating bin dir "+binDir+": "+err.Error(), output.ExitConfig)
				return nil
			}
			if err := os.WriteFile(wrapperPath, []byte(content), 0o755); err != nil {
				ctx.Error("writing wrapper "+wrapperPath+": "+err.Error(), output.ExitConfig)
				return nil
			}

			// Verify the account works against the server.
			status := "verified"
			if c, err := client.New(alias, ctx); err == nil {
				var u api.User
				if err := c.GetJSON(c.URL("user"), &u); err != nil {
					status = "configured (verify failed: " + err.Error() + ")"
				} else if u.Login != "" {
					status = "verified (user: " + u.Login + ")"
				}
			} else {
				status = "configured (verify failed: " + err.Error() + ")"
			}

			ctx.Println("Account", alias, "configured:", server)
			ctx.Println("  Wrapper:", wrapperPath)
			ctx.Println("  Token:  ", auth.NewToken(alias).Status())
			ctx.Println("  Status: ", status)
			ctx.Println("")
			ctx.Println("Next:")
			ctx.Println("  wpci-" + alias + " doctor")
			ctx.Println("  wpci-" + alias + " repo ls")
			return nil
		},
		SilenceUsage: true,
	}
	flags := cmd.Flags()
	flags.StringVar(&server, "server", "", "Woodpecker server URL (required)")
	flags.StringVar(&alias, "alias", "local", "account alias (wrapper named wpci-<alias>)")
	flags.StringVar(&binDir, "bin-dir", "", "directory to install the wrapper (default first writable of $HOME/bin, $HOME/.local/bin)")
	flags.StringVar(&token, "token", "", "bearer token (prefer --token-stdin)")
	flags.BoolVar(&tokenStdin, "token-stdin", false, "read token from stdin")
	flags.StringVar(&apiBase, "api-base", "/api", "API base path")
	flags.BoolVar(&tlsSkip, "tls-skip-verify", false, "skip TLS verification")
	flags.IntVar(&timeout, "timeout", 30, "HTTP timeout in seconds")
	return cmd
}

// defaultBinDir returns the first writable bin directory on the user's PATH,
// preferring $HOME/bin then $HOME/.local/bin.
func defaultBinDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	for _, c := range []string{filepath.Join(home, "bin"), filepath.Join(home, ".local", "bin")} {
		if isWritableDir(c) {
			return c
		}
	}
	return filepath.Join(home, "bin")
}

func isWritableDir(dir string) bool {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return false
	}
	probe := filepath.Join(abs, ".wpci-write-probe")
	if err := os.WriteFile(probe, []byte{}, 0o644); err != nil {
		return false
	}
	_ = os.Remove(probe)
	return true
}
