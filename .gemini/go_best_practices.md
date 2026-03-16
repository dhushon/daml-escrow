# Go Engineering Guardrails

## Language Philosophy

Go services should be:

- small
- explicit
- deterministic

Avoid unnecessary abstraction.

------------------------------------------------------------------------

## Project Layout

Standard layout:

/cmd /internal /pkg /api /config

Example:

/cmd/escrow-api/main.go /internal/ledger /internal/oracle
/internal/services

------------------------------------------------------------------------

## Dependency Rules

Allowed:

- standard library
- minimal third-party dependencies

Avoid:

- large frameworks
- reflection-heavy libraries

Preferred:

- chi (HTTP routing)
- zap (logging)
- viper (config)

------------------------------------------------------------------------

## Concurrency

Use Go concurrency patterns:

- worker pools
- channels
- context cancellation

Never create unbounded goroutines.

------------------------------------------------------------------------

## Error Handling

Errors must:

- be explicit
- include context

Example:

return fmt.Errorf("ledger submission failed: %w", err)

------------------------------------------------------------------------

## Testing

Minimum requirements:

- unit tests
- integration tests against ledger sandbox

Coverage target: 80%
