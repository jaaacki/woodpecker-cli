package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newRepoCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository operations",
	}
	cmd.AddCommand(newRepoListCommand(alias, newCtx))
	cmd.AddCommand(newRepoShowCommand(alias, newCtx))
	cmd.AddCommand(newRepoSearchCommand(alias, newCtx))
	cmd.AddCommand(newRepoLookupCommand(alias, newCtx))
	return cmd
}

func newRepoListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var repos []api.Repo
			if err := c.GetJSON(c.URL("repos"), &repos); err != nil {
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

func newRepoShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <owner/repo>",
		Short: "Show repository details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repo, err := c.ResolveRepo(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(repo)
				return nil
			}
			ctx.Println(output.JSONString(repo))
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoSearchCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			query := args[0]
			urlStr := client.SetQuery(c.URL("repos"), map[string][]string{"search": {query}})
			var repos []api.Repo
			if err := c.GetJSON(urlStr, &repos); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(repos)
				return nil
			}
			if len(repos) == 0 {
				ctx.Println("No repositories match.")
				return nil
			}
			rows := make([][]string, 0, len(repos))
			for _, r := range repos {
				rows = append(rows, []string{fmt.Sprintf("%d", r.ID), r.FullName, r.SCM})
			}
			ctx.PrintTable([]string{"ID", "REPO", "SCM"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoLookupCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "lookup <owner/repo>",
		Short: "Look up a repository by full name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repo, err := c.ResolveRepo(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(repo)
				return nil
			}
			rows := [][]string{
				{"ID", fmt.Sprintf("%d", repo.ID)},
				{"FullName", repo.FullName},
				{"Owner", repo.Owner},
				{"Name", repo.Name},
				{"SCM", repo.SCM},
				{"Active", client.FormatBool(repo.Active)},
				{"DefaultBranch", repo.DefaultBranch},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}
