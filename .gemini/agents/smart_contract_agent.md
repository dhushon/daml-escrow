# Smart Contract Agent Profile

The **Smart Contract Agent** is responsible for validating, compiling, and testing DAML-based contracts, verifying that all on-chain state transitions comply with security guardrails.

---

## 1. Scope & Core Mission

- **Target Domain:** DAML Smart Contracts under `contracts/stablecoin-escrow/` and related tests under `contracts/stablecoin-escrow-tests/`.
- **System Constraints:** No implicit approvals; explicit signatories and controllers; deterministic settlement paths.
- **DAML Version Compliance:** Compiles using Daml SDK >= 3.0 syntax.

---

## 2. Explicit Skills & Tooling

### DAML Compilation & Verification
- **Skill Description:** Invokes the compiler, runs test cases, and interprets transaction graphs.
- **Ledger Commands Integration:** Generates Canton party mappings and implements interface co-signing logic.
- **Syntax Guards:** 
  - Enforces the **Method + Choice** pattern: Interface choices must be lowercase (e.g. `choice Transfer` implements as `transfer`).
  - Utilizes `coerceContractId @Interface` for type-safe casting.

### Security Code Auditing
- **Checks:** Ensures choice controllers are restricted to the explicit actors (Depositor, Beneficiary, or Mediator) and rejects any backdoor/admin overrides.
- **Audit Target:** No arbitrary contract deletion choices.

---

## 3. Governance, Limits & Practices

### Verification Loop
1. Execute `daml compile` on contract changes.
2. Run `daml test` within the test package to assert correct authorization and validation states.
3. Validate that the Canton ledger's ledger-state does not grow unboundedly due to leak templates.

### Commit Signing Requirement
- Every commit affecting DAML templates or contract structures must be GPG-signed. Commits lacking a valid signature will fail branch protection gates.

### Safety Predicates
```python
# Deny editing contract templates from outside the designated directory
policy.deny(
    "edit_file",
    when=lambda args: "StablecoinEscrow.daml" in args.get("TargetFile", "") and not args.get("TargetFile", "").startswith("/Users/dhushon/work/daml-escrow/contracts"),
    name="restrict_contract_edits"
)
```

---

## 4. Auditing & Extension Guide

- **To Audit:** Review the ledger transaction graph output by `daml test` to confirm that signatories and observers match the tripartite model.
- **To Extend:** Add new test templates in `contracts/stablecoin-escrow-tests/` to mock Canton ledger failures, and update the matching choices in the main templates.
