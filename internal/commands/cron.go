package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newCronCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cron",
		Short: "Cron job operations",
	}
	cmd.AddCommand(newCronListCommand(alias, newCtx))
	cmd.AddCommand(newCronShowCommand(alias, newCtx))
	cmd.AddCommand(newCronAddCommand(alias, newCtx))
	cmd.AddCommand(newCronEditCommand(alias, newCtx))
	cmd.AddCommand(newCronDeleteCommand(alias, newCtx))
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


func newCronShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <owner/repo> <name>",
		Short: "Show a cron job",
		Args:  cobra.ExactArgs(2),
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
			var cron api.Cron
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "cron", args[1])
			if err := c.GetJSON(urlStr, &cron); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(cron)
				return nil
			}
			rows := [][]string{
				{"ID", fmt.Sprintf("%d", cron.ID)},
				{"Name", cron.Name},
				{"Schedule", cron.Schedule},
				{"Branch", cron.Branch},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newCronAddCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <owner/repo> <name>",
		Short: "Add a cron job",
		Args:  cobra.ExactArgs(2),
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
			repoID, err := c.RepoID(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			schedule, _ := cmd.Flags().GetString("schedule")
			branch, _ := cmd.Flags().GetString("branch")
			cron := api.Cron{Name: args[1], Schedule: schedule, Branch: branch}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "cron")
			var created api.Cron
			if err := c.PostJSON(urlStr, cron, &created); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(created)
				return nil
			}
			ctx.Println("Created cron", created.Name)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("schedule", "", "Cron schedule expression")
	fs.String("branch", "", "Branch to build")
	return cmd
}

func newCronEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <owner/repo> <name>",
		Short: "Edit a cron job",
		Args:  cobra.ExactArgs(2),
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
			repoID, err := c.RepoID(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			patch := api.Cron{Name: args[1]}
			fs := cmd.Flags()
			if fs.Changed("schedule") {
				v, _ := fs.GetString("schedule")
				patch.Schedule = v
			}
			if fs.Changed("branch") {
				v, _ := fs.GetString("branch")
				patch.Branch = v
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "cron", args[1])
			var updated api.Cron
			if err := c.PatchJSON(urlStr, patch, &updated); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(updated)
				return nil
			}
			ctx.Println("Updated cron", updated.Name)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("schedule", "", "Cron schedule expression")
	fs.String("branch", "", "Branch to build")
	return cmd
}

func newCronDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <owner/repo> <name>",
		Short: "Delete a cron job",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			repoFullName := args[0]
			cronName := args[1]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, repoFullName+"/"+cronName) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := c.RepoID(repoFullName)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "cron", cronName)
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted cron", cronName)
			return nil
		},
		SilenceUsage: true,
	}
}
