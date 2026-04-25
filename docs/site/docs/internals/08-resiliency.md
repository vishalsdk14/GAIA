# Chapter 8: Resiliency & Escalation

A primary mission of the GAIA Kernel is to ensure that a task reaches its goal even when agents are unreliable. This chapter covers the Kernel's "Immune System"—the mechanisms of retry, fallback, and escalation.

---

## 8.1 The Failure Taxonomy

Not all failures are equal. GAIA classifies failures into three categories to determine the correct response:

| Type | Examples | Recovery Strategy |
| :--- | :--- | :--- |
| **Soft Failure** | Timeout, Rate limit, 503 error | **Retry** with exponential backoff. |
| **Hard Failure** | Schema violation, 403 Forbidden | **Fallback** to a different agent. |
| **Policy Violation** | Unauthorized IO, Budget breach | **Abort** and notify the administrator. |

---

## 8.2 The 4-Tier Escalation Path

When a step fails, the Kernel follows a strict escalation ladder (see [failure-handling.md](../../specs/failure-handling.md)):

### Tier 1: Retry (Same Agent)
If the error is retryable and the capability is **Idempotent**, the Kernel re-dispatches the step to the same agent with an exponential backoff.
*   *Default*: Max 3 attempts.

### Tier 2: Fallback (Different Agent)
If retries fail, the Kernel queries the Registry for a secondary agent that provides the same capability. This handles scenarios where one specific agent instance is down or degraded.

### Tier 3: Replan (New Strategy)
If no other agent can fulfill the capability, the Kernel invokes the **Planner**. It provides the "Failure Context" (e.g., *"The weather agent is unavailable"*) and asks: *"Can we achieve the goal another way?"*
*   *Example*: Use a "Web Search" capability to find the weather instead.

### Tier 4: Abort (Safe Shutdown)
If all strategies fail, the task is marked as `failed`. The Kernel ensures a graceful shutdown, releasing all resources and logging the full escalation trace for developer analysis.

---

## 8.3 Idempotency & Safety

The Kernel is "Safety-First" when it comes to retries.
*   **Idempotent** capabilities (e.g., `read_file`) are automatically retried.
*   **Mutating** capabilities (e.g., `send_payment`) are **never** retried unless the agent explicitly flags the error as `retryable: true` and provides a transaction token.

---

## 8.4 Circuit Breakers

To protect the ecosystem from "Cascading Failures," the Kernel implements circuit breakers at two levels:

1.  **Agent Level**: If an agent fails N times in a row, it is moved to `degraded` or `quarantined` status in the Registry.
2.  **Capability Level**: If all agents for a specific capability are failing, the Kernel temporarily stops routing requests to that capability to prevent Planner busy-loops.

---

## 8.5 Deterministic Recovery

Because every step result and state snapshot is durable (Tier 2 State), a GAIA task can be resumed even after a total Kernel crash. Upon restart, the Kernel:
1.  Scans the database for `executing` tasks.
2.  Reloads the latest state snapshot.
3.  Identifies the last successful step and resumes the control loop from the next pending step.

---

## 8.6 Related Specifications

*   [Failure Handling Spec](../../specs/failure-handling.md)
*   [Lifecycles Spec (Step Lifecycle)](../../specs/lifecycles.md)
*   [Control Loop Spec (Phase 8)](../../specs/control-loop.md)
