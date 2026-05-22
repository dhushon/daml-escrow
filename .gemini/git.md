# Git Workflow & Standards

## Commit Signing (DCO)

- **Mandatory Sign-off:** Every commit MUST be signed off using the `-s` or `--signoff` flag (e.g., `git commit -s -m "message"`).
- **Purpose:** This ensures compliance with the Developer Certificate of Origin (DCO), asserting that the contributor has the right to submit the code.
- **Verification:** CI and repository hooks may reject unsigned commits.

## Commit Messages

- **Prefixes:** Use conventional commit prefixes: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`.
- **Imperative Mood:** Describe what the commit does, not what you did (e.g., "add escrow logic" not "added escrow logic").

## Pull Request & Merge Policy

- **PR-First Mandate:** ALL changes to the codebase (features, refactors, documentation, tests) MUST be made via a formal Pull Request.
- **Direct Pushes:** Direct pushes to `origin/main` are strictly prohibited for all non-bugfix work.
- **Bug Fix Exception:** For critical bug fixes, you MUST ask for explicit user permission before pushing directly to `origin/main`.
- **Branch Naming:** Use `feature/*` for new work and `fix/*` for bug fixes.
