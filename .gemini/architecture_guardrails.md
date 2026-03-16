# Architecture Guardrails

These guardrails define the **core architectural principles** for the
escrow platform.

## 1. Contract-First Design

All business logic involving funds MUST live in DAML contracts.

Backend services:

- orchestrate
- validate inputs
- relay transactions

They must NEVER implement financial logic.

## 2. Ledger as Source of Truth

The distributed ledger is the **single source of truth**.

Applications must treat the ledger as authoritative.

No off-ledger balance tracking is allowed.

## 3. Privacy by Default

Escrow contracts must restrict visibility to:

- buyer
- seller
- mediator
- issuer

No escrow information should leak to unrelated parties.

## 4. Stateless Services

All Go services must be stateless.

State must reside in:

- the ledger
- durable event streams

## 5. Event-Driven Architecture

Services should subscribe to ledger events rather than polling.

Preferred technologies:

- Kafka
- NATS
- Cloud PubSub

## 6. Deterministic Escrow Flows

Escrow state transitions must always follow deterministic rules:

Created → Locked → Delivered → Released/Refunded
