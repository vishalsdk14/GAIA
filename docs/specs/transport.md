# Transport Layer Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 7, 15, 21.4](../design.md)

---

## Purpose

This document specifies the **Transport Layer** — the abstraction that makes agent invocation uniform regardless of whether the agent is local (IPC), remote (HTTP/gRPC), A2A-compatible, or MCP-compatible. It defines the adapter interface, routing logic, and protocol-specific wire formats.

---

## Sections to Define

### 1. Transport Abstraction Interface

* `invoke(agent, payload) → Response`
* How the Transport Layer selects the correct adapter
* Protocol resolution from `agent.protocol` field

---

### 2. Native Adapter

* Local transport: direct function call / IPC
* Remote transport: HTTP POST / gRPC call
* Wire format (GAIA's internal JSON schema)
* Timeout handling

---

### 3. A2A Adapter

* See [A2A Integration](../protocols/a2a-integration.md) for full protocol details
* JSON-RPC 2.0 methods: `message/send`, `message/stream`, `tasks/get`, `tasks/cancel`
* Agent Card → Manifest translation
* A2A Artifact → GAIA Response normalization

---

### 4. MCP Adapter

* See [MCP Integration](../protocols/mcp-integration.md) for full protocol details
* JSON-RPC 2.0 methods: `tools/call`, `tools/list`, `resources/read`
* MCP Tool → GAIA Capability translation
* MCP content → GAIA Response normalization

---

### 5. Adapter Interface Contract

* Every adapter must implement:
  * `discover(url) → Manifest` — discover agent capabilities
  * `invoke(agent, payload) → Response` — send a request
  * `cancel(agent, step_id) → void` — cancel an in-flight step (best-effort)
  * `health(agent) → HealthStatus` — check agent health

---

### 6. Local vs. Remote Routing

* Agent classification: `location`, `transport`, `compute_class`
* Routing preferences based on latency requirements
* Failure characteristics per transport type

---

### 7. Adding New Adapters

* How to implement a new protocol adapter
* Registration with the Transport Layer
* Testing requirements

---

## Related Documents

* [A2A Integration](../protocols/a2a-integration.md) — A2A adapter details
* [MCP Integration](../protocols/mcp-integration.md) — MCP adapter details
* [Native Protocol](../protocols/native-protocol.md) — native adapter details
* [Data Model & Schemas](schemas.md) — AgentManifest `transport` and `protocol` fields
* [Registry Spec](registry.md) — agent classification and routing strategy
* [Building Adapters Guide](../guides/building-adapters.md) — contributor guide for new adapters

---

## TODO

- [ ] Define formal adapter interface (TypeScript/Go)
- [ ] Document A2A wire format with examples
- [ ] Document MCP wire format with examples
- [ ] Specify error mapping per protocol
- [ ] Add sequence diagrams for each transport type
