# FIAT-SETTLEMENT.md — Multi-Rail Disbursement Architecture

## Purpose

Today `Disburse` assumes a single execution path: on-chain transfer via
BitGo or Circle. This document proposes the abstraction needed to add a
fiat leg (ACH, RTP, FedNow, wire) without touching contract-level authority
logic, and evaluates a payments-orchestration provider as the fiat-side
component.

Companion to `ESCROW-PROCESS-EXTENSIONS.md` DIRECTIVE 14.

---

## Why this matters

The platform currently treats stablecoin as the only settlement medium.
Real institutional escrow, payroll-adjacent disbursement, regulated payouts,
and any counterparty without a wallet still needs fiat rails. Building a
custom ACH/wire integration from scratch means originator relationships,
NACHA file formatting, and a sanctions-screening pipeline that has nothing
to do with DAML. That's a distraction from the ledger work this project is
actually good at.

## Design principle

The DAML contract stays the single source of truth for state and party
authority, regardless of which rail executes a payment. The ledger should
never need to know the mechanics of ACH batch windows or wire cutoffs. It
only needs to know: who authorized this disbursement, how much, and to
which of two buckets, on-chain or off-chain.

## Proposed component: RailRouter

A new Go package, `internal/railrouter`, sitting behind the existing
`Disburse` handler.

```
RailRouter.Route(settlementTerms, rail SettlementRail) error
  - rail == Stablecoin  → existing BitGo/Circle StablecoinFactory path
  - rail == Fiat        → new FiatProvider interface, first implementation
                           targets a payments orchestration API
```

`FiatProvider` interface, minimal surface:

```go
type FiatProvider interface {
    InitiateTransfer(ctx context.Context, req TransferRequest) (TransferRef, error)
    GetStatus(ctx context.Context, ref TransferRef) (TransferStatus, error)
    RegisterWebhook(ctx context.Context, handler WebhookHandler) error
}
```

This keeps the fiat provider swappable. Modern Treasury is the leading
candidate given its stated support for ACH, RTP, FedNow, and wires through
one API, but the interface should not assume any single vendor stays fixed.

## Why a payments-orchestration provider specifically

Three things it removes from this project's build list, worth naming
plainly.

Ledger and reconciliation. A payments orchestration platform typically
ships its own ledger for tracking money movement across rails, separate
from your DAML ledger. You are not trying to replace that, you are
treating it as the system of record for the fiat leg only, while DAML
stays system of record for escrow state and authority.

Compliance screening on the fiat leg. Sanctions screening, KYC checks on
external accounts, these belong at the payments layer, not encoded as a
DAML choice guard. Your `REGULATORY_CONFORMANCE.md` currently covers GDPR
and data sovereignty. It does not yet cover payout-side AML, and it
shouldn't need to if this sits at the provider.

Rail abstraction that matches your own philosophy. You already built a
pluggable custody factory for stablecoin providers. A `FiatProvider`
interface is the same pattern, applied one layer over.

## Open questions for the PR discussion

Where does the webhook confirmation land. Likely a new endpoint under
`/cmd`, parallel to the existing Oracle webhook pattern already in the
system stack diagram.

Does `ConfirmFiatSettlement` require the same OIDC-scoped authority as
other Issuer choices, or a separate service-principal scope for
webhook-originated calls. Recommend the latter, distinct from human-driven
Issuer actions, given `IDENTITY.md` already distinguishes Contributor from
Deployment Service principals.

Whether `FIAT_PENDING` needs its own row in the Operational Velocity
dashboard's stage-duration heatmap. Probably yes. Fiat settlement latency
is a genuinely different distribution than on-chain confirmation, and
collapsing them into one SETTLED bucket would hide that.

---

*End of FIAT-SETTLEMENT.md — new file, proposed for future-roadmap PR*
