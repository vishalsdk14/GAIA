# Data Model & Schema Definitions

> **Status**: 🟢 In Progress
>
> **Source**: [design.md — Sections 3.2, 4.1, 8.1, 9](../design.md)

---

## Purpose

This document defines the **canonical JSON Schemas** for every data object in the GAIA kernel. These schemas are the wire protocol — the single source of truth that ensures all components (kernel, adapters, agents) speak the same language.

---

## 1. AgentManifest

The **Agent Manifest** is the "Digital Identity" submitted by every agent during the Handshake phase. It defines the agent's capabilities, its communication protocol, and its security constraints.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/agent-manifest.json",
  "title": "AgentManifest",
  "description": "The authoritative registration record for a GAIA agent",
  "type": "object",
  "properties": {
    "agent_id": {
      "type": "string",
      "description": "Unique identifier for the agent (reverse domain notation recommended)",
      "examples": ["com.example.coder-agent"]
    },
    "version": {
      "type": "string",
      "pattern": "^\\d+\\.\\d+\\.\\d+$",
      "description": "Semantic version of the agent"
    },
    "transport": {
      "type": "string",
      "enum": ["http", "ipc", "websocket", "grpc"],
      "description": "The underlying network transport used by the agent"
    },
    "protocol": {
      "type": "string",
      "enum": ["native", "a2a", "mcp"],
      "description": "The communication protocol dialect"
    },
    "endpoint": {
      "type": "string",
      "format": "uri",
      "description": "The base URL or pipe address for the agent"
    },
    "health_endpoint": {
      "type": "string",
      "format": "uri",
      "description": "Endpoint for heartbeat and health checks"
    },
    "capabilities": {
      "type": "array",
      "minItems": 1,
      "items": {
        "$ref": "#/$defs/capability"
      }
    },
    "auth": {
      "type": "object",
      "properties": {
        "type": { "type": "string", "enum": ["none", "bearer", "mTLS", "api_key"] },
        "secret_ref": { "type": "string", "description": "Reference to the secret in the Kernel vault" }
      },
      "required": ["type"]
    }
  },
  "required": ["agent_id", "version", "transport", "protocol", "endpoint", "capabilities"],
  "$defs": {
    "capability": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "pattern": "^[a-z0-9_]+$" },
        "description": { "type": "string" },
        "input_schema": { "type": "object", "description": "JSON Schema for expected input" },
        "output_schema": { "type": "object", "description": "JSON Schema for guaranteed output" },
        "constraints": {
          "type": "object",
          "properties": {
            "read_only": { "type": "boolean", "default": true },
            "mutates_state": { "type": "boolean", "default": false },
            "external_io": { "type": "boolean", "default": false }
          }
        }
      },
      "required": ["name", "description", "input_schema", "output_schema"]
    }
  }
}
```

---

## 2. Task

The **Task** object is the root state for a user goal. It tracks the overall progress, the evolved plan, and the global context.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/task.json",
  "title": "Task",
  "description": "The root object representing a user goal and its execution state",
  "type": "object",
  "properties": {
    "task_id": { "type": "string", "format": "uuid" },
    "goal": { "type": "string", "description": "The original natural language goal" },
    "status": {
      "type": "string",
      "enum": ["pending", "planning", "executing", "completed", "failed", "cancelled"]
    },
    "plan": {
      "type": "array",
      "items": { "$ref": "https://gaia-kernel.org/schemas/step.json" }
    },
    "metadata": {
      "type": "object",
      "additionalProperties": true
    },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "finished_at": { "type": "string", "format": "date-time" }
  },
  "required": ["task_id", "goal", "status", "created_at", "updated_at"]
}
```

---

## 3. Step

An individual unit of work within a plan.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/step.json",
  "title": "Step",
  "type": "object",
  "properties": {
    "step_id": { "type": "string" },
    "capability": { "type": "string", "description": "The capability required for this step" },
    "input": { "type": "object", "description": "The input data, potentially containing interpolations" },
    "depends_on": {
      "type": "array",
      "items": { "type": "string" },
      "description": "List of step_ids this step depends on"
    },
    "status": {
      "type": "string",
      "enum": ["pending", "running", "pending_async", "done", "failed"]
    },
    "assigned_agent": { "type": "string", "description": "The agent_id selected for this step" },
    "output": { "type": "object" },
    "error": {
      "type": "object",
      "properties": {
        "code": { "type": "string" },
        "message": { "type": "string" }
      }
    },
    "retry_count": { "type": "integer", "default": 0 }
  },
  "required": ["step_id", "capability", "input", "status"]
}
```

---

## 4. Request

The message sent from the Kernel to an Agent to trigger a capability invocation.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/request.json",
  "title": "Request",
  "type": "object",
  "properties": {
    "request_id": { "type": "string", "format": "uuid" },
    "task_id": { "type": "string", "format": "uuid" },
    "step_id": { "type": "string" },
    "capability": { "type": "string" },
    "input": { "type": "object" },
    "mode": {
      "type": "string",
      "enum": ["sync", "async"],
      "default": "sync"
    },
    "timeout_ms": { "type": "integer", "minimum": 1000 }
  },
  "required": ["request_id", "task_id", "step_id", "capability", "input"]
}
```

---

## 5. Response

The standardized output returned by an Agent after processing a Request.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/response.json",
  "title": "Response",
  "type": "object",
  "properties": {
    "request_id": { "type": "string", "format": "uuid" },
    "success": { "type": "boolean" },
    "output": {
      "type": "object",
      "description": "Must conform to the output_schema defined in the agent's manifest for this capability"
    },
    "error": { "$ref": "https://gaia-kernel.org/schemas/error.json" },
    "metrics": {
      "type": "object",
      "properties": {
        "duration_ms": { "type": "integer" },
        "cost_estimate": { "type": "number" },
        "tokens_used": { "type": "integer" }
      }
    }
  },
  "required": ["request_id", "success"]
}
```

---

## 6. Error

The structured failure object used throughout the system.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/error.json",
  "title": "Error",
  "type": "object",
  "properties": {
    "code": {
      "type": "string",
      "enum": [
        "SCHEMA_VIOLATION",
        "TIMEOUT",
        "POLICY_DENIED",
        "CAPABILITY_NOT_FOUND",
        "AGENT_UNAVAILABLE",
        "EXECUTION_FAILED",
        "INTERNAL_ERROR"
      ]
    },
    "message": { "type": "string" },
    "retryable": { "type": "boolean", "default": false },
    "details": { "type": "object", "additionalProperties": true }
  },
  "required": ["code", "message"]
}
```

---

## Related Documents

* [Lifecycle State Machines](lifecycles.md) — valid status transitions
* [Communication Spec](communication.md) — message flow using these schemas
* [Error Code Catalog](../reference/error-codes.md) — all error codes
* [Event Catalog](../reference/event-catalog.md) — all event types

---

## TODO

- [x] Define AgentManifest schema
- [x] Define Task schema
- [x] Define Step schema
- [x] Define Request/Response schemas
- [x] Define Error schema
- [ ] Define Event schema
- [ ] Define Snapshot schema
