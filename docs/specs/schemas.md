# Data Model & Schema Definitions

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 3.2, 4.1, 8.1, 9](../design.md)

---

## Purpose

This document defines the **canonical JSON Schemas** for every data object in the GAIA kernel. These schemas are the wire protocol — the single source of truth that ensures all components (kernel, adapters, agents) speak the same language.

---

## Schemas to Define

### 1. AgentManifest

The "Digital ID" submitted by every agent during the Handshake.

* **Design reference**: Section 4.1
* Fields: `agent_id`, `version`, `base_url`, `transport`, `protocol`, `invoke`, `capabilities[]`, `health_endpoint`, `auth`
* Formal JSON Schema with required fields, types, enums, and constraints.

---

### 2. Capability

A single capability within an Agent Manifest.

* **Design reference**: Section 4.1 (nested within manifest)
* Fields: `name`, `description`, `input_schema`, `output_schema`, `constraints[]`, `idempotent`

---

### 3. Task

The global state object for a goal.

* **Design reference**: Section 9.1
* Fields: `task_id`, `goal`, `status`, `plan[]`, `current_step`, `created_at`, `updated_at`
* Status enum: `pending | planning | executing | completed | failed | cancelled`

---

### 4. Step

An individual unit of work within a plan.

* **Design reference**: Section 9.2
* Fields: `step_id`, `capability`, `input`, `depends_on[]`, `status`, `output`, `error`, `assigned_agent`, `retry_count`
* Status enum: `pending | running | pending_async | done | failed`

---

### 5. Request

The message sent from the kernel to an agent.

* **Design reference**: Section 3.2
* Fields: `type`, `from`, `capability`, `input`, `task_id`, `step_id`, `mode`

---

### 6. Response

The standardized output returned by any agent.

* **Design reference**: Section 3.2
* Fields: `success`, `output`, `error`, `metrics`

---

### 7. Error

The structured failure object.

* **Design reference**: Section 3.2
* Fields: `code`, `message`, `retryable`
* Code enum: `SCHEMA_VIOLATION | TIMEOUT | POLICY_DENIED | INTERNAL | UNKNOWN`
* See also: [Error Code Catalog](../reference/error-codes.md)

---

### 8. Event

Asynchronous event emitted by the kernel.

* **Design reference**: Section 3.2, 3.6
* Fields: `type`, `name`, `payload`, `task_id`, `step_id`
* See also: [Event Catalog](../reference/event-catalog.md)

---

### 9. AgentRecord

The kernel's internal record for a registered agent.

* **Design reference**: Section 9.3
* Fields: `agent_id`, `status`, `trust_score`, `registered_at`, `last_health_check`, `rolling_metrics`
* See also: [Lifecycle State Machines](lifecycles.md) for agent status transitions

---

### 10. RetryPolicy

Per-step retry configuration.

* **Design reference**: Section 8.1
* Fields: `max_attempts`, `backoff`, `base_delay_ms`, `max_delay_ms`
* See also: [Failure Handling Spec](failure-handling.md)

---

### 11. Snapshot

State checkpoint for tiered state management.

* **Design reference**: Section 2.5
* Fields: `summary`, `key_state`, `checkpoint_step`
* See also: [State Management Spec](state-management.md)

---

## Format

Each schema will be defined as:
1. **Illustrative JSON example** (for readability)
2. **Formal JSON Schema** (for validation)
3. **Field-by-field table** (type, required, constraints, description)

---

## Related Documents

* [Lifecycle State Machines](lifecycles.md) — valid status transitions
* [Communication Spec](communication.md) — message flow using these schemas
* [Error Code Catalog](../reference/error-codes.md) — all error codes
* [Event Catalog](../reference/event-catalog.md) — all event types

---

## TODO

- [ ] Define all schemas with formal JSON Schema syntax
- [ ] Add validation examples (valid + invalid payloads)
- [ ] Cross-reference with lifecycle state machines
