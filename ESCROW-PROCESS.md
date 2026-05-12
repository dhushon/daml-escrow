# ESCROW PROCESS DIRECTIVES
## DAML Design Scaffolding — Tokenized Reserve Escrow

> **Purpose:** Process scaffolding directives for DAML contract design generation.
> Covers a bilateral escrow lifecycle supporting stablecoin and tokenized digital assets,
> with mediator-governed dispute resolution and partial settlement support.

---

## DIRECTIVE 01 — PARTIES & ROLES

```
PARTIES:
  - Depositor   : initiates escrow, deposits stablecoin, ratifies settlement
  - Beneficiary : counterparty, receives disbursement, ratifies settlement
  - Issuer      : escrow holder, custodies digital asset, executes disbursement
  - Mediator    : conditional activator, proposes settlement on dispute only
  - Observer    : (optional) read-only visibility e.g. regulator, auditor

AUTHORITY MODEL:
  - Issuer      : SIGNATORY on all states (holds asset, must authorize release)
  - Depositor   : SIGNATORY on creation and settlement ratification
  - Beneficiary : SIGNATORY on creation and settlement ratification
  - Mediator    : CONTROLLER on [ConfirmConditions, ProposeSettlement] only
  - Observer    : OBSERVER only, no choice rights
```

---

## DIRECTIVE 02 — ASSET MODEL

```
ASSET:
  - AssetType   : stablecoin | tokenized_reserve | digital_asset
  - AssetId     : unique token identifier
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
STATE SEQUENCE:
  DRAFT → FUNDED → ACTIVE → SETTLED | EXPIRED | CANCELLED

STATE DEFINITIONS:
  DRAFT       : Terms agreed, asset not yet deposited
  FUNDED      : Asset locked by Issuer, contract enforceable
  ACTIVE      : Awaiting condition fulfilment or dispute
  DISPUTED    : Party has raised dispute, Mediator is now active
  PROPOSED    : Mediator has submitted settlement proposal
  SETTLED     : Disbursement executed, contract closed
  EXPIRED     : Deadline passed without resolution, asset returned to Depositor
  CANCELLED   : Mutual withdrawal pre-condition test, asset returned to Depositor

RULE: No disbursement choice may be exercised before state = FUNDED
RULE: Mediator choices are disabled unless state = DISPUTED
```

---

## DIRECTIVE 04 — CONTRACT TERMS

```
TERMS BLOCK:
  - ConditionDescription  : Text (human-readable obligation)
  - ConditionType         : Binary | Partial | Milestone
  - EvidenceRequired      : DocumentRef | OracleSignal | MediatorAttestation
  - ExpiryDate            : Date (mandatory)
  - GracePeriodDays       : Int (optional extension window)
  - DisputeWindowDays     : Int (how long parties have to raise dispute post-condition)
  - PartialSchedule       : [(Decimal, ConditionRef)] (for milestone-based release)

RULE: ExpiryDate must be set at DRAFT state
RULE: PartialSchedule amounts must sum to total escrowed Amount
```

---

## DIRECTIVE 05 — CHOICES (Actions / Transitions)

