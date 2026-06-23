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

func newAgentCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent operations",
	}
	cmd.AddCommand(newAgentListCommand(alias, newCtx))
	cmd.AddCommand(newAgentShowCommand(alias, newCtx))
	cmd.AddCommand(newAgentCreateCommand(alias, newCtx))
	cmd.AddCommand(newAgentEditCommand(alias, newCtx))
	cmd.AddCommand(newAgentDeleteCommand(alias, newCtx))
	cmd.AddCommand(newAgentTasksCommand(alias, newCtx))
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

func newAgentShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			id := args[0]
			var agent api.Agent
			urlStr := c.URL("agents", id)
			if err := c.GetJSON(urlStr, &agent); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(agent)
				return nil
			}
			rows := [][]string{
				{"ID", fmt.Sprintf("%d", agent.ID)},
				{"Name", agent.Name},
				{"Platform", agent.Platform},
				{"Backend", agent.Backend},
				{"Capacity", fmt.Sprintf("%d", agent.Capacity)},
				{"NoSchedule", client.FormatBool(agent.NoSchedule)},
				{"Version", agent.Version},
			}
			ctx.PrintTable([]string{"KEY", "VALUE"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newAgentCreateCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new agent",
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
			name, _ := cmd.Flags().GetString("name")
			platform, _ := cmd.Flags().GetString("platform")
			backend, _ := cmd.Flags().GetString("backend")
			capacity, _ := cmd.Flags().GetInt64("capacity")
			noSchedule, _ := cmd.Flags().GetBool("no-schedule")
			agent := api.Agent{
				Name:       name,
				Platform:   platform,
				Backend:    backend,
				Capacity:   capacity,
				NoSchedule: noSchedule,
			}
			var created api.Agent
			if err := c.PostJSON(c.URL("agents"), agent, &created); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(created)
				return nil
			}
			ctx.Println("Created agent", created.ID)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("name", "", "Agent name")
	fs.String("platform", "", "Agent platform")
	fs.String("backend", "", "Agent backend")
	fs.Int64("capacity", 0, "Agent capacity")
	fs.Bool("no-schedule", false, "Disable scheduling on this agent")
	return cmd
}

func newAgentEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an agent",
		Args:  cobra.ExactArgs(1),
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
			id := args[0]
			patch := api.Agent{}
			fs := cmd.Flags()
			if fs.Changed("name") {
				patch.Name, _ = fs.GetString("name")
			}
			if fs.Changed("platform") {
				patch.Platform, _ = fs.GetString("platform")
			}
			if fs.Changed("backend") {
				patch.Backend, _ = fs.GetString("backend")
			}
			if fs.Changed("capacity") {
				patch.Capacity, _ = fs.GetInt64("capacity")
			}
			if fs.Changed("no-schedule") {
				patch.NoSchedule, _ = fs.GetBool("no-schedule")
			}
			var updated api.Agent
			urlStr := c.URL("agents", id)
			if err := c.PatchJSON(urlStr, patch, &updated); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(updated)
				return nil
			}
			ctx.Println("Updated agent", updated.ID)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("name", "", "Agent name")
	fs.String("platform", "", "Agent platform")
	fs.String("backend", "", "Agent backend")
	fs.Int64("capacity", 0, "Agent capacity")
	fs.Bool("no-schedule", false, "Disable scheduling on this agent")
	return cmd
}

func newAgentDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			id := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, id) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			urlStr := c.URL("agents", id)
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted agent", id)
			return nil
		},
		SilenceUsage: true,
	}
}

func newAgentTasksCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "tasks <id>",
		Short: "List tasks running on an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			id := args[0]
			var tasks []api.Task
			urlStr := c.URL("agents", id, "tasks")
			if err := c.GetJSON(urlStr, &tasks); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(tasks)
				return nil
			}
			if len(tasks) == 0 {
				ctx.Println("No tasks found.")
				return nil
			}
			rows := make([][]string, 0, len(tasks))
			for _, t := range tasks {
				rows = append(rows, []string{
					strconv.FormatInt(t.ID, 10),
					t.RepositoryName,
					t.Status,
				})
			}
			ctx.PrintTable([]string{"ID", "REPOSITORY", "STATUS"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}
