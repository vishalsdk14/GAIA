# Security & Policy Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Sections 5, 8](../design.md)

---

## Purpose

This document outlines the GAIA Kernel's security architecture. It formalizes the Policy Engine DSL used to define cross-cutting constraints, the exact failure modes for policy violations, and the immutable Audit Logging system required for system traceability.

---

## 1. The Policy Engine

The Policy Engine acts as the definitive gatekeeper in Phase 5 of the Control Loop. Every inbound Request from the Planner and every outbound Request to an Agent must pass policy evaluation.

### 1.1 Policy DSL

GAIA uses **Common Expression Language (CEL)** for policy definitions. CEL provides a fast, safe, and non-Turing complete environment to evaluate boolean rules against the task and step context.

**Example Policies:**

* *Cost Control*: Restrict an agent from exceeding a specific budget.
  ```cel
  task.metrics.cost_estimate + step.metrics.cost_estimate < 100.0
  ```
* *Sandbox Enforcement*: Prevent agents from modifying global state unless explicitly granted the `state:write` scope.
  ```cel
  !("state:write" in agent.auth.scopes) ? !capability.constraints.mutates_state : true
  ```
* *Approval Gates*: Force human approval if an external IO capability is invoked in production.
  ```cel
  (capability.constraints.external_io && env == "production") ? false : true
  ```

Policies are stored dynamically in the Registry and evaluated in microseconds.

---

## 2. Policy Failure Modes

When the Policy Engine evaluates a rule that returns `false` or encounters an invalid state, it triggers one of six specific failure modes (referenced in control-loop.md Phase 5.1).

1. **`POLICY_DENIED`**: 
   * **Trigger**: A CEL rule evaluates to false (e.g., budget exceeded, scope missing).
   * **Action**: Step fails immediately. Escalates to Fallback/Replan.
2. **`SCHEMA_VIOLATION`**:
   * **Trigger**: The step input fails JSON Schema validation against the agent's `input_schema`.
   * **Action**: Step fails. Agent may be quarantined (see failure-handling.md).
3. **`UNAUTHORIZED`**:
   * **Trigger**: Agent credentials (JWT, API key) are invalid or expired.
   * **Action**: Agent connection is rejected. Status → `disconnected`.
4. **`CAPABILITY_FORBIDDEN`**:
   * **Trigger**: The agent attempts to invoke a capability it hasn't registered for.
   * **Action**: Request dropped. Agent trust score is heavily penalized.
5. **`BUDGET_EXHAUSTED`**:
   * **Trigger**: The global task budget is depleted.
   * **Action**: Task halts. Status → `failed`. (Terminal).
6. **`APPROVAL_REQUIRED`**:
   * **Trigger**: A CEL rule determines human oversight is needed.
   * **Action**: The loop yields. Step status remains `pending`. Emits `STEP_APPROVAL_REQUIRED`. Execution pauses until an admin explicitly clears the flag via the API.

---

## 3. Audit Logging

To guarantee traceability for all agent actions (design.md Section 5.3), the kernel implements a strict, tamper-proof Audit Log.

### 3.1 Audit Events
*Every* state transition (Task, Step, Agent, Plan) is written to the Audit Log. This log is a superset of the Event Bus.

### 3.2 Immutability and Retention
* **Format**: W3C ActivityStreams 2.0 JSON or standard NDJSON.
* **Storage**: Appended to a write-only datastore (e.g., AWS S3 with Object Lock, or a WORM drive).
* **Integrity**: Each log entry contains a cryptographic hash of the previous entry, establishing an unbreakable chain of custody.
* **Retention**: Configurable, but defaults to 365 days for compliance tracking.

### 3.3 Log Entry Schema (Abstract)
```json
{
  "log_id": "uuid",
  "timestamp": "iso8601",
  "actor": "kernel | agent_id | admin_id",
  "action": "STEP_STARTED | POLICY_DENIED | ...",
  "resource": "step_id | task_id",
  "context": { "cel_rule_failed": "..." },
  "hash": "sha256",
  "prev_hash": "sha256"
}
```

---

## Related Documents

* [Control Loop Spec](control-loop.md) — Phase 5 Policy checks.
* [Failure Handling Spec](failure-handling.md) — Escalation paths for `POLICY_DENIED`.
* [Schemas Spec](schemas.md) — Capability constraints evaluated by CEL.
