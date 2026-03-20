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

-   **Cryptographic Signatures:** Every incoming webhook MUST include a signature (e.g., HMAC-SHA256) verified against a pre-shared secret or public key.
-   **State Cross-Referencing:** Trigger logic MUST fetch the current ledger state of the target contract to verify logical consistency (e.g., matching milestone indices) before executing a choice.
-   **Whitelisting:** Only registered oracle providers are permitted to trigger state changes.
-   **Replay Protection:** Use nonces or timestamps within the signed payload to prevent replay attacks.
