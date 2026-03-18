# PLAN.md --- Stablecoin Escrow Platform

## Project Objective

Build a privacy-preserving **stablecoin escrow platform** using **DAML
smart contracts** and **Go-based services** for integration, APIs, and
orchestration.

------------------------------------------------------------------------

## Phase 2 --- Backend Platform (Go) (IN PROGRESS)

Goals: Create service layer to interact with the ledger.

### Completed Work
1. **REST API:** Swagger UI integrated on port 8080.
2. **JSON V2 Client:** Implemented `JsonLedgerClient` with dynamic party resolution and robust event extraction.
3. **Multi-Milestone Support:** Handle dynamic milestone lists.
4. **Mediator Services:** `ResolveDispute` choice implemented.
5. **Stablecoin Settlement:** Added `EscrowSettlement` lifecycle for Central Bank payout finalization.
6. **Integration Testing:** 100% coverage of core lifecycles.

### Detailed Next Steps (Closing Phase 2)

**Task 2.11: Awareness & Analytics (NEW)**
- Refactor `ListEscrows` to support role-based filtering (Buyer vs. Seller vs. Bank view).
- Implement `GetMetrics` to provide server-side aggregation (Total Under Management, Forward Gains).
- Add `GET /metrics` and role filters to `GET /escrows` API.

**Task 2.12: Final Phase 2 Cleanup**
- Refactor `internal/ledger/daml_client.go` to match the expanded interface.
- Standardize error handling for "Contract Not Found" vs "Ledger Error".

------------------------------------------------------------------------

## Phase 3 --- Oracle Integrations (NEXT)
...
