# Gemini AI Agent Registry & Governance Protocol

This directory contains the specifications, explicit skills, boundaries, and auditing practices for the AI coding agents contributing to the Stablecoin Escrow platform. 

The goal of this protocol is to establish a secure, auditable, and extensible multi-agent framework utilizing the **Google Antigravity (AGY) SDK** and standard workspace capabilities.

---

## Agent Directory Index

We define five discrete agent profiles to govern work in this repository:

1. [Smart Contract Agent](file:///Users/dhushon/work/daml-escrow/.gemini/agents/smart_contract_agent.md)
   - **Focus:** DAML on-chain ledger modeling, transaction mechanics, and Canton node integration.
2. [Go Backend Service Agent](file:///Users/dhushon/work/daml-escrow/.gemini/agents/go_backend_agent.md)
   - **Focus:** Standard Go HTTP handlers, specific validation DTOs, PostgreSQL dynamic configurations, and Okta OIDC validation.
3. [Astro Frontend UX Agent](file:///Users/dhushon/work/daml-escrow/.gemini/agents/frontend_ux_agent.md)
   - **Focus:** HTML5 semantic layouts, premium HSL glassmorphism, zero-JS Astro components, and native SVG metrics rendering.
4. [Pipeline Verification Agent](file:///Users/dhushon/work/daml-escrow/.gemini/agents/pipeline_verification_agent.md)
   - **Focus:** Compile-time TS validation, Vitest mocking, Playwright E2E simulation suites, and Git hooks enforcement.
5. [Antigravity SDK Agent Template](file:///Users/dhushon/work/daml-escrow/.gemini/agents/antigravity_sdk_agent.md)
   - **Focus:** Programmatic creation of custom Python agents using the native `google-antigravity` package.

---

## Explicit Skills & Tool Matrix

Each agent is equipped with explicit toolkits and system-provided plugins to perform its role under strict authorization:

| Tool / Skill | Scope | Governed By | Auditing Practice |
| :--- | :--- | :--- | :--- |
| **`chrome-devtools`** | Frontend E2E browser inspection and debugging. | devtools plugin | Console log analysis and request tracing checks. |
| **`a11y-debugging`** | Accessibility auditing & WCAG compliance. | devtools plugin | Tap target spacing, color contrast, and ARIA roles verification. |
| **`modern-web-guidance`** | Lookup of modern HTML5 / CSS v4 specifications. | web guidance plugin | Checking for obsolete APIs and layout paradigms. |
| **`run_command`** | System command and test suite execution. | `policy.confirm_run_command()` | Strict Git signature checking and command whitelist lambda. |
| **`edit_file` / `write_to_file`** | File write and code modification. | `policy.workspace_only()` | Path validation matching workspace boundaries. |

---

## Governance Rules & Limits

To align with the project's security guardrails, all agent instances MUST adhere to the following constraints:

### 1. Fail-Safe Defaults
All tool configuration policies default to **"Deny All"**. Agents are explicitly denied shell command execution (`run_command`) in non-interactive mode. For interactive execution, developers must prompt the user via `ask_user` hook confirmations.

### 2. Workspace Containment
Any filesystem operations (`view_file`, `edit_file`, `write_to_file`) are strictly confined to the workspace root:
`/Users/dhushon/work/daml-escrow`

### 3. PII & Compliance Safeguards
Agents are forbidden from creating API query parameters containing plain-text email addresses, user IDs, or credentials. Identity discovery and saving configuration details must be routed through `POST` request bodies or secure JWT context attributes to prevent log leakage.

---

## Auditing and Extending Agent Skills

When auditing or adding new skills:
1. **Define the Skill:** Place a new `SKILL.md` instruction file under a dedicated directory.
2. **Set Predicates:** Add Lambda predicates to validate tool inputs (e.g. denying shell execution containing `rm` or variables injection).
3. **Register:** Add the new agent or skill details to this registry.
4. **Compile check:** Run the verification pipeline to ensure no compile errors are introduced.
