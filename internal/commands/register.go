package commands

import (
	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// ContextFactory builds an output.Context from CLI flags.
type ContextFactory func() output.Context

// RegisterAlias adds all area commands to an alias command.
func RegisterAlias(aliasCmd *cobra.Command, alias string, newCtx ContextFactory) {
	aliasCmd.AddCommand(newRepoCommand(alias, newCtx))
	aliasCmd.AddCommand(newPipelineCommand(alias, newCtx))
	aliasCmd.AddCommand(newCronCommand(alias, newCtx))
	aliasCmd.AddCommand(newSecretCommand(alias, newCtx))
	aliasCmd.AddCommand(newRegistryCommand(alias, newCtx))
	aliasCmd.AddCommand(newOrgCommand(alias, newCtx))
	aliasCmd.AddCommand(newAgentCommand(alias, newCtx))
	aliasCmd.AddCommand(newQueueCommand(alias, newCtx))
	aliasCmd.AddCommand(newUserCommand(alias, newCtx))
	aliasCmd.AddCommand(NewDoctorCommand(func() string { return alias }))
	aliasCmd.AddCommand(NewVersionCommand())
}
