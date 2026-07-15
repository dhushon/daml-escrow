# FRONTEND-PROCESS.md

## UI/UX Scaffolding Directives — Complex Escrow Frontend

> **Purpose:** Companion to `ESCROW-PROCESS.md`. Defines the frontend
> component and state-handling requirements needed so the Astro dashboard
> reflects the reworked contract model rather than the single-buyer,
> single-seller, single-condition assumptions it currently encodes.
>
> Scope: `frontend/` (Astro components, `api.ts` client, `/metrics`
> dashboard). No backward compatibility constraint applies; existing
> components may be restructured or replaced outright.

---

## DIRECTIVE F1 — STATE MACHINE & COMPONENT MAPPING

```
CURRENT (api.ts, EscrowCard.astro):
  assumes DRAFT → FUNDED → ACTIVE → DISPUTED → PROPOSED → SETTLED only

REQUIRED:
  - Add ARBITRATION and FIAT_PENDING as first-class render states
  - EscrowCard must render a distinct visual treatment for ARBITRATION
    (binding-decision pending, mediator authority suspended) versus
    DISPUTED (mediator still active)
  - FIAT_PENDING must render as visually distinct from SETTLED; a contract
    is not "done" from the user's perspective until fiat confirmation lands
  - Remove any component logic keyed on a single `buyer` / `seller` field;
    replace with `buyerSet` / `sellerSet` arrays throughout
```

---

## DIRECTIVE F2 — MULTI-PARTY CONSENT VIEW

```
COMPONENT: PartySetRoster (new)

REQUIRED FIELDS PER MEMBER:
  - PartyId, DisplayName, Role (BuyerSet | SellerSet)
  - RatificationStatus : Pending | Signed
  - SignedAt : timestamp (if Signed)

REQUIRED BEHAVIOR:
  - Render each set as a roster, not a single name field
  - Show a threshold progress indicator (e.g. "2 of 3 signed, 2 required")
    against ConsentThreshold, not just a binary signed/unsigned toggle
  - RatifySettlement action button is disabled per-set until that set's
    ConsentThreshold is met, independently of the other set
```

---

## DIRECTIVE F3 — MILESTONE BOARD

```
COMPONENT: MilestoneBoard (replaces single condition-status indicator)

REQUIRED BEHAVIOR:
  - Render Milestones as a list of independent cards, each with its own
    Status (Pending | Verified | Released | Rejected)
  - Verify and Release actions are scoped to a single MilestoneId; a
    Verifier or Issuer acting on one milestone must not require or imply
    action on any other milestone in the same contract
  - A dispute raised against one MilestoneId shows a flag on that card
    only; sibling milestone cards remain actionable
  - Progress summary at the top of the board (e.g. "3 of 5 milestones
    released, $42,000 of $100,000 disbursed") derived from Milestone data,
    not from contract-level state alone
```

---

## DIRECTIVE F4 — PARENT / CHILD ESCROW NAVIGATION

```
COMPONENT: EscrowRollup (new)

REQUIRED BEHAVIOR:
  - A parent escrow renders a summary card linking to each child contract
  - Child status rolls up into a parent-level progress indicator, computed
    according to the parent's AggregationRule (AllChildrenSettled shows
    "3 of 4 settled"; AnyChildSettled shows first-settled prominently)
  - Navigating into a child escrow shows a breadcrumb back to the parent;
    a child never renders as if it were a standalone top-level escrow when
    a ParentRef is present
```

---

## DIRECTIVE F5 — ARBITRATION FLOW UI

```
REQUIRED BEHAVIOR:
  - After RejectionCount reaches EscalationThreshold, surface an
    "Escalate to Arbitration" action to eligible parties (BuyerSet |
    SellerSet members), not just a passive counter
  - Once in ARBITRATION, hide Mediator-facing actions entirely; show only
    the Arbitrator's pending-decision view
  - RenderArbitrationDecision outcome displays a "binding decision, no
    further ratification required" notice, distinct from the mediated
    settlement flow's ratify/reject buttons, so users don't expect a vote
```

---

## DIRECTIVE F6 — RAIL SELECTION & FIAT_PENDING UX

