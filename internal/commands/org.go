package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
	"github.com/jaaacki/woodpecker-cli/internal/safety"
)

func newOrgCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization operations",
	}
	cmd.AddCommand(newOrgListCommand(alias, newCtx))
	cmd.AddCommand(newOrgDeleteCommand(alias, newCtx))
	return cmd
}

func newOrgListCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			var orgs []api.Org
			if err := c.GetJSON(c.URL("orgs"), &orgs); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			if ctx.JSON {
				ctx.Data(orgs)
				return nil
			}
			if len(orgs) == 0 {
				ctx.Println("No organizations found.")
				return nil
			}
			rows := make([][]string, 0, len(orgs))
			for _, o := range orgs {
				rows = append(rows, []string{fmt.Sprintf("%d", o.ID), o.Name})
			}
			ctx.PrintTable([]string{"ID", "NAME"}, rows)
			return nil
		},
		SilenceUsage: true,
	}
}


func newOrgDeleteCommand(alias string, newCtx ContextFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := newCtx()
			name := args[0]
			gate := safety.NewGate(writeFlagFromCmd(cmd), confirmFlagFromCmd(cmd))
			if !gate.CheckWrite(ctx) || !gate.CheckConfirm(ctx, name) {
				return nil
			}
			c, err := client.New(alias, ctx)
			if err != nil {
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}
			orgID, err := c.OrgID(name)
			if err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			urlStr := c.URL("orgs", fmt.Sprintf("%d", orgID))
			if _, err := c.Delete(urlStr); err != nil {
				ctx.Error(err.Error(), client.ExitForError(err))
				return nil
			}
			ctx.Println("Deleted org", name)
			return nil
		},
		SilenceUsage: true,
	}
}
