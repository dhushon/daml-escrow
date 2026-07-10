# ESCROW PROCESS DIRECTIVES

## DAML Design Scaffolding — Tokenized Reserve Escrow

> **Purpose:** Process scaffolding directives for DAML contract design generation.
> Covers a multi-party escrow lifecycle supporting stablecoin, tokenized digital
> assets, and fiat settlement, with mediator-governed dispute resolution,
> tiered arbitration escalation, milestone-independent partial settlement,
> and nested contract composition.
>
> Fiat-rail implementation detail lives in `FIAT-SETTLEMENT.md`. Frontend
> component requirements live in `FRONTEND-PROCESS.md`. This document
> defines the contract-level directives only.

---

## DIRECTIVE 01 — PARTIES & ROLES

```
PARTIES:
  - BuyerSet    : NonEmpty [Party]   (initiates escrow, deposits asset,
                                       ratifies settlement)
  - SellerSet   : NonEmpty [Party]   (counterparty, receives disbursement,
                                       ratifies settlement)
  - Issuer      : Party              (escrow holder, custodies asset,
                                       executes disbursement)
  - Mediator    : Party              (conditional activator, proposes
                                       settlement on dispute only)
  - Arbitrator  : Party (optional)   (binding decision authority on
                                       escalated disputes; may be the same
                                       Party as Mediator, but authority is
                                       granted separately)
  - Verifier    : Party (optional)   (milestone evidence verification;
                                       defaults to Mediator if unset)
  - Observer    : Party (optional)   (read-only visibility, e.g. regulator,
                                       auditor)

AUTHORITY MODEL:
  - Issuer      : SIGNATORY on all states (holds asset, must authorize release)
  - BuyerSet    : SIGNATORY on creation and settlement ratification, subject
                  to ConsentThreshold (see DIRECTIVE 10)
  - SellerSet   : SIGNATORY on creation and settlement ratification, subject
                  to ConsentThreshold
  - Mediator    : CONTROLLER on [ProposeSettlement] only
  - Arbitrator  : CONTROLLER on [RenderArbitrationDecision] only, and only
                  once state == ARBITRATION
  - Verifier    : CONTROLLER on [VerifyMilestone] only
  - Observer    : OBSERVER only, no choice rights

RULE: A single-buyer, single-seller escrow is simply BuyerSet/SellerSet
      with one member each and ConsentThreshold = 1. There is no separate
      "simple" template; the generalized model is the only model.
```

---

## DIRECTIVE 02 — ASSET MODEL

```
ASSET:
  - AssetType   : stablecoin | tokenized_reserve | digital_asset | fiat
  - AssetId     : unique token identifier (null if AssetType == fiat)
  - Amount      : Decimal (supports partial)
  - Currency    : ISO code or on-chain denomination
  - CustodyRef  : issuer-side ledger reference confirming lock

CONSTRAINTS:
  - Amount must be > 0 at FUNDED state
  - Partial settlement amounts must sum to <= total escrowed Amount
  - Asset must remain locked until a terminal choice executes
```

---

## DIRECTIVE 03 — CONTRACT STATES (Lifecycle Stages)

