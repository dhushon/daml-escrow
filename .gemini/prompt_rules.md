# prompt_rules.md

Guidelines for writing prompts that give **Gemini coding agents high
autonomy without losing safety**.

------------------------------------------------------------------------

## 1. Provide Architecture Context

Agents perform best when they know:

- system architecture
- repo structure
- constraints

Example:

"Implement a new oracle ingestion service following the
service_templates.md pattern."

------------------------------------------------------------------------

## 2. Always Specify Boundaries

Example:

Allowed: - Go service changes - API layer modifications

Not allowed: - escrow contract modifications

------------------------------------------------------------------------

## 3. Require Deterministic Outputs

Example prompt:

"Generate a deterministic escrow milestone service using Go following
repo guardrails."

------------------------------------------------------------------------

## 4. Encourage Small Changes

Large prompts cause unstable outputs.

Prefer:

"Add milestone approval endpoint"

instead of

"Implement the entire payment system."

------------------------------------------------------------------------

## 5. Require Tests

Always request:

- unit tests
- edge cases
- failure conditions

------------------------------------------------------------------------

## 6. Require Explanation

Example:
"Explain design decisions and reference guardrails used."
