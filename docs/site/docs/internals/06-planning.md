# Chapter 6: Planning & Interpolation

This chapter explores how GAIA transforms a high-level Goal into a sequence of executable actions and how data flows between those actions with surgical precision.

---

## 6.1 The Planner Interface

The **Planner** is an isolated component that acts as the "Architect" of the system. While the Kernel provides the infrastructure, the Planner provides the strategy.

### Bounded Context
To prevent hallucinations and context drift, GAIA never sends the full task history to the Planner. Instead, it sends:
1.  The original **Goal**.
2.  The current **Active State** (Tier 1).
3.  A **Capability Manifest** (a filtered list of names and descriptions of attached tools).

### 1-3 Step Incrementalism
GAIA encourages **Incremental Planning**. The Planner is instructed to generate only the next 1–3 steps. Once those steps are executed, the Kernel returns to the Planner with the results to decide the next move. This makes the system extremely resilient to unexpected agent outputs.

---

## 6.2 The Scheduler & DAG Resolution

Once a plan is generated, it is passed to the **Scheduler**. The Scheduler's job is to determine the order of execution.

### Dependency Graph (DAG)
Each step can declare a list of dependencies (`depends_on`). The Scheduler builds a Directed Acyclic Graph:
*   Steps with no dependencies run in **parallel** (up to the Kernel's concurrency limit).
*   Steps that depend on others wait until their parents reach the `done` status.

The Scheduler ensures that the Kernel never attempts to execute a step before its required data is ready.

---

## 6.3 Interpolation: The Data Binding Engine

**Interpolation** is the mechanism that moves data between steps. GAIA uses a simple but powerful `{{...}}` syntax.

### The Flow:
1.  **Plan Generation**: The Planner includes markers in the step input:
    `"city": "{{step_1.output.location}}"`
2.  **Wait**: Step 2 sits in the queue until Step 1 completes.
3.  **Resolve**: In **Phase 4** of the control loop, the Kernel's Interpolation Engine scans the input and swaps the marker for the actual value from Step 1's output.
4.  **Dispatch**: The agent receives a clean, concrete value: `"city": "London"`.

### Priority Order:
1.  `{{step_id.output}}` (Previous step results)
2.  `{{state.field}}` (Managed active state)
3.  `{{const.field}}` (Task constants)

---

## 6.4 Circular Dependency Detection

Before a plan is accepted, the Kernel performs a **Cycle Check**. It scans for back-edges in the dependency graph (both explicit `depends_on` and implicit interpolation links).

If a plan looks like this:
*   Step A depends on Step B
*   Step B depends on Step A
The Kernel **rejects** the plan (`PLAN_REJECTED`) and triggers the Planner Failure Recovery strategy (Phase 2.3).

---

## 6.5 Zero-Allocation Performance

Interpolation happens on the Kernel's hot path. To maintain performance, GAIA avoids traditional JSON unmarshalling for this phase. Instead, it uses a **byte-level traversal** engine that replaces markers directly in the JSON buffer without generating garbage collection overhead.

---

## 6.6 Related Specifications

*   [Planning Spec](../../specs/planning.md)
*   [Control Loop Spec (Phase 2, 3, 4)](../../specs/control-loop.md)
*   [State Management Spec](../../specs/state-management.md)
