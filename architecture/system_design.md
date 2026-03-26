# System Design --- Stablecoin Escrow Platform

## 1. Overview
The Stablecoin Escrow Platform is a privacy-preserving system designed for institutional multi-party agreements using **DAML** smart contracts and **Canton** distributed ledger technology. It utilizes the **CIP-0056** token standard for secure stablecoin pledging and settlement.

## 2. Components

### 2.1. DAML Ledger (Canton)
- **Escrow Contracts:** Manage the lifecycle (Draft -> Funded -> Active -> Settled).
- **CIP-0056 Assets:** Uses stablecoins like **USDCx** (via BitGo/Circle) for collateral/escrow assets.
- **Privacy:** Native selective disclosure via the Canton Synchronizer.

### 2.2. Backend Services (Go)
- **escrow-api:** REST interface with DTO validation.
- **ledger-client:** Integration layer for Daml JSON API V2.
- **Identity Bridge:** Maps Okta/OIDC subjects to Daml User IDs.

### 2.3. Analytics & Infrastructure
- **Noves API:** Real-time and indexed data for tracking token holdings and transaction metrics.
- **Splice Validator APIs:** High-level endpoints for external signing and automated escrow agent workflows.
- **Privy/Dfns:** Wallet-as-a-Service integration for secure participant key management (CIP-0056 compatible).

## 3. High-Assurance Workflow

The escrow lifecycle follows a strict sequence to ensure funds safety and regulatory conformance:

1.  **Locking:** Use a Daml script/command to lock stablecoins (**USDCx**) from a Buyer into a smart contract. The **Issuer (Bank)** acts as a signatory to confirm custody.
2.  **Validation:** Use the **Noves API** to confirm the deposit and verify on-ledger holdings before activating the escrow.
3.  **Active Management:** Escrow remains live until milestone conditions are met.
4.  **Release/Revert:** Upon fulfillment (or mediation), the Daml contract automatically executes a transfer to the Seller or returns funds to the Buyer.

## 4. Privacy & Authorization Model
- **Sovereignty:** Bank, Buyer, and Seller operate on separate Canton nodes for maximum data isolation.
- **Authority:** Signatories include Buyer, Seller, and Issuer.
- **Mediator:** Controls the activation of terminal settlement choices during a dispute.

## 5. Ledger Topology (Phase 6: Distributed)
For high-assurance testing and production parity, we utilize separate participant nodes:
- **Bank Node (Port 7575):** Manages stablecoin issuance and settlement.
- **Buyer Node (Port 7576):** Manages buyer-side escrow proposals and funding.
- **Seller Node (Port 7577):** Manages seller-side acceptance and work evidence.
