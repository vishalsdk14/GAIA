# Capability Registry Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 4, 6.5](../design.md)

---

## Purpose

This document specifies the **Capability Registry** — the authoritative directory that maps capabilities to agents. It covers registration, deregistration, querying, and the routing algorithm used by the Dispatcher.

---

## Routing Constraints

1. **Capability Primacy**: The registry maps to capabilities, not agents. Agents are transient; capabilities are persistent.
2. **Health-Weighted Selection**: The dispatcher must prioritize agents with a trust score > 0.9 and latency within the 95th percentile of the capability average.
3. **Fallback Determinism**: The fallback chain must be pre-calculated or deterministic to prevent infinite "hunting" for agents.
4. **Handshake Integrity**: No agent shall be registered without a valid health endpoint and verified mTLS/OAuth credentials.

---

## Sections to Define

### 1. Registry Data Model

* Capability → Agent[] mapping
* Per-agent metadata stored: version, health, trust score, SLA
* Registry index structure

---

### 2. Registration (Handshake)

* Full CONNECT → VALIDATE → SANDBOX → REGISTER → READY flow
* Manifest validation rules
* Schema validation of capability definitions
* Auth verification
* Sandbox assignment logic

---

### 3. Deregistration (Disconnect)

* DRAIN → REASSIGN → DEREGISTER → CLOSED flow
* In-flight step reassignment logic
* Crash detection via health check failure
* Cleanup of registry entries

---

### 4. Registry Queries

* `lookup(capability) → Agent[]` — returns all agents for a capability
* `select(capability, constraints) → Agent` — returns best agent
* `list_capabilities() → Capability[]` — returns all available capabilities

---

### 5. Agent Selection Algorithm (Dispatcher)

* Scoring: health × SLA × trust × policy
* Tiebreaking rules
* Fallback chain: primary → secondary → replan
* Local vs. remote preference logic (Section 15)

---

### 6. Hot-Swap Support

* What happens when a new agent registers for an already-served capability?
* Can in-flight steps be redirected?
* Version conflict resolution

---

## Related Documents

* [Data Model & Schemas](schemas.md) — AgentManifest and AgentRecord schemas
* [Lifecycle State Machines](lifecycles.md) — agent status transitions
* [Transport Spec](transport.md) — local vs. remote routing
* [Failure Handling Spec](failure-handling.md) — health monitoring and enforcement
* [Native Protocol](../protocols/native-protocol.md) — native registration flow
* [A2A Integration](../protocols/a2a-integration.md) — Agent Card discovery
* [MCP Integration](../protocols/mcp-integration.md) — tool discovery

---

## TODO

- [ ] Define registry data model formally
- [ ] Specify agent selection scoring algorithm
- [ ] Document hot-swap behavior
- [ ] Add sequence diagrams for registration and disconnect
