package commands

import (
	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newQueueCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue",
		Short: "Queue operations",
	}
	cmd.AddCommand(newQueueInfoCommand(alias, newCtx))
	cmd.AddCommand(newQueuePauseCommand(alias, newCtx))
	cmd.AddCommand(newQueueResumeCommand(alias, newCtx))
	return cmd
}

func newQueueInfoCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show queue statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var info api.QueueInfo
			if err := c.GetJSON(c.URL("queue", "info"), &info); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(info)
				return nil
			}
			rows := [][]string{
				{"Workers", client.FormatNumber(int64(info.Stats.Workers))},
				{"Pending", client.FormatNumber(int64(info.Stats.Pending))},
				{"WaitingOnDeps", client.FormatNumber(int64(info.Stats.WaitingOnDeps))},
				{"Running", client.FormatNumber(int64(info.Stats.Running))},
				{"Total", client.FormatNumber(int64(info.Stats.Total))},
				{"Paused", client.FormatBool(info.Stats.Paused)},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newQueuePauseCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "pause",
		Short: "Pause the global queue",
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
			if _, err := c.Post(c.URL("queue", "pause"), nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Queue paused")
			return nil
		},
		SilenceUsage: true,
	}
}

func newQueueResumeCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "resume",
		Short: "Resume the global queue",
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
			if _, err := c.Post(c.URL("queue", "resume"), nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Queue resumed")
			return nil
		},
		SilenceUsage: true,
	}
}
