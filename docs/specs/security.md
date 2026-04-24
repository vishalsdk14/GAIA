# Security & Policy Engine Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 5, 13](../design.md)

---

## Purpose

This document specifies the **Policy Engine** (the "Firewall") and the **Security Model** — how the kernel authenticates agents, authorizes actions, enforces sandboxes, and prevents data leakage.

---

## Enforcement Principles

1. **Isolation First**: No agent shall have network or filesystem access beyond what is explicitly granted in its sandbox profile.
2. **Mediation Mandate**: All inter-agent data flow must pass through the kernel. Direct memory sharing or IPC between agents is a protocol violation.
3. **Schema Enforcement**: Every byte of input and output must be validated. Non-compliant agents are immediately transitioned to the `quarantined` state.
4. **Least Privilege**: Capabilities are granted with the minimum necessary scope (e.g., `read_only` by default).

---

## Sections to Define

### 1. Identity & Authentication

* Per-agent credentials: mTLS, signed tokens, OAuth
* Credential lifecycle (issuance, rotation, revocation)
* Agent identity verification on every request

---

### 2. Authorization (Scopes)

* Per-capability invocation rights
* Scope format: `capability:invoke`, `capability:read`, etc.
* Scope validation in the Validation Pipeline (Section 5.2)

---

### 3. Policy Rules

* Policy definition format (declarative rules)
* Examples:
  * "Agent X can only invoke read_only capabilities"
  * "No agent can spend more than $100 per task"
  * "Capabilities with `external_io` require human approval"
* Policy evaluation order (deny-by-default)

---

### 4. Sandbox Model

* What does "sandboxed execution" mean concretely?
* Network policy: deny-by-default egress
* Resource limits: CPU, memory, execution time
* Filesystem access restrictions

---

### 5. Data Minimization

* Agents receive only step-local input
* No global state leakage
* Policy exceptions (explicit opt-in)

---

### 6. Rate Limiting

* Per-agent rate limits
* Per-capability rate limits
* Rate limit response format

---

### 7. Audit & Compliance

* What is logged? (every message, every policy decision)
* Audit log format
* Retention policy
* Tamper-proof logging requirements

---

### 8. Threat Model

* Malicious agent scenarios
* Data exfiltration prevention
* Denial of service protection
* Trust score manipulation

---

## Related Documents

* [Data Model & Schemas](schemas.md) — AgentManifest auth field, Error schema
* [Communication Spec](communication.md) — policy enforcement in message flow
* [Registry Spec](registry.md) — sandbox assignment during registration
* [Failure Handling Spec](failure-handling.md) — enforcement actions on violations
* [Security Policy](../../SECURITY.md) — vulnerability disclosure process

---

## TODO

- [ ] Define policy rule format (DSL or JSON)
- [ ] Specify sandbox implementation requirements
- [ ] Document rate limit algorithm
- [ ] Create threat model matrix
- [ ] Define audit log schema
