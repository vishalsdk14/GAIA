# Data Model & Schema Definitions

> **Status**: 🟢 In Progress
>
> **Source**: [design.md — Sections 2.5, 3.2, 4.1, 8.1, 9](../design.md)

---

## Purpose

This document defines the **canonical JSON Schemas** for every data object in the GAIA kernel. These schemas are the wire protocol — the single source of truth that ensures all components (kernel, adapters, agents) speak the same language.

All schemas follow [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12/schema).

---

## Errata (Intentional Divergences from design.md)

The following fields have been intentionally refined from their original definition in design.md to improve correctness and expressiveness:

| Field | design.md | schemas.md | Rationale |
| :--- | :--- | :--- | :--- |
| `constraints` | Array of strings | Object with boolean flags | Allows a capability to express multiple constraints simultaneously (e.g., `mutates_state: true` AND `external_io: true`) |
| `base_url` | `base_url` | `endpoint` | `endpoint` is transport-agnostic — supports IPC pipes, WebSocket URIs, and HTTP URLs |
| `transport` | `http \| grpc \| local` | `http \| grpc \| ipc \| websocket` | `local` renamed to `ipc` for precision; `websocket` added for streaming-first agents |
| `error.code: INTERNAL` | `INTERNAL` | `INTERNAL_ERROR` | Consistent with `_ERROR` / `_VIOLATION` suffix pattern across the enum |

---

## 1. AgentManifest

