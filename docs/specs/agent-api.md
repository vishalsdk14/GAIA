# Agent API Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Section 5.4](../design.md)

---

## Purpose

The Agent API is the internal interface exposed by the GAIA Kernel specifically for connected Agents. Currently, its primary purpose is to expose the **Managed Agent State** (Tier 4) Key-Value store, allowing agents to persist data across tasks without provisioning their own external databases.

This ensures the Kernel retains absolute visibility and control over all agent storage (Auditability and the "Kill Switch").

---

## 1. Authentication & Authorization

All endpoints in the Agent API require authentication.

*   **Authentication Method**: Agents must use the exact authentication mechanism they defined in their `AgentManifest` during the handshake (e.g., `mTLS` client certificates, or an `api_key` in the `Authorization: Bearer <token>` header).
*   **Authorization Scope**: To access the Managed State API, the agent must have requested `state_requirements.required = true` in its manifest, and the Kernel must have approved the quota.
*   **Isolation**: The API is strictly siloed. An agent can *only* read and write to its own namespace (partitioned by its `agent_id`). Any attempt to access another agent's namespace will result in a `403 Forbidden` and trigger a `POLICY_DENIED` audit event.

---

## 2. Managed State Endpoints (Tier 4)

The Managed State API provides a simple Key-Value Document store. Keys are strings, and Values must be valid JSON objects.

### 2.1 Store State (Write)
Saves or overwrites a JSON document for a specific key.

*   **Endpoint**: `PUT /internal/v1/state/{key}`
*   **Headers**: `Content-Type: application/json`
*   **Request Body**: Any valid JSON object.
*   **Response**:
    *   `200 OK`: Successfully stored.
    *   `413 Payload Too Large`: The payload exceeds the agent's approved `max_bytes` quota.
    *   `400 Bad Request`: Invalid JSON.

### 2.2 Retrieve State (Read)
Fetches the JSON document associated with a key.

*   **Endpoint**: `GET /internal/v1/state/{key}`
*   **Response**:
    *   `200 OK`: Returns the JSON document.
    *   `404 Not Found`: Key does not exist.

### 2.3 Delete State
Removes a specific key-value pair from the agent's namespace.

*   **Endpoint**: `DELETE /internal/v1/state/{key}`
*   **Response**:
    *   `204 No Content`: Successfully deleted (or key didn't exist).

### 2.4 List Keys
Retrieves a paginated list of all keys currently stored by the agent.

*   **Endpoint**: `GET /internal/v1/state`
*   **Query Parameters**: `?limit=100&offset=0`
*   **Response**:
    ```json
    {
      "keys": ["user_prefs_123", "cache_xyz"],
      "total_keys": 2,
      "bytes_used": 1024,
      "bytes_quota": 50000000
    }
    ```

---

## 3. Quota Enforcement

The Kernel's Policy Engine intercepts every `PUT` request. It calculates the byte size of the incoming JSON payload.
If `current_bytes_used + new_payload_bytes > max_bytes` (as defined in the `AgentManifest`), the request is rejected with `413 Payload Too Large`.

---

## 4. The "Kill Switch" (Kernel Admin Action)

Because the state is managed by the GAIA Kernel, administrators have absolute control over agent data.
If an agent is flagged as malicious or deleted from the Registry, the Kernel executes a hard cascade delete:
`DELETE FROM tier_4_state WHERE agent_id = ?`

This ensures no orphaned data remains and prevents persistent malware.

---

## Related Documents

* [State Management Spec](state-management.md) — Tier 4 definitions.
* [Schemas Spec](schemas.md) — `AgentManifest` state requirements.
* [Security Spec](security.md) — Policy Engine rules for storage limits.
