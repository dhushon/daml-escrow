# Architecture Evolution: Modular & Professional Escrow

This document outlines the plan for evolving the Stablecoin Escrow platform into a modular, production-grade system using Daml Interfaces and Daml Finance primitives.

---

## 1. Modular Interface Strategy
To support multiple business domains (Leasing, Supply Chain, Grants) without a monolithic contract, we will move to an **Interface-First** model.

### Core Interface: `interface Escrow`
Instead of domain-specific fields, the core interface will define the **capabilities** required for any escrow:
*   **Choice `ApproveMilestone`**: Signals completion of a phase.
*   **Choice `RaiseDispute`**: Freezes funds for mediation.
*   **Method `getMetadata`**: Returns the `EscrowMetadata` (SchemaURL + Payload).
*   **Method `getAmount`**: Returns the financial obligation.

### Benefits
The Go API can then interact with the `Escrow` interface ID. It doesn't need to know if it's dealing with a `LeaseContract` or a `GrantContract`, as long as they both implement the `Escrow` interface.

---

## 2. Professional Asset Layer (Daml Finance)
Currently, "Stablecoin" is represented by a `Text` field. To mature this, we plan to integrate specific `daml-finance` modules compatible with SDK 3.4.11.

### Target Dependencies
*   `daml-finance-interface-holding`: Standard interface for owning assets.
*   `daml-finance-interface-settlement`: Standard for multi-party atomic swaps.
*   `daml-finance-interface-lifecycle`: For managing the "event" clock of milestones.

### Evolution Path
1.  **Phase 3.5 (Internal):** Replace `EscrowSettlement` text fields with a `daml-finance` Instruction.
2.  **Phase 4.0 (Integration):** Use real Stablecoin holdings (e.g., ERC-20 mirrors) as the underlying value.

---

## 3. Deployment & Packaging
To maintain simplicity while supporting extensions, we will move to a multi-package structure:

1.  **`escrow-core`**:
    *   Contains the `interface Escrow`.
    *   Contains generic `EscrowMetadata` types.
    *   *No business logic.*
2.  **`escrow-common`**:
    *   Standard implementations of the interface for simple use cases.
3.  **`escrow-extensions`**:
    *   Domain-specific packages (e.g., `escrow-leasing`) that depend on `escrow-core`.

---

## 4. Summary of Change impact
*   **Daml Code:** Logic moves from `StablecoinEscrow.daml` into an `Interface` definition + implementation templates.
*   **Go Code:** The `JsonLedgerClient` will query the interface name (`Escrow`) instead of the template name (`StablecoinEscrow`).
*   **Metadata:** Stays exactly as implemented in Phase 3 (JSON Schema driven), as this is already highly decoupled.

---
**Status:** PROPOSAL ONLY. No changes have been made to the codebase.
