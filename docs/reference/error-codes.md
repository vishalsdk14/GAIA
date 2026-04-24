# Error Code Catalog

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 3.2, 8, 10](../design.md)

---

## Purpose

This document is the **authoritative catalog** of every error code in the GAIA kernel. Agents, adapters, and clients should reference this to understand failure semantics.

---

## Error Codes to Define

### Kernel Error Codes

| Code | Retryable | Description | Typical Cause |
| :--- | :---: | :--- | :--- |
| `SCHEMA_VIOLATION` | No | | |
| `TIMEOUT` | Yes | | |
| `POLICY_DENIED` | No | | |
| `INTERNAL` | No | | |
| `UNKNOWN` | No | | |

---

### Planning Error Codes

| Code | Retryable | Description | Typical Cause |
| :--- | :---: | :--- | :--- |
| `PLANNER_TIMEOUT` | Yes | | |
| `PLANNER_MALFORMED_OUTPUT` | Yes | | |
| `PLANNER_EMPTY_PLAN` | No | | |
| `PLANNER_UNKNOWN_CAPABILITY` | Yes | | |
| `PLANNER_UNAVAILABLE` | No | | |

---

### Transport Error Codes

| Code | Retryable | Description | Typical Cause |
| :--- | :---: | :--- | :--- |
| `AGENT_UNREACHABLE` | Yes | | |
| `AGENT_TIMEOUT` | Yes | | |
| `TRANSPORT_ERROR` | Yes | | |
| `PROTOCOL_MISMATCH` | No | | |

---

### Registration Error Codes

| Code | Retryable | Description | Typical Cause |
| :--- | :---: | :--- | :--- |
| `INVALID_MANIFEST` | No | | |
| `AUTH_FAILED` | No | | |
| `DUPLICATE_AGENT` | No | | |

---

## TODO

- [ ] Fill in all descriptions and typical causes
- [ ] Add HTTP status code mappings for Client API
- [ ] Document error response format with examples
