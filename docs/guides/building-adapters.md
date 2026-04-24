# Building Protocol Adapters for GAIA

> **Status**: 🔲 Not Started — will be written when the adapter interface is finalized.

---

## Purpose

This guide will walk a contributor through **building a new protocol adapter** for the GAIA kernel. If a new agent protocol emerges (beyond A2A and MCP), this guide explains how to integrate it.

---

## Planned Sections

### 1. What is a Protocol Adapter?
* A Transport Layer component
* Translates between GAIA's internal format and an external protocol
* Must implement the standard adapter interface

### 2. Adapter Interface Contract
* `discover(url) → Manifest`
* `invoke(agent, payload) → Response`
* `cancel(agent, step_id) → void`
* `health(agent) → HealthStatus`

### 3. Discovery & Manifest Translation
* How to map external agent metadata to GAIA's Manifest schema
* Capability extraction strategies
* Schema inference when output schemas are missing

### 4. Invocation Mapping
* Translating GAIA payloads to the external wire format
* Handling sync and async modes
* Timeout enforcement

### 5. Response Normalization
* Converting external responses to GAIA's Response schema
* Error mapping

### 6. Testing
* Adapter conformance test suite
* Mock agents for testing
* Integration test requirements

### 7. Registration
* How to register your adapter with the Transport Layer
* Configuration requirements

---

## TODO

- [ ] Write after transport layer spec is finalized
- [ ] Create adapter conformance test suite
- [ ] Include example adapter implementation
