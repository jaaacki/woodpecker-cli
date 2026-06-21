package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newCronCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cron",
		Short: "Cron job operations",
	}
	cmd.AddCommand(newCronListCommand(alias, newCtx))
	return cmd
}

func newCronListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls <owner/repo>",
		Short: "List cron jobs for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := c.RepoID(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			var crons []api.Cron
			urlStr := c.URL("repos", fmt.Sprintf("%d", repoID), "cron")
			if err := c.GetJSON(urlStr, &crons); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(crons)
				return nil
			}
			if len(crons) == 0 {
				ctx.Println("No cron jobs found.")
				return nil
			}
			rows := make([][]string, 0, len(crons))
			for _, cr := range crons {
				rows = append(rows, []string{
					fmt.Sprintf("%d", cr.ID),
					cr.Name,
					cr.Schedule,
					cr.Branch,
				})
			}
			ctx.PrintTable([]string{"ID", "NAME", "SCHEDULE", "BRANCH"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}
