# Orchestrator Design (Control Plane)

## 1. Definition

This system is **not an agent**.
It is a **control plane (orchestration kernel)** that:

* Accepts goals
* Decomposes them into executable steps
* Routes steps to capabilities
* Maintains global state
* Enforces constraints
* Handles failures
* Mediates all communication

> Equivalent to: an execution kernel for goal-directed systems

---

## 2. Core Responsibilities

### 2.1 Goal Management

* Accept structured goals
* Assign task IDs
* Track lifecycle:

  * pending → planning → executing → completed/failed

---

### 2.2 Planning

* Convert goal → structured plan
* Implement via LLM or rules
* Output must be machine-readable (JSON)

**Constraint:** no execution logic here

---

### 2.3 Scheduling

* Determine:

  * execution order
  * dependencies
  * parallelism

Deterministic system logic

---

### 2.4 Execution Control

* Invoke capabilities
* Track:

  * start time
  * outputs
  * failures

Must be deterministic

---

### 2.5 State Management (UPDATED)

#### Tiered State Model

##### Tier 1 — Active State (Hot)

* Small, structured, always loaded
* Used by planner + execution

```json
{
  "key_state": {...},
  "current_step": 3
}
```

---

##### Tier 2 — Task History (Warm)

* Append-only log of:

  * steps
  * results
  * messages
* Stored in DB
* Not passed to planner directly

---

##### Tier 3 — Archived State (Cold)

* Older segments
* Compressed or summarized

---

#### Snapshotting Strategy

Trigger when:

* step count threshold reached OR
* state size exceeds limit

Snapshot format:

```json
{
  "summary": "...",
  "key_state": {...},
  "checkpoint_step": 10
}
```

Post-snapshot:

* prune intermediate state
* retain snapshot + recent delta

---

#### Planner Input Rule

Planner receives only:

* active state
* latest snapshot
* current goal

Never full history.

---

### 2.6 Policy Enforcement

* Permissions
* Constraints
* Safety rules

Examples:

* restrict actions
* require approvals
* enforce limits (time, cost)

---

### 2.7 Failure Handling

Explicit strategies:

* retry
* fallback
* replan
* abort

---

### 2.8 Observability

* structured logs
* execution traces
* metrics (latency, failures)

---

### 2.9 Extensibility Interface

* capability registration
* schema validation
* versioning

---

### 2.10 Communication Control

All inter-agent communication is **mediated by the orchestrator**.

No direct agent-to-agent interaction is allowed.

---

## 3. Communication Model

### 3.1 Principles

* No peer-to-peer communication
* No free-form messages
* All communication is:

  * typed
  * structured
  * traceable

---

### 3.2 Message Types

#### Request (synchronous)

```json
{
  "type": "REQUEST",
  "from": "agent_id",
  "capability": "string",
  "input": {...},
  "task_id": "uuid"
}
```

---

#### Event (asynchronous)

```json
{
  "type": "EVENT",
  "name": "string",
  "payload": {...},
  "task_id": "uuid"
}
```

---

### 3.3 Communication Flow

```
Agent → Orchestrator → Routing Layer → Target Capability → Response → Orchestrator → Agent
```

Events:

```
Agent → Orchestrator → Event Bus → Subscribers
```

---

### 3.4 Communication Components

```
Orchestrator Kernel
├── Request Router
├── Event Bus
├── Policy Engine
├── State Store
└── Audit Log
```

---

### 3.5 Routing Rules

#### Request Routing

* capability → agent selection
* based on:

  * availability
  * policy
  * priority

#### Event Routing

* publish/subscribe model
* filtered by event type

---

### 3.6 Policy Enforcement in Communication

Every message validated for:

* permission
* scope
* safety

---

### 3.7 Traceability

All messages must include:

* `task_id`
* optional `step_id`

---

### 3.8 Anti-Patterns

* direct agent-to-agent calls
* unstructured communication
* uncontrolled subscriptions

---

## 4. Capability & Agent Registration (Handshake)

### 4.1 Agent Manifest (required on attach)

```json
{
  "agent_id": "string",
  "version": "semver",
  "capabilities": [
    {
      "name": "string",
      "input_schema": { "...": "JSON Schema" },
      "output_schema": { "...": "JSON Schema" },
      "constraints": ["read_only | mutates_state | external_io"],
      "timeouts_ms": 15000,
      "idempotent": true
    }
  ],
  "health_endpoint": "string",
  "auth": {
    "type": "mTLS | token",
    "scopes": ["capability:invoke"]
  }
}
```

---

### 4.2 Registration Flow

1. **CONNECT** → submit manifest
2. **VALIDATE** → schemas, constraints, auth
3. **SANDBOX ASSIGNMENT** → runtime limits + network policy
4. **REGISTER** → add to Capability Registry (capability → agent bindings)
5. **READY** → agent eligible for dispatch

---

### 4.3 Capability Registry (authoritative)

Maps capability → [agents]

Stores:
* versions
* health
* trust score
* SLA (latency, success rate)

---

## 5. Trust Layer (Orchestrator as Firewall)

### 5.1 No Peer-to-Peer (enforced)