The **Agent Manifest** is the "Digital Identity" submitted by every agent during the Handshake phase (design.md Section 4.1). It defines the agent's capabilities, invocation contract, communication protocol, and security constraints.

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
      "enum": ["http", "grpc", "ipc", "websocket"],
      "description": "The underlying network transport used by the agent"
    },
    "protocol": {
      "type": "string",
      "enum": ["native", "a2a", "mcp"],
      "description": "The communication protocol dialect"
    },
    "endpoint": {
      "type": "string",
      "description": "The base URL, pipe address, or WebSocket URI for the agent"
    },
    "health_endpoint": {
      "type": "string",
      "description": "Endpoint for heartbeat and health checks"
    },
    "invoke": {
      "type": "object",
      "description": "Invocation contract: default timeout and async support",
      "properties": {
        "timeout_ms": {
          "type": "integer",
          "minimum": 1000,
          "default": 15000,
          "description": "Default timeout for capability invocations"
        },
        "async_supported": {
          "type": "boolean",
          "default": false,
          "description": "Whether the agent supports async (polling/streaming) invocations"
        }
      },
      "required": ["timeout_ms"]
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
      "description": "Authentication and authorization configuration",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["none", "bearer", "mTLS", "oauth", "api_key"]
        },
        "secret_ref": {
          "type": "string",
          "description": "Reference to the secret in the Kernel vault"
        },
        "scopes": {
          "type": "array",
          "items": { "type": "string" },
          "description": "Authorized scopes for this agent (e.g., 'capability:invoke')"
        }
      },
      "required": ["type"]
    }
  },
  "required": [
    "agent_id",
    "version",
    "transport",
    "protocol",
    "endpoint",
    "invoke",
    "capabilities"
  ],
  "$defs": {
    "capability": {
      "type": "object",
      "description": "A single capability offered by an agent",
      "properties": {
        "name": {
          "type": "string",
          "pattern": "^[a-z0-9_.-]+$",
          "description": "Machine-readable capability identifier"
        },
        "description": {
          "type": "string",
          "description": "Human-readable description of what this capability does"
        },
        "input_schema": {
          "$ref": "https://json-schema.org/draft/2020-12/schema",
          "description": "JSON Schema defining the expected input structure"
        },
        "output_schema": {
          "$ref": "https://json-schema.org/draft/2020-12/schema",
          "description": "JSON Schema defining the guaranteed output structure"
        },
        "idempotent": {
          "type": "boolean",
          "default": false,
          "description": "If true, this capability is safe to retry without side effects"
        },
        "constraints": {
          "type": "object",
          "description": "Behavioral constraints declared by the agent",
          "properties": {
            "read_only": {
              "type": "boolean",
              "default": true,
              "description": "If true, the capability does not modify external state"
            },
            "mutates_state": {
              "type": "boolean",
              "default": false,
              "description": "If true, the capability modifies external state"
            },
            "external_io": {
              "type": "boolean",
              "default": false,
              "description": "If true, the capability performs network or filesystem I/O"
            }
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

The **Task** object is the root state for a user goal (design.md Section 9.1). It tracks the overall progress, the evolved plan, and the global context.

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
    "goal": {
      "type": "string",
      "description": "The original natural language goal (immutable after creation)"
    },
    "status": {
      "type": "string",
      "enum": ["pending", "planning", "executing", "completed", "failed", "cancelled"]
    },
    "plan": {
      "type": "array",
      "items": { "$ref": "https://gaia-kernel.org/schemas/step.json" }
    },
    "current_step": {
      "type": "integer",
      "minimum": 0,
      "description": "Index of the currently active step in the plan (O(1) lookup)"
    },
    "metadata": {
      "type": "object",
      "additionalProperties": true,
      "description": "Extensible key-value store for client-provided context"
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

An individual unit of work within a plan (design.md Section 9.2).

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/step.json",
  "title": "Step",
  "type": "object",
  "properties": {
    "step_id": { "type": "string" },
    "capability": {
      "type": "string",
      "description": "The capability required for this step (must exist in the Capability Registry)"
    },
    "input": {
      "description": "The input data (any JSON value), potentially containing interpolation references (e.g., {{step_1.output.url}})"
    },
    "depends_on": {
      "type": "array",
      "items": { "type": "string" },
      "default": [],
      "description": "List of step_ids that must complete before this step can execute"
    },
    "status": {
      "type": "string",
      "enum": ["pending", "running", "pending_async", "done", "failed"]
    },
    "assigned_agent": {
      "type": "string",
      "description": "The agent_id selected by the Capability Registry for this step"
    },
    "output": { "description": "The output data (any JSON value) returned by the agent" },
    "error": { "$ref": "https://gaia-kernel.org/schemas/error.json" },
    "retry_count": { "type": "integer", "default": 0 }
  },
  "required": ["step_id", "capability", "input", "status"]
}
```

---

## 4. Request

The message sent from the Kernel to an Agent to trigger a capability invocation (design.md Section 3.2).

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/request.json",
  "title": "Request",
  "type": "object",
  "properties": {
    "type": {
      "const": "REQUEST",
      "description": "Message type discriminator"
    },
    "request_id": { "type": "string", "format": "uuid" },
    "from": {
      "type": "string",
      "description": "The originator of the request (kernel or agent_id) for audit attribution"
    },
    "task_id": { "type": "string", "format": "uuid" },
    "step_id": { "type": "string" },
    "capability": { "type": "string" },
    "input": { "description": "The fully resolved input data (any JSON value)" },
    "mode": {
      "type": "string",
      "enum": ["sync", "async"],
      "default": "sync"
    },
    "timeout_ms": {
      "type": "integer",
      "minimum": 1000,
      "description": "Per-request timeout override (falls back to agent invoke.timeout_ms)"
    }
  },
  "required": ["type", "request_id", "from", "task_id", "step_id", "capability", "input"]
}
```

---

## 5. Response

The standardized output returned by an Agent after processing a Request (design.md Section 3.2).

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
      "description": "Must conform to the output_schema defined in the agent's manifest for this capability"
    },
    "error": { "$ref": "https://gaia-kernel.org/schemas/error.json" },
    "job_id": {
      "type": "string",
      "description": "Async tracking ID returned when mode is async (agent ACK). Used by kernel to correlate async completion events."
    },
    "metrics": {
      "type": "object",
      "properties": {
        "duration_ms": { "type": "integer" },
        "cost_estimate": { "type": "number" },
        "tokens_used": { "type": "integer" }
      }
    }
  },
  "allOf": [
    {
      "if": { "properties": { "success": { "const": true } } },
      "then": { "required": ["request_id", "success", "output"] },
      "else": { "required": ["request_id", "success", "error"] }
    }
  ]
}
```

---

## 6. Error

The structured failure object used throughout the system (design.md Section 3.2).

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
        "INTERNAL_ERROR",
        "UNKNOWN"
      ],
      "description": "Machine-readable error classification"
    },
    "message": {
      "type": "string",
      "description": "Human-readable error description"
    },
    "retryable": {
      "type": "boolean",
      "default": false,
      "description": "If true, the kernel may retry this step per the RetryPolicy"
    },
    "details": {
      "type": "object",
      "additionalProperties": true,
      "description": "Optional structured context (e.g., validation errors, stack traces)"
    }
  },
  "required": ["code", "message"]
}
```

---

## 7. Event

Asynchronous event emitted by the Kernel via the Event Bus (design.md Sections 3.2, 3.6).

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/event.json",
  "title": "Event",
  "type": "object",
  "properties": {
    "type": {
      "const": "EVENT",
      "description": "Message type discriminator"
    },
    "name": {
      "type": "string",
      "enum": [
        "TASK_CREATED",
        "TASK_PLANNING",
        "TASK_EXECUTING",
        "TASK_COMPLETED",
        "TASK_FAILED",
        "TASK_CANCELLED",
        "STEP_STARTED",
        "STEP_APPROVAL_REQUIRED",
        "STEP_COMPLETED",
        "STEP_FAILED",
        "PLAN_GENERATED",
        "PLAN_REJECTED",
        "REPLAN_TRIGGERED",
        "AGENT_REGISTERED",
        "AGENT_DEGRADED",
        "AGENT_QUARANTINED",
        "AGENT_BLACKLISTED",
        "AGENT_EJECTED",
        "AGENT_DISCONNECTED"
      ],
      "description": "The event name from the Event Catalog"
    },
    "payload": {
      "type": "object",
      "additionalProperties": true,
      "description": "Event-specific data"
    },
    "task_id": { "type": "string", "format": "uuid" },
    "step_id": { "type": "string" },
    "timestamp": { "type": "string", "format": "date-time" }
  },
  "required": ["type", "name", "task_id", "timestamp"]
}
```

---

## 8. AgentRecord

The Kernel's internal record for a registered agent (design.md Section 9.3). This is the runtime counterpart of the AgentManifest — it tracks the agent's health and behavioral metrics after registration.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/agent-record.json",
  "title": "AgentRecord",
  "type": "object",
  "properties": {
    "agent_id": { "type": "string" },
    "status": {
      "type": "string",
      "enum": ["connecting", "active", "degraded", "quarantined", "blacklisted", "disconnected", "rejected"],
      "description": "Current lifecycle status of the agent (see lifecycles.md Section 3)"
    },
    "trust_score": {
      "type": "number",
      "minimum": 0.0,
      "maximum": 1.0,
      "description": "Composite trust score used for routing decisions"
    },
    "registered_at": { "type": "string", "format": "date-time" },
    "last_health_check": { "type": "string", "format": "date-time" },
    "rolling_metrics": {
      "type": "object",
      "properties": {
        "success_rate": {
          "type": "number",
          "minimum": 0.0,
          "maximum": 1.0
        },
        "p95_latency_ms": {
          "type": "integer",
          "minimum": 0
        },
        "error_counts": {
          "type": "object",
          "additionalProperties": { "type": "integer" },
          "description": "Error counts keyed by error code (e.g., {\"TIMEOUT\": 2, \"SCHEMA_VIOLATION\": 0})"
        }
      }
    }
  },
  "required": ["agent_id", "status", "trust_score", "registered_at"]
}
```

---

## 9. RetryPolicy

Per-step retry configuration (design.md Section 8.1).

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/retry-policy.json",
  "title": "RetryPolicy",
  "type": "object",
  "properties": {
    "max_attempts": {
      "type": "integer",
      "minimum": 0,
      "default": 3,
      "description": "Maximum number of retry attempts (0 = no retries)"
    },
    "backoff": {
      "type": "string",
      "enum": ["none", "linear", "exponential"],
      "default": "exponential"
    },
    "base_delay_ms": {
      "type": "integer",
      "minimum": 0,
      "default": 500
    },
    "max_delay_ms": {
      "type": "integer",
      "minimum": 0,
      "default": 10000
    }
  },
  "required": ["max_attempts"]
}
```

---

## 10. Snapshot

State checkpoint for tiered state management (design.md Section 2.5).

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/snapshot.json",
  "title": "Snapshot",
  "type": "object",
  "properties": {
    "summary": {
      "type": "string",
      "description": "LLM-generated summary of the state at checkpoint time"
    },
    "key_state": {
      "type": "object",
      "additionalProperties": true,
      "description": "Essential state variables preserved across pruning"
    },
    "checkpoint_step": {
      "type": "integer",
      "minimum": 0,
      "description": "The step index at which this snapshot was taken"
    },
    "created_at": { "type": "string", "format": "date-time" }
  },
  "required": ["summary", "key_state", "checkpoint_step", "created_at"]
}
```

