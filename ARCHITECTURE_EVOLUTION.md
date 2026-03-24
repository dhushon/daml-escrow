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
3.  **Applies Corporate Policy:** Can enforce that only an `@datacloud.com` authenticated Okta user can claim the role of "Seller Approver."

### C. Negative Outcomes & Resolution
- **Term Deadlock:** If Seller Legal (`SA`) finds terms non-compliant, the proposal is archived. The system tracks this as a "Lost Opportunity" in metrics.
- **Milestone Gridlock:** If Buyer Approver (`BA`) rejects work but Seller refuses to redo it, the **Mediator Process Lead** (`ML`) uses the Evidence metadata stored on-ledger to determine a fair payout ratio.
