# Product Design Record: wpci

## Problem

Existing local Woodpecker access tends to become a set of server-specific
wrappers. Those wrappers are useful, but they drift, expose only small slices of
the API, and are awkward for AI agents that need predictable commands and JSON
output.

`wpci` should be a neutral, unofficial client library and CLI that can talk to
any configured Woodpecker endpoint by account alias.

## User Ask

Design a public, release-installable Woodpecker API client:

```sh
wpci <account-alias> <api-area> <action> [params/options]
```

The priority is ease of use for humans and AI agents. Avoid unnecessary
frameworks and invented concepts. Stay close to Woodpecker's upstream API and
CLI vocabulary.

## Users

- Human operators inspecting and managing Woodpecker CI.
- AI coding agents that need reliable CI inspection and controlled operations.
- Scripts that need stable JSON output and non-interactive setup.

## Core Requirements

1. Implement in Go.
2. Expose one binary named `wpci`.
3. Multi-account configuration.
4. Credentials stored outside repos.
5. Safe token handling and redacted diagnostics.
6. Broad Woodpecker API coverage.
7. Human-readable default output.
8. `--json` output for every command.
9. Explicit write gates for mutating operations.
10. Public release installation via `install.sh` and `install.ps1`.

## Implementation Defaults

Use these defaults unless a later implementation issue has a stronger reason to
change them:

- Language: Go.
- CLI framework: Cobra.
- HTTP client: Go standard `net/http`.
- JSON: Go standard `encoding/json`.
- Config format: JSON files.
- Binary name: `wpci`.
- Module path: `github.com/jaaacki/woodpecker-cli`.
- Internal package layout:

```text
cmd/wpci/          main package
internal/cli/      Cobra command tree
internal/config/   account and token storage
internal/client/   HTTP transport, auth, pagination, errors
internal/wood/     Woodpecker API methods
internal/output/   tables, JSON, raw output
internal/safety/   --write and --confirm checks
```

Avoid code generation in v1. Use the OpenAPI document as the parity checklist,
not as a generated client source, unless a later spike proves generation is
cleaner.

## Configuration

File layout:

```text
~/.config/wpci/
  accounts/
    <alias>.json
  tokens/
    <alias>
```

Permissions:

```text
~/.config/wpci          700
~/.config/wpci/tokens/* 600
```

Account config schema:

```json
{
  "alias": "home",
  "server": "https://ci.example.com",
  "api_base": "/api",
  "auth": "bearer-token",
  "tls_skip_verify": false,
  "timeout_seconds": 30
}
```

Rules:

- `account add` creates the account file and normalizes trailing slashes from
  `server`.
- `api_base` defaults to `/api`.
- `auth` defaults to `bearer-token`.
- `timeout_seconds` defaults to `30`.
- Never store the token in the account JSON.
- `account show` may print account config but must never print token contents.
- `token status` prints present/missing and permission status only.

## CLI Shape

Provisioning:

```sh
wpci account add <alias> --server <url>
wpci account ls
wpci account show <alias>
wpci account rm <alias>
wpci account test <alias>
wpci account token set <alias> [--stdin]
wpci account token status <alias>
wpci account token rm <alias>
```

Inspection:

```sh
wpci <alias> info
wpci <alias> version
wpci <alias> whoami
wpci <alias> doctor
```

Repo and pipeline examples:

```sh
wpci <alias> repo ls
wpci <alias> repo search <text>
wpci <alias> repo show <owner/repo>
wpci <alias> repo branches <owner/repo>
wpci <alias> repo perms <owner/repo>

wpci <alias> pipeline ls <owner/repo>
wpci <alias> pipeline last <owner/repo>
wpci <alias> pipeline show <owner/repo> <number>
wpci <alias> pipeline ps <owner/repo> <number>
wpci <alias> pipeline log show <owner/repo> <number> <step>
wpci <alias> pipeline config <owner/repo> <number>
wpci <alias> pipeline metadata <owner/repo> <number>
```

Scoped resources:

```sh
wpci <alias> secret ls global
wpci <alias> secret ls org <org>
wpci <alias> secret ls repo <owner/repo>

wpci <alias> registry ls global
wpci <alias> registry ls org <org>
wpci <alias> registry ls repo <owner/repo>
```

## Safety Model

Read-only commands should work with no extra flag.

Mutating commands require `--write`:

```sh
wpci home pipeline stop jaaacki/project 123 --write
```

Destructive commands require `--write` and `--confirm <target>`:

```sh
wpci home repo rm jaaacki/old-project --write --confirm jaaacki/old-project
```

Admin commands should be available, but not easy to run accidentally.

Default gates:

- Read-only commands require no gate.
- Any `POST`, `PATCH`, or `DELETE` endpoint requires `--write`.
- Destructive deletes require `--write --confirm <target>`.
- Queue pause/resume, user mutation, agent mutation, forge mutation, and global
  secret/registry mutation require `--write --confirm <target>`.
- If `--json` is set, write-gate failures must be JSON errors.

## Output Model

Default output is concise human-readable text or tables.

Every command supports:

```sh
--json
--raw
```

Exit codes:

```text
0 success
1 generic runtime error
2 usage/argument error
3 config or credential error
4 authentication/authorization error
5 remote API error
6 safety gate error
```

Agent-friendly errors should be structured:

