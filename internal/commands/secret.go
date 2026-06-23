package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newSecretCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret operations",
	}
	cmd.AddCommand(newSecretListCommand(alias, newCtx))
	cmd.AddCommand(newSecretShowCommand(alias, newCtx))
	cmd.AddCommand(newSecretAddCommand(alias, newCtx))
	cmd.AddCommand(newSecretEditCommand(alias, newCtx))
	cmd.AddCommand(newSecretDeleteCommand(alias, newCtx))
	return cmd
}

func newSecretListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls global|org <org>|repo <owner/repo>",
		Short: "List secrets by scope",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}

			scope := args[0]
			var urlStr string
			switch scope {
			case "global":
				urlStr = c.URL("secrets")
			case "org":
				if len(args) < 2 {
					ctx.Error("org scope requires an organization name", output.ExitUsage)
					return nil
				}
				orgID, err := c.OrgID(args[1])
				if err != nil {
					ctx.Error(err.Error(), client.ExitForError(err))
					return nil
				}
				urlStr = c.URL("orgs", fmt.Sprintf("%d", orgID), "secrets")
			case "repo":
				if len(args) < 2 {
					ctx.Error("repo scope requires owner/repo", output.ExitUsage)
					return nil
				}
				repoID, err := c.RepoID(args[1])
				if err != nil {
					ctx.Error(err.Error(), client.ExitForError(err))
					return nil
				}
				urlStr = c.URL("repos", fmt.Sprintf("%d", repoID), "secrets")
			default:
				ctx.Error("scope must be global, org, or repo", output.ExitUsage)
				return nil
			}

			var secrets []api.Secret
			if err := c.GetJSON(urlStr, &secrets); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(secrets)
				return nil
			}
			if len(secrets) == 0 {
				ctx.Println("No secrets found.")
				return nil
			}
			rows := make([][]string, 0, len(secrets))
			for _, s := range secrets {
				rows = append(rows, []string{fmt.Sprintf("%d", s.ID), s.Name})
			}
			ctx.PrintTable([]string{"ID", "NAME"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

// secretScopeURL resolves the base scope path and secret name for a secret scope.
// args is the command args excluding the secret name (i.e. [global|org <org>|repo <owner/repo>]).
func secretScopeURL(c *client.Client, args []string) (scope string, name string, err error) {
	if len(args) < 1 {
		return "", "", fmt.Errorf("scope must be global, org, or repo")
	}
	switch args[0] {
	case "global":
		if len(args) != 2 {
			return "", "", fmt.Errorf("global scope requires a secret name")
		}
		return "secrets", args[1], nil
	case "org":
		if len(args) != 3 {
			return "", "", fmt.Errorf("org scope requires an org and a secret name")
		}
		orgID, err := c.OrgID(args[1])
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("orgs/%d/secrets", orgID), args[2], nil
	case "repo":
		if len(args) != 3 {
			return "", "", fmt.Errorf("repo scope requires an owner/repo and a secret name")
		}
		repoID, err := c.RepoID(args[1])
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("repos/%d/secrets", repoID), args[2], nil
	default:
		return "", "", fmt.Errorf("scope must be global, org, or repo")
	}
}

func newSecretShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show global|org <org>|repo <owner/repo> <name>",
		Short: "Show a secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			scope, name, err := secretScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			var secret api.Secret
			urlStr := c.URL(scope, name)
			if err := c.GetJSON(urlStr, &secret); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(secret)
				return nil
			}
			rows := [][]string{
				{"ID", fmt.Sprintf("%d", secret.ID)},
				{"Name", secret.Name},
				{"Images", fmt.Sprintf("%v", secret.Images)},
				{"Events", fmt.Sprintf("%v", secret.Events)},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newSecretAddCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add global|org <org>|repo <owner/repo> <name>",
		Short: "Add a secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			scope, name, err := secretScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			value, err := secretValueFromFlags(cmd)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			images, _ := cmd.Flags().GetStringSlice("image")
			events, _ := cmd.Flags().GetStringSlice("event")
			secret := api.Secret{Name: name, Value: value, Images: images, Events: events}
			urlStr := c.URL(scope)
			var created api.Secret
			if err := c.PostJSON(urlStr, secret, &created); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(created)
				return nil
			}
			ctx.Println("Created secret", created.Name)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("value", "", "Secret value (use --value-stdin to avoid shell history)")
	fs.Bool("value-stdin", false, "Read secret value from stdin")
	fs.StringSlice("image", nil, "Limit secret to images")
	fs.StringSlice("event", nil, "Limit secret to events")
	return cmd
}

func newSecretEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit global|org <org>|repo <owner/repo> <name>",
		Short: "Edit a secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			scope, name, err := secretScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			patch := api.Secret{Name: name}
			fs := cmd.Flags()
			if fs.Changed("value") || fs.Changed("value-stdin") {
				value, err := secretValueFromFlags(cmd)
				if err != nil {
					ctx.Error(err.Error(), output.ExitUsage)
					return nil
				}
				patch.Value = value
			}
			if fs.Changed("image") {
				patch.Images, _ = fs.GetStringSlice("image")
			}
			if fs.Changed("event") {
				patch.Events, _ = fs.GetStringSlice("event")
			}
			urlStr := c.URL(scope, name)
			var updated api.Secret
			if err := c.PatchJSON(urlStr, patch, &updated); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(updated)
				return nil
			}
			ctx.Println("Updated secret", updated.Name)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("value", "", "Secret value (use --value-stdin to avoid shell history)")
	fs.Bool("value-stdin", false, "Read secret value from stdin")
	fs.StringSlice("image", nil, "Limit secret to images")
	fs.StringSlice("event", nil, "Limit secret to events")
	return cmd
}

func newSecretDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete global|org <org>|repo <owner/repo> <name>",
		Short: "Delete a secret",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			scope, name, err := secretScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL(scope, name)
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted secret", name)
			return nil
		},
		SilenceUsage: true,
	}
}

func secretValueFromFlags(cmd *cobra.Command) (string, error) {
	fs := cmd.Flags()
	stdin, _ := fs.GetBool("value-stdin")
	if stdin {
		reader := bufio.NewReader(os.Stdin)
		value, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("reading value from stdin: %w", err)
		}
		return value, nil
	}
	value, _ := fs.GetString("value")
	return value, nil
}
