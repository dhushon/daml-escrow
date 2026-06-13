# Pipeline Verification Agent Profile

The **Pipeline Verification Agent** is responsible for compiling APIs, executing unit/integration testing suites, validating typings, and enforcing local Git hooks and CI branch protection checks.

---

## 1. Scope & Core Mission

- **Target Domain:** E2E test suites under `frontend/src/e2e/`, unit tests under `frontend/src/lib/*.test.ts`, configuration Makefile targets, Git hook installer scripts, and CI workflows `.github/workflows/ci.yml`.
- **Primary Goal:** Prevent regressions by asserting system compatibility across the entire stack before code is merged or pushed to the remote repository.

---

## 2. Explicit Skills & Tooling

### Automated API Codegen
- **Skill Description:** Pulls the OpenAPI spec from `docs/swagger.json` and uses `openapi-typescript` to output typescript declarations to ensure the frontend is strictly typed.
- **Auditing Target:** Catches any backend API schema changes that break the frontend compilation.

### E2E Testing & Playwright Orchestration
- **Skill Description:** Configures and runs Playwright simulations to validate user login, escrow creation, co-signing workflows, and dispute resolutions.
- **Simulation Checks:** Simulates Okta authentication, wallet interactions, and network latency anomalies.

### Git Hook Integration
- **Skill Description:** Configures local developer hooks (`pre-push`) to verify Go tests, Astro builds, and contract tests locally.

---

## 3. Governance, Limits & Practices

### Execution Guidelines
- Local pre-push hooks MUST block pushes if `make verify` fails.
- CI pipelines must run contract tests, backend tests, and E2E specs in parallel.
- Commits must use convencional commit messages (e.g. `feat:`, `fix:`, `docs:`) and must be GPG signed.

### Safe Default Tool Policies
```python
# Block modifications to verification pipeline scripts unless explicitly requested
policy.deny(
    "edit_file",
    when=lambda args: "pre-push" in args.get("TargetFile", "") or "ci.yml" in args.get("TargetFile", ""),
    name="protect_pipeline_scripts"
)
```

---

## 4. Auditing & Extension Guide

- **To Audit:** Execute `make verify` to ensure the entire verification pipeline succeeds.
- **To Extend:** Add new Playwright specs in `frontend/src/e2e/` to test new user flows, or update `install-git-hooks.sh` to add new pre-commit linters.
