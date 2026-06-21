# Roadmap

## Milestone 0: Project Shape

- Create public repository.
- Document PDR, README, and release install flow.
- Open tracking issue for implementation.

## Milestone 1: Minimal Multi-Account CLI

- Initialize Go module `github.com/jaaacki/woodpecker-cli`.
- Add Cobra command tree and `cmd/wpci/main.go`.
- Implement `wpci account add/ls/show/rm/test`.
- Implement `wpci account token set/status/rm`.
- Store configs under `~/.config/wpci/accounts`.
- Store tokens under `~/.config/wpci/tokens` with owner-only permissions.
- Implement `wpci <alias> doctor --json`.
- Define stable JSON success/error wrappers and exit codes.

## Milestone 2: Read-Only API Coverage

- Implement HTTP client, auth, error normalization, and JSON output.
- Implement user/version/repo/pipeline/log read commands.
- Implement cron, secret list, registry list, org list, user list, agent list,
  and queue info.
- Resolve repo slugs through `/repos/lookup/{repo_full_name}` where supported,
  with `/repos` fallback for older deployments.
- Add subprocess tests for `--help`, `--json`, config errors, auth errors, and
  representative read-only commands with fixture HTTP servers.

## Milestone 3: Release Installation

- Add build and release workflow.
- Publish platform artifacts.
- Publish `checksums.txt`.
- Make `install.sh` and `install.ps1` install latest or requested release.
- Test installer behavior against a draft/test release before publishing install
  commands as stable.

## Milestone 4: Controlled Writes

- Add `--write` gate for mutating pipeline, repo, cron, secret, and registry
  commands.
- Add `--confirm <target>` for destructive commands.
- Add admin command coverage where the token has permission.

## Milestone 5: API Parity Hardening

- Compare implemented commands against upstream OpenAPI.
- Add compatibility probes for Woodpecker version/API differences.
- Add integration tests against fixture responses and, where practical, a local
  Woodpecker test instance.