Agents cannot address each other. All calls must be: `REQUEST(capability, input, task_id)`

---

### 5.2 Validation Pipeline (every request)

1. Agent → Orchestrator
2. Auth check (identity, scopes)
3. Schema validation (input)
4. Policy check (permissions, side-effects)
5. Routing decision
6. Invoke target agent
7. Schema validation (output)
8. Return / emit events

---

### 5.3 Data Minimization

* Agents receive only step-local input
* No global state leakage unless explicitly allowed by policy

---

## 6. Planning & Dispatching (UPDATED)

### 6.1 Capability-Level Planning

Each planning call includes:
* **Goal**: The objective to decompose.
* **Active State**: Latest snapshot + current delta.
* **Capability Manifest**: Curated list of currently attached agents only.

Planner operates on capability abstraction and never references agent IDs.

---

### 6.2 Incremental Planning

Planner generates **partial plans** (1–3 steps).

```json
{
  "steps": [
    { "capability": "read_pdf", "input": {...} },
    { "capability": "summarize_text", "input": {...} }
  ],
  "has_more": false
}
```

---

### 6.3 Routing (Dispatcher logic)

For each step, select agent by:
* capability match
* health score
* latency/SLA
* policy constraints

**Fallback chain**: primary → secondary → replan

---

## 7. Internal Architecture

```
Orchestrator Kernel
├── Goal Manager
├── Planner Interface
├── Scheduler
├── Execution Engine
├── State Store
├── Policy Engine
├── Capability Registry
├── Request Router
├── Event Bus
└── Audit Log
```

---

## 8. Data Model

### 8.1 Request (enforced)

```json
{
  "type": "REQUEST",
  "capability": "string",
  "input": {...},
  "task_id": "uuid",
  "step_id": "uuid"
}
```

---

### 8.2 Step

```json
{
  "step_id": "uuid",
  "capability": "string",
  "input": {...},
  "status": "pending | done | failed"
}
```

---

### 8.3 Response (strict)

```json
{
  "success": true,
  "output": {...},
  "error": null,
  "metrics": {
    "latency_ms": 1200
  }
}
```

---

## 9. Control Loop (with dynamic attachment)

```
on agent_connect(manifest):
    validate → sandbox → register

while task not complete:

    if no pending steps:
        plan = planner(goal, state, capabilities)

    step = next_step(plan)

    if violates_policy(step):
        halt / require approval

    agent = route(step.capability)
    result = execute(agent, step.input)

    if success:
        validate_output()
        update_state()
    else:
        classify_failure()
        apply_enforcement() # quarantine if needed
        retry / fallback / replan

    continue
```

---

## 10. Isolation & “Ejection” (Dirty Agent Handling)

### 10.1 Failure Taxonomy

* **Soft failure**: timeout, transient error
* **Hard failure**: schema violation, malformed output
* **Policy violation**: unauthorized action attempt

---

### 10.2 Enforcement Actions

| Condition | Action |
| :--- | :--- |
| Repeated timeouts | degrade priority |
| Schema violations | immediate quarantine |
| Policy violation | blacklist |
| Crash/health down | temporary eject |

---

### 10.3 Quarantine / Blacklist Model

```json
{
  "agent_id": "...",
  "status": "active | degraded | quarantined | blacklisted",
  "reason": "schema_violation",
  "since": "timestamp"
}
```

---

### 10.4 Health Monitoring

* Periodic **health_endpoint** checks.
* Rolling metrics: success rate, **p95 latency**, and error types.
* Used as the primary signal for routing and dispatcher decisions.

---

## 11. Security Model

* **Identity**: per-agent credentials (mTLS or signed tokens)
* **Scopes**: per-capability invocation rights
* **Rate limits**: per-agent + per-capability
* **Network policy**: sandboxed egress (deny by default)

---

## 12. Non-Negotiables (for stability)

* Capability-first (never agent-first) planning
* Strict schema validation (in/out)
* Centralized mediation (no bypass)
* Tiered trust (active → degraded → quarantined → blacklisted)
* Bounded planner context (manifest is curated)

**Bottom line**: To make “plug-in agents” viable, the orchestrator must behave like a **capability router + policy firewall + execution kernel**. If the handshake is strict, validation is hard, and isolation is automatic, system integrity is maintained regardless of which agents attach.

---

## 13. Design Constraints

### Deterministic Execution

* Same input → same result

---

### LLM Isolation

* LLM only plans/replans

---

### Strict Schemas

* Validate all inputs/outputs

---

### Centralized Communication

* All interactions mediated

---

## 14. Conceptual Equivalents

* OS kernel
* Workflow engines (Temporal, Airflow)
* Distributed control systems

---

## 15. Validation Criteria

System is correct if:

* tasks can be paused/resumed
* execution is replayable
* messages are traceable
* planner is replaceable

---

## Summary

You are building:

> a deterministic orchestration kernel with a probabilistic planner, mediated communication, and bounded state

System stability depends on:

* strict contracts
* centralized control
* controlled planning
* observable execution

---

# Orchestrator Addendum: Execution Gaps + Local/Remote Strategy