```
STATE SEQUENCE (primary path):
  DRAFT → FUNDED → ACTIVE → SETTLED | EXPIRED | CANCELLED

DISPUTE / ESCALATION PATH:
  ACTIVE → DISPUTED → PROPOSED → SETTLED
  PROPOSED → DISPUTED                        (rejection loops back)
  DISPUTED → ARBITRATION                     (on repeated rejection)
  ARBITRATION → SETTLED                      (binding decision)

FIAT SETTLEMENT SUB-STATE:
  SETTLED → FIAT_PENDING → [contract closed]
  (only reachable when SettlementRail == Fiat; see DIRECTIVE 14)

STATE DEFINITIONS:
  DRAFT        : Terms agreed, asset not yet deposited
  FUNDED       : Asset locked by Issuer, contract enforceable
  ACTIVE       : Awaiting milestone verification or dispute
  DISPUTED     : Party has raised dispute, Mediator is now active
  PROPOSED     : Mediator has submitted settlement proposal
  ARBITRATION  : Mediation exhausted, Arbitrator has binding authority
  SETTLED      : Disbursement decision finalized, executing or executed
  FIAT_PENDING : Disbursement instruction emitted to fiat rail, awaiting
                 external confirmation (ledger-recorded intent, not yet
                 confirmed money movement)
  EXPIRED      : Deadline passed without resolution, asset returned to
                 BuyerSet
  CANCELLED    : Mutual withdrawal pre-condition test, asset returned to
                 BuyerSet

RULE: No disbursement choice may be exercised before state = FUNDED
RULE: Mediator choices are disabled unless state = DISPUTED
RULE: Arbitrator choices are disabled unless state = ARBITRATION
RULE: A SETTLED contract with SettlementRail == Fiat is not closed until
      it exits FIAT_PENDING via ConfirmFiatSettlement
```

---

## DIRECTIVE 04 — CONTRACT TERMS

```
TERMS BLOCK:
  - ExpiryDate            : Date (mandatory)
  - GracePeriodDays       : Int (optional extension window)
  - DisputeWindowDays     : Int (how long parties have to raise dispute
                            post-milestone)
  - Milestones            : NonEmpty [MilestoneBlock] (see DIRECTIVE 11;
                            a single-condition escrow is simply one
                            Milestone carrying the full Amount)
  - ClearingMode          : Progressive | AllOrNone (see DIRECTIVE 16)
  - AccrualPolicy         : None | FixedRate | ReferenceRate (see DIRECTIVE 13)
  - EscalationThreshold   : Int, default 2 (rejected proposals before
                            ARBITRATION becomes reachable)

RULE: ExpiryDate must be set at DRAFT state
RULE: Milestone amounts must sum to total escrowed Amount
RULE: There is no contract-wide "condition" separate from Milestones; all
      release logic routes through the Milestone list, even when it has
      exactly one entry
```

---

## DIRECTIVE 05 — CHOICES (Actions / Transitions)

