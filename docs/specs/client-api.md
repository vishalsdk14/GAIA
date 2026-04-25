# Client API Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Section 3.4](../design.md)

---

## Purpose

This document defines the external-facing Application Programming Interface (API) used by clients (web apps, mobile apps, or other backend systems) to interact with the GAIA Kernel. It covers task submission, state querying, and real-time event streaming.

---

## Implementation Status

| Endpoint | Method | Status | Description |
| :--- | :--- | :--- | :--- |
| `/api/v1/tasks` | `POST` | ✅ Complete | Submits goal to Orchestrator |
| `/api/v1/tasks/{id}` | `GET` | ✅ Complete | Returns current task status/plan |
| `/api/v1/tasks/{id}` | `DELETE` | 🔲 Pending | Cancellation logic |
| `/api/v1/tasks/{id}/state` | `GET` | 🔲 Pending | Tier 1 state retrieval |
| `/api/v1/registry/agents` | `GET` | ✅ Complete | Lists all registered agents |
| `/api/v1/registry/capabilities`| `GET` | ✅ Complete | Lists available capability names |
| `/api/v1/stream` | `WS` | ✅ Complete | Real-time event streaming |
| `/api/v1/tasks/{id}/steps/{stepID}/approve` | `POST` | ✅ Complete | Manual approval gate |
| `/api/v1/admin/audit-logs` | `GET` | 🔲 Pending | Audit trail retrieval |

---

## 1. REST Endpoints

The Kernel exposes a standard HTTP REST API for synchronous operations. All endpoints expect and return `application/json`.

### 1.1 Task Management

* **`POST /api/v1/tasks`**
  * **Purpose**: Submit a new natural language goal.
  * **Request**: `{ "goal": "Summarize the latest financial news", "metadata": {} }`
  * **Response**: `201 Created`. Returns the newly generated `Task` object (status: `pending`).

* **`GET /api/v1/tasks/{task_id}`**
  * **Purpose**: Retrieve the current state of a task, including its active plan and current step index.
  * **Response**: Returns the full `Task` schema.

* **`DELETE /api/v1/tasks/{task_id}`**
  * **Purpose**: Cancel a running task.
  * **Response**: `202 Accepted`. Transitions task to `cancelled`.

### 1.2 State & Approvals

* **`GET /api/v1/tasks/{task_id}/state`**
  * **Purpose**: Fetch the current Tier 1 `ActiveState` (accumulated step outputs).

* **`POST /api/v1/tasks/{task_id}/steps/{step_id}/approve`**
  * **Purpose**: Manually unblock a step that is in the `AWAITING_APPROVAL` mode.
  * **Response**: `200 OK`. Step transitions back to `pending`.

### 1.3 Registry & Administration

* **`GET /api/v1/registry/capabilities`**
  * **Purpose**: List all active capabilities available to the Planner.
  
* **`GET /api/v1/registry/agents`**
  * **Purpose**: List all connected agents and their `AgentRecord` health metrics.

---

## 2. WebSocket Streaming Model

For real-time observability, clients should connect via WebSockets to consume the Event Bus.

* **Endpoint**: `wss://kernel-host/api/v1/stream?task_id={uuid}`
* **Behavior**: 
  1. The client upgrades the connection.
  2. The kernel streams all `Event` schemas (schemas.md Section 7) associated with the `task_id` in causal order.
  3. If a connection drops, the client can reconnect and pass `?since_sequence={N}` to replay missed events from the durable log.

**Stream Messages:**
```json
{
  "type": "EVENT",
  "name": "STEP_COMPLETED",
  "task_id": "123e4567-e89b-12d3-a456-426614174000",
  "step_id": "step_2",
  "sequence_number": 42,
  "payload": { ... }
}
```

---

## 3. Client Authentication

The GAIA Kernel API is secured via JWT (JSON Web Tokens).

1. **Token Provisioning**: Clients obtain a JWT from an external Identity Provider (e.g., Auth0, Keycloak).
2. **Bearer Auth**: All REST and WebSocket connections must include the token:
   `Authorization: Bearer <token>`
3. **Scope Validation**: The kernel validates that the client has the necessary scopes (e.g., `tasks:write` to create tasks, `admin:read` to view the registry).

*Note: Agent authentication is handled separately during the Registry Handshake via mTLS or Agent API Keys (see registry.md).*

---

## Related Documents

* [Control Loop Spec](control-loop.md) — How the API integrates into Phase 1 of the loop.
* [Communication Spec](communication.md) — Event streaming guarantees.
* [Security Spec](security.md) — Approval overrides via the API.
