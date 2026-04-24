# Building Agents for GAIA

> **Status**: 🔲 Not Started — will be written when the native protocol is finalized.

---

## Purpose

This guide will walk a developer through **building an agent that plugs into the GAIA kernel**. It covers the manifest, the invocation contract, health checks, and best practices.

---

## Planned Sections

### 1. What is a GAIA Agent?
* A stateless capability provider
* Receives structured input, returns structured output
* Managed by the kernel (not self-orchestrating)

### 2. Choose Your Protocol
* Native (simplest, recommended for new agents)
* A2A (if your agent already speaks A2A)
* MCP (if your agent is an MCP tool server)

### 3. Define Your Manifest
* Agent ID, version, capabilities
* Input/output schemas
* Constraints and idempotency
* Health endpoint

### 4. Implement the Invocation Endpoint
* Request payload handling
* Sync vs. async response
* Output schema compliance

### 5. Implement Health Checks
* Health endpoint contract
* What the kernel checks and when

### 6. Handle Cancellation
* CANCEL message handling (best-effort)
* Graceful cleanup

### 7. Testing Your Agent
* Manifest validation tool
* Mock kernel for local testing
* Schema compliance tests

### 8. Best Practices
* Keep agents stateless
* Validate your own output before returning
* Set realistic timeouts
* Handle idempotency correctly

---

## TODO

- [ ] Write after native protocol spec is finalized
- [ ] Include code examples in TypeScript and Python
