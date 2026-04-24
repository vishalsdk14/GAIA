# Communication & Event System Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 3: Communication Model](../design.md)

---

## Purpose

This document specifies the **message formats, routing rules, event catalog, and subscription model** for all communication within the GAIA kernel. All inter-agent communication is mediated — this spec defines exactly how.

---

## Sections to Define

### 1. Message Types

* Request schema (with all fields)
* Response schema (with error object)
* Event schema
* Interrupt/Cancel messages

---

### 2. Communication Flow

* Request routing pipeline (8-step validation from Section 5.2)
* Event routing pipeline
* Sequence diagrams for both flows

---

### 3. Event Catalog (Complete)

* All system events with payload schemas
* STEP_COMPLETED, STEP_FAILED, TASK_COMPLETED, TASK_FAILED, TASK_CANCELLED
* AGENT_REGISTERED, AGENT_EJECTED, PLAN_GENERATED
* Custom event extensibility rules

---

### 4. Subscription Model

* Who can subscribe to what?
* Client subscriptions (WebSocket/SSE)
* Audit Log (mandatory subscriber)
* Agent restrictions (cannot subscribe)

---

### 5. Event Bus Architecture

* Internal implementation model
* Delivery guarantees (at-least-once? exactly-once?)
* Ordering guarantees
* Back-pressure handling

---

### 6. Traceability

* Required fields on all messages (`task_id`, `step_id`)
* Correlation ID propagation
* Audit log format

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Request, Response, Event, Error schemas
* [Event Catalog](../reference/event-catalog.md) — complete event type reference
* [Error Code Catalog](../reference/error-codes.md) — all error codes
* [Security Spec](security.md) — policy enforcement in communication
* [Control Loop Spec](control-loop.md) — how the loop emits events

---

## TODO

- [ ] Define complete message schemas with examples
- [ ] Create sequence diagrams for all communication flows
- [ ] Specify delivery guarantees for the Event Bus
- [ ] Document audit log format and retention policy
