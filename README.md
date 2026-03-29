# Stablecoin Escrow Platform (DAML-Based)

## Overview

This project implements a **high-assurance, privacy-preserving stablecoin escrow platform** using **DAML (Digital Asset Modeling Language)** and a **Canton distributed ledger**. It follows a rigorous formal escrow process designed for tokenized reserves.

👉 **[Get Started with Installation & Setup](./GET_STARTED.md)**

------------------------------------------------------------------------

## High-Assurance Architecture

### System Stack

```mermaid
graph TD
    User((User / Admin)) <-->|Port 8080| Astro[Astro Frontend]
    Astro <-->|REST / JSON| Go[Go Backend API]
    Go <-->|Port 8081| JSON[Daml JSON API V2]
    JSON <-->|Port 7575| Canton[Canton Ledger]
    Canton <--> DB[(Postgres DB)]
    Oracle((Oracle Service)) -->|Webhook| Go
```

### Canton Network & Token Standards

This platform is built on the **Canton Network**, a privacy-enabled, interoperable blockchain designed for institutional finance. It leverages industry-standard protocols to ensure secure B2B stablecoin pledging and escrow:

* **CIP-0056 Token Standard:** Implements the "holding", "lockable", and "transferable" interfaces required for secure, interoperable stablecoin movement (e.g., USDCx via BitGo/Circle).
* **Canton OpenZeppelin Stablecoin/CDP Module:** Utilizes production-ready Daml templates for Collateralized Debt Positions (CDP) and standard CIP-0056 holding mechanisms.
* **Validator APIs (Splice):** Employs high-level validator endpoints for automated escrow workflows and external party signing (e.g., trusted escrow agents).
* **Noves Data & Analytics:** Integrates real-time indexed data for tracking token holdings, transaction history, and wallet metrics across the Canton Network.

### Escrow Lifecycle (Formal Model)

Refined per `ESCROW-PROCESS.md` to ensure bilateral consent and tripartite authority.

```mermaid
stateDiagram-v2
    [*] --> DRAFT: ProposeEscrow (B+S+I)
    DRAFT --> FUNDED: Fund (Buyer)
    FUNDED --> ACTIVE: Activate (Issuer)
    ACTIVE --> SETTLED: ConfirmConditions (Mediator)
    ACTIVE --> DISPUTED: RaiseDispute (Buyer/Seller)
    DISPUTED --> PROPOSED: ProposeSettlement (Mediator)
    PROPOSED --> DISPUTED: Reject (Buyer/Seller)
    PROPOSED --> SETTLED: Ratify + Finalize (Buyer+Seller)
    SETTLED --> [*]: Disburse (Issuer)
```

------------------------------------------------------------------------

## Key Features

### 1. Robust State Machine (Phase 5)

Strict transition guards ensure funds cannot be released until conditions are met or bilateral agreement is reached in a dispute.

* **DRAFT:** Terms agreed, but asset not yet deposited.
* **FUNDED:** Asset locked by Issuer, awaiting activation.
* **ACTIVE:** Escrow is live and conditions are being monitored.
* **DISPUTED:** Adjudication phase initiated.
* **PROPOSED:** Mediated settlement awaiting party ratification.

### 2. Distributed Sovereignty (Phase 6)

The platform has transitioned from a single-node sandbox to a **tripartite distributed topology**, enforcing strict data sovereignty and authority:

*   **Tripartite Authority Model:** Every escrow requires co-signature from **Buyer**, **Seller**, and **Issuer (Bank)**. This prevents unauthorized state transitions and ensures institutional compliance.
*   **Multi-Node Isolation:** Each participant operates their own **Canton Node**, ensuring that private contract data (e.g., specific terms or evidence) only resides on the participants' infrastructure.
*   **Intelligent Routing:** The Go backend utilizes a `MultiLedgerClient` to intelligently route commands to the specific node hosting the primary submitter, maintaining zero-trust isolation.

### 3. CIP-0056 & Institutional Tokens

Native support for **CIP-0056** ensures the platform can interoperate with real stablecoins:
*   **Lockable Interface:** Assets are cryptographically locked in the escrow contract, preventing double-spending while the escrow is `ACTIVE`.
*   **Transferable Interface:** Final settlement triggers authoritative transfers using standardized token choices, ensuring compatibility with major institutional issuers.

### 4. Self-Healing Integration

The Go backend features a **Dynamic Discovery Engine** that automatically resolves Package IDs and Party IDs at runtime, ensuring the stack is environment-agnostic and resilient to ledger resets.

------------------------------------------------------------------------

## Analytics & Operational Velocity (Phase 6.3)

The platform integrates a high-assurance analytics layer powered by **Noves-ready logic** to provide real-time visibility into the escrow lifecycle.

### 1. The Operational Velocity Dashboard
Accessible via `/metrics`, this dashboard visualizes the platform's efficiency using:
*   **Stage Duration Heatmap:** Tracks the average minutes spent in each escrow state (`DRAFT`, `FUNDED`, `ACTIVE`, `PROPOSED`), identifying systemic bottlenecks.
*   **Conversion Funnel:** Visualizes the "drop-off" and success rate from initial proposal through to final settlement.
*   **System Health:** Real-time monitoring of P95/P99 latencies, command success rates, and ACS (Active Contract Set) size.

------------------------------------------------------------------------

## Repository Structure

* `/cmd`: Entry points for API and Oracle Simulator.
* `/internal`: Modular Go backend (Ledger Client, Service Layer, REST Handlers).
* `/contracts`: Multi-package Daml structure (Interfaces, Implementation, Tests).
* `/frontend`: Astro-based dashboard with DataCloud LNF styling.
* `ESCROW-PROCESS.md`: The formal process specification.
* `REGULATORY_CONFORMANCE.md`: Details on GDPR/CCPA and data sovereignty.