```
CHOICE: Fund
  CONTROLLER  : any member of BuyerSet
  GUARD       : state == DRAFT
  EFFECT      : state → FUNDED, CustodyRef recorded by Issuer

CHOICE: VerifyMilestone
  CONTROLLER  : Verifier (defaults to Mediator)
  GUARD       : state == ACTIVE, target Milestone.Status == Pending
  EFFECT      : Milestone.Status → Verified

CHOICE: ReleaseMilestone
  CONTROLLER  : Issuer
  GUARD       : target Milestone.Status == Verified
  EFFECT      : transfers Milestone.Amount only; remaining escrowed balance
                stays locked; Milestone.Status → Released; if this was the
                last unreleased Milestone, state → SETTLED

CHOICE: RaiseDispute
  CONTROLLER  : any member of BuyerSet | SellerSet
  GUARD       : state == ACTIVE, within DisputeWindowDays
  EFFECT      : state → DISPUTED
  NOTE        : may target a specific MilestoneId; disputes on one
                Milestone do not block VerifyMilestone/ReleaseMilestone on
                other, unrelated Milestones in the same contract

CHOICE: ProposeSettlement
  CONTROLLER  : Mediator
  GUARD       : state == DISPUTED
  EFFECT      : state → PROPOSED, SettlementTerms recorded

CHOICE: RatifySettlement
  CONTROLLER  : >= ConsentThreshold of BuyerSet AND >= ConsentThreshold of
                SellerSet, independently evaluated
  GUARD       : state == PROPOSED
  EFFECT      : state → SETTLED, triggers Disburse

CHOICE: RejectSettlement
  CONTROLLER  : any member of BuyerSet | SellerSet
  GUARD       : state == PROPOSED
  EFFECT      : state → DISPUTED, RejectionCount += 1 (returns to mediator)

CHOICE: EscalateToArbitration
  CONTROLLER  : any member of BuyerSet | SellerSet
  GUARD       : state == DISPUTED, RejectionCount >= EscalationThreshold
  EFFECT      : state → ARBITRATION, Mediator authority suspended for this
                contract, Arbitrator authority activated

CHOICE: RenderArbitrationDecision
  CONTROLLER  : Arbitrator
  GUARD       : state == ARBITRATION
  EFFECT      : state → SETTLED, decision is binding; does NOT require
                RatifySettlement from BuyerSet/SellerSet

CHOICE: Disburse
  CONTROLLER  : Issuer
  GUARD       : state == SETTLED
  EFFECT      : IF SettlementRail == Stablecoin: executes on-chain transfer
                per CIP-0056 logic, per WeightedSplit if set (DIRECTIVE 10)
                IF SettlementRail == Fiat: state → FIAT_PENDING, emits a
                DisbursementInstruction event; execution happens off-ledger
                via the rail integration described in FIAT-SETTLEMENT.md

CHOICE: ConfirmFiatSettlement
  CONTROLLER  : Issuer, via service-principal / webhook identity, not a
                human-driven Issuer action
  GUARD       : state == FIAT_PENDING, DisbursementInstruction unconfirmed
  EFFECT      : records external settlement confirmation, contract closes

CHOICE: Cancel
  CONTROLLER  : >= ConsentThreshold of BuyerSet AND >= ConsentThreshold of
                SellerSet
  GUARD       : state ∈ [DRAFT, FUNDED, ACTIVE]
  EFFECT      : state → CANCELLED, full return to BuyerSet

CHOICE: ExpireEscrow
  CONTROLLER  : Issuer | any party (time-triggered)
  GUARD       : state ∈ [FUNDED, ACTIVE, DISPUTED], currentDate > ExpiryDate
  EFFECT      : state → EXPIRED, full return to BuyerSet

CHOICE: SpawnChildEscrow
  CONTROLLER  : Issuer, on behalf of a Parent contract
  GUARD       : Parent.state == ACTIVE
  EFFECT      : creates a new EscrowContract, ParentRef recorded on the
                child (see DIRECTIVE 12)

CHOICE: SettleParent
  CONTROLLER  : Issuer
  GUARD       : AggregationRule condition met across ChildContractIds
  EFFECT      : Parent.state → SETTLED
```

---

## DIRECTIVE 06 — SETTLEMENT TERMS MODEL

```
SETTLEMENT BLOCK:
  - SettlementType  : FullRelease | FullReturn | PartialSplit
  - BuyerReturn     : Decimal  (0.0 if FullRelease)
  - SellerPayment   : Decimal  (0.0 if FullReturn)
  - MediatorFee     : Decimal  (optional, deducted before split)
  - WeightedSplit   : [(Party, Decimal)] (optional; shares must sum to 1.0
                      across SellerSet; defaults to equal split if unset
                      and SellerSet has more than one member)
  - SettlementRail  : Stablecoin | Fiat (see DIRECTIVE 14)

CONSTRAINT:
  BuyerReturn + SellerPayment + MediatorFee == escrowed Amount
```

---

## DIRECTIVE 07 — EVENTS & AUDIT LOG

```
EVENTS (append-only, DAML via archive + create pattern):
  - EscrowCreated           : timestamp, parties, asset, terms
  - EscrowFunded             : timestamp, CustodyRef, Amount
  - MilestoneVerified        : timestamp, MilestoneId, VerifierId, EvidenceRef
  - MilestoneReleased        : timestamp, MilestoneId, Amount
  - DisputeRaised            : timestamp, RaisingParty, Reason, MilestoneId?
  - SettlementProposed       : timestamp, SettlementTerms
  - SettlementRatified       : timestamp, BuyerSigs, SellerSigs
  - SettlementRejected       : timestamp, RejectingParty, RejectionCount
  - EscalatedToArbitration   : timestamp, RejectionCount
  - ArbitrationDecided       : timestamp, ArbitratorId, Decision
  - Disbursed                : timestamp, BuyerReturn, SellerPayment,
                                MediatorFee, SettlementRail
  - FiatSettlementConfirmed  : timestamp, DisbursementInstructionRef
  - ChildEscrowSpawned       : timestamp, ParentRef, ChildContractId
  - ParentSettled            : timestamp, AggregationRule, ChildContractIds
  - EscrowExpired            : timestamp, ReturnAmount
  - EscrowCancelled          : timestamp, ReturnAmount

RULE: Every state transition must emit a corresponding event contract
```

