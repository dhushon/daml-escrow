# Stablecoin & Token Integration Guardrails (CIP-0056)

This document defines the requirements for integrating real institutional stablecoins (BitGo, Circle, institutional tokens) into the escrow platform using the **CIP-0056** standard.

## 1. CIP-0056 Interface Mandate

- **Cryptographic Holdings:** Move away from "numeric balance" fields. All digital assets MUST be represented as **Daml Interface instances** (`Token.CIP56.Holding`).
- **Standard Interactions:** All token operations MUST use the standardized interface choices:
    - **Lockable:** For freezing assets during the `ACTIVE` escrow phase.
    - **Transferable:** For authoritative disbursement or refund.
    - **Burnable/Mintable:** For lifecycle management by the Issuer.

## 2. Vault-Centric Logic

- **Holding-Based Escrow:** The Escrow contract MUST NOT store a "balance" number. It MUST store a **Contract ID Reference** or a **Template Lock** on a real CIP-0056 holding.
- **Zero-Trust Settlement:** Transfers must be driven by the ledger's authority, not by "Instructional" API calls that trust the backend to move numbers.

## 3. Mock & Provider Synchronicity

- **Parallel Evolution:** Every change to the `StablecoinProvider` interface MUST be reflected immediately in:
    - `JsonStablecoinProvider` (Real Implementation)
    - `MockStablecoinProvider` (Testing Mock)
- **Interface Integrity:** The `MockStablecoinProvider` MUST implement 100% of the methods in the `StablecoinProvider` interface.

## 4. Testing Strategy & Hierarchy

- **Unit Tests:** Use `MockStablecoinProvider` for fast, logic-only verification.
- **Integration Tests (`-tags=integration`):** 
    - **MANDATORY:** Test the `MockStablecoinProvider` during integration runs to ensure that the service layer interacts correctly with the stablecoin interface even when the live provider is unavailable.
- **Stablecoin Integration Tests (`-tags=stablecoin,bitgo,circle`):**
    - Targets real/simulated **BitGo/Circle** environments.
    - Uses the `-tags=bitgo` or `-tags=circle` build tags to select specific provider implementations.
    - Validates actual DAR interactions with the CIP-0056 templates on a live ledger.
    - Requires connectivity to the institutional token issuer's API or a high-fidelity simulator.

## 5. Security & Vault Protection

- **No Shared Secrets:** Stablecoin provider API keys MUST never be stored in the ledger or configuration files. Use HSM or Cloud Secret Manager.
- **Double-Spend Protection:** The platform MUST verify that a holding is **Locked** before transitioning an escrow to `ACTIVE`.
