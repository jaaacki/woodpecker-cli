package commands

import (
	"github.com/spf13/cobra"

	"github.com/jaaacki/woodpecker-cli/internal/api"
	"github.com/jaaacki/woodpecker-cli/internal/client"
	"github.com/jaaacki/woodpecker-cli/internal/output"
)

// apiParity holds the advisory API coverage counters used by doctor --json.
type apiParity struct {
	Implemented int `json:"implemented"`
	Spec        int `json:"spec"`
}

// doctorResult is the structured doctor report.
type doctorResult struct {
	Ok                 bool     `json:"ok"`
	Server             string   `json:"server"`
	Version            any      `json:"version"`
	User               any      `json:"user"`
	WriteSupport       bool     `json:"write_support"`
	OpenAPIParityScore apiParity `json:"openapi_parity_score"`
	CompatibilityNotes []string `json:"compatibility_notes"`
}

// NewDoctorCommand returns `wpci <alias> doctor` or `wpci doctor`. The alias
// resolver is injected because the standalone `wpci doctor` form is not used in
// normal operation.
func NewDoctorCommand(aliasResolver func() string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate account and token against the server",
		Long:  "Calls /version and /user to verify connectivity, token validity, and account configuration, and reports API parity.",
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
				if ctx.JSON {
					ctx.Data(map[string]any{
						"ok":                   false,
						"error":                err.Error(),
						"write_support":        true,
						"openapi_parity_score": parityScore(),
						"compatibility_notes":  compatibilityNotes(),
					})
					return nil
				}
				ctx.Error(err.Error(), output.ExitConfig)
				return nil
			}

			var version api.Version
			var versionUnavailable bool
			versionErr := c.GetJSON(c.URL("version"), &version)
			if versionErr != nil {
				var apiErr api.APIError
				if apiErrOk := api.AsAPIError(versionErr); apiErrOk && apiErr.NotFound() {
					versionUnavailable = true
					versionErr = nil
				}
			}

			var user api.User
			userErr := c.GetJSON(c.URL("user"), &user)

			result := doctorResult{
				Ok:                 versionErr == nil && userErr == nil,
				Server:             c.Account.Server,
				Version:            versionOrNote(version, versionUnavailable, versionErr),
				User:               userOrError(user, userErr),
				WriteSupport:       true,
				OpenAPIParityScore: parityScore(),
				CompatibilityNotes: compatibilityNotes(),
			}

			if !result.Ok {
				if ctx.JSON {
					ctx.Data(result)
					return nil
				}
				ctx.Println("Account", alias, "has problems")
				if versionErr != nil {
					ctx.Println("  /version:", versionErr.Error())
				}
				if userErr != nil {
					ctx.Println("  /user:", userErr.Error())
				}
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
			ctx.Println("  Write support: enabled")
			ctx.Println("  API parity:", result.OpenAPIParityScore.Implemented, "/", result.OpenAPIParityScore.Spec)
			return nil
		},
		SilenceUsage: true,
	}
	return cmd
}

func versionOrNote(version api.Version, unavailable bool, err error) any {
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	if unavailable {
		return map[string]any{"ok": true, "value": "unavailable", "note": "Woodpecker 3.x exposes no /version endpoint"}
	}
	return map[string]any{"ok": true, "source": version.Source, "version": version.Version}
}

func userOrError(user api.User, err error) any {
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	return map[string]any{"ok": true, "value": user}
}

func parityScore() apiParity {
	// Advisory counters: total paths in the Woodpecker 3.x OpenAPI spec, and the
	// number currently implemented by wpci.
	return apiParity{Implemented: 45, Spec: 65}
}

func compatibilityNotes() []string {
	return []string{
		"Targets Woodpecker 3.x endpoints; older servers may return 404 for /version and new write paths.",
		"Write commands require --write; destructive commands additionally require --confirm <target>.",
	}
}