---

## DIRECTIVE 08 — GUARD RULES & INVARIANTS

```
INVARIANTS (must hold at all times):
  I1: escrowed PRINCIPAL Amount never changes after FUNDED, except via an
      explicit TopUp choice under a future CDP-backed escrow mode. Accrued
      yield (DIRECTIVE 13) is tracked as a separate ledger field and is
      not subject to this invariant.
  I2: sum of all Released Milestone amounts + remaining locked balance ==
      escrowed principal Amount at all times
  I3: Mediator may not act before state == DISPUTED; Arbitrator may not act
      before state == ARBITRATION
  I4: No party may act on an EXPIRED, CANCELLED, or fully closed SETTLED
      contract (a SETTLED contract with SettlementRail == Fiat is not
      "fully closed" until FIAT_PENDING resolves)
  I5: Issuer is signatory on every contract instance
  I6: Settlement ratification requires explicit positive consent meeting
      ConsentThreshold from both BuyerSet and SellerSet, independently
      evaluated, EXCEPT decisions rendered via RenderArbitrationDecision,
      which are binding without bilateral ratification
  I7: ExpiryDate must be in the future at DRAFT creation
  I8: A child contract's parties must be a subset of, or explicitly
      authorized by, its Parent's parties (no privilege escalation through
      nesting)
  I9: AccrualPolicy is immutable once set at FUNDED
```

---

## DIRECTIVE 09 — DAML DESIGN HINTS

```
TEMPLATE STRUCTURE (suggested):
  - EscrowProposal    : DRAFT state, created by BuyerSet + SellerSet co-sign
  - EscrowContract    : FUNDED → ACTIVE states, Issuer as signatory
  - MilestoneRecord   : child contract per MilestoneBlock, tracks Status
                        independently of contract-wide state
  - DisputeRecord     : DISPUTED state, child contract, optionally scoped
                        to a MilestoneId
  - SettlementRecord  : PROPOSED state, child contract
  - ArbitrationRecord : ARBITRATION state, child contract
  - DisbursementOrder : SETTLED, executed by Issuer, branches on
                        SettlementRail
  - ParentEscrow      : optional wrapper contract referencing
                        ChildContractIds for nested composition

DAML PATTERNS TO APPLY:
  - Use Propose/Accept for multi-party creation (BuyerSet proposes,
    SellerSet accepts, subject to ConsentThreshold)
  - Use Role contracts for Mediator, Arbitrator, and Verifier to control
    activation scope independently
  - Use nonconsuming choice for read/query actions (e.g. GetStatus)
  - Use fetchByKey for CustodyRef, AssetId, and MilestoneId lookups
  - Represent Milestones as a list of child contracts, never a flat
    payload field, so VerifyMilestone/ReleaseMilestone can target one
    without touching the others
  - Keep FiatProvider interaction entirely in the Go service layer; the
    DAML template only records DisbursementInstruction and
    FiatSettlementConfirmed events, per FIAT-SETTLEMENT.md
```

---

## DIRECTIVE 10 — MULTI-PARTY & SYNDICATED ESCROW

```
CONSENT MODEL:
  - ConsentThreshold : Nat, minimum signatures required from a party set

RULE: RatifySettlement and Cancel require signatures from >= ConsentThreshold
      members of BOTH BuyerSet and SellerSet, independently evaluated
RULE: Disbursement to SellerSet members follows WeightedSplit if set,
      otherwise defaults to equal split across the set
```

