# Failure Handling & Retry Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 8, 10, 12](../design.md)

---

## Purpose

This document specifies the **complete failure handling strategy** — from individual step retries to agent ejection to planner failure recovery. It defines the escalation path, backoff algorithms, and circuit breaker logic.

---

## Sections to Define

### 1. Failure Taxonomy

* **Soft failure**: timeout, transient network error
* **Hard failure**: schema violation, malformed output
* **Policy violation**: unauthorized action attempt
* **Planner failure**: LLM timeout, hallucinated capability, empty plan

---

### 2. Retry Policy

* Per-step configuration: `max_attempts`, `backoff`, `base_delay_ms`, `max_delay_ms`
* Backoff algorithms: exponential, linear, constant
* Idempotent vs. non-idempotent retry rules
* Default retry policy (when none specified)

---

### 3. Escalation Path

* retry → fallback agent → replan → abort
* When does each escalation trigger?
* How is the planner invoked during replan? (failure context injection)
* Maximum replan count (circuit breaker)

---

### 4. Agent Enforcement

* Enforcement actions per failure type:
  * Repeated timeouts → degrade priority
  * Schema violations → immediate quarantine
  * Policy violation → blacklist
  * Crash/health down → temporary eject
* Trust score calculation
* Restoration rules (can a quarantined agent recover?)

---

### 5. Health Monitoring

* Health endpoint check frequency
* Rolling metrics: success rate, p95 latency, error type counts
* Health-based routing weight adjustment

---

### 6. Circuit Breaker Pattern

* When does the kernel stop retrying a specific agent?
* When does the kernel stop retrying a specific capability?
* When does the kernel stop replanning entirely?

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Error, RetryPolicy, AgentRecord schemas
* [Lifecycle State Machines](lifecycles.md) — agent status transitions on failure
* [Control Loop Spec](control-loop.md) — failure handling within the loop
* [Registry Spec](registry.md) — agent selection after degradation
* [Error Code Catalog](../reference/error-codes.md) — all error codes and retryability

---

## TODO

- [ ] Define backoff algorithm formally
- [ ] Specify trust score calculation formula
- [ ] Document agent restoration criteria
- [ ] Define circuit breaker thresholds
- [ ] Add failure handling sequence diagrams
