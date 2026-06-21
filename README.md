# woodpecker-cli

An AI-friendly, multi-account Woodpecker CI API client and CLI.

This project is not a replacement for Woodpecker CI. It is a neutral client
library and command-line wrapper for working with one or more Woodpecker
servers from humans, scripts, and AI agents.

Target command shape:

```sh
wpci <account-alias> <api-area> <action> [params/options]
```

Examples:

```sh
wpci home repo ls
wpci home repo show jaaacki/emby-processor
wpci home pipeline last jaaacki/emby-processor --branch main
wpci lab pipeline log show sparkfn/media-butler 42 build
wpci home doctor --json
```

## Goals

- Support multiple Woodpecker endpoints by account alias.
- Store credentials outside application repositories.
- Keep commands close to the upstream Woodpecker API and CLI vocabulary.
- Make inspection easy by default for humans and AI agents.
- Provide JSON output everywhere.
- Support write/admin operations only behind explicit gates such as `--write`
  and `--confirm`.
- Install from public releases with curl or PowerShell.

## Non-Goals

- No browser automation.
- No token values in command output, repo files, logs, or issue templates.
- No large plugin framework until real usage demands it.
- No one-off wrapper per Woodpecker server.

## Planned Install Flow

Unix-like systems:

```sh
curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.ps1 | iex
```

Install a specific version:

```sh
WPCI_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
```

The installer is intended to download release artifacts, verify checksums when
available, and install `wpci` into a user-writable bin directory. It should not
ask for Woodpecker credentials.

## Planned Account Setup

```sh
wpci account add home --server https://ci.example.com
wpci account token set home
wpci home doctor
```

Agent-friendly setup:

```sh
wpci account add home --server https://ci.example.com
printf '%s\n' "$WOODPECKER_TOKEN" | wpci account token set home --stdin
wpci home doctor --json
```

## Initial Command Surface

Account provisioning:

```sh
wpci account add <alias> --server <url>
wpci account ls
wpci account show <alias>
wpci account rm <alias>
wpci account test <alias>
wpci account token set <alias>
wpci account token status <alias>
wpci account token rm <alias>
```

Core inspection:

```sh
wpci <alias> info
wpci <alias> version
wpci <alias> whoami
wpci <alias> doctor
```

Woodpecker API areas:

```sh
wpci <alias> repo ...
wpci <alias> pipeline ...
wpci <alias> cron ...
wpci <alias> secret ...
wpci <alias> registry ...
wpci <alias> org ...
wpci <alias> user ...
wpci <alias> agent ...
wpci <alias> queue ...
```

## Upstream References

- Woodpecker CI: https://woodpecker-ci.org/
- API reference: https://woodpecker-ci.org/api
- OpenAPI YAML: https://woodpecker-ci.org/redocusaurus/plugin-redoc-0.yaml
- Upstream repository: https://github.com/woodpecker-ci/woodpecker
- Swagger generation docs:
  https://woodpecker-ci.org/docs/next/development/swagger

## Status

Design/scaffold phase. See `docs/PDR.md` and `docs/ROADMAP.md`.

