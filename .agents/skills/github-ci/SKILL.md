---
name: github-ci
description: >
  Manages the full GitHub workflow for the daml-escrow project: branching, signed commits, pull request creation via `gh`, CI pipeline monitoring, and local pre-push verification. Activate whenever the user wants to create a branch, stage changes, commit, open a PR, check CI status, fix a failing CI job, or review the merge checklist for this project.
---

# GitHub & CI Workflow Skill

This skill governs every interaction with source control and the GitHub Actions CI pipeline for the `daml-escrow` project.
Always follow these rules in full, in the order given.

---

## 1. Pre-Commit Local Verification (MANDATORY — do this before every commit)

Run the full local verification stack.
**Do NOT commit or push until all three pass.**

```bash
# 1. Backend unit + integration tests
go test -v ./...

# 2. Frontend build verification (no broken Astro/TS)
cd frontend && npm run build && cd ..

# 3. Frontend unit tests
cd frontend && npm run test && cd ..

# 4. Full integration smoke-test
make integration-test
```

If any step fails, diagnose and fix the issue before proceeding.
Use `go test -v -run TestName ./pkg/...` to isolate a single failing Go test.

---

## 2. Branching Rules

| Work type | Branch prefix | Example |
|-----------|---------------|---------|
| New feature | `feature/` | `feature/fiat-settlement-api` |
| Bug fix | `fix/` | `fix/acs-offset-resolution` |
| Docs / refactor | `chore/` or `docs/` | `docs/update-readme` |
| Security patch | `sec/` | `sec/resolve-esbuild-cve` |

- **NEVER push directly to `main` or `develop`.**
- Bug-fix direct pushes to `main` require **explicit user confirmation** before executing.
- Always branch from `develop` unless the user specifies otherwise.

```bash
git checkout develop
git pull origin develop
git checkout -b feature/<short-description>
```

---

## 3. Commit Standards

### Mandatory Flags

Every commit MUST include **both**:

| Flag | Purpose |
|------|---------|
| `-s` / `--signoff` | Developer Certificate of Origin (DCO) compliance |
| `-S` | GPG signature (only "Verified" commits merge to `main`) |

```bash
git commit -S -s -m "feat: add fiat settlement webhook handler"
```

### Conventional Commit Prefixes

```
feat:     new feature or capability
fix:      bug fix
refactor: code change with no behaviour delta
chore:    build tooling, dependencies, CI config
docs:     documentation only
test:     adding or updating tests
sec:      security vulnerability remediation
```

- Use **imperative mood**: "add X" not "added X".
- Keep the subject line <= 72 characters.
- Add a blank line + body for anything non-trivial.

---

## 4. Staging & Pushing Changes

```bash
# Review what's changed
git status
git diff --stat

# Stage selectively — never use `git add .` blindly
git add <specific-files>

# Confirm staged set
git diff --cached --stat

# Commit with mandatory flags
git commit -S -s -m "<type>: <description>"

# Push the branch
git push origin feature/<short-description>
```

---

## 5. Pull Request Creation

Use the GitHub CLI (`gh`). The PR description MUST include all four sections below.

```bash
gh pr create \
  --base develop \
  --title "<type>: <short description>" \
  --body "$(cat <<'EOF'
## Summary
<!-- One paragraph: what changed and why -->

## Architecture Impact
<!-- Did this change any DAML contracts, Go service boundaries, ledger topology, or Astro SSR routes? -->

## Changes Made
<!-- Bullet list of specific files and what changed -->

## Test Evidence
<!-- Paste relevant test output snippets -->
- [ ] go test -v ./... — all passing
- [ ] cd frontend && npm run build — build clean
- [ ] cd frontend && npm run test — all passing
- [ ] make integration-test — lifecycle passes

## Security Considerations
<!-- Any new endpoints, secrets, DAML contract choices, or identity flows introduced? -->
EOF
)"
```

> **Note:** For `main`-targeting PRs (hotfix / release), set `--base main` and get explicit user approval first.

---

## 6. CI Pipeline Overview

Defined in `.github/workflows/ci.yml`.

| Job | Runner | What it does |
|-----|--------|-------------|
| `lint` | ubuntu-latest | `golangci-lint` on Go source |
| `unit-tests` | ubuntu-latest | `go test -v ./...` |
| `frontend-build` | ubuntu-latest | `npm install`, `npm run build`, `npm run test` |
| `daml-tests` | ubuntu-latest | DPM install SDK 3.4.11, `dpm build --all`, `dpm test` |

Triggers on PRs to: `main`, `develop`, `feat/phase3-oracle-integration`, `feat/bitgo-stablecoin-integration`.

---

## 7. Monitoring CI After PR Creation

```bash
# List all checks for the PR
gh pr checks <pr-number>

# Stream live logs for a failing job
gh run view --log-failed

# Watch until all checks complete
gh pr checks <pr-number> --watch
```

**Failure handling loop:**
1. Pull failing logs: `gh run view <run-id> --log-failed`
2. Diagnose root cause (lint error / test failure / build error / DAML compile error).
3. Fix locally -> re-run local verification (Section 1) -> commit fix -> push.
4. Repeat until `gh pr checks <pr-number>` exits 0 with all jobs green.

---

## 8. Common CI Failure Patterns & Fixes

### Go lint failure (`golangci-lint`)
```bash
# Run locally to reproduce
golangci-lint run ./...
# Auto-fix where possible
golangci-lint run --fix ./...
```

### Go test failure
```bash
go test -v -count=1 ./...
# Run a specific package
go test -v -run TestFunctionName ./internal/services/...
```

### Frontend build failure
```bash
cd frontend
npm install        # ensure lockfile is in sync
npm run build      # reproduce the Astro build error
```

### DAML compile / test failure
```bash
cd contracts
dpm build --all
cd stablecoin-escrow-tests
dpm test
```

### Dependency vulnerability (`npm audit`)
```bash
cd frontend
npm audit
# Pin overrides in package.json under the "overrides" key, then:
npm install
```

---

## 9. Merge Checklist

Before requesting review or merging, confirm:

- [ ] All local verification steps pass (Section 1)
- [ ] All CI jobs green (`gh pr checks <pr-number>`)
- [ ] PR description has all four required sections (Section 5)
- [ ] No secrets, `.env` values, or credentials in any committed file
- [ ] For smart contract changes: minimum **2 reviewers** assigned
- [ ] For service code changes: minimum **1 reviewer** assigned
- [ ] Commits are GPG-signed (`git log --show-signature -1`)
- [ ] DCO sign-off present on every commit (`git log --oneline` shows `Signed-off-by:`)

---

## 10. Useful Reference Commands

```bash
# Check current branch and status
git status && git log --oneline -5

# Verify GPG signature on last commit
git log --show-signature -1

# List open PRs
gh pr list

# View PR status
gh pr view <pr-number>

# Add more commits to an open PR
git add <files> && git commit -S -s -m "fix: ..."
git push origin $(git branch --show-current)

# Sync branch with develop
git fetch origin && git rebase origin/develop
```

---

## 11. References

- `.github/workflows/ci.yml` — CI pipeline definition
- `.gemini/git.md` — Git workflow standards
- `.gemini/repo_rules.md` — Repository guardrails
- `.gemini/security_guardrails.md` — Security rules (commit signing, secrets)
- `.gemini/agents/pipeline_verification_agent.md` — Pipeline verification agent profile
