## CI: Docker-only Woodpecker, no host Go runtime

- This repo uses Woodpecker CI (local `wpci`). Config lives in `.woodpecker/`.
- The product is a CLI binary shipped via GitHub Releases — NOT a container image. "Docker-only" means CI runs in throwaway `golang` containers, not that the product is containerized.
- Do NOT run `go test`, `go build`, `go vet`, `go mod`, or `gofmt` checks on the host. Do NOT install Go or project deps on the host.
- Validate by pushing to a PR branch and reading CI only through `wpci`. Do not run raw `woodpecker-cli`, open the Woodpecker UI, or paste tokens.
- Running a produced binary for e2e on the dev machine is allowed. Running `go test`/`go build` on the host is not.
- PR tier (`.woodpecker/pr.yml`): fmt + vet + build + unit tests, `event: pull_request`, fast.
- DEV tier (`.woodpecker/dev.yml`): tests in container + build binary → rolling `dev` GitHub prerelease for dev e2e.
- MAIN tier (`.woodpecker/release.yml`): tag `v*` → cross-compile + checksums → GitHub Release.
- Secrets (e.g. `github_token`) live in Woodpecker, not in YAML or repo files.

## Workflow

- One issue → one worktree under `.claude/worktrees/<issue-slug>`. PR target: `dev`. Do not commit to `dev` or `main`.
- PRs and `dev` build through Woodpecker CI; do not merge until CI passes.
- After merge: cleanup worktrees and branches; verify with `git worktree list` and `git branch -vv`.
- Development history lives in GitHub issues/PRs/commits/CI — not local markdown. No `PLAN.md`/`STATUS.md`/`NOTES.md`.
