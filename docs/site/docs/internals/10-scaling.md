# Chapter 10: Performance & Scaling

As a high-performance orchestration kernel, GAIA is designed to minimize the overhead of mediation while maximizing the concurrency of the agentic swarm.

---

## 10.1 Concurrency Model

The GAIA Kernel uses Go's native **Goroutines** to handle massive parallelism.

*   **Task Isolation**: Each task runs in its own goroutine. A complex task with a 5-minute wait time does not block other tasks.
*   **Step Parallelism**: The **Scheduler** automatically dispatches all "Ready" steps (those with satisfied dependencies) in parallel.
*   **Agent Parallelism**: A single agent can serve requests from multiple different tasks concurrently, up to the `max_concurrent_per_agent` limit defined in its manifest.

---

## 10.2 Hot-Path Optimization

The "Hot Path" of the Kernel—the cycle of Interpolation, Policy Check, and Dispatch—is optimized for sub-millisecond latency.

### Zero-Allocation Interpolation
The Kernel avoids traditional JSON unmarshalling for data binding (e.g., swapping `{{step.output}}`). Instead, it uses **byte-level traversal** (via libraries like `tidwall/gjson`). This avoids heap allocations and keeps Garbage Collection (GC) pauses near zero.

### Compiled Policy Cache
CEL policies are pre-compiled into an Abstract Syntax Tree (AST) during the agent handshake. Evaluating a rule against a request is a simple tree traversal, taking only a few microseconds.

---

## 10.3 Scaling the Kernel

GAIA supports three scaling strategies:

### 1. Vertical Scaling
Thanks to Go's efficient memory management, a single GAIA Kernel can manage thousands of tasks on a single multi-core server. 

### 2. The Hybrid Model (Local vs. Remote)
The Kernel can route requests based on latency requirements:
*   **Fast Path**: Local agents connected via **IPC** or direct function calls for low-latency utility tasks.
*   **Remote Path**: Scalable agents running in **Kubernetes** or serverless environments for heavy compute tasks.

### 3. Horizontal Scaling (Future)
By using an external **State Store** (e.g., PostgreSQL instead of SQLite), multiple GAIA Kernel instances can share the same task database, allowing for a cluster-based orchestration model.

---

## 10.4 Resource Quotas

To prevent a "Runaway Task" from consuming all Kernel resources, GAIA enforces:
*   **Task Budget**: Maximum total compute time and step count per task.
*   **Agent Budget**: Maximum concurrent requests and state storage per agent.
*   **Kernel Load Shedding**: If the Kernel's memory pressure exceeds a threshold, it will return `503 Service Unavailable` for new tasks until the load drops.

---

## 10.5 Related Specifications

*   [Tech Stack Spec](../../specs/tech-stack.md)
*   [Control Loop Spec](../../specs/control-loop.md)
*   [State Management Spec](../../specs/state-management.md)