This document defines **missing connectors** required for implementation:

* Invocation model
* Data flow between steps
* Async execution
* Cancellation & interrupts
* Client boundary
* Local vs Remote execution strategy

---

## 1. Invocation Model

### 1.1 Agent Manifest (Updated)

```json
{
  "agent_id": "string",
  "base_url": "https://agent-host",
  "transport": "http | grpc | local",
  "invoke": {
    "timeout_ms": 15000,
    "async_supported": true
  },
  "capabilities": [...]
}
```

---

### 1.2 Invocation Interface

#### Unified call

```text
invoke(agent, payload)
```

#### Payload

```json
{
  "task_id": "uuid",
  "step_id": "uuid",
  "capability": "string",
  "input": {...},
  "mode": "sync | async"
}
```

---

### 1.3 Transport Resolution

```text
if agent.transport == "local":
    direct function / IPC call
else:
    HTTP/gRPC request
```

---

## 2. Data Flow Between Steps

### 2.1 Step Binding Syntax

Planner uses explicit references:

```json
{
  "steps": [
    {
      "id": "step_1",
      "capability": "read_pdf",
      "input": {"url": "..."}
    },
    {
      "id": "step_2",
      "capability": "summarize",
      "input": {
        "text": "{{step_1.output.text}}"
      }
    }
  ]
}
```

---

### 2.2 Interpolation Engine

Before execution:

* resolve `{{...}}`

#### Sources

1. previous step outputs
2. active state
3. constants

---

### 2.3 Rules

* Only completed steps can be referenced
* Invalid references → plan rejection

---

## 3. Async Execution Model

### 3.1 Modes

| Mode  | Behavior                         |
| ----- | -------------------------------- |
| sync  | blocking response                |
| async | immediate ACK + later completion |

---

### 3.2 Async Flow

#### Invocation

```json
{
  "mode": "async"
}
```

#### Agent ACK

```json
{
  "status": "accepted",
  "job_id": "xyz"
}
```

#### Completion Event

```json
{
  "type": "STEP_COMPLETED",
  "task_id": "...",
  "step_id": "...",
  "output": {...}
}
```

---

### 3.3 Orchestrator Handling

* mark step `pending_async`
* wait for completion event
* enforce timeout

---

## 4. Cancellation & Interrupts

### 4.1 Task State

```json
{
  "status": "running | cancelled | failed | completed"
}
```

---

### 4.2 Interrupt Event

```json
{
  "type": "INTERRUPT",
  "task_id": "...",
  "reason": "user_cancel | system_shutdown"
}
```

---

### 4.3 Control Loop Rule

```text
if task.status == "cancelled":
    abort execution
```

---

### 4.4 Agent Cancellation (Optional)

```json
{
  "type": "CANCEL",
  "task_id": "...",
  "step_id": "..."
}
```

---

## 5. Client Interface

### 5.1 Submit Goal

```http
POST /tasks
```

```json
{
  "goal": {...}
}
```

---

### 5.2 Get Status

```http
GET /tasks/{task_id}
```

---

### 5.3 Cancel Task

```http
POST /tasks/{task_id}/cancel
```

---

### 5.4 Optional Streaming

* WebSocket / SSE for real-time updates

---

## 6. Local vs Remote Execution

### 6.1 Agent Classification

```json
{
  "agent_id": "...",
  "location": "local | remote",
  "transport": "local | http | grpc",
  "avg_latency_ms": 20,
  "compute_class": "light | heavy"
}
```

---

### 6.2 Routing Strategy

#### Prefer Local Agents When:

* low-latency required
* short execution
* high interaction frequency

#### Prefer Remote Agents When:

* long-running tasks
* heavy computation
* async workloads

---

### 6.3 Dispatcher Logic

```text
select_agent(capability):
    if low_latency_required:
        choose local
    else if heavy_task:
        choose remote
    else:
        choose best available (latency + health)
```

---

### 6.4 Transport Abstraction

Execution must be uniform:

```text
invoke(agent, payload):
    route via transport layer
```

* local → direct call / IPC
* remote → HTTP/gRPC

---

### 6.5 Failure Characteristics

| Type               | Local   | Remote   |
| ------------------ | ------- | -------- |
| latency            | low     | higher   |
| scaling            | limited | scalable |
| failure isolation  | low     | high     |
| network dependency | none    | yes      |

---

### 6.6 Hybrid Model

```text
Orchestrator
    ├── Local Agents (fast path)
    └── Remote Agents (scalable path)
```

---

## 7. Integrated Control Loop

```text
while task not complete:

    if task.status == "cancelled":
        abort
        break

    if no pending steps:
        plan = planner(...)

    step = next_step(plan)

    resolve_input(step)

    agent = route(step.capability)

    result = invoke(agent, step)

    if async:
        wait_for_event()
    else:
        process_result()

    if failure:
        retry / fallback / replan
```

---

## Summary

This addendum defines:

* explicit invocation contract
* deterministic data flow between steps
* async execution model
* cancellation mechanism
* client interface boundary
* hybrid local/remote execution strategy

System is now:

* executable
* transport-agnostic
* scalable
* interruption-safe
