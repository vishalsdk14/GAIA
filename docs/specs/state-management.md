# State Management Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Section 2.5](../design.md)

---

## Purpose

This document specifies the GAIA kernel's Tiered State Architecture. It defines how step outputs are accumulated during task execution, how concurrent writes are managed to prevent race conditions, and the exact rules for snapshotting and pruning the in-memory state to prevent memory bloat.

---

## 1. The Three Tiers of State

The kernel manages state across three distinct tiers to balance fast resolution against memory constraints.

### 1.1 Tier 1: Active State (Hot)
* **Medium**: In-memory.
* **Format**: Fast JSON access. Event-sourced delta log (to prevent concurrent write races).
* **Usage**: Interpolation resolution by the planner and execution engine.
* **Scope**: Only the current task iteration.

### 1.2 Tier 2: Task History (Warm)
* **Medium**: Persistent Document Store / DB.
* **Format**: Full `Task`, `PlanRecord`, and `Step` schemas.
* **Usage**: Resuming interrupted tasks, debugging, and audit logging.
* **Scope**: Full task history (up to TTL).

### 1.3 Tier 3: Archived State (Cold)
* **Medium**: Persistent Storage + Vector DB.
* **Format**: LLM-summarized context.
* **Usage**: Long-term agent memory across tasks.
* **Scope**: Cross-task history.

---

## 2. Active State Schema (Tier 1)

The Active State is the context injected into the Planner and used for step interpolation.

### 2.1 Schema Definition

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/active-state.json",
  "title": "ActiveState",
  "type": "object",
  "properties": {
    "task_id": { "type": "string", "format": "uuid" },
    "accumulated_outputs": {
      "type": "object",
      "description": "A collapsed view of all step outputs. Keyed by step_id.",
      "additionalProperties": true
    },
    "delta_log": {
      "type": "array",
      "description": "Append-only log of step completions to prevent concurrent write races.",
      "items": {
        "type": "object",
        "properties": {
          "step_id": { "type": "string" },
          "output": { "description": "Any JSON" },
          "timestamp": { "type": "string", "format": "date-time" }
        }
      }
    },
    "metadata": {
      "type": "object",
      "properties": {
        "state_size_bytes": { "type": "integer" },
        "step_count": { "type": "integer" },
        "last_snapshot_generation": { "type": "integer" }
      }
    }
  },
  "required": ["task_id", "accumulated_outputs", "delta_log", "metadata"]
}
```

---

## 3. Concurrency & Locking (Event Sourcing)

Tasks often execute steps in parallel (e.g., `max_concurrent_steps = 10`). Directly mutating `accumulated_outputs` leads to race conditions.

### 3.1 Write Path (Agent Completion)
1. When an agent returns an output, the kernel strictly acquires a lock on the Active State.
2. The output is **appended** to the `delta_log`.
3. The lock is released. This operation is $O(1)$.

### 3.2 Read Path (Interpolation / Snapshot)
1. Before interpolating variables or taking a snapshot, the kernel "collapses" the `delta_log`.
2. For each entry in `delta_log`, its `output` is merged into `accumulated_outputs` under the key `step_id`.
3. The `delta_log` is then cleared.

---

## 4. Snapshotting & Pruning

To prevent the `ActiveState` from growing indefinitely during massive tasks, the kernel enforces strict limits.

### 4.1 Triggering a Snapshot
A Tier 1 → Tier 2 snapshot is triggered if either of these conditions is met:
1. `metadata.state_size_bytes` > **500 KB** (configurable default).
2. `metadata.step_count` since last snapshot > **50** steps.

### 4.2 The Snapshot Process
1. Collapse the `delta_log` into `accumulated_outputs`.
2. Write the full `ActiveState` payload to Tier 2 (Task History database).
3. Increment `last_snapshot_generation`.

### 4.3 Pruning (Tier 1 Eviction)
After a snapshot is confirmed in Tier 2:
1. The kernel prunes `accumulated_outputs` to retain only the outputs of steps that are explicitly marked as "contextual" by the Planner, or outputs from the most recent 10 steps.
2. `metadata.state_size_bytes` is recalculated.

If a subsequent step requests an interpolated variable that has been pruned from Tier 1, the kernel performs a lazy fetch from Tier 2.

---

## 5. Recovery Protocol

If the kernel crashes mid-task, the recovery sequence is:
1. Load the latest Snapshot from Tier 2.
2. Replay any pending events from the Event Bus that occurred after the snapshot timestamp.
3. Re-populate the Tier 1 `ActiveState`.
4. Resume the task loop.

---

## Related Documents

* [Planning Spec](planning.md) — Active state injection to the Planner.
* [Control Loop Spec](control-loop.md) — Snapshot trigger integration (Phase 7).
* [Data Model & Schemas](schemas.md) — Base Task and Step definitions.
