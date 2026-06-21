package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newRegistryCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Registry credential operations",
	}
	cmd.AddCommand(newRegistryListCommand(alias, newCtx))
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
