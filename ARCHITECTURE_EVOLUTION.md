# Architecture Evolution: Multi-Actor Lifecycle

This document elaborates on the detailed roles and state transitions within the Stablecoin Escrow platform.

## 1. Role-Based Workflow Matrix

The following diagram illustrates the granular interactions between Buyer, Seller, and Mediator roles across the contract lifecycle.

```mermaid
sequenceDiagram
    autonumber
    participant BC as Buyer (Preparer/Legal)
    participant BA as Buyer (Approver/Payer)
    participant L as Daml Ledger
    participant SA as Seller (Approver/Legal)
    participant SR as Seller (Refunder)
    participant ML as Mediator (Process Lead)
    participant MS as Mediator (Settler/Bank)

    Note over BC, BA: Phase 1: Drafting & Internal Review
    BC->>L: Create Escrow Draft (Proposal/Invite)
    L-->>BA: Notify for Internal Approval
    BA->>L: Approve for External Dispatch
    
    Note over L, SA: Phase 2: Counterparty Acceptance
    L-->>SA: Notify (External Offer)
    alt Positive Path: Acceptance
        SA->>L: Accept Terms (Choice: Accept)
        L->>L: Transition to ACTIVE Escrow
    else Negative Path: Rejection
        SA->>L: Reject Terms
        L->>L: Archive Proposal (OFFER_REJECTED)
    end

    Note over BA, SA: Phase 3: Execution & Milestone Approval
    SA->>L: Submit Evidence / Milestone Work
    L-->>BA: Notify for Review
    alt Positive Path: Approval
        BA->>L: Approve Milestone (Choice: ApproveMilestone)
        L->>MS: Trigger Settlement Request
        MS->>L: Execute Payment (Choice: Settle)
        L-->>SA: Funds Received
    else Negative Path: Dispute
        BA->>L: Reject Milestone / Raise Dispute
        L->>ML: Assign Mediator Process Lead
        ML->>L: Investigate & Resolve (Choice: ResolveDispute)
        L->>MS: Settled based on Resolution
    end

    Note over SA, SR: Phase 4: Termination & Refunds
    alt Seller Initiated Refund
        SR->>L: Return Funds (Choice: SellerRefund)
        L-->>BA: Funds Returned to Payer
    end
```

## 2. Decision Logic & Branching

### A. The Preparer-Approver Loop (Internal Governance)
By separating the **Contract Preparer** from the **Buyer Approver**, we enforce a "four-eyes" principle. A preparer (typically a procurement officer) can define terms, but only an authorized officer (Payer) can commit funds to the ledger.

### B. Business Email Logic (Onboarding)
When an invitation is issued to `user@datacloud.com`, the platform:
1.  **Extracts Domain:** Validates the suffix `datacloud.com`.
2.  **Associates Organization:** Automatically tags the invitation with the "DataCloud LLC" metadata.
3.  **Applies Corporate Policy:** Can enforce that only an `@datacloud.com` authenticated user can claim the role of "Seller Approver."

### C. Negative Outcomes & Resolution
- **Term Deadlock:** If Seller Legal (`SA`) finds terms non-compliant, the proposal is archived. The system tracks this as a "Lost Opportunity" in metrics.
- **Milestone Gridlock:** If Buyer Approver (`BA`) rejects work but Seller refuses to redo it, the **Mediator Process Lead** (`ML`) uses the Evidence metadata stored on-ledger to determine a fair payout ratio.

## Phase 5: High-Assurance Identity & Adjudication (Completed)

### Architectural Shift: The Adjudicator Model
Moved from a simple "Buyer releases funds" model to a mediated "State Actor" model. Stakeholders (Buyer, Seller, Issuer) sign the agreement, while an independent Adjudicator (Mediator) authoritatively backs the evidence of completion.

```mermaid
sequenceDiagram
    participant B as Buyer
    participant S as Seller
    participant M as Mediator (Adjudicator)
    participant L as Ledger
    participant CB as Central Bank (Issuer)

    Note over B,M: 1. Authoritative Appointment
    B->>L: Invite Seller (Signatories: B, M, CB)
    S->>L: Claim Invite (Verified by OIDC Email)
    L->>L: Create Proposal (Signatories: M, CB)
    S->>L: Accept Proposal
    L->>L: Create Active Escrow (Signatories: B, S, CB)

    Note over M,L: 2. Mediated Completion
    M->>L: Approve Milestone (Evidence of Completion)
    L->>L: Create EscrowSettlement (Signatory: CB)

    Note over CB,L: 3. Final Settlement
    CB->>L: Settle Payment
    L->>L: Funds Transferred to Seller
```

### Key Technical Enhancements
1.  **Dynamic Discovery:** Eliminated hardcoded IDs. The backend now resolves Package and Party IDs at runtime via `ledger-state.json` or active metadata discovery.
2.  **OIDC-Daml Mapping:** Unified external identities (Google sub) with internal ledger handles (Daml User ID), ensuring authoritative matching.
3.  **Thread-Safe Multi-Tenancy:** Refactored the ledger client to support concurrent, independent user sessions without identity bleed.
4.  **Stakeholder Parity:** Both stakeholders can authoritatively raise disputes, ensuring balanced power dynamics.
