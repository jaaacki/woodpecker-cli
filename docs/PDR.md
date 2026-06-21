# Product Design Record: wpci

## Problem

Existing local Woodpecker access tends to become a set of server-specific
wrappers. Those wrappers are useful, but they drift, expose only small slices of
the API, and are awkward for AI agents that need predictable commands and JSON
output.

`wpci` should be a neutral client library and CLI that can talk to any configured
Woodpecker endpoint by account alias.

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

1. Multi-account configuration.
2. Credentials stored outside repos.
3. Safe token handling and redacted diagnostics.
4. Broad Woodpecker API coverage.
5. Human-readable default output.
6. `--json` output for every command.
7. Explicit write gates for mutating operations.
8. Public release installation via `install.sh` and `install.ps1`.

## Configuration

Suggested file layout:

```text
~/.config/wpci/
  accounts/
    <alias>.json
  tokens/
    <alias>
```

Suggested permissions:

```text
~/.config/wpci          700
~/.config/wpci/tokens/* 600
```

Account config should include:

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

## Output Model

Default output is concise human-readable text or tables.

Every command supports:

```sh
--json
--raw
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

## API Coverage Target

Use the upstream OpenAPI specification as the coverage baseline:

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

