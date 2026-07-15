# AI-SERVICES.md — AI-Assisted Authoring & Advisory Services

## Purpose

Extends the intelligent ingest engine already established for legacy
document ingestion into a broader set of AI-assisted services: contract
construction, translation, change detection, and mediation advisory.
None of these grant AI a signatory or controller role on any DAML choice.
AI produces drafts, translations, diffs, and recommendations; a Party
still exercises every choice that changes contract state. The one runtime
exception, trusted-signer auto-approval, is a signature check, not an AI
decision, and is defined in `ESCROW-PROCESS.md` DIRECTIVE 20, not here.

---

## A1 — Novel Contract Construction

```
INPUT   : natural-language brief describing the deal (parties, asset,
          conditions, jurisdiction)
OUTPUT  : a populated EscrowProposal in DRAFT state — Parties,
          Milestones, ClearingMode, SettlementTerms — presented to the
          proposing party for review before Fund is reachable

RULE: AI-constructed terms enter the same DRAFT state and require the
      same BuyerSet/SellerSet co-signature as a manually drafted contract;
      construction assistance does not shortcut DIRECTIVE 01's authority
      model
RULE: The brief itself is retained alongside the contract as provenance,
      distinct from the terms it produced, so a later dispute can
      reference what was actually asked for versus what was drafted
```

## A2 — Template Transfer

```
INPUT   : an existing, previously-ratified contract (or contract type)
          plus a new counterparty/deal context
OUTPUT  : a new DRAFT carrying the prior contract's structure, Milestone
          shape, and ClearingMode, with party-specific and deal-specific
          fields replaced

RULE: Transfer never carries over party identities, amounts, or
      AttestedNote history from the source contract; only the
      structural shape (Milestone count, EvidenceRequired types,
      ClearingMode, AccrualPolicy shape) transfers
```

## A3 — Hybrid Extension of an Existing Contract

```
INPUT   : an ACTIVE or FUNDED contract plus a requested change (add a
          Milestone, adjust AccrualPolicy, add a SharedMediator, etc.)
OUTPUT  : a proposed amendment, NOT an in-place mutation

RULE: Hybrid extension must route through an explicit amendment choice
      requiring the same ConsentThreshold ratification as original terms
      (see ESCROW-PROCESS.md DIRECTIVE 10); AI proposing a change is not
      different from a party proposing a change, and does not bypass
      INVARIANT I9's immutability rules for AccrualPolicy or ClearingMode
RULE: An extension proposal that would violate an existing invariant
      (e.g. retroactively changing AccrualPolicy after FUNDED) is
      rejected at the drafting stage, before it ever reaches a party for
      signature, so parties aren't asked to ratify something the ledger
      would refuse anyway
```

## A4 — Legal Translation

```
INPUT   : ratified contract terms in their authoring language
OUTPUT  : a rendered translation for a counterparty's working language

RULE: A translation is always a derived, informational artifact. The
      DAML contract terms in their original authoring language remain
      the sole authoritative version for any dispute, mediation, or
      arbitration
RULE: Every translated document is labeled with the source contract's
      hash or ContractId and a "translation, not the governing text"
      notice, rendered wherever the translation is displayed, not just
      in a footnote
```

## A5 — Authoring & Change Detection

```
AUTHORING:
  Drafting assistance for TERMS BLOCK language, Milestone descriptions,
  and SettlementTerms, building on the existing Gemini-based ingest
  engine's document classification and extraction capability.

CHANGE DETECTION:
  INPUT   : a newly uploaded document (e.g. a revised shipping agreement,
            an amended purchase order) plus the current contract terms
  OUTPUT  : a structured diff — which fields changed, which stayed the
            same — presented to the parties before any amendment
            (DIRECTIVE A3) is proposed

RULE: Change detection output is advisory input to a human decision about
      whether to propose an amendment; it never triggers A3 automatically
RULE: A detected change that would affect an already-Released Milestone's
      terms is flagged distinctly from a change affecting a still-Pending
      Milestone, since the former may require a dispute rather than an
      amendment
```

## A6 — Mediation Advisory

```
INPUT   : DisputeRecord, the Milestone(s) in question, associated
          evidence (DocumentRef, OracleSignal, AttestedNote)
OUTPUT  : an AIRecommendation (ESCROW-PROCESS.md DIRECTIVE 20) — a
          summary of the evidence, a confidence indicator, and a
          rationale, presented to Mediator before ProposeSettlement and
          to Arbitrator before RenderArbitrationDecision

RULE: AIRecommendation is generated fresh for ARBITRATION review even if
      one was already produced during mediation; escalation means the
      prior recommendation didn't resolve the dispute, and stale
      reasoning shouldn't be presented as current
RULE: Mediator and Arbitrator actions are recorded as their own; the
      event log never attributes a ProposeSettlement or
      RenderArbitrationDecision outcome to the AI, only to the Party who
      exercised the choice
```

---

*End of AI-SERVICES.md*
