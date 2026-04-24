# Lifecycle State Machines

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 2.1, 9.1, 9.2, 11](../design.md)

---

## Purpose

This document defines the **formal state machines** for every stateful entity in the GAIA kernel. Each state machine specifies the valid states, transitions, triggers, and guards to prevent race conditions and illegal state changes.

---

## State Machines to Define

### 1. Task Lifecycle

States: `pending → planning → executing → completed | failed | cancelled`

* What triggers each transition?
* Can a task go from `executing` back to `planning` (replan)?
* What happens if a task is cancelled while in `planning`?
* State diagram (Mermaid)

---

### 2. Step Lifecycle

States: `pending → running → pending_async → done | failed`

* When does a step move to `pending_async`?
* Can a failed step return to `pending` (retry)?
* What happens to a `running` step when the task is cancelled?
* State diagram (Mermaid)

---

### 3. Agent Lifecycle

States: `connecting → active → degraded → quarantined → blacklisted → disconnected`

* What triggers degradation? (Section 10.2)
* Can a quarantined agent be restored to active?
* What is the disconnect flow? (Section 4.3: DRAIN → REASSIGN → DEREGISTER → CLOSED)
* State diagram (Mermaid)

---

### 4. Plan Lifecycle

States: `generating → valid → executing → completed | failed | replanning`

* When does replanning occur?
* How does the plan interact with the Task lifecycle?

---

## Format

Each state machine will include:
1. **State diagram** (Mermaid syntax)
2. **Transition table** (from → to, trigger, guard conditions)
3. **Edge cases** (concurrent transitions, race conditions)
4. **Invariants** (properties that must always hold)

---

## Related Documents

* [Data Model & Schemas](schemas.md) — schema definitions for Task, Step, AgentRecord
* [Control Loop Spec](control-loop.md) — how the loop drives state transitions
* [Failure Handling Spec](failure-handling.md) — agent degradation and quarantine triggers
* [Registry Spec](registry.md) — registration and disconnect flows

---

## TODO

- [ ] Define Task state machine with Mermaid diagram
- [ ] Define Step state machine with Mermaid diagram
- [ ] Define Agent state machine with Mermaid diagram
- [ ] Document all edge cases and race conditions
- [ ] Cross-reference with Control Loop (design.md Section 16)