Use case: M&A escrows, real estate closings, and syndicated lending deals
where more than one buyer or seller sits on a side of the transaction, or
where a single disbursement pool splits across multiple beneficiaries.

---

## DIRECTIVE 11 — MILESTONE-INDEPENDENT PARTIAL RELEASE

```
MILESTONE BLOCK:
  - MilestoneId       : Text (unique within contract)
  - Amount            : Decimal
  - ConditionRef      : Text
  - EvidenceRequired  : DocumentRef | OracleSignal | MediatorAttestation
  - VerifiedBy        : Party (Verifier, defaults to Mediator)
  - Status            : Pending | Verified | Released | Rejected
```

Construction draws, earnout tranches, and licensing payments clear on
independent schedules, without waiting on unrelated conditions elsewhere in
the same agreement. See DIRECTIVE 05 for the `VerifyMilestone` /
`ReleaseMilestone` choice pair.

---

## DIRECTIVE 12 — NESTED / COMPOSED ESCROW CONTRACTS

```
PARENT ESCROW:
  - ChildContractIds  : [ContractId EscrowContract]
  - AggregationRule   : AllChildrenSettled | AnyChildSettled | Custom
  - SharedExpiryDate  : Date (optional, overrides child ExpiryDate if earlier)
  - SharedMediator    : Party (optional, propagates to children at creation)
```

Use case: phased M&A where each closing condition is its own escrow,
sharing one overall deal expiry and one mediator, but independently
fundable and disputable. See DIRECTIVE 05 for `SpawnChildEscrow` and
`SettleParent`, and INVARIANT I8 for the authorization constraint.

---

## DIRECTIVE 13 — ACCRUAL & YIELD POLICY

```
ACCRUAL BLOCK:
  - AccrualPolicy     : None | FixedRate | ReferenceRate(OracleRef)
  - AccrualRateBps    : Int (basis points, if FixedRate)
  - AccrualBeneficiary: BuyerSet | SellerSet | Split(Decimal, Decimal) | Issuer
  - AccrualStartState : FUNDED | ACTIVE (when the accrual clock starts)
```

Tokenized reserves sitting in escrow for months represent real opportunity
cost. Silence on who owns that yield becomes a dispute source later.
`AccrualPolicy` is immutable once set at FUNDED (INVARIANT I9), specifically
to avoid retroactive disputes over terms.

---

## DIRECTIVE 14 — MULTI-RAIL DISBURSEMENT (FIAT + STABLECOIN)

```
RAIL BLOCK (part of SETTLEMENT BLOCK, DIRECTIVE 06):
  - SettlementRail    : Stablecoin | Fiat
  - FiatRailProvider  : Text (e.g. "modern-treasury"; null if Stablecoin)
  - FiatDestination   : ExternalAccountRef (ABA/routing, IBAN, etc.)
  - FiatRailType      : ACH | RTP | FedNow | Wire (if FiatRailProvider set)
```

The DAML contract is the source of truth for STATE and AUTHORITY. It is
explicitly NOT the source of truth for fiat settlement finality, that lives
with the fiat rail provider and is reconciled asynchronously through the
`FIAT_PENDING` sub-state (DIRECTIVE 03) and the `ConfirmFiatSettlement`
choice (DIRECTIVE 05). Implementation detail lives in `FIAT-SETTLEMENT.md`.

---

## DIRECTIVE 15 — TIERED DISPUTE ESCALATION

```
ESCALATION:
  - EscalationThreshold : Int, default 2 (rejected mediation proposals
                           before ARBITRATION becomes reachable)
  - Arbitrator           : distinct role from Mediator (DIRECTIVE 01); a
                           contract may name the same Party for both, but
                           the authority grant is separate
```

A rejected settlement proposal loops back to DISPUTED, but that loop needs
a forcing function. `EscalateToArbitration` and `RenderArbitrationDecision`
(DIRECTIVE 05) add a binding exit ramp for disputes mediation cannot
resolve.

