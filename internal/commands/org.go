package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

func newOrgCommand(alias string, newCtx ContextFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Organization operations",
	}
	cmd.AddCommand(newOrgListCommand(alias, newCtx))
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