```
COMPONENT: SettlementRailSelector (new, appears at settlement-terms entry)

REQUIRED FIELDS:
  - Rail toggle: Stablecoin | Fiat
  - If Fiat: FiatRailType (ACH | RTP | FedNow | Wire), external account
    entry (routing/account or IBAN depending on locale)

REQUIRED BEHAVIOR:
  - FIAT_PENDING state shows a status indicator distinct from a loading
    spinner; label it by what's actually happening ("Awaiting bank
    confirmation") rather than a generic "Processing"
  - Poll or subscribe for FiatSettlementConfirmed rather than assuming
    completion on Disburse; do not optimistically render SETTLED-closed
    before ConfirmFiatSettlement actually fires
  - Surface FiatRailProvider-reported failures distinctly from DAML-level
    guard failures; these are different failure classes with different
    remediation paths for the user
```

---

## DIRECTIVE F7 — DASHBOARD EXTENSION (/metrics)

```
REQUIRED ADDITIONS to Operational Velocity dashboard:
  - Milestone-level funnel: proposed → verified → released, per milestone,
    not just contract-level DRAFT → SETTLED funnel
  - Separate latency band for FIAT_PENDING duration, distinct from
    on-chain confirmation latency; these have different expected
    distributions and should not be averaged together
  - Arbitration rate: percentage of DISPUTED contracts that escalate to
    ARBITRATION rather than resolving via mediation, as a signal of
    mediator effectiveness
```

---

## DIRECTIVE F8 — CLEARING MODE, FUNDING THRESHOLD & LAPSED ACCESS UI

```
COMPONENT: ClearingModeBanner (new)

REQUIRED BEHAVIOR:
  - When ClearingMode == AllOrNone, the MilestoneBoard (DIRECTIVE F3) must
    visually communicate that no individual card can release on its own;
    show one combined progress bar toward ReleaseAll instead of per-card
    release buttons
  - When ClearingMode == Progressive, retain the per-card release behavior
    from DIRECTIVE F3 unchanged

COMPONENT: RecurringEntitlementStatus (new, for DIRECTIVE 18 contracts)

REQUIRED BEHAVIOR:
  - Render ACTIVE / LAPSED as an access-status indicator ("Access active"
    / "Access suspended, payment overdue"), not as a contract-health
    warning; a LAPSED recurring contract is an expected, recoverable
    state, not an error condition
  - Show the current period's due date and GraceWindowDays countdown
  - ResumePeriod availability should be immediately visible once payment
    clears, so the user isn't left wondering whether access will return
    automatically or requires a separate action
```

---

## DIRECTIVE F9 — SHIPMENT TRACKING VISUALIZATION

```
COMPONENT: ShipmentMilestoneTracker (new, for DIRECTIVE 19 contracts)

REQUIRED BEHAVIOR:
  - Render geofence-based Milestones with a map or route indicator showing
    the vessel's position relative to the relevant boundary, not just a
    text status; distance-to-boundary is more useful to the user than a
    bare Pending/Verified label
  - Render customs-based Milestones as a distinct step type from geofence
    Milestones within the same MilestoneBoard (DIRECTIVE F3), since they
    represent different evidence sources
  - A Milestone stuck in Pending due to a stale SourceFeed (DIRECTIVE 19)
    must render differently from one legitimately awaiting an event still
    in the future; show last-updated timestamp for the feed itself, not
    just the Milestone status, so the user can tell the two apart
```

---

## DIRECTIVE F10 — AI RECOMMENDATION & AUTO-APPROVAL DISPLAY

```
REQUIRED BEHAVIOR:
  - An AIRecommendation (ESCROW-PROCESS.md DIRECTIVE 20) renders in a
    visually distinct panel from the Mediator's or Arbitrator's actual
    decision; never merge the two into one card, since one is advisory
    and the other is the binding action
  - A Milestone cleared via AutoVerifyMilestone shows "auto-cleared,
    signed by [TrustedSigner]" — never "AI-approved" or any label
    implying the AI made the clearing decision
  - Attach the source AttestedNote or DocumentRef inline wherever an
    AIRecommendation is shown, so the human reviewer sees the underlying
    evidence next to the summary, not the summary in isolation
```

---

*End of FRONTEND-PROCESS.md*