```
CHOICE: Fund
  CONTROLLER  : Depositor
  GUARD       : state == DRAFT
  EFFECT      : state → FUNDED, CustodyRef recorded by Issuer

CHOICE: ConfirmConditions
  CONTROLLER  : Mediator
  GUARD       : state == ACTIVE
  EFFECT      : state → triggers Release or Partial Release

CHOICE: RaiseDispute
  CONTROLLER  : Depositor | Beneficiary
  GUARD       : state == ACTIVE, within DisputeWindowDays
  EFFECT      : state → DISPUTED

CHOICE: ProposeSettlement
  CONTROLLER  : Mediator
  GUARD       : state == DISPUTED
  EFFECT      : state → PROPOSED, SettlementTerms recorded

CHOICE: RatifySettlement
  CONTROLLER  : Depositor AND Beneficiary (both required)
  GUARD       : state == PROPOSED
  EFFECT      : state → SETTLED, triggers Disburse

CHOICE: RejectSettlement
  CONTROLLER  : Depositor | Beneficiary
  GUARD       : state == PROPOSED
  EFFECT      : state → DISPUTED (returns to mediator)

CHOICE: Disburse
  CONTROLLER  : Issuer
  GUARD       : state == SETTLED
  EFFECT      : executes asset transfer per settlement terms (full | partial | return)

CHOICE: Cancel
  CONTROLLER  : Depositor AND Beneficiary (both required)
  GUARD       : state ∈ [DRAFT, FUNDED, ACTIVE]
  EFFECT      : state → CANCELLED, full return to Depositor

CHOICE: ExpireEscrow
  CONTROLLER  : Issuer | any party (time-triggered)
  GUARD       : state ∈ [FUNDED, ACTIVE, DISPUTED], currentDate > ExpiryDate
  EFFECT      : state → EXPIRED, full return to Depositor
```

---

## DIRECTIVE 06 — SETTLEMENT TERMS MODEL

```
SETTLEMENT BLOCK:
  - SettlementType  : FullRelease | FullReturn | PartialSplit
  - DepositorReturn     : Decimal  (0.0 if FullRelease)
  - BeneficiaryPayment   : Decimal  (0.0 if FullReturn)
  - MediatorFee     : Decimal  (optional, deducted before split)

CONSTRAINT:
  DepositorReturn + BeneficiaryPayment + MediatorFee == escrowed Amount
```

---

## DIRECTIVE 07 — EVENTS & AUDIT LOG

```
EVENTS (append-only, DAML via archive + create pattern):
  - EscrowCreated       : timestamp, parties, asset, terms
  - EscrowFunded        : timestamp, CustodyRef, Amount
  - ConditionsConfirmed : timestamp, MediatorId, EvidenceRef
  - DisputeRaised       : timestamp, RaisingParty, Reason
  - SettlementProposed  : timestamp, SettlementTerms
  - SettlementRatified  : timestamp, DepositorSig, BeneficiarySig
  - SettlementRejected  : timestamp, RejectingParty
  - Disbursed           : timestamp, DepositorReturn, BeneficiaryPayment, MediatorFee
  - EscrowExpired       : timestamp, ReturnAmount
  - EscrowCancelled     : timestamp, ReturnAmount

RULE: Every state transition must emit a corresponding event contract
```

---

## DIRECTIVE 08 — GUARD RULES & INVARIANTS

```
INVARIANTS (must hold at all times):
  I1: escrowed Amount never changes after FUNDED
  I2: sum of all disbursements == escrowed Amount at SETTLED
  I3: Mediator may not act before state == DISPUTED
  I4: No party may act on an EXPIRED or CANCELLED or SETTLED contract
  I5: Issuer is signatory on every contract instance
  I6: Settlement ratification requires explicit positive consent from both Depositor and Beneficiary
  I7: ExpiryDate must be in the future at DRAFT creation
```

---

## DIRECTIVE 09 — DAML DESIGN HINTS

```
TEMPLATE STRUCTURE (suggested):
  - EscrowProposal    : DRAFT state, created by Depositor + Beneficiary co-sign
  - EscrowContract    : FUNDED → ACTIVE states, Issuer as signatory
  - DisputeRecord     : DISPUTED state, child contract
  - SettlementRecord  : PROPOSED state, child contract
  - DisbursementOrder : SETTLED, executed by Issuer

DAML PATTERNS TO APPLY:
  - Use Propose/Accept for bilateral creation (Depositor proposes, Beneficiary accepts)
  - Use Role contracts for Mediator to control activation scope
  - Use nonconsuming choice for read/query actions (e.g. GetStatus)
  - Use fetchByKey for CustodyRef and AssetId lookups
  - Represent PartialSchedule as [(Decimal, Text)] list in template payload
```

---

*End of ESCROW-PROCESS.md — v1.0*
