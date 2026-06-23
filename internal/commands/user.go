package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newUserCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "User and current-user operations",
	}

	cmd.AddCommand(newUserShowCommand(alias, newCtx))
	cmd.AddCommand(newUserReposCommand(alias, newCtx))
	cmd.AddCommand(newUserListCommand(alias, newCtx))
	cmd.AddCommand(newUserCreateCommand(alias, newCtx))
	cmd.AddCommand(newUserEditCommand(alias, newCtx))
	cmd.AddCommand(newUserDeleteCommand(alias, newCtx))
	cmd.AddCommand(newUserTokenCommand(alias, newCtx))
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


func newUserListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List users (admin)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var users []api.User
			if err := c.GetJSON(c.URL("users"), &users); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(users)
				return nil
			}
			if len(users) == 0 {
				ctx.Println("No users found.")
				return nil
			}
			rows := make([][]string, 0, len(users))
			for _, u := range users {
				rows = append(rows, []string{fmt.Sprintf("%d", u.ID), u.Login, u.Email})
			}
			ctx.PrintTable([]string{"ID", "LOGIN", "EMAIL"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}

func newUserCreateCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <login>",
		Short: "Create a user (admin)",
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
			login := args[0]
			email, _ := cmd.Flags().GetString("email")
			user := api.User{Login: login, Email: email}
			var created api.User
			if err := c.PostJSON(c.URL("users"), user, &created); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(created)
				return nil
			}
			ctx.Println("Created user", created.Login)
			return nil
		},
		SilenceUsage: true,
	}
	cmd.Flags().String("email", "", "User email")
	return cmd
}

func newUserEditCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <login>",
		Short: "Edit a user (admin)",
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
			login := args[0]
			patch := api.User{Login: login}
			fs := cmd.Flags()
			if fs.Changed("email") {
				patch.Email, _ = fs.GetString("email")
			}
			if fs.Changed("active") {
				patch.Active, _ = fs.GetBool("active")
			}
			if fs.Changed("admin") {
				patch.Admin, _ = fs.GetBool("admin")
			}
			var updated api.User
			urlStr := c.URL("users", login)
			if err := c.PatchJSON(urlStr, patch, &updated); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(updated)
				return nil
			}
			ctx.Println("Updated user", updated.Login)
			return nil
		},
		SilenceUsage: true,
	}
	fs := cmd.Flags()
	fs.String("email", "", "User email")
	fs.Bool("active", false, "Mark user as active")
	fs.Bool("admin", false, "Grant admin privileges")
	return cmd
}

func newUserDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <login>",
		Short: "Delete a user (admin)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			login := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, login) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			urlStr := c.URL("users", login)
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted user", login)
			return nil
		},
		SilenceUsage: true,
	}
}

func newUserTokenCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Current-user token operations",
	}
	cmd.AddCommand(newUserTokenShowCommand(alias, newCtx))
	cmd.AddCommand(newUserTokenResetCommand(alias, newCtx))
	return cmd
}

func newUserTokenShowCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the current user token",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var token api.Token
			if err := c.PostJSON(c.URL("user", "token"), nil, &token); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(token)
				return nil
			}
			ctx.Println(token.Value)
			return nil
		},
		SilenceUsage: true,
	}
}

func newUserTokenResetCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset the current user token",
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
			if _, err := c.Delete(c.URL("user", "token")); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Token reset")
			return nil
		},
		SilenceUsage: true,
	}
}
