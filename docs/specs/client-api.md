# Client API Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 14: Client Interface](../design.md)

---

## Purpose

This document specifies the **Client-facing REST API** — the boundary between external users/applications and the GAIA kernel. It defines endpoints, request/response schemas, authentication, and streaming protocols.

---

## Endpoints to Define

### 1. Submit Goal

```http
POST /tasks
```

* Request body schema
* Response schema (task_id, initial status)
* Validation rules
* Rate limiting

---

### 2. Get Task Status

```http
GET /tasks/{task_id}
```

* Response schema (full task object with steps)
* Filtering options (steps only, status only)

---

### 3. Cancel Task

```http
POST /tasks/{task_id}/cancel
```

* Idempotency (cancelling an already cancelled task)
* Response schema
* Side effects (interrupt propagation)

---

### 4. List Tasks

```http
GET /tasks
```

* Pagination
* Filtering by status
* Sorting options

---

### 5. Streaming (Real-time Updates)

* WebSocket endpoint specification
* SSE endpoint specification
* Event format for streaming
* Connection lifecycle (subscribe, heartbeat, disconnect)

---

### 6. Agent Management (Admin)

```http
POST /agents/register
GET /agents
GET /agents/{agent_id}
DELETE /agents/{agent_id}
```

* Admin authentication requirements
* Agent status and health queries

---

### 7. Authentication

* Client authentication methods
* API key management
* Token format

---

### 8. Error Responses

* Standard error response format
* HTTP status code mapping
* Error code catalog

---

## Related Documents

* [Data Model & Schemas](schemas.md) — Task, Error, and Response schemas
* [Communication Spec](communication.md) — event streaming for clients
* [Event Catalog](../reference/event-catalog.md) — events available via WebSocket/SSE
* [Error Code Catalog](../reference/error-codes.md) — HTTP status code mapping
* [Getting Started Guide](../guides/getting-started.md) — first API usage walkthrough

---

## TODO

- [ ] Define OpenAPI 3.0 specification
- [ ] Document all request/response examples
- [ ] Specify rate limiting per endpoint
- [ ] Define WebSocket/SSE protocol details
- [ ] Add authentication flow diagrams
