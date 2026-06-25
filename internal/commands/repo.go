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

func newRepoCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo",
		Short: "Repository operations",
	}
	cmd.AddCommand(newRepoListCommand(alias, newCtx))
	cmd.AddCommand(newRepoShowCommand(alias, newCtx))
	cmd.AddCommand(newRepoSearchCommand(alias, newCtx))
	cmd.AddCommand(newRepoLookupCommand(alias, newCtx))
	cmd.AddCommand(newRepoEnableCommand(alias, newCtx))
	cmd.AddCommand(newRepoDisableCommand(alias, newCtx))
	cmd.AddCommand(newRepoEditCommand(alias, newCtx))
	cmd.AddCommand(newRepoDeleteCommand(alias, newCtx))
	cmd.AddCommand(newRepoRepairCommand(alias, newCtx))
	cmd.AddCommand(newRepoChownCommand(alias, newCtx))
	cmd.AddCommand(newRepoMoveCommand(alias, newCtx))
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
			urlStr, err := client.SetQuery(c.URL("repos"), map[string][]string{"search": {query}})
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
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

func newRepoEnableCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "enable <forge-remote-id>",
		Short: "Enable (activate) a repository in Woodpecker",
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
			forgeRemoteID := args[0]
			urlStr, err := client.SetQuery(c.URL("repos"), map[string][]string{"forge_remote_id": {forgeRemoteID}})
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
			var repo api.Repo
			if err := c.PostJSON(urlStr, nil, &repo); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Enabled repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoDisableCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <owner/repo>",
		Short: "Disable (deactivate) a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, args[0]) {
				return nil
			}
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
			// Woodpecker deactivates a repo via DELETE with remove=false.
			urlStr, err := client.SetQuery(
				c.URL("repos", strconv.FormatInt(repo.ID, 10)),
				map[string][]string{"remove": {"false"}},
			)
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Disabled repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <owner/repo>",
		Short: "Edit repository settings",
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
			repo, err := c.ResolveRepo(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			patch := buildRepoPatch(cmd)
			urlStr := c.URL("repos", strconv.FormatInt(repo.ID, 10))
			if err := c.PatchJSON(urlStr, patch, &repo); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(repo)
				return nil
			}
			ctx.Println("Updated repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.Bool("active", false, "Activate or deactivate the repository")
	fs.Bool("allow-pr", false, "Allow pull request pipelines")
	fs.String("config-file", "", "Custom pipeline config file path")
	fs.String("default-branch", "", "Default branch name")
	fs.Int64("timeout", 0, "Pipeline timeout in minutes")
	fs.String("visibility", "", "Repository visibility (public, private, internal)")
	fs.Bool("private", false, "Mark repository as private")
	fs.Bool("protected", false, "Mark repository as protected")
	fs.StringSlice("cancel-prev", nil, "Events that cancel previous pipelines")
	fs.Bool("trusted-network", false, "Allow network access in untrusted containers")
	fs.Bool("trusted-volumes", false, "Allow volume mounts in untrusted containers")
	fs.Bool("trusted-security", false, "Allow security-escalating options")
	return cmd
}

func newRepoDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <owner/repo>",
		Short: "Delete a repository from Woodpecker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			repoFullName := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, repoFullName) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repo, err := c.ResolveRepo(repoFullName)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			urlStr, err := client.SetQuery(
				c.URL("repos", strconv.FormatInt(repo.ID, 10)),
				map[string][]string{"remove": {"true"}},
			)
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoRepairCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "repair <owner/repo>",
		Short: "Repair repository webhooks",
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
			repo, err := c.ResolveRepo(args[0])
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			urlStr := c.URL("repos", strconv.FormatInt(repo.ID, 10), "repair")
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Repaired repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoChownCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "chown <owner/repo> <user-id>",
		Short: "Change repository owner",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			repoFullName := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, repoFullName) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repo, err := c.ResolveRepo(repoFullName)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			userID := args[1]
			urlStr, err := client.SetQuery(
				c.URL("repos", strconv.FormatInt(repo.ID, 10), "chown"),
				map[string][]string{"user_id": {userID}},
			)
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Changed owner for repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func newRepoMoveCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "move <owner/repo> <to-owner/repo>",
		Short: "Move a repository to a new owner",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			repoFullName := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, repoFullName) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			repo, err := c.ResolveRepo(repoFullName)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			to := args[1]
			urlStr, err := client.SetQuery(
				c.URL("repos", strconv.FormatInt(repo.ID, 10), "move"),
				map[string][]string{"to": {to}},
			)
			if err != nil {
				ctx.Error(err.Error(), output.ExitRuntime)
				return nil
			}
			if _, err := c.Post(urlStr, nil); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Moved repository", repo.FullName)
			return nil
		},
		SilenceUsage: true,
	}
}

func writeFlagFromCmd(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("write")
	return v
}

func confirmFlagFromCmd(cmd *cobra.Command) string {
	v, _ := cmd.Flags().GetString("confirm")
	return v
}

func boolPtr(b bool) *bool {
	return &b
}

func buildRepoPatch(cmd *cobra.Command) api.RepoPatch {
	var patch api.RepoPatch
	fs := cmd.Flags()
	if fs.Changed("active") {
		v, _ := fs.GetBool("active")
		patch.Active = boolPtr(v)
	}
	if fs.Changed("allow-pr") {
		v, _ := fs.GetBool("allow-pr")
		patch.AllowPull = boolPtr(v)
	}
	if fs.Changed("private") {
		v, _ := fs.GetBool("private")
		patch.Private = boolPtr(v)
	}
	if fs.Changed("protected") {
		v, _ := fs.GetBool("protected")
		patch.Protected = boolPtr(v)
	}
	if fs.Changed("config-file") {
		v, _ := fs.GetString("config-file")
		patch.ConfigFile = &v
	}
	if fs.Changed("default-branch") {
		v, _ := fs.GetString("default-branch")
		patch.DefaultBranch = &v
	}
	if fs.Changed("timeout") {
		v, _ := fs.GetInt64("timeout")
		patch.Timeout = &v
	}
	if fs.Changed("visibility") {
		v, _ := fs.GetString("visibility")
		patch.Visibility = &v
	}
	if fs.Changed("cancel-prev") {
		v, _ := fs.GetStringSlice("cancel-prev")
		patch.CancelPrev = v
	}
	if fs.Changed("trusted-network") || fs.Changed("trusted-volumes") || fs.Changed("trusted-security") {
		patch.Trusted = &api.Trusted{}
		if fs.Changed("trusted-network") {
			v, _ := fs.GetBool("trusted-network")
			patch.Trusted.Network = &v
		}
		if fs.Changed("trusted-volumes") {
			v, _ := fs.GetBool("trusted-volumes")
			patch.Trusted.Volumes = &v
		}
		if fs.Changed("trusted-security") {
			v, _ := fs.GetBool("trusted-security")
			patch.Trusted.Security = &v
		}
	}
	return patch
}
