# Canton Ledger Management Guide (Daml 3.4.11)

This guide documents the setup, configuration, and integration of the Canton 3.x ledger for the Stablecoin Escrow project.

## 1. Prerequisites & Installation

### Java Environment
- **Requirement:** JDK 17 (LTS) is mandatory at `/opt/homebrew/opt/openjdk@17`.
- **Verification:** `java -version` should report 17.x.

### Daml Package Manager (DPM)
Daml 3.x uses `dpm` instead of the legacy `daml` assistant.
- **Binary:** `/Users/dhushon/.dpm/bin/dpm`

---

## 2. Configuration

### Canton Configuration (`contracts/sandbox.conf`)
Minimal HOCON file to enable the HTTP JSON API:
```hocon
canton {
  participants {
    sandbox {
      http-ledger-api {
        port = 7575
        address = 127.0.0.1
      }
    }
  }
}
```

### Bootstrap Script (`contracts/init.canton`)
Automates environment readiness:
1. Connects `sandbox` to `mysynchronizer`.
2. Allocates and **Authorizes** parties (`CentralBank`, `Buyer`, `Seller`, `EscrowMediator`) on the topology.
3. Maps human-readable User IDs (e.g., `"Buyer"`) to Canton Party IDs.
4. Grants `actAs` rights for multi-party operations.

---

## 3. Launching the Ledger

### Full Start Sequence
```bash
make sandbox
# Wait for port 6865
make ledger-setup
```

---

## 4. Integration Strategy

### User-Based Identities
The Go backend (`JsonLedgerClient`) interacts with the ledger using **User IDs**.
- It dynamically resolves User IDs to raw Canton Party IDs at startup via `/v2/parties`.
- **Authorization:** `Buyer` user can `actAs` both `Buyer` and `CentralBank`.

### JSON API V2
- **Create:** `/v2/commands/submit-and-wait`
- **Query:** `/v2/state/active-contracts` (Requires `filtersByParty` or `filtersForAnyParty`)

---

## 5. Verified Synthetic Transaction (JSON V2)

To verify the ledger manually, use this exact structure:

```bash
# 1. Get current Party IDs
curl -s http://localhost:7575/v2/parties

# 2. Create Escrow (Replace IDs with output from step 1)
curl -X POST http://localhost:7575/v2/commands/submit-and-wait \
  -H "Content-Type: application/json" \
  -d '{
    "commandId": "test-001",
    "actAs": ["BUYER_ID", "CB_ID"],
    "userId": "Buyer",
    "commands": [
      {
        "CreateCommand": {
          "templateId": "ec35fce924adbefbae43d1f546879c29fdc42b9efac531f4de8eaeb39a5693c1:StablecoinEscrow:StablecoinEscrow",
          "createArguments": {
            "issuer": "CB_ID",
            "buyer": "BUYER_ID",
            "seller": "SELLER_ID",
            "mediator": "MEDIATOR_ID",
            "totalAmount": "100.0000000000",
            "currency": "USD",
            "description": "Manual Test",
            "milestones": [{"label":"Full","amount":"100.0000000000","completed":false}],
            "currentMilestoneIndex": 0
          }
        }
      }
    ]
  }'
```

---

## 6. Debugging

### View Party Mappings
```bash
curl -s http://localhost:7575/v2/parties | jq
```

### View Update Stream (Flat)
```bash
curl -X POST http://localhost:7575/v2/updates/flats \
  -H "Content-Type: application/json" \
  -d '{"beginExclusive": 0, "filter": {"filtersByParty": {"BUYER_ID": {"cumulative": []}}}}'
```
