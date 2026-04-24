# Control Loop Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 16: Authoritative Control Loop](../design.md)

---

## Purpose

This document provides a **detailed, implementation-ready specification** of the GAIA kernel's control loop — the "heartbeat" of the entire system. It expands the pseudocode from design.md into formal logic with error handling, concurrency rules, and timing constraints.

---

## Kernel Invariants

1. **Progress Guarantee**: The loop must either make progress (complete a step) or transition to a terminal failure/replan state. Infinite loops are forbidden.
2. **State Consistency**: No state update shall occur without a corresponding audit log entry.
3. **Atomic Transitions**: Task and Step status changes must be atomic; partial or "in-between" states shall not be exposed to the Event Bus.
4. **Deny-by-Default**: No step shall be dispatched without passing the Policy Engine validation phase.

---

## Sections to Define

### 1. Loop Entry Conditions

* When does the loop start? (on Task creation)
* What pre-conditions must be met?

---

### 2. Planning Phase

* Planner invocation contract
* Input assembly (goal + active state + capability manifest)
* Output validation (plan schema check)
* Planner failure recovery (Section 12)

---

### 3. Step Scheduling (DAG Resolution)

* How `depends_on` is resolved
* Parallel dispatch strategy
* Maximum concurrent steps limit
* Step readiness algorithm

---

### 4. Input Resolution (Interpolation)

* `{{step_id.output.field}}` resolution
* `{{state.field}}` resolution
* `{{const.field}}` resolution
* Error on unresolvable references

---

### 5. Policy Check

* Pre-dispatch policy validation
* Approval workflows (human-in-the-loop)
* Halt vs. skip behavior

---

### 6. Dispatch & Invocation

* Sync vs. async mode selection
* Transport routing
* Timeout enforcement

---

### 7. Result Processing

* Output schema validation
* State update logic
* Event emission

---

### 8. Failure Handling (in-loop)

* Failure classification
* Retry policy application
* Agent enforcement (degrade/quarantine/blacklist)
* Escalation: retry → fallback → replan → abort

---

### 9. Async Completion Handling

* How pending_async steps are tracked
* Timeout enforcement for async steps
* Event-driven completion

---

### 10. Loop Termination

* Success condition (all steps done, `has_more == false`)
* Failure condition (unrecoverable error)
* Cancellation condition (interrupt received)

---

## Related Documents

* [Planning Spec](planning.md) — planner interface and interpolation engine
* [Lifecycle State Machines](lifecycles.md) — Task and Step state transitions
* [Failure Handling Spec](failure-handling.md) — retry, escalation, enforcement
* [Registry Spec](registry.md) — agent selection and routing
* [Transport Spec](transport.md) — invocation and protocol adapters
* [Security Spec](security.md) — policy checks and sandbox enforcement

---

## TODO

- [ ] Convert pseudocode to formal flowchart (Mermaid)
- [ ] Define concurrency limits and thread safety rules
- [ ] Document timing constraints (max loop iteration time)
- [ ] Add sequence diagrams for sync and async flows
