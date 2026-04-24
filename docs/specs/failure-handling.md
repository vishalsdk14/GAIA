# Failure Handling & Retry Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Sections 8, 10, 12](../design.md)

---

## Purpose

This document specifies the **complete failure handling strategy** for the GAIA kernel — from individual step retries to agent ejection, and up to planner failure recovery. It defines the formal escalation path, backoff algorithms, agent enforcement rules, and circuit breaker logic that ensures the kernel remains stable under degraded conditions.

---

## 1. Failure Taxonomy

All errors returned by agents or internal systems must map to a specific error code (schemas.md Section 6). These codes are grouped into three failure categories (design.md Section 10.1):

### 1.1 Soft Failure (Retryable)
Transient errors that may resolve if the operation is attempted again.
* **Codes**: `TIMEOUT`, `AGENT_UNAVAILABLE`, `INTERNAL_ERROR`, `UNKNOWN`
* **Agent Impact**: Degrades trust score and priority over time.
* **Action**: Retry (if policy allows).

### 1.2 Hard Failure (Non-Retryable)
Deterministic errors where the input/output structurally violates contracts, or the agent logic is definitively broken for the given input.
* **Codes**: `SCHEMA_VIOLATION`, `EXECUTION_FAILED`, `CAPABILITY_NOT_FOUND`
* **Agent Impact**: Immediate quarantine.
* **Action**: Fallback or replan immediately.

### 1.3 Policy Violation (Terminal)
Malicious or unauthorized behavior detected by the Policy Engine or Sandbox.
* **Codes**: `POLICY_DENIED`
* **Agent Impact**: Immediate blacklist.
* **Action**: Fallback or replan immediately.

---

## 2. Retry Policy

Every step execution is governed by a `RetryPolicy` (schemas.md Section 9). If no policy is provided in the plan, the kernel applies the following defaults:

### 2.1 Default Retry Rules

| Property | Default Value | Description |
| :--- | :--- | :--- |
| `max_attempts` | 3 | Maximum number of retry attempts |
| `backoff` | `exponential` | Backoff algorithm |
| `base_delay_ms` | 500 | Initial delay before first retry |
| `max_delay_ms` | 10000 | Maximum delay bound |

### 2.2 Idempotency Constraints (design.md Section 8.2)

Retries are strictly governed by the capability's manifest definition:
1. **Safe to retry**: Capabilities marked `idempotent: true` OR where `constraints.mutates_state: false`.
2. **Blocked**: Capabilities marked `idempotent: false` AND `constraints.mutates_state: true`.
   * The kernel will **override** any defined RetryPolicy and force `max_attempts = 0` for these steps to prevent dirty writes or duplicate transactions.

### 2.3 Backoff Algorithms

Given $attempt$ (1-indexed), the delay before the next retry is calculated as:

* **none**: $delay = 0$ (immediate retry)
* **linear**: $delay = \min(base\_delay\_ms \times attempt, max\_delay\_ms)$
* **exponential**: $delay = \min(base\_delay\_ms \times 2^{(attempt-1)} + jitter, max\_delay\_ms)$
  * *Jitter*: A random variation of $\pm 20\%$ to prevent thundering herd problems.

---

## 3. Agent Enforcement (design.md Section 10.2)

When an agent returns a failure or fails a health check, the kernel applies immediate enforcement actions affecting the agent's lifecycle (lifecycles.md Section 3).

### 3.1 Trust Score & Degradation

The `trust_score` (0.0 to 1.0) is a rolling composite metric. It acts as the primary weighting factor for the Capability Registry's routing decisions.

$$ Trust Score = (Success Rate \times 0.6) + (Latency Score \times 0.3) + (Availability \times 0.1) $$

* **Latency Score**: Normalized against the agent's declared SLA (`avg_latency_ms`).
* **Degraded State**: If the `trust_score` drops below `0.70` (configurable), the agent transitions to `degraded`. It receives a massive routing penalty but remains eligible if no fallback exists.

### 3.2 Enforcement Matrix

| Trigger | Agent Action | Emitted Event | Resulting State |
| :--- | :--- | :--- | :--- |
| Consecutive timeouts > 3 | Degrade priority | `AGENT_DEGRADED` | `degraded` |
| `SCHEMA_VIOLATION` | Immediate Quarantine | `AGENT_QUARANTINED` | `quarantined` |
| `POLICY_DENIED` | Immediate Blacklist | `AGENT_BLACKLISTED` | `blacklisted` |
| Health check unreachable | Temporary Eject | `AGENT_EJECTED` | `disconnected` |

