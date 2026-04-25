# Planning & Interpolation Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Sections 6, 12, 16](../design.md)

---

## Purpose

This document specifies the **Planner Interface** and the **Interpolation Engine** — the kernel components responsible for decomposing high-level goals into executable units of work and managing the data flow between them. 

The GAIA kernel treats the Planner as a pluggable, isolated module that operates on capability abstractions rather than direct agent invocations.

---

## 1. Planner Interface Contract

The Planner is invoked by the Control Loop during Phase 2 (Planning) or Phase 10.2 (Incremental Planning).

### 1.1 Input Assembly (The "Context Window")

Every planning call MUST receive the following structured context (design.md Section 6.1):

| Component | Source | Description |
| :--- | :--- | :--- |
| **Goal** | `task.goal` | The immutable user objective. |
| **Active State** | State Store (Tier 1) | The current accumulated results and context. |
| **Capability Manifest** | Registry | A strictly filtered list of names, descriptions, and input/output schemas for currently `active` capabilities. |
| **Failure Context** | Control Loop | (Optional) The error and step trace that triggered a replan. |

### 1.2 Output Specification (PlanRecord)

The Planner must return a `PlanRecord` (schemas.md Section 11) conforming to:
* **Steps**: A list of 1–3 `Step` objects.
* **has_more**: Boolean indicating if the goal requires further decomposition after these steps.
* **Metadata**: Generation counter for tracking replan iterations.

---

## 2. Incremental Planning (design.md Section 6.2)

GAIA enforces **Incremental Planning** to ensure bounded context and system adaptability.

### 2.1 The "1-3 Step" Rule
The Planner is encouraged to generate small plan segments. This prevents:
1. **Context Drift**: Long plans becoming irrelevant as world state changes.
2. **Hallucination**: LLMs losing coherence in deep dependency chains.
3. **Latency**: Faster response times for the first executable steps.

### 2.2 Re-invocation Triggers
The kernel re-invokes the planner when:
1. All steps in the current plan are `done` AND `has_more == true`.
2. A step failure triggers the `replan` tier of the escalation path (failure-handling.md Section 4).

---

## 3. Interpolation Engine (design.md Section 6.4)

The Interpolation Engine is responsible for binding data between steps using the `{{...}}` syntax.

### 3.1 Syntax and Sources
Interpolation is resolved by the Kernel **before** dispatching a step to an agent.

| Pattern | Source | Priority |
| :--- | :--- | :--- |
| `{{step_id.output.field}}` | Output of a previously completed step | 1 |
| `{{state.field}}` | Current value in the Kernel's Active State | 2 |
| `{{const.field}}` | Task metadata/constants | 3 |

### 3.2 Resolution Algorithm
1. **Scan**: Identify all `{{...}}` markers in the `step.input` object.
2. **Lookup**: Resolve the reference against the source priority list.
3. **Cast**: Ensure the resolved value matches the type expected by the capability's `input_schema`.
4. **Replace**: Swap the marker for the concrete value.

### 3.3 Validation Rules
**At Planning Time (Phase 2):**
* Planner strictly checks that `{{step_id...}}` references point to valid, existing steps in the plan.
* If unresolvable, the Planner rejects the plan generation (`PLAN_REJECTED`).

**At Execution Time (Phase 4):**
* Interpolation resolves actual runtime values.
* If a reference is unresolvable or missing from state, the step fails immediately with `EXECUTION_FAILED` (triggers step escalation).

---

## 4. Circular Dependency Detection

The kernel validates every plan to ensure it is a Directed Acyclic Graph (DAG).

1. **Extract Interpolation Links**: Scan all `{{step_id.output}}` markers inside each `step.input` and synthesize them into implicit dependencies.
2. **Build Adjacency Graph**: Combine explicit `depends_on` lists with the implicit interpolation links.
3. **Cycle Check**: Perform a Depth-First Search (DFS) for back-edges on the combined graph.
4. **Rejection**: If any cycle is detected (whether explicit via `depends_on` or implicit via interpolation loops), emit `PLAN_REJECTED` and trigger Planner Failure Recovery.

---

## 5. Planner Failure Recovery (design.md Section 12)

The kernel implements a specialized circuit breaker for the Planner to handle non-deterministic LLM behavior.

| Failure Mode | Kernel Behavior | Recovery Strategy |
| :--- | :--- | :--- |
| **Timeout/Rate Limit** | Soft Failure | Retry with exponential backoff (max 3). |
| **Malformed Output** | Hard Failure | Retry once with "Correction Prompt" (force JSON). |
| **Hallucination** | Contextual Failure | Identify unknown capability; retry with manifest filtered to ONLY valid tools. |
| **Empty Plan** | Logical Failure | Terminate task. Emit `TASK_FAILED`. |

---

## 6. Replanning Logic

Replanning is an escalation tier used when execution hits a roadblock.

### 6.1 Failure Context Injection
When replanning, the Planner receives the "Failure Context":
* The step that failed.
* The error code (e.g., `CAPABILITY_NOT_FOUND`).
* The partial results of other successful steps.

### 6.2 Replanning Circuit Breaker
To prevent infinite "I will try again" loops:
* **Max Replans**: 2 per task.
* **Pruning**: The Planner is instructed that the failed capability is unavailable for this task iteration.

---

## Related Documents

* [Data Model & Schemas](schemas.md) — PlanRecord and Step schema definitions.
* [Control Loop Spec](control-loop.md) — The exact phases where planning and interpolation occur.
* [Failure Handling Spec](failure-handling.md) — Detailed escalation path (retry → fallback → replan).
* [Registry Spec](registry.md) — How the Capability Manifest is curated and filtered.

---

## TODO

- [x] Define planner input/output schemas formally.
- [x] Specify interpolation engine algorithm and priority.
- [x] Document circular dependency and interpolation validation.
- [x] Define replan loop limit and circuit breaker logic.
