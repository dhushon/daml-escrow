# Contract Strategy: Persistent & Extensible Escrow

This document details the architectural strategy for the Stablecoin Escrow platform, focusing on the decoupling of generic financial dynamics from industry-specific business metadata.

------------------------------------------------------------------------

## 1. Core Escrow Dynamics (The "Stable" Layer)

The platform utilizes a generic Daml template (`StablecoinEscrow`) to manage the primary financial lifecycle. By keeping this layer agnostic of business domain, we ensure high performance, auditability, and reuse.

### Structural Elements
- **Parties:** `Buyer`, `Seller`, `Issuer` (Central Bank), and `Mediator`.
- **Financials:** `totalAmount`, `currency`.
- **Workflow:** `currentMilestoneIndex` tracks progress through an array of `Milestones`.
- **State:** Explicit transitions between `Active` and `Disputed` (via the `DisputedEscrow` template).

### Mechanics
1. **Approval:** The `Buyer` (or an authorized Oracle acting as Buyer) exercises `ApproveMilestone`.
2. **Settlement:** Each approval triggers the creation of an `EscrowSettlement` contract, signaling the `Issuer` to release stablecoins.
3. **Dispute:** If terms are not met, the `Buyer` exercises `RaiseDispute`, freezing further milestones until the `Mediator` resolves the split.

------------------------------------------------------------------------

## 2. Schema-Driven Metadata (The "Flexible" Layer)

Domain-specific "Contracted Elements" (e.g., Serial Numbers, Parcel IDs, Tracking Numbers) are not hardcoded into the Daml templates. Instead, they are stored as a JSON-serialized string in the `metadata` field.

### The Metadata Container
```go
type EscrowMetadata struct {
	SchemaURL string                 `json:"schemaUrl"`
	Payload   map[string]interface{} `json:"payload"`
}
```

### Strategic Benefits
- **Evolution:** New business domains can be added by creating a new JSON Schema without changing the smart contracts.
- **Validation:** Metadata can be validated off-chain (in the Go API) or by Oracles using standard JSON Schema tools.
- **Privacy:** Detailed metadata can be hashed or encrypted before being stored in the Daml `metadata` field if absolute privacy is required.

------------------------------------------------------------------------

## 3. Business Domain Examples

### Domain A: High-Value Equipment Leasing
Focuses on security deposits and usage-based milestones.
- **Schema:** `leasing.json`
- **Metadata Example:**
```json
{
  "schemaUrl": "https://stablecoin-escrow.io/schemas/leasing.v1.json",
  "payload": {
    "assetId": "CAT-320-XYZ",
    "assetType": "Machinery",
    "securityDepositAmount": 15000.0,
    "leaseTermDays": 365,
    "usageLimits": { "maxHours": 2000 }
  }
}
```

### Domain B: International Supply Chain
Focuses on logistics events and proof of delivery.
- **Schema:** `supply-chain.json`
- **Metadata Example:**
```json
{
  "schemaUrl": "https://stablecoin-escrow.io/schemas/supply-chain.v1.json",
  "payload": {
    "trackingNumber": "FEDEX-998877",
    "carrier": "FedEx",
    "billOfLading": "hash:a1b2c3...",
    "origin": "CN-SHA",
    "destination": "US-LAX"
  }
}
```

### Domain C: Federal Grants (GREAT Act)
Focuses on accountability and standardized identifier compliance.
- **Schema:** `grants.json`
- **Metadata Example:**
```json
{
  "schemaUrl": "https://stablecoin-escrow.io/schemas/grants.v1.json",
  "payload": {
    "opportunityId": "O-HHS-2026",
    "assistanceListing": "93.000",
    "uei": "ABC123XYZ456",
    "grantAwardNumber": "G-2026-001"
  }
}
```

------------------------------------------------------------------------

## 4. Implementation Mechanics

1. **API Ingestion:** The Go API receives a `CreateEscrowRequest` containing the `Metadata` object.
2. **Persistence:** The API marshals the `Metadata` into a JSON string and submits it to the Daml ledger.
3. **Oracle Trigger:** When an Oracle sends a webhook, it identifies the contract by `escrowId`. The service fetches the contract, retrieves the `SchemaURL` from the metadata, and validates the event against that schema before advancing the milestone.