### 3.3 Restoration Criteria

* **Degraded**: Automatically recovers to `active` if rolling metrics push `trust_score` back above `0.85` over the last 100 requests or 10 health checks.
* **Quarantined**: Requires manual intervention. An admin must review the schema violation and clear the quarantine via the Kernel Admin API.
* **Blacklisted**: Terminal. The agent credentials are wiped, and it cannot re-register without new certificates/tokens.

---

## 4. The Escalation Path (design.md Section 8.3)

When a step fails, the kernel walks a formal escalation ladder. If any tier succeeds, the task continues. If a tier exhausts its limits, the kernel escalates to the next tier.

```mermaid
flowchart TD
    FAIL([Step Fails]) --> IS_RETRYABLE{Retryable error & \n attempts < max?}
    
    IS_RETRYABLE -->|Yes| CHECK_SAFE{Idempotent or \n pure function?}
    CHECK_SAFE -->|Yes| RETRY[Retry (same or different agent)]
    RETRY --> WAIT[Apply Backoff]
    WAIT --> EXEC[Execute Step]
    EXEC -->|Success| DONE([Done])
    EXEC -->|Fail| IS_RETRYABLE

    IS_RETRYABLE -->|No| FALLBACK
    CHECK_SAFE -->|No| FALLBACK

    FALLBACK{Fallback Agent \n available?}
    FALLBACK -->|Yes| ASSIGN[Assign to fallback agent]
    ASSIGN --> EXEC
    
    FALLBACK -->|No| REPLAN{Replan count < \n MAX_REPLANS?}
    
    REPLAN -->|Yes| CALL_PLANNER[Invoke Planner with \n Failure Context]
    CALL_PLANNER --> PLAN_OK{Plan Valid?}
    PLAN_OK -->|Yes| NEW_STEPS[Execute new plan]
    PLAN_OK -->|No| ABORT

    REPLAN -->|No| ABORT
    ABORT([Abort: Task Failed])
```

---

## 5. Planner Failure Handling (design.md Section 12)

The planner itself (the LLM) is an external dependency that can fail. The control loop isolates planner failures from task execution.

| Failure Mode | Recovery Strategy | Kernel Action |
| :--- | :--- | :--- |
| **Timeout / Rate Limit** | Transient | Retry with backoff (max 3 attempts). |
| **Malformed JSON** | Structural | Retry once with a stricter system prompt demanding JSON. |
| **Empty Plan** | Logical | Abort. Emit `TASK_FAILED("planner returned empty plan")`. |
| **Hallucinated Capability** | Contextual | Reject plan. Retry once with a strictly filtered capability manifest. |
| **Retries Exhausted** | Terminal | Abort. Emit `TASK_FAILED("planner unavailable")`. |

---

## 6. Circuit Breakers

To prevent catastrophic loops and resource exhaustion, the kernel enforces the following global circuit breakers per Task:

1. **Maximum Replans**: A task may only trigger the `replan` escalation tier `MAX_REPLANS` times (default: 2). Exceeding this triggers an immediate `TASK_FAILED`.
2. **Maximum Planner Retries**: The planner invocation loop will hard-abort after 3 consecutive failures (timeouts or malformed output) within a single planning phase.
3. **Capability Blackhole**: If *all* registered agents for a specific capability are quarantined or disconnected during a task, any pending step requiring that capability immediately escalates to `replan`.

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Error, RetryPolicy, AgentRecord schemas
* [Lifecycle State Machines](lifecycles.md) — agent status transitions and task replan loops
* [Control Loop Spec](control-loop.md) — exact integration of the escalation path into the loop
* [Registry Spec](registry.md) — agent routing weights based on trust score
* [Error Code Catalog](../reference/error-codes.md) — full list of kernel error codes

---

## TODO

- [x] Define backoff algorithm formally with jitter
- [x] Specify trust score calculation formula
- [x] Document agent restoration criteria for all states
- [x] Define circuit breaker thresholds (replan limits)
- [x] Add formal escalation path flowchart (Mermaid)
