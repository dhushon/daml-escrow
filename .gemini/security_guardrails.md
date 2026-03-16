# security_guardrails.md

Security rules for escrow and financial contract systems.

------------------------------------------------------------------------

# Critical Security Principles

1.  **Funds safety first**
2.  **Least privilege**
3.  **Explicit authorization**
4.  **Auditability**
5.  **Fail-safe defaults**

------------------------------------------------------------------------

# Smart Contract Rules

Contracts must enforce:

-   explicit controllers
-   no implicit approvals
-   deterministic settlement

Never allow:

-   arbitrary contract deletion
-   unauthorized state transitions

------------------------------------------------------------------------

# Service Security

Services must:

-   validate all input
-   verify identities
-   avoid trust in client state

------------------------------------------------------------------------

# Secrets Management

Secrets must never appear in:

-   code
-   config files
-   logs

Use:

-   environment variables
-   vault systems
-   KMS

------------------------------------------------------------------------

# Oracle Security

External triggers must be validated using:

-   signatures
-   whitelisted providers
-   replay protection
