# Google A2A Protocol Integration

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 21.2: A2A Integration](../design.md)

---

## Purpose

This document provides the **complete specification** for how the GAIA kernel interacts with Google's Agent-to-Agent (A2A) protocol. It covers discovery, manifest translation, invocation mapping, and output normalization.

---

## Sections to Define

### 1. A2A Protocol Overview

* Protocol purpose: agent-to-agent delegation
* Architecture: client-server over JSON-RPC 2.0 / HTTPS
* Key concepts: Agent Card, Task, Message, Artifact, Parts

---

### 2. Agent Discovery

* Fetching `/.well-known/agent.json`
* Agent Card schema
* Periodic re-discovery for capability updates

---

### 3. Agent Card → GAIA Manifest Translation

* Field-by-field mapping table
* Handling of skills → capabilities conversion
* Auth method mapping (OAuth, API Key)
* Input/Output mode → JSON Schema inference

---

### 4. Task Lifecycle Mapping

* GAIA Step states ↔ A2A Task states
* `input-required` handling (GAIA has no equivalent — how to bridge?)
* Streaming task updates via SSE

---

### 5. Invocation

* `message/send` for synchronous requests
* `message/stream` for async/streaming requests
* Message Parts construction (TextPart, DataPart, FilePart)
* Request serialization

---

### 6. Response Normalization

* A2A Artifacts → GAIA step output
* Part type handling:
  * TextPart → `output.text`
  * FilePart → `output.file_url` or base64
  * DataPart → `output.data`
* Multi-part artifact assembly

---

### 7. Error Mapping

* A2A error responses → GAIA Error schema
* A2A task failure → GAIA step failure

---

### 8. Authentication

* OAuth flow for A2A agents
* Token management and refresh
* mTLS support

---

## TODO

- [ ] Document complete Agent Card schema
- [ ] Implement field mapping with edge cases
- [ ] Handle `input-required` state (human-in-the-loop bridge)
- [ ] Test with reference A2A agents
- [ ] Document streaming protocol details