---

## DIRECTIVE 16 — CLEARING AUTHORITY & CLEARING MODE

```
CLEARING AUTHORITY:
  - Verifier is the clearing party (DIRECTIVE 01). It evaluates evidence
    of term fulfillment in the ordinary course and authorizes release.
    Mediator and Arbitrator do not perform clearing; they only activate
    when a party contests a clearing decision Verifier already made, or
    the absence of one.

CLEARING MODE (contract-level policy, DIRECTIVE 04):
  - Progressive : each Milestone clears and releases independently as it
                  is Verified, per DIRECTIVE 11. This is the default.
  - AllOrNone   : no release fires until every Milestone in the contract
                  reaches Verified. A single ReleaseAll choice replaces
                  per-milestone ReleaseMilestone for this contract.

CHOICE: ReleaseAll (only valid when ClearingMode == AllOrNone)
  CONTROLLER  : Issuer
  GUARD       : every Milestone.Status == Verified
  EFFECT      : executes the Effect (DIRECTIVE 17) attached to each
                Milestone in a single atomic transaction; state → SETTLED

RULE: ClearingMode is set at DRAFT and is immutable thereafter, for the
      same reason AccrualPolicy is immutable (INVARIANT I9) — changing the
      release model mid-contract is itself a dispute waiting to happen
```

---

## DIRECTIVE 17 — EFFECT ABSTRACTION

```
A Milestone's "release" is not always a fund transfer. Generalize what
fires when a Milestone or ReleaseAll clears:

EFFECT BLOCK (attached to each MilestoneBlock, DIRECTIVE 11):
  - EffectType  : FundTransfer | AccessGrant | AccessRevoke |
                  Custom(WebhookRef)
  - EffectTarget: Party | ExternalSystemRef (e.g. license server,
                  customs broker system)
  - EffectPayload: Decimal (FundTransfer) | AccessScopeRef (AccessGrant/
                  Revoke) | Text (Custom)

RULE: A single Milestone may carry more than one Effect (e.g. a customs
      clearance Milestone fires a FundTransfer of the remaining balance
      to Seller and, separately, a Custom effect notifying a freight
      forwarder system that the shipment is clear)
RULE: FundTransfer effects still obey DIRECTIVE 14's rail split
      (Stablecoin vs Fiat); non-fund Effects are not rail-dependent and
      execute via the Go service layer's external-system integration,
      not via Disburse
RULE: Effect execution failure (e.g. the license server is unreachable)
      does not roll back Milestone.Status; it is logged as a distinct
      EffectFailed event and retried, since the clearing decision itself
      already stands
```

This is the piece that makes the two scenarios below expressible without
inventing a separate contract type for each.

---

## DIRECTIVE 18 — RECURRING ESCROW & EXTERNAL ACCESS GATING

```
For cases where escrow governs ongoing access rather than a one-time
disbursement, e.g. software support subscriptions, SaaS entitlements,
recurring licensing fees.

RECURRING BLOCK (extends DIRECTIVE 04's TERMS BLOCK):
  - RecurrencePolicy    : None | Recurring(PeriodDays, PeriodAmount)
  - CurrentPeriodId     : Text (auto-generated, one per cycle)
  - GraceWindowDays     : Int (days past PeriodDays before lapse fires)

STATE ADDITION:
  ACTIVE ⇄ LAPSED
  (a recurring contract oscillates between ACTIVE and LAPSED; it does not
  route through DISPUTED for a missed payment, since non-payment is not a
  disagreement about facts, it's simply a fact)

CHOICE: OpenPeriod
  CONTROLLER  : Issuer (time-triggered at PeriodDays boundary)
  GUARD       : RecurrencePolicy == Recurring, state == ACTIVE
  EFFECT      : creates a new MilestoneBlock scoped to CurrentPeriodId
                carrying PeriodAmount, Status == Pending

CHOICE: LapsePeriod
  CONTROLLER  : Issuer (time-triggered)
  GUARD       : currentDate > period end + GraceWindowDays,
                CurrentPeriodId Milestone.Status != Released
  EFFECT      : state → LAPSED, fires an AccessRevoke Effect
                (DIRECTIVE 17) targeting the external license or
                entitlement system

CHOICE: ResumePeriod
  CONTROLLER  : Issuer
  GUARD       : state == LAPSED, outstanding period Milestone reaches
                Released (payment caught up)
  EFFECT      : state → ACTIVE, fires an AccessGrant Effect

RULE: A recurring contract has no single ExpiryDate representing final
      settlement; ExpiryDate instead represents the term's end date, after
      which no further OpenPeriod cycles fire and the contract proceeds to
      SETTLED via its final period's Milestone
RULE: LAPSED does not return escrowed principal to BuyerSet the way
      EXPIRED or CANCELLED do; it only suspends the AccessGrant Effect
      until payment resumes
RULE: Disputing a specific period's charge uses RaiseDispute scoped to
      that period's MilestoneId (DIRECTIVE 05), same as any other
      milestone dispute; LapsePeriod itself is not disputable, only the
      underlying charge is
```

