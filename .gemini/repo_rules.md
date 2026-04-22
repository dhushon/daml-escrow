# Repository Guardrails

## Branching Strategy

main → production ready

develop → integration branch

feature/\* → new work

------------------------------------------------------------------------

## Commit Standard

Use conventional commits:

feat: fix: docs: refactor: test: chore:

**MANDATORY:** All commits MUST be GPG-signed.

------------------------------------------------------------------------

## Pull Request Rules

PR must include:

- description
- architecture impact
- tests
- security considerations

**Verified Commits:** Only "Verified" commits are eligible for merge into the `main` branch.

------------------------------------------------------------------------

## Code Review Requirements

Minimum:

2 reviewers for smart contract changes

1 reviewer for service code

## Pre-Commit Verification

**MANDATORY:** Before committing and pushing code, developers MUST verify the full stack locally to prevent CI failures and maintain trunk stability.

1.  **Frontend**: Execute `npm run build` in the `frontend` directory to verify JavaScript/Astro build stability.
2.  **Backend Logic**: Execute `go test -v ./...` to verify business logic and mock-based provider interactions.
3.  **End-to-End Integrity**: Execute `make integration-test` to verify ledger-to-service synchronization and multi-actor co-signing.

------------------------------------------------------------------------

## CI Requirements

Every PR must run:

- contract compilation
- Go linting
- unit tests
- integration tests
