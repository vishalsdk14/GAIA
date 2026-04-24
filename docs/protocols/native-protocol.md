# GAIA Native Protocol Specification

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Sections 3, 4, 7](../design.md)

---

## Purpose

This document specifies **GAIA's native protocol** — the default communication format for agents built specifically for the GAIA kernel. Unlike A2A and MCP (which are external protocols consumed via adapters), the native protocol is GAIA's "first-class" wire format.

---

## Sections to Define

### 1. Protocol Overview

* Direct JSON over HTTP or local IPC
* No external protocol dependency
* Simplest path for GAIA-native agents

---

### 2. Agent Registration (Native)

* Agent submits manifest via `POST /agents/register`
* Full manifest schema (see [schemas.md](../specs/schemas.md))
* Handshake flow (design.md Section 4.2)

---

### 3. Invocation Contract

* Request payload format
* Response format
* Error format
* Sync and async modes

---

### 4. Health Check Protocol

* `GET {health_endpoint}` contract
* Expected response format
* Timeout and failure behavior

---

### 5. Cancellation Protocol

* CANCEL message format
* Best-effort semantics (agent may ignore)
* Kernel does not wait for confirmation

---

### 6. Disconnect Protocol

* DRAIN signal format
* Graceful shutdown sequence
* Crash detection (no signal)

---

### 7. SDK Guidelines

* What a "GAIA-native agent SDK" should provide
* Required interfaces to implement
* Example agent structure

---

## TODO

- [ ] Define complete wire format with examples
- [ ] Specify health check response schema
- [ ] Document SDK interface requirements
- [ ] Create example agent implementation