---

## DIRECTIVE 19 — GEOSPATIAL & CUSTOMS EVENT TRIGGERS (Import/Export)

```
For cases where milestone clearing depends on a physical or regulatory
event, e.g. a vessel crossing territorial waters, entry filed with
customs, customs clearance granted, rather than a document a Verifier
reviews by hand.

ORACLE PRINCIPAL:
  - Verifier (DIRECTIVE 16) may be a Party representing an automated
    oracle service, not only a human. This does not change its authority
    grant; VerifyMilestone still requires the Verifier's signature, that
    signature is simply produced by a signed feed rather than a UI click.
  - Oracle signatures use the existing asymmetric HSM-backed signing
    infrastructure (GCP KMS), consistent with the oracle pattern already
    in place for stablecoin triggers.

GEOFENCE EVIDENCE (extends EvidenceRequired: OracleSignal, DIRECTIVE 11):
  - BoundaryType   : TerritorialWaters | ContiguousZone | CustomsZone |
                      PortOfEntry
  - DistanceNM     : Decimal (e.g. 12, 24, 100), null if PortOfEntry
  - Direction      : Inbound | Outbound
  - VesselRef      : IMO number or equivalent tracking identifier
  - SourceFeed     : AIS provider reference

CUSTOMS EVIDENCE (extends EvidenceRequired: OracleSignal, DIRECTIVE 11):
  - CustomsEventType : EntryFiled | InspectionPassed | Cleared
  - Jurisdiction      : ISO country code
  - DeclarationRef    : customs filing identifier
  - SourceFeed         : customs authority API or licensed customs broker
                          integration

EXAMPLE MILESTONE SEQUENCE (Progressive ClearingMode):
  Milestone 1 : BoundaryType = TerritorialWaters, DistanceNM = 100,
                Direction = Inbound → releases an agreed partial percentage
  Milestone 2 : CustomsEventType = EntryFiled → releases a further
                percentage
  Milestone 3 : CustomsEventType = Cleared → releases the remainder
                (or the full balance, if Milestones 1-2 were withheld
                rather than partially released, per SettlementTerms)

RULE: An oracle-sourced VerifyMilestone still produces a normal
      MilestoneVerified event (DIRECTIVE 07); it is not a distinct event
      type, so audit and dispute handling do not need to special-case it
RULE: If SourceFeed stops reporting (vessel goes dark, customs API is
      unreachable), the Milestone simply stays Pending; this is not an
      automatic dispute. A party may still RaiseDispute if it believes an
      oracle-reported crossing or filing is incorrect, same as disputing
      any other milestone
RULE: GeofenceEvidence and CustomsEvidence are evaluated independently;
      a contract may require both a boundary crossing and a customs event
      on the same Milestone (AND), or split them across separate
      Milestones as in the example above (sequential)
```

---

*End of ESCROW-PROCESS.md*
