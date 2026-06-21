package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newAgentCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent operations",
	}
	cmd.AddCommand(newAgentListCommand(alias, newCtx))
	return cmd
}

func newAgentListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var agents []api.Agent
			if err := c.GetJSON(c.URL("agents"), &agents); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(agents)
				return nil
			}
			if len(agents) == 0 {
				ctx.Println("No agents found.")
				return nil
			}
			rows := make([][]string, 0, len(agents))
			for _, a := range agents {
				rows = append(rows, []string{
					fmt.Sprintf("%d", a.ID),
					a.Name,
					client.FormatBool(!a.NoSchedule),
					a.Platform,
					a.Version,
				})
			}
			ctx.PrintTable([]string{"ID", "NAME", "SCHEDULABLE", "PLATFORM", "VERSION"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}
