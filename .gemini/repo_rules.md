# Repository Guardrails

## Branching Strategy

main → production ready

develop → integration branch

feature/\* → new work

------------------------------------------------------------------------

## Commit Standard

Use conventional commits:

feat: fix: docs: refactor: test:

Example:

feat: add escrow milestone contract

------------------------------------------------------------------------

## Pull Request Rules

PR must include:

- description
- architecture impact
- tests
- security considerations

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
