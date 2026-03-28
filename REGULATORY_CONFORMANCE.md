# Regulatory Conformance & Compliance Framework

This document outlines the architectural decisions and technical controls implemented to meet global financial and data privacy regulations (e.g., GDPR, CCPA, SOC2, Basel III).

---

## 1. Funds Safety & Asset Segregation

**Requirement:** Regulated assets (stablecoins) must be held in a manner that prevents unauthorized movement or commingling.

*   **Tripartite Authority:** Every escrow contract requires co-signatures from the **Buyer**, **Seller**, and **Issuer (Bank)**. No single party can unilaterally move funds once locked.
*   **Deterministic Settlement:** Funds are only released upon the fulfillment of on-ledger conditions verified by an independent **Mediator (Adjudicator)**.
*   **Issuer Custody:** The stablecoin remains under the custody of the Issuer node until the terminal state is reached, ensuring "Settlement Finality."

## 2. Privacy & Data Sovereignty (GDPR/CCPA)

**Requirement:** Personal and transactional data must only be stored where there is a "Legal Basis" or "Need-to-Know."

*   **Node-Level Isolation:** By utilizing separate Canton participant nodes for Bank, Buyer, and Seller, we ensure that data is physically segregated. 
    *   The **Seller Node** never receives data about the **Buyer's** unrelated escrows.
    *   The **Bank Node** only sees the financial metadata required for settlement, not the private trade secrets (Evidence/Metadata) stored on participant nodes.
*   **Right to be Forgotten:** Archive patterns in DAML allow for the logical deletion of contract data while maintaining a cryptographic audit trail for regulators.

## 3. The "Four-Eyes" Principle & Identity

**Requirement:** High-value financial actions must involve multiple authorized individuals or entities to prevent fraud.

*   **Role-Based Routing:** The API Gateway routes requests to specific participant nodes based on **Verified OIDC Claims** (e.g., `@bank.com` users route to the Bank Node).
*   **Identity Pinning:** User identities (Okta/Google sub) are "pinned" to specific DAML Party IDs on their respective nodes. This prevents "Identity Bleed" where a user could masquerade as a different role on a different node.

## 4. Auditability & Forensic Logging

**Requirement:** All state changes must be immutable and reconstructible for a minimum of 7 years.

*   **Distributed Ledger Audit Trail:** Every state transition is recorded as a "Command" on the ledger, including the cryptographic signature of the initiator.
*   **Cross-Node Reconciliation:** Regulators can query the separate nodes and reconcile the states to ensure no node has been tampered with. If the Bank node reports "SETTLED" but the Buyer node reports "ACTIVE," the system triggers a **Consensus Alert**.

## 5. Fail-Safe Defaults & Continuity

**Requirement:** The system must remain stable or fail into a safe state during network partitions.

*   **Atomic Transactions:** Commands in DAML are atomic. If a multi-party signature fails (e.g., the Seller node is offline), the transaction is never committed.
*   **Timeout-Based Refund:** Every escrow includes a mandatory `ExpiryDate`. If a deadlock occurs or a node goes permanently offline, the Issuer node can authoritatively execute an "Expiry" choice to return funds to the Buyer after the deadline passes.

---

*Version: 1.0.0*
*Last Updated: 2026-03-26*
