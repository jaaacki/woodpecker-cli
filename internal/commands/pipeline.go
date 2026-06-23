package commands

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newPipelineCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Pipeline operations",
	}
	cmd.AddCommand(newPipelineListCommand(alias, newCtx))
	cmd.AddCommand(newPipelineLastCommand(alias, newCtx))
	cmd.AddCommand(newPipelineShowCommand(alias, newCtx))
	cmd.AddCommand(newPipelineConfigCommand(alias, newCtx))
	cmd.AddCommand(newPipelineMetadataCommand(alias, newCtx))
	cmd.AddCommand(newPipelinePsCommand(alias, newCtx))
	cmd.AddCommand(newPipelineLogCommand(alias, newCtx))
	cmd.AddCommand(newPipelineRunCommand(alias, newCtx))
	cmd.AddCommand(newPipelineRestartCommand(alias, newCtx))
	cmd.AddCommand(newPipelineApproveCommand(alias, newCtx))
	cmd.AddCommand(newPipelineDeclineCommand(alias, newCtx))
	cmd.AddCommand(newPipelineCancelCommand(alias, newCtx))
	return cmd
}

func resolveRepoID(c *client.Client, fullName string, ctx output.Context) (int64, error) {
	return c.RepoID(fullName)
}

func parsePipelineNumber(s string, ctx output.Context) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid pipeline number: %s", s)
	}
	return n, nil
}

func newPipelineListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	var branch string
	cmd := &cobra.Command{
		Use:   "ls <owner/repo>",
		Short: "List pipelines for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			params := url.Values{}
			if branch != "" {
				params.Set("branch", branch)
			}
			urlStr := client.SetQuery(c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines"), params)
			var pipelines []api.Pipeline
			if err := c.GetJSON(urlStr, &pipelines); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(pipelines)
				return nil
			}
			if len(pipelines) == 0 {
				ctx.Println("No pipelines found.")
				return nil
			}
			rows := make([][]string, 0, len(pipelines))
			for _, p := range pipelines {
				rows = append(rows, []string{
					fmt.Sprintf("%d", p.Number),
					p.Event,
					p.Branch,
					p.Status,
					client.FormatTime(p.Created),
				})
			}
			ctx.PrintTable([]string{"NUMBER", "EVENT", "BRANCH", "STATUS", "CREATED"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVar(&branch, "branch", "", "Filter by branch")
	return cmd
}

func newPipelineLastCommand(alias string, newCtx ContextFactory) *cobra.Command {
	var branch string
	cmd := &cobra.Command{
		Use:   "last <owner/repo>",
		Short: "Show the latest pipeline for a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			params := url.Values{}
			if branch != "" {
				params.Set("branch", branch)
			}
			urlStr := client.SetQuery(c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines"), params)
			var pipelines []api.Pipeline
			if err := c.GetJSON(urlStr, &pipelines); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if len(pipelines) == 0 {
				ctx.Println("No pipelines found.")
				return nil
			}
			pipeline := pipelines[0]
			if ctx.JSON {
				ctx.Data(pipeline)
				return nil
			}
			ctx.Println(output.JSONString(pipeline))
			return nil
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVar(&branch, "branch", "", "Filter by branch")
	return cmd
}

func newPipelineShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <owner/repo> <number>",
		Short: "Show a pipeline by number",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			var pipeline api.Pipeline
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10))
			if err := c.GetJSON(urlStr, &pipeline); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(pipeline)
				return nil
			}
			ctx.Println(output.JSONString(pipeline))
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineConfigCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "config <owner/repo> <number>",
		Short: "Show compiled pipeline YAML configuration",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			var cfgs []api.PipelineConfig
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10), "config")
			if err := c.GetJSON(urlStr, &cfgs); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			var combined strings.Builder
			for _, c2 := range cfgs {
				if combined.Len() > 0 {
					combined.WriteString("\n---\n")
				}
				combined.WriteString(c2.Data)
			}
			cfgData := combined.String()
			if ctx.Raw {
				ctx.RawBytes([]byte(cfgData))
				return nil
			}
			if ctx.JSON {
				ctx.Data(cfgs)
				return nil
			}
			ctx.Println(cfgData)
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineMetadataCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "metadata <owner/repo> <number>",
		Short: "Show pipeline metadata",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			var meta api.PipelineMetadata
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10), "metadata")
			if err := c.GetJSON(urlStr, &meta); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(meta)
				return nil
			}
			ctx.Println(output.JSONString(meta))
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelinePsCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ps <owner/repo> <number>",
		Short: "List workflow steps for a pipeline",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			steps, err := fetchSteps(c, repoID, number)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(steps)
				return nil
			}
			if len(steps) == 0 {
				ctx.Println("No steps found.")
				return nil
			}
			rows := make([][]string, 0, len(steps))
			for _, s := range steps {
				rows = append(rows, []string{
					fmt.Sprintf("%d", s.ID),
					s.Name,
					s.State,
					client.FormatTime(s.Started),
					client.FormatTime(s.Stopped),
					fmt.Sprintf("%d", s.ExitCode),
				})
			}
			ctx.PrintTable([]string{"ID", "NAME", "STATE", "STARTED", "STOPPED", "EXIT"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

// fetchSteps returns the flattened step list for a pipeline. In Woodpecker 3.x
// the steps are embedded under each workflow's Children, so this GETs the
// pipeline and flattens them.
func fetchSteps(c *client.Client, repoID, number int64) ([]api.Step, error) {
	var p api.Pipeline
	urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10))
	if err := c.GetJSON(urlStr, &p); err != nil {
		return nil, err
	}
	var steps []api.Step
	for _, wf := range p.Workflows {
		steps = append(steps, wf.Children...)
	}
	return steps, nil
}

func newPipelineLogCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Pipeline log operations",
	}
	cmd.AddCommand(newPipelineLogShowCommand(alias, newCtx))
	return cmd
}

func newPipelineLogShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show <owner/repo> <number> <step>",
		Short: "Show logs for a pipeline step",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			stepName := args[2]

			steps, err := fetchSteps(c, repoID, number)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			var stepID int64 = -1
			for _, s := range steps {
				if s.Name == stepName {
					stepID = s.ID
					break
				}
			}
			if stepID < 0 {
				ctx.Error("step not found: "+stepName, output.ExitUsage)
				return nil
			}

			var logs []api.Log
			logsURL := c.URL("repos", strconv.FormatInt(repoID, 10), "logs", strconv.FormatInt(number, 10), strconv.FormatInt(stepID, 10))
			if err := c.GetJSON(logsURL, &logs); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(logs)
				return nil
			}
			if ctx.Raw {
				for _, line := range logs {
					ctx.RawBytes(line.Data)
				}
				return nil
			}
			for _, line := range logs {
				ctx.Printf("%s", string(line.Data))
			}
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineRunCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <owner/repo>",
		Short: "Trigger a new pipeline",
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
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			branch, _ := cmd.Flags().GetString("branch")
			vars, _ := cmd.Flags().GetStringToString("var")
			opts := api.PipelineOptions{Branch: branch, Variables: vars}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines")
			var pipeline api.Pipeline
			if err := c.PostJSON(urlStr, opts, &pipeline); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(pipeline)
				return nil
			}
			ctx.Println("Triggered pipeline", pipeline.Number)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("branch", "", "Branch to build (defaults to repository default branch)")
	fs.StringToString("var", nil, "Pipeline variables (KEY=VALUE)")
	return cmd
}

func newPipelineRestartCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "restart <owner/repo> <number>",
		Short: "Restart an existing pipeline",
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
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10))
			var pipeline api.Pipeline
			if err := c.PostJSON(urlStr, nil, &pipeline); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(pipeline)
				return nil
			}
			ctx.Println("Restarted pipeline", pipeline.Number)
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineApproveCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "approve <owner/repo> <number>",
		Short: "Approve a blocked pipeline",
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
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10), "approve")
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Approved pipeline", number)
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineDeclineCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "decline <owner/repo> <number>",
		Short: "Decline a blocked pipeline",
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
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10), "decline")
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Declined pipeline", number)
			return nil
		},
		SilenceUsage: true,
	}
}

func newPipelineCancelCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel <owner/repo> <number>",
		Short: "Cancel a running or pending pipeline",
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
			repoID, err := resolveRepoID(c, args[0], ctx)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			number, err := parsePipelineNumber(args[1], ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitUsage)
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repoID, 10), "pipelines", strconv.FormatInt(number, 10), "cancel")
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Cancelled pipeline", number)
			return nil
		},
		SilenceUsage: true,
	}
}
