# Planning & Interpolation Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 6, 12](../design.md)

---

## Purpose

This document specifies the **Planner Interface** and the **Interpolation Engine** — how goals become plans, how plans become executable steps, and how data flows between steps via `{{...}}` references.

---

## Sections to Define

### 1. Planner Interface Contract

* Input: goal + active state + capability manifest
* Output: partial plan (1–3 steps) with `has_more` flag
* Output schema validation
* Planner is a pluggable component (LLM, rules-based, hybrid)

---

### 2. Incremental Planning

* Why partial plans? (bounded context, adaptability)
* When does the kernel re-invoke the planner?
* How does `has_more` affect the control loop?

---

### 3. Capability Manifest Curation

* What goes into the manifest sent to the planner?
* How are unhealthy/quarantined agents filtered?
* Format of the capability list (names + descriptions + schemas)

---

### 4. Interpolation Engine

* `{{step_id.output.field}}` — step output references
* `{{state.field}}` — active state references
* `{{const.field}}` — constant references
* Resolution priority order
* Nested field access (dot notation)

---

### 5. Interpolation Validation

* Only `done` steps can be referenced
* Circular dependency detection
* Missing reference → plan rejection
* Type checking (does the resolved value match the input schema?)

---

### 6. Planner Failure Handling

* LLM timeout / rate limit → retry with backoff
* Malformed output → retry once with stricter prompt
* Empty plan → TASK_FAILED
* Hallucinated capability → reject plan, retry with filtered manifest
* All retries exhausted → TASK_FAILED

---

### 7. Planner Replanning

* When does replanning trigger? (step failure after retry exhaustion)
* What context does the planner receive on replan?
* How does the kernel prevent infinite replan loops?

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Step schema with `depends_on` and interpolation
* [Control Loop Spec](control-loop.md) — how planning integrates into the loop
* [State Management Spec](state-management.md) — planner input assembly
* [Registry Spec](registry.md) — capability manifest curation
* [Failure Handling Spec](failure-handling.md) — replan triggers

---

## TODO

- [ ] Define planner input/output schemas formally
- [ ] Specify interpolation engine algorithm
- [ ] Document circular dependency detection
- [ ] Define replan loop limit and circuit breaker
