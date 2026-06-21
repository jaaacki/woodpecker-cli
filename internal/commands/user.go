package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newUserCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User and current-user operations",
	}

	cmd.AddCommand(newUserShowCommand(alias, newCtx))
	cmd.AddCommand(newUserReposCommand(alias, newCtx))
	return cmd
}

func newUserShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var user api.User
			if err := c.GetJSON(c.URL("user"), &user); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(user)
				return nil
			}
			ctx.Println(output.JSONString(user))
			return nil
		},
		SilenceUsage: true,
	}
}

func newUserReposCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "repos",
		Short: "List repositories for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var repos []api.Repo
			if err := c.GetJSON(c.URL("user", "repos"), &repos); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(repos)
				return nil
			}
			if len(repos) == 0 {
				ctx.Println("No repositories found.")
				return nil
			}
			rows := make([][]string, 0, len(repos))
			for _, r := range repos {
				rows = append(rows, []string{
					fmt.Sprintf("%d", r.ID),
					r.FullName,
					r.SCM,
					client.FormatBool(r.Active),
					r.DefaultBranch,
				})
			}
			ctx.PrintTable([]string{"ID", "REPO", "SCM", "ACTIVE", "BRANCH"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}
