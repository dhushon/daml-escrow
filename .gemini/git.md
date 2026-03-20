# Git Workflow & Standards

## Commit Signing (DCO)
- **Mandatory Sign-off:** Every commit MUST be signed off using the `-s` or `--signoff` flag (e.g., `git commit -s -m "message"`).
- **Purpose:** This ensures compliance with the Developer Certificate of Origin (DCO), asserting that the contributor has the right to submit the code.
- **Verification:** CI and repository hooks may reject unsigned commits.

## Commit Messages
- **Prefixes:** Use conventional commit prefixes: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`.
- **Imperative Mood:** Describe what the commit does, not what you did (e.g., "add escrow logic" not "added escrow logic").
