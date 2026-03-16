# System Design --- Stablecoin Escrow Platform

## 1. Overview

The Stablecoin Escrow Platform is a privacy-preserving system designed for multi-party agreements using DAML smart contracts and Go-based integration services.

## 2. Components

### 2.1. DAML Ledger (Canton)

- **Escrow Contracts:** Manage the lifecycle of an escrow (Created -> Locked -> Delivered -> Released/Refunded).
- **Settlement Contracts:** Record final payments to recipients.
- **Privacy:** Native selective disclosure; only signatories and observers see the contract.

### 2.2. Backend Services (Go)

- **escrow-api:** REST interface for clients (Buyer, Seller, Mediator).
- **ledger-client:** Integration layer for DAML Ledger API (gRPC).
- **oracle-service:** Automated triggers for milestone completions.

### 2.3. Frontend Applications

- **Buyer Portal:** Escrow creation and approval.
- **Seller Portal:** Progress tracking and delivery notification.
- **Mediator Dashboard:** Dispute resolution and settlement.

## 3. Data Flow

1. **Creation:** Buyer initiates escrow via `escrow-api`.
2. **Locking:** Issuer (e.g., Stablecoin Custodian) and Buyer sign the contract.
3. **Execution:** Seller works; Buyer approves milestones.
4. **Settlement:** Ledger releases funds to Seller/Buyer based on approvals or mediation.

## 4. Privacy Model

DAML's "need-to-know" privacy ensures that financial details are only visible to the involved parties:

- **Signatories:** Buyer, Issuer.
- **Observers:** Seller, Mediator.

No other party on the Canton network has visibility into the transaction.

## 5. Ledger Topology (Local Development)

For local development, we utilize the **DAML Sandbox**:

- In-memory ledger.
- Lightweight and fast.
- Compatible with production Canton deployments.
