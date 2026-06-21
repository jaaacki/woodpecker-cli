# woodpecker-cli

An AI-friendly, multi-account Woodpecker CI API client and CLI.

This is an unofficial project. It is not affiliated with, endorsed by, or a
replacement for Woodpecker CI. It is a neutral client library and command-line
wrapper for working with one or more Woodpecker servers from humans, scripts,
and AI agents.

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

## Install

One line. The installer detects OS/arch, downloads the matching release binary,
verifies its sha256 checksum, and puts `wpci` on a user-writable bin directory.

Unix-like systems (macOS/Linux):

```sh
curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.ps1 | iex
```

Pin a specific release:

```sh
WPCI_VERSION=v0.0.1 curl -fsSL https://raw.githubusercontent.com/jaaacki/woodpecker-cli/main/install.sh | sh
```

Env knobs: `WPCI_VERSION`, `WPCI_INSTALL_DIR`, `WPCI_REPO`. The installer never
asks for Woodpecker credentials.

## Account Setup

A fresh install has no accounts. Configure a server, then validate:

```sh
# one command: server + token in one shot (token via stdin)
printf '%s' "$WPCI_TOKEN" | wpci account add home --server https://ci.example.com --token-stdin
wpci home doctor --json
```

If `~/.local/bin` is not on your PATH the installer prints the `export PATH=…`
line to add it. Windows uses `wpci account add home --server <url> --token-stdin`.


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

## Credits

- Woodpecker CI and its maintainers provide the upstream CI system, API, CLI
  vocabulary, API documentation, and OpenAPI specification this project targets:
  https://github.com/woodpecker-ci/woodpecker
- CLI-Anything by HKUDS influenced the agent-native CLI requirements,
  especially consistent help output, JSON-first automation, subprocess testing,
  and deterministic setup: https://github.com/HKUDS/CLI-Anything
- This repository is an independent implementation and should preserve upstream
  attribution anywhere Woodpecker or CLI-Anything design ideas are referenced.

## Status

Usable and installable from releases (latest `v0.0.2`). Multi-account
configuration and the core read-only API surface (repo, pipeline, org, agent,
queue, cron, registry, secret listing) are implemented and covered by tests.
Ships with a three-tier Woodpecker CI (PR / dev / release) and a one-line
curl installer (macOS, Linux, Windows). Write operations land behind the
`--write` flag. See `docs/PDR.md` and `docs/ROADMAP.md` for what remains —
milestone 4 (controlled writes) and milestone 5 (API-parity hardening).
