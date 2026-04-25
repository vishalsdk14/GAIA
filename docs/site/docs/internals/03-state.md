# Chapter 3: Memory & State Management

In the GAIA Kernel, state is not just a variable—it is a tiered, durable, and isolated resource. GAIA treats state like an Operating System treats memory pages, providing different "tiers" of access speed and persistence.

---

## 3.1 The 5-Tier State Model

GAIA uses a hierarchical state model to balance the need for fast planning context with the requirement for long-term auditing and legal compliance.

| Tier | Type | Location | Purpose |
| :--- | :--- | :--- | :--- |
| **Tier 1** | **Active State** | RAM (Goroutine-local) | Hot variables used by the Planner for the current step. |
| **Tier 2** | **Task History** | SQLite (Local) | Durable log of every step result and state snapshot for recovery. |
| **Tier 3** | **Archived State** | Cold Storage (S3/Disk) | Compressed history of completed tasks for long-term reference. |
| **Tier 4** | **Managed State** | SQLite (Partitioned) | Persistent "memory" provided to agents, isolated by `AgentID`. |
| **Tier 5** | **Audit Log** | WORM Storage | Tamper-proof, cryptographically chained log of every transition. |

---

## 3.2 Tier 4: Agent State Isolation

One of GAIA's most powerful security features is **Tier 4 Managed State**. Agents are forbidden from hosting their own external databases. Instead, they interact with the Kernel's State API.

### How it works:
1.  An agent requests to `GET` or `SET` a key.
2.  The Kernel identifies the agent via its **mTLS Certificate** or **AgentID Header**.
3.  The Kernel performs a **Partition Check**: Agents can only access keys within their own namespace.
4.  The Kernel enforces **Quotas**: Limits on the total bytes an agent can store (defined in the `AgentManifest`).

This ensures that an agent cannot "leak" data between different tasks or users unless explicitly allowed by the Policy Engine.

---

## 3.3 Snapshotting & Recovery

To keep the Planner's context window small and efficient, GAIA implements a **Snapshotting Strategy** (see [state-management.md](../../specs/state-management.md)).

### The Trigger:
A snapshot is triggered when:
*   A specific number of steps are completed (e.g., every 10 steps).
*   The cumulative size of the Active State exceeds a threshold.

### The Process:
1.  The Kernel summarizes the recent execution history.
2.  The "Hot" state is pruned of unnecessary intermediate results.
3.  The snapshot is committed to Tier 2 (SQLite).
4.  If the Kernel crashes, it can resume any task by reloading the latest Tier 2 snapshot.

---

## 3.4 Data Integrity & Determinism

All state updates in GAIA are **Append-Only** in the database. We never overwrite a step's result. This ensures a perfect "Time-Travel Debugging" capability where a developer can see exactly what the system knew at Step 14 of a 100-step task.

### The Audit Chain (Tier 5):
Every state transition is hashed using **SHA-256**, and each hash includes the hash of the previous entry. This creates a "Blockchain-lite" structure that makes the Audit Log tamper-proof. If a single byte of task history is modified by an attacker, the chain will break, and the Kernel will raise a security alert.

---

## 3.5 Related Specifications

For deeper implementation details, refer to:
*   [State Management Spec](../../specs/state-management.md)
*   [Schemas Spec (State Objects)](../../specs/schemas.md)
*   [Security Spec (Audit Chaining)](../../specs/security.md)
