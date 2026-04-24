# State Management Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 2.5: State Management](../design.md)

---

## Purpose

This document specifies the **Tiered State Model** — how the kernel stores, snapshots, prunes, and recovers task state to keep the planner's context window bounded while ensuring no data is lost.

---

## Sections to Define

### 1. Tiered State Architecture

* Tier 1 — Active State (Hot): always in memory, used by planner + execution
* Tier 2 — Task History (Warm): append-only log in DB, not passed to planner
* Tier 3 — Archived State (Cold): compressed, summarized, long-term storage

---

### 2. Active State Schema

* What fields are stored in active state?
* Size limits
* Update rules (who can write, when)

---

### 3. Snapshotting Strategy

* Trigger conditions (step count threshold, state size limit)
* Snapshot format: `summary`, `key_state`, `checkpoint_step`
* Post-snapshot pruning rules
* Snapshot → recent delta retention

---

### 4. Planner Input Assembly

* Active state + latest snapshot + current goal
* Never full history
* How is the snapshot summarized for the planner?

---

### 5. State Recovery

* How is state restored after a kernel restart?
* Checkpoint-based recovery
* Idempotency guarantees

---

### 6. State Storage Backend

* In-memory for active state
* Database requirements for warm state
* Archival storage options for cold state
* Pluggable storage interface

---

### 7. Concurrency & Locking

* How is state updated during parallel step execution?
* Optimistic vs. pessimistic locking
* State merge conflicts

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Snapshot and Task schemas
* [Planning Spec](planning.md) — planner input assembly from state
* [Control Loop Spec](control-loop.md) — when state updates occur

---

## TODO

- [ ] Define active state schema formally
- [ ] Specify snapshot trigger thresholds
- [ ] Document state recovery procedure
- [ ] Define storage backend interface
- [ ] Address concurrency during parallel execution
