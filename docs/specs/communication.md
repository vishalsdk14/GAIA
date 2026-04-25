# Communication Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Sections 3, 5](../design.md)

---

## Purpose

This document defines the communication paradigms used by the GAIA kernel. It formalizes how asynchronous capabilities deliver their results (the callback mechanisms), and establishes the strict delivery and ordering guarantees for the Event Bus.

---

## 1. Async Delivery Mechanisms

When a step is executed in `async` mode, the agent immediately returns a `Response` acknowledging the request and providing a `job_id`. The final output must be delivered via an `AsyncCompletion` payload (schemas.md Section 12). 

The kernel supports three delivery models to accommodate different agent architectures.

### Option A: Agent-Initiated Callback (Push)
* **Best For**: Low-latency agents, webhook-capable systems.
* **Flow**:
  1. Kernel sends `Request` with a unique callback URL in `metadata.reply_to`.
  2. Agent returns ACK Response with `job_id`.
  3. Agent processes task.
  4. Agent sends HTTP POST to `reply_to` with the `AsyncCompletion` payload.

### Option B: Kernel Polling (Pull)
* **Best For**: Stateless agents, unreliable networks, long-running batch jobs.
* **Flow**:
  1. Kernel sends `Request`.
  2. Agent returns ACK Response with `job_id`.
  3. Kernel polls the agent's endpoint: `GET /agent/{agent_id}/jobs/{job_id}` at exponential backoff:
     * **Initial delay**: 100ms
     * **Multiplier**: 2x
     * **Max delay**: 30s
     * **Jitter**: ±20%
     * *(Example sequence: 100ms, 200ms, 400ms, 800ms, 1.6s, ..., 30s)*
  4. Agent returns HTTP 202 until done, then returns 200 OK with `AsyncCompletion` payload.

### Option C: Event Bus Subscription (Publish-Subscribe)
* **Best For**: Event-driven native architectures, MCP integrations.
* **Flow**:
  1. Kernel sends `Request`.
  2. Agent returns ACK Response with `job_id`.
  3. Agent publishes the `AsyncCompletion` payload to the kernel's message broker (e.g., Kafka, Redis PubSub) on a predefined topic (`gaia.agent.completions`).
  4. Kernel consumes the event and resumes the step.

---

## 2. Event Bus Guarantees

The GAIA Kernel uses a strict event-sourcing model. Every state transition emits an `Event` (schemas.md Section 7).

### 2.1 Event Ordering (Causality)
To ensure clients can accurately reconstruct state, events are causally ordered per-task.
* **Monotonic Counter**: Every event includes a `sequence_number`. The counter starts at `1` when the `TASK_CREATED` event is fired and strictly increments by 1 for every subsequent event in that `task_id`.
* **Hash Chaining**: Every event includes a `previous_event_id` field containing the cryptographic hash (SHA-256) of the preceding event payload. This forms a tamper-evident chain (similar to a blockchain).

*Note: Events across different `task_id`s have no strict ordering guarantees relative to one another.*

### 2.2 Durability & Replay
* **Append-Only Immutable Log**: All events are persisted to a durable write-ahead log (WAL) *before* they are emitted to subscribers. Once written, an event cannot be altered or deleted.
* **Replayability**: If a client disconnects or crashes, it can request a replay from the kernel: `GET /tasks/{task_id}/events?since_sequence=N`. This ensures zero data loss for downstream observers.
* **Atomicity**: State updates within the kernel and event emissions are executed within a distributed transaction. If the event cannot be persisted to the log, the state transition is rolled back.

---

## 3. Communication Boundaries

To adhere to the "Orchestrator as Firewall" principle (design.md Section 5):

1. **No Peer-to-Peer**: Agents must never communicate directly. All inter-agent data transfer occurs strictly through step outputs mediated by the kernel.
2. **Strict Ingress**: All inbound agent communications (Callbacks, Polling Responses, PubSub events) must pass through the Validation Pipeline (Schema validation, Auth check) before the payload is allowed to mutate the active state.

---

## Related Documents

* [Schemas Spec](schemas.md) — Event and AsyncCompletion payload schemas.
* [Control Loop Spec](control-loop.md) — How the kernel processes async completions (Phase 9).
* [Transport Spec](transport.md) — The physical wire protocols mapping to these models.