```json
{
  "ok": false,
  "error": {
    "kind": "auth_failed",
    "message": "HTTP 401 from /api/user",
    "hint": "Run: wpci account token set home"
  }
}
```

For `--json`, successful command output should be wrapped unless `--raw` is also
set:

```json
{
  "ok": true,
  "data": {}
}
```

`--raw` prints the upstream API response body where practical.

## Agent-Native Lessons

CLI-Anything by HKUDS is credited as the source of several agent-native CLI
design prompts considered here: https://github.com/HKUDS/CLI-Anything

Useful patterns from CLI-Anything:

- Every command must have standard `--help` discoverability.
- Every command must support `--json` for machine consumption.
- The installed CLI should be tested through subprocess calls, not only library
  unit tests.
- Setup should be deterministic: install, add account, set token, run
  `doctor --json`.
- The CLI should delegate to real Woodpecker APIs instead of reimplementing CI
  concepts.

Patterns not needed for the first version:

- A generated-harness framework.
- A plugin registry.
- A stateful REPL.
- Preview/live-preview workflows.

## API Coverage Target

Use the upstream OpenAPI specification as the coverage baseline:

- Woodpecker CI: https://woodpecker-ci.org/
- API reference: https://woodpecker-ci.org/api
- OpenAPI YAML: https://woodpecker-ci.org/redocusaurus/plugin-redoc-0.yaml
- Upstream repository: https://github.com/woodpecker-ci/woodpecker
- Swagger docs: https://woodpecker-ci.org/docs/next/development/swagger

Initial endpoint families to cover:

- user
- version
- repos
- pipelines
- logs and log streams
- cron
- secrets
- registries
- orgs
- users
- agents
- queue

Command-to-endpoint mapping should live in code comments or docs near the
implementation. Use these endpoint mappings for the first pass:

```text
whoami/info                 GET /user
version                     GET /version
repo ls                     GET /repos
repo show                   GET /repos/lookup/{repo_full_name}, fallback GET /repos
repo branches               GET /repos/{repo_id}/branches
repo perms                  GET /repos/{repo_id}/permissions
pipeline ls                 GET /repos/{repo_id}/pipelines
pipeline show               GET /repos/{repo_id}/pipelines/{pipeline_number}
pipeline config             GET /repos/{repo_id}/pipelines/{pipeline_number}/config
pipeline metadata           GET /repos/{repo_id}/pipelines/{pipeline_number}/metadata
pipeline log show           GET /repos/{repo_id}/logs/{pipeline_number}/{step_id}
cron ls                     GET /repos/{repo_id}/cron
secret ls global            GET /secrets
secret ls org               GET /orgs/{org_id}/secrets
secret ls repo              GET /repos/{repo_id}/secrets
registry ls global          GET /registries
registry ls org             GET /orgs/{org_id}/registries
registry ls repo            GET /repos/{repo_id}/registries
org ls                      GET /orgs
user ls                     GET /users
agent ls                    GET /agents
queue info                  GET /queue/info
```

Resolve `<owner/repo>` to `repo_id` once per command. Prefer
`/repos/lookup/{repo_full_name}` and fall back to scanning `/repos` for
deployments that do not support lookup.

## Distribution

The repo should publish GitHub releases with platform artifacts:

```text
wpci-darwin-arm64
wpci-darwin-amd64
wpci-linux-amd64
wpci-linux-arm64
wpci-windows-amd64.exe
checksums.txt
```

Installers:

```sh
curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
```

```powershell
irm https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.ps1 | iex
```

The installer should fetch releases, not execute arbitrary source code as the
application runtime.

No release artifacts exist yet. Public install commands must be presented as the
target release flow until the first release is published.

Release defaults:

- Build with GoReleaser if it stays simple; otherwise use a GitHub Actions
  matrix running `go build`.
- Publish static binaries named:
  - `wpci-darwin-arm64`
  - `wpci-darwin-amd64`
  - `wpci-linux-amd64`
  - `wpci-linux-arm64`
  - `wpci-windows-amd64.exe`
- Publish `checksums.txt` with SHA-256 hashes.
- Installers install to `~/.local/bin` by default.
- `WPCI_INSTALL_DIR` overrides install location.
- `WPCI_VERSION` chooses a release tag; default is latest.

## Credits and Attribution

- Woodpecker CI and its maintainers provide the upstream CI system, API, CLI
  vocabulary, API documentation, and OpenAPI specification this project targets:
  https://github.com/woodpecker-ci/woodpecker
- CLI-Anything by HKUDS influenced the agent-native CLI requirements,
  especially standard help output, JSON-first automation, subprocess tests, and
  deterministic setup: https://github.com/HKUDS/CLI-Anything
- This project is an independent, unofficial implementation. Do not imply
  affiliation with Woodpecker CI or HKUDS/CLI-Anything.

## Real Open Questions

These are the few decisions that should stay open until implementation pressure
forces an answer:

- Should the public repository remain named `woodpecker-cli`, or should it be
  renamed to reduce confusion with Woodpecker's official CLI?
- Should log streaming use `/stream/logs/...` in v1, or should v1 start with
  non-streaming log fetch only?

## Decisions Made

- v1 includes write/admin commands gated by `--write` and `--confirm <target>`.
- Secret values can be provided with `--value` for scripts or `--value-stdin`
  to avoid shell history.
- API parity probes in `doctor --json` are advisory; commands still surface
  upstream 404/403 errors directly.