---

## 11. PlanRecord

The kernel's internal tracking object for a plan segment generated by the Planner (derived from design.md Section 6.2). This schema gives the Plan lifecycle (lifecycles.md Section 4) a concrete data representation.

### Schema

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://gaia-kernel.org/schemas/plan-record.json",
  "title": "PlanRecord",
  "type": "object",
  "properties": {
    "plan_id": { "type": "string", "format": "uuid" },
    "task_id": { "type": "string", "format": "uuid" },
    "status": {
      "type": "string",
      "enum": ["generating", "valid", "rejected", "executing", "completed", "failed", "replanning"],
      "description": "Current lifecycle status of the plan (see lifecycles.md Section 4)"
    },
    "steps": {
      "type": "array",
      "items": { "$ref": "https://gaia-kernel.org/schemas/step.json" }
    },
    "has_more": {
      "type": "boolean",
      "default": false,
      "description": "If true, the planner intends to generate additional steps in a subsequent iteration (incremental planning)"
    },
    "generation": {
      "type": "integer",
      "minimum": 1,
      "default": 1,
      "description": "Plan generation counter. Increments on each replan."
    },
    "created_at": { "type": "string", "format": "date-time" }
  },
  "required": ["plan_id", "task_id", "status", "steps", "has_more", "created_at"]
}
```

---

## Related Documents

* [Lifecycle State Machines](lifecycles.md) — valid status transitions for Task, Step, AgentRecord, and PlanRecord
* [Communication Spec](communication.md) — message flow using Request, Response, and Event schemas
* [Failure Handling Spec](failure-handling.md) — RetryPolicy usage and escalation paths
* [State Management Spec](state-management.md) — Snapshot triggers and tiered storage
* [Error Code Catalog](../reference/error-codes.md) — all error codes with retryability
* [Event Catalog](../reference/event-catalog.md) — all event types with payload definitions

---

## TODO

- [x] Define AgentManifest schema (with invoke, idempotent, scopes)
- [x] Define Task schema (with current_step)
- [x] Define Step schema
- [x] Define Request schema (with type, from)
- [x] Define Response schema (with job_id for async)
- [x] Define Error schema (with UNKNOWN)
- [x] Define Event schema
- [x] Define AgentRecord schema (with full lifecycle enum)
- [x] Define RetryPolicy schema
- [x] Define Snapshot schema
- [x] Define PlanRecord schema (with has_more, status)
- [x] Document errata (intentional divergences from design.md)
