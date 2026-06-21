package commands

import (
	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// NewDoctorCommand returns `wpci <alias> doctor` or `wpci doctor`. The alias
// resolver is injected because the standalone `wpci doctor` form is not used in
// normal operation.
func NewDoctorCommand(aliasResolver func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate account and token against the server",
		Long:  "Calls /version and /user to verify connectivity, token validity, and account configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := NewContextFromCmd(cmd)
			alias := ""
			if aliasResolver != nil {
				alias = aliasResolver()
			}
			if alias == "" {
				ctx.Error("doctor requires an account alias", output.ExitUsage)
				return nil
			}

			c, err := client.New(alias, ctx)
			if err != nil {
				code := output.ExitConfig
				if _, ok := err.(api.APIError); ok {
					code = output.ExitAPI
				}
				ctx.Error(err.Error(), code)
				return nil
			}

			var version api.Version
			// /api/version is best-effort: Woodpecker 3.x dropped it (the route
			// returns the SPA), so a version probe failure is reported as
			// "unavailable" but does not fail the health check. Auth is verified
			// via /user below.
			versionErr := c.GetJSON(c.URL("version"), &version)
			versionUnavailable := versionErr != nil

			var user api.User
			userErr := c.GetJSON(c.URL("user"), &user)

			result := map[string]any{
				"alias":  alias,
				"server": c.Account.Server,
				"version": map[string]any{
					"ok":          versionErr == nil,
					"value":       version,
					"unavailable": versionUnavailable,
					"error":       errString(versionErr),
				},
				"user": map[string]any{
					"ok":    userErr == nil,
					"value": user,
					"error": errString(userErr),
				},
			}

			if userErr != nil {
				if ctx.JSON {
					ctx.Data(result)
					return nil
				}
				ctx.Println("Account", alias, "has problems")
				ctx.Println("  /user:", userErr.Error())
				return nil
			}

			if ctx.JSON {
				ctx.Data(result)
				return nil
			}
			ctx.Println("Account", alias, "is healthy")
			ctx.Println("  Server:", c.Account.Server)
			if versionUnavailable {
				ctx.Println("  Version: unavailable (Woodpecker 3.x exposes no /version endpoint)")
			} else {
				ctx.Println("  Version:", version.Source, version.Version)
			}
			ctx.Println("  User:", user.Login, user.Email)
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
