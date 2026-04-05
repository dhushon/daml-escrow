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

------------------------------------------------------------------------

## CI Requirements

Every PR must run:

- contract compilation
- Go linting
- unit tests
- integration tests
