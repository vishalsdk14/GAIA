# Anthropic MCP Protocol Integration

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 21.3: MCP Integration](../design.md)

---

## Purpose

This document provides the **complete specification** for how the GAIA kernel interacts with Anthropic's Model Context Protocol (MCP). It covers tool discovery, capability translation, invocation, resource access, and output normalization.

---

## Sections to Define

### 1. MCP Protocol Overview

* Protocol purpose: agent-to-tool/data connectivity
* Architecture: client-server over JSON-RPC 2.0 (stdio or HTTP+SSE)
* Key concepts: Tools, Resources, Prompts, Sampling

---

### 2. Connection Lifecycle

* GAIA as MCP Client connecting to MCP Servers
* Connection initialization and capability negotiation
* Session management (stateful connections)
* Reconnection strategy

---

### 3. Tool Discovery

* `tools/list` call and response handling
* Dynamic tool re-discovery (tools can change at runtime)
* `listChanged` notification handling

---

### 4. MCP Tool → GAIA Capability Translation

* Field-by-field mapping table
* `tool.name` → `capability.name`
* `tool.inputSchema` → `capability.input_schema`
* `tool.annotations` → `capability.constraints`
* Output schema inference (MCP tools don't declare output schemas)

---

### 5. Invocation

* `tools/call` request format
* Argument serialization
* Timeout handling

---

### 6. Response Normalization

* MCP content types → GAIA step output:
  * `text` content → `output.text`
  * `image` content → `output.image_url` or base64
  * `resource` content → `output.resource`
* `isError` flag → GAIA Error schema mapping

---

### 7. Resource Integration

* `resources/read` for contextual data injection
* Resource URI → Interpolation Engine (`{{mcp.resource_uri}}`)
* Resource subscription for live data

---

### 8. Prompt Integration

* Can MCP Prompts be used as planning templates?
* `prompts/list` and `prompts/get` integration

---

### 9. Security Considerations

* MCP's trust model (tool descriptions are untrusted)
* GAIA's policy enforcement on MCP tools
* User consent requirements

---

## TODO

- [ ] Document complete MCP tool schema
- [ ] Implement output schema inference strategy
- [ ] Handle dynamic tool list changes
- [ ] Test with reference MCP servers
- [ ] Document resource subscription model
