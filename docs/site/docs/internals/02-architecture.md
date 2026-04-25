# Chapter 2: Kernel Architecture

The GAIA Kernel is implemented in **Go (Golang)**, chosen for its high-performance concurrency model and industry-standard orchestration capabilities. This chapter explores the internal components of the Kernel and the "Heartbeat" that drives it: the 10-Phase Control Loop.

---

## 2.1 The Process Model

The Kernel operates as a long-running daemon. It is designed to manage thousands of concurrent **Tasks**, each of which may contain multiple parallel **Steps**.

### Goroutine Per Task
Each task is assigned its own **Control Loop** goroutine upon submission. This ensures that a slow-running task (e.g., waiting for an async agent) never blocks the execution of other tasks.

### Thread-Safe Internal Bus
All internal components (Orchestrator, Scheduler, Policy Engine) communicate via an asynchronous **Event Bus**. This ensures that state transitions are decoupled from execution, allowing for high throughput and clean observability.

---

## 2.2 The 10-Phase Control Loop

The Control Loop is the definitive state machine for the Kernel. Every task follows these ten phases:

1.  **Submission**: Validation of the goal and creation of the `TaskID`.
2.  **Planning**: Invocation of the pluggable Planner to generate a `PlanRecord`.
3.  **Scheduling**: Building the **Directed Acyclic Graph (DAG)** of step dependencies.
4.  **Interpolation**: Resolving `{{step.output}}` markers into concrete data.
5.  **Policy Check**: Evaluating CEL rules (e.g., budget, safety, human-approval).
6.  **Routing**: Selecting the healthiest agent from the Registry for a specific capability.
7.  **Dispatch**: Transmitting the request via the appropriate **Transport Adapter**.
8.  **Execution**: The period during which the agent performs the work (Sync or Async).
9.  **Validation**: Post-execution schema check of the agent's output.
10. **State Update**: Committing the results to the State Store and checking if more steps are needed.

For a detailed technical breakdown of each phase, see the [Control Loop Specification](../../specs/control-loop.md).

---

## 2.3 Internal Components

The Kernel is organized into several modular packages:

### `pkg/core` (The Brain)
Contains the **Orchestrator** (manages the loop), the **Planner Interface** (LLM bridge), and the **Scheduler** (DAG resolver). This package is responsible for all decision-making.

### `pkg/policy` (The Firewall)
Implements the **CEL Engine**. Every message moving in or out of the Kernel is intercepted here. It is designed for microsecond evaluation to minimize latency.

### `pkg/state` (The Memory)
Manages the **Tiered State Model**. It handles the transition from Tier 1 (In-memory hot state) to Tier 2 (SQLite persistence) and manages state isolation for agents.

### `pkg/registry` (The Catalog)
Maintains the authoritative list of agents, their capabilities, and their **Trust Scores**. It performs the "Handshake" when an agent attaches to the Kernel.

### `pkg/api` (The Gateway)
Provides the RESTful interface for clients and the **WebSocket Event Stream** for real-time observability.

---

## 2.4 The Determinism Invariant

A critical rule of the GAIA Architecture is that **given the same plan and the same agent responses, the Kernel must produce the same state transitions.**

This is achieved by:
*   **Atomic Transitions**: Status changes (e.g., `running` -> `done`) are atomic and logged.
*   **Immutable Goal**: The original goal can never be modified after submission.
*   **Schema Enforcement**: No unstructured data is allowed to bypass the validation phase.

This determinism allows GAIA tasks to be **paused, resumed, and audited** with absolute reliability.
