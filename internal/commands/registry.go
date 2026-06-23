package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newRegistryCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Registry credential operations",
	}
	cmd.AddCommand(newRegistryListCommand(alias, newCtx))
	cmd.AddCommand(newRegistryShowCommand(alias, newCtx))
	cmd.AddCommand(newRegistryAddCommand(alias, newCtx))
	cmd.AddCommand(newRegistryEditCommand(alias, newCtx))
	cmd.AddCommand(newRegistryDeleteCommand(alias, newCtx))
	return cmd
}

func newRegistryListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls global|org <org>|repo <owner/repo>",
		Short: "List registries by scope",
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
				urlStr = c.URL("registries")
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
				urlStr = c.URL("orgs", fmt.Sprintf("%d", orgID), "registries")
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
				urlStr = c.URL("repos", fmt.Sprintf("%d", repoID), "registries")
			default:
				ctx.Error("scope must be global, org, or repo", output.ExitUsage)
				return nil
			}

			var registries []api.Registry
			if err := c.GetJSON(urlStr, &registries); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(registries)
				return nil
			}
			if len(registries) == 0 {
				ctx.Println("No registries found.")
				return nil
			}
			rows := make([][]string, 0, len(registries))
			for _, r := range registries {
				rows = append(rows, []string{fmt.Sprintf("%d", r.ID), r.Address, r.Username})
			}
			ctx.PrintTable([]string{"ID", "ADDRESS", "USERNAME"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}


// registryScopeURL resolves the base scope path and registry address for a registry scope.
func registryScopeURL(c *client.Client, args []string) (scope string, address string, err error) {
	if len(args) < 1 {
		return "", "", fmt.Errorf("scope must be global, org, or repo")
	}
	switch args[0] {
	case "global":
		if len(args) != 2 {
			return "", "", fmt.Errorf("global scope requires an address")
		}
		return "registries", args[1], nil
	case "org":
		if len(args) != 3 {
			return "", "", fmt.Errorf("org scope requires an org and an address")
		}
		orgID, err := c.OrgID(args[1])
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("orgs/%d/registries", orgID), args[2], nil
	case "repo":
		if len(args) != 3 {
			return "", "", fmt.Errorf("repo scope requires an owner/repo and an address")
		}
		repoID, err := c.RepoID(args[1])
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("repos/%d/registries", repoID), args[2], nil
	default:
		return "", "", fmt.Errorf("scope must be global, org, or repo")
	}
}

func newRegistryShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show global|org <org>|repo <owner/repo> <address>",
		Short: "Show a registry credential",
		Args:  cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			scope, address, err := registryScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			var registry api.Registry
			urlStr := c.URL(scope, address)
			if err := c.GetJSON(urlStr, &registry); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(registry)
				return nil
			}
			rows := [][]string{
				{"ID", fmt.Sprintf("%d", registry.ID)},
				{"Address", registry.Address},
				{"Username", registry.Username},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRegistryAddCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add global|org <org>|repo <owner/repo> <address>",
		Short: "Add a registry credential",
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
			scope, address, err := registryScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			registry := api.Registry{Address: address, Username: username, Password: password}
			urlStr := c.URL(scope)
			var created api.Registry
			if err := c.PostJSON(urlStr, registry, &created); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(created)
				return nil
			}
			ctx.Println("Created registry", created.Address)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("username", "", "Registry username")
	fs.String("password", "", "Registry password")
	return cmd
}

func newRegistryEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit global|org <org>|repo <owner/repo> <address>",
		Short: "Edit a registry credential",
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
			scope, address, err := registryScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			patch := api.Registry{Address: address}
			fs := cmd.Flags()
			if fs.Changed("username") {
				patch.Username, _ = fs.GetString("username")
			}
			if fs.Changed("password") {
				patch.Password, _ = fs.GetString("password")
			}
			urlStr := c.URL(scope, address)
			var updated api.Registry
			if err := c.PatchJSON(urlStr, patch, &updated); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(updated)
				return nil
			}
			ctx.Println("Updated registry", updated.Address)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("username", "", "Registry username")
	fs.String("password", "", "Registry password")
	return cmd
}

func newRegistryDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete global|org <org>|repo <owner/repo> <address>",
		Short: "Delete a registry credential",
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
			scope, address, err := registryScopeURL(c, args)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL(scope, address)
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted registry", address)
			return nil
		},
		SilenceUsage: true,
	}
}
