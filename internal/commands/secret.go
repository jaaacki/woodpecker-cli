package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newSecretCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Secret operations",
	}
	cmd.AddCommand(newSecretListCommand(alias, newCtx))
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
