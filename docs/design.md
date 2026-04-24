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

  * pending → planning → executing → completed/failed/cancelled

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
  * dependencies (DAG)
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

### 2.5 State Management

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

#### Request (synchronous or async)

```json
{
  "type": "REQUEST",
  "from": "agent_id",
  "capability": "string",
  "input": {...},
  "task_id": "uuid",
  "step_id": "uuid",
  "mode": "sync | async"
}
```

---

#### Response (strict)

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

Must conform to `output_schema`.
Non-conforming → hard failure.

---

#### Event (asynchronous)

```json
{
  "type": "EVENT",
  "name": "string",
  "payload": {...},
  "task_id": "uuid",
  "step_id": "uuid"
}
```

---

#### Error Object

When `success` is false, the `error` field must contain:

```json
{
  "error": {
    "code": "SCHEMA_VIOLATION | TIMEOUT | POLICY_DENIED | CAPABILITY_NOT_FOUND | AGENT_UNAVAILABLE | EXECUTION_FAILED | INTERNAL_ERROR | UNKNOWN",
    "message": "human-readable description",
    "retryable": true,
    "details": {}
  }
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

### 3.6 Event Catalog

#### System Events (emitted by Kernel)

| Event Name       | When                                    |
| :--------------- | :-------------------------------------- |
| STEP_COMPLETED   | step finished successfully              |
| STEP_FAILED      | step failed after exhausting retries    |
| TASK_COMPLETED   | all steps done                          |
| TASK_FAILED      | unrecoverable failure                   |
| TASK_CANCELLED   | user or system cancelled                |
| AGENT_REGISTERED | new agent attached                      |
| AGENT_EJECTED    | agent quarantined/blacklisted           |
| PLAN_GENERATED   | planner produced new steps              |

#### Subscribers

* **Client**: subscribes via WebSocket/SSE to task-level events.
* **Audit Log**: subscribes to all events (mandatory).
* **Agents**: cannot subscribe to events. They only respond to requests.

---

### 3.7 Policy Enforcement in Communication

Every message validated for:

* permission
* scope
* safety

---

### 3.8 Traceability

All messages must include:

* `task_id`
* `step_id`

---

### 3.9 Anti-Patterns

* direct agent-to-agent calls
* unstructured communication
* uncontrolled subscriptions

---

## 4. Capability & Agent Registration (Handshake)

### 4.1 Agent Manifest (canonical, required on attach)

```json
{
  "agent_id": "string",
  "version": "semver",
  "endpoint": "https://agent-host",
  "transport": "http | grpc | ipc | websocket",
  "protocol": "native | a2a | mcp",
  "invoke": {
    "timeout_ms": 15000,
    "async_supported": true
  },
  "capabilities": [
    {
      "name": "string",
      "description": "what this capability does",
      "input_schema": { "...": "JSON Schema" },
      "output_schema": { "...": "JSON Schema" },
      "constraints": {
        "read_only": true,
        "mutates_state": false,
        "external_io": false
      },
      "idempotent": true
    }
  ],
  "health_endpoint": "/health",
  "auth": {
    "type": "none | bearer | mTLS | oauth | api_key",
    "scopes": ["capability:invoke"]
  }
}
```

**Protocol values:**
* `native` — GAIA's internal protocol (default).
* `a2a` — Google Agent-to-Agent protocol. Manifest auto-derived from Agent Card.
* `mcp` — Anthropic Model Context Protocol. Capabilities derived from MCP Tool definitions.

---

### 4.2 Registration Flow

1. **CONNECT** → submit manifest
2. **VALIDATE** → schemas, constraints, auth
3. **SANDBOX ASSIGNMENT** → runtime limits + network policy
4. **REGISTER** → add to Capability Registry (capability → agent bindings)
5. **READY** → agent eligible for dispatch

---

### 4.3 Disconnect / Graceful Detach

1. **DRAIN** → agent signals intent to disconnect.
2. **REASSIGN** → kernel reassigns any in-flight steps to fallback agents.
3. **DEREGISTER** → remove agent from Capability Registry.
4. **CLOSED** → agent is fully detached.

If an agent disconnects without signaling (crash), the kernel detects via failed health checks and triggers **temporary eject** (see Section 10).

---

### 4.4 Capability Registry (authoritative)

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

## 6. Planning & Dispatching

### 6.1 Planner Inputs

Each planning call includes:
* **Goal**: The objective to decompose.
* **Active State**: Latest snapshot + current delta.
* **Capability Manifest**: Curated list of currently attached capabilities only.

Planner operates on capability abstraction and never references agent IDs.

---

### 6.2 Incremental Planning

Planner generates **partial plans** (1–3 steps).

```json
{
  "steps": [
    {
      "id": "step_1",
      "capability": "read_pdf",
      "input": { "url": "..." },
      "depends_on": []
    },
    {
      "id": "step_2",
      "capability": "summarize_text",
      "input": { "text": "{{step_1.output.text}}" },
      "depends_on": ["step_1"]
    }
  ],
  "has_more": false
}
```

---

### 6.3 Parallel Execution (DAG)

Steps declare dependencies via `depends_on`.

* Steps with **no unmet dependencies** can run in parallel.
* The Scheduler builds a DAG from `depends_on` and dispatches ready steps concurrently.
* A step is "ready" when all entries in its `depends_on` list are `done`.

Example (parallel):

```json
{
  "steps": [
    { "id": "a", "capability": "fetch_weather", "depends_on": [] },
    { "id": "b", "capability": "fetch_traffic", "depends_on": [] },
    { "id": "c", "capability": "plan_route", "depends_on": ["a", "b"] }
  ]
}
```

Steps `a` and `b` execute in parallel. Step `c` waits for both.

---

### 6.4 Data Flow Between Steps (Interpolation)

Planner uses explicit `{{...}}` references to bind step outputs.

#### Interpolation Sources (priority order)

1. Previous step outputs (`{{step_id.output.field}}`)
2. Active state (`{{state.field}}`)
3. Constants (`{{const.field}}`)

#### Rules

* Only `done` steps can be referenced.
* Invalid or circular references → plan rejection.
* Interpolation is resolved by the Kernel **before** dispatching to the agent.

---

### 6.5 Routing (Dispatcher logic)

For each step, select agent by:
* capability match
* health score
* latency/SLA
* policy constraints

**Fallback chain**: primary → secondary → replan

---

## 7. Invocation Model

### 7.1 Unified Call

```text
invoke(agent, payload)
```

### 7.2 Payload

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

### 7.3 Transport Resolution

```text
if agent.transport == "ipc":
    direct function / IPC call
else:
    HTTP/gRPC/WebSocket request to agent.endpoint
```

---

### 7.4 Async Execution

| Mode  | Behavior                         |
| ----- | -------------------------------- |
| sync  | blocking response                |
| async | immediate ACK + later completion |

#### Async Flow

Agent ACK:

```json
{
  "status": "accepted",
  "job_id": "xyz"
}
```

Completion (agent → Orchestrator):

```json
{
  "type": "STEP_COMPLETED",
  "task_id": "...",
  "step_id": "...",
  "output": {...}
}
```

Orchestrator handling:

* mark step `pending_async`
* wait for completion event
* enforce timeout (from manifest `invoke.timeout_ms`)

---

## 8. Retry & Failure Policy

### 8.1 Retry Configuration (per-step)

```json
{
  "retry": {
    "max_attempts": 3,
    "backoff": "exponential",
    "base_delay_ms": 500,
    "max_delay_ms": 10000
  }
}
```

---

### 8.2 Retry Rules

* Only retryable errors (where `error.retryable == true`) trigger retries.
* Idempotent capabilities (from manifest) are always safe to retry.
* Non-idempotent capabilities with `mutates_state` constraint require explicit retry policy; default is **no retry**.

---

### 8.3 Escalation Path

```text
retry (up to max_attempts)
  → fallback agent (if available)
    → replan (invoke planner with failure context)
      → abort task
```

---

## 9. Data Model

### 9.1 Task

```json
{
  "task_id": "uuid",
  "goal": {...},
  "status": "pending | planning | executing | completed | failed | cancelled",
  "plan": [...],
  "current_step": 2,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

---

### 9.2 Step

```json
{
  "step_id": "uuid",
  "capability": "string",
  "input": {...},
  "depends_on": ["step_id", ...],
  "status": "pending | running | pending_async | done | failed",
  "output": {...},
  "error": null,
  "assigned_agent": "agent_id",
  "retry_count": 0
}
```

---

### 9.3 Agent Record

```json
{
  "agent_id": "...",
  "status": "active | degraded | quarantined | blacklisted",
  "trust_score": 0.95,
  "registered_at": "timestamp",
  "last_health_check": "timestamp",
  "rolling_metrics": {
    "success_rate": 0.98,
    "p95_latency_ms": 450,
    "error_counts": { "TIMEOUT": 2, "SCHEMA_VIOLATION": 0 }
  }
}
```

---

## 10. Isolation & "Ejection" (Dirty Agent Handling)

### 10.1 Failure Taxonomy

* **Soft failure**: timeout, transient error
* **Hard failure**: schema violation, malformed output
* **Policy violation**: unauthorized action attempt

---

### 10.2 Enforcement Actions

| Condition          | Action               |
| :----------------- | :------------------- |
| Repeated timeouts  | degrade priority     |
| Schema violations  | immediate quarantine |
| Policy violation   | blacklist            |
| Crash/health down  | temporary eject      |

---

### 10.3 Quarantine / Blacklist Model

* **quarantined**: not used for routing, still observable for debugging.
* **blacklisted**: fully blocked from the kernel.

---

### 10.4 Health Monitoring

* Periodic **health_endpoint** checks.
* Rolling metrics: success rate, **p95 latency**, and error types.
* Used as the primary signal for routing and dispatcher decisions.

---

## 11. Cancellation & Interrupts

### 11.1 Task States

```json
{
  "status": "pending | planning | executing | completed | failed | cancelled"
}
```

---

### 11.2 Interrupt Event

```json
{
  "type": "INTERRUPT",
  "task_id": "...",
  "reason": "user_cancel | system_shutdown"
}
```

---

### 11.3 Control Loop Cancellation

```text
if task.status == "cancelled":
    send CANCEL to in-flight agents (best-effort)
    abort execution
```

---

### 11.4 Agent Cancellation (best-effort)

```json
{
  "type": "CANCEL",
  "task_id": "...",
  "step_id": "..."
}
```

Agent may ignore this. Kernel does not wait for confirmation.

---

## 12. Planner Failure Handling

### 12.1 Failure Modes

* **LLM timeout / rate limit**: transient, retryable.
* **Malformed output**: planner returns non-JSON or invalid plan schema.
* **Empty plan**: planner returns zero steps.
* **Hallucinated capability**: planner references a capability not in the manifest.

---

### 12.2 Recovery Strategy

```text
planner_result = planner(goal, state, capabilities)

if timeout or rate_limit:
    retry with backoff (max 3 attempts)

if malformed output:
    retry once with stricter prompt

if empty plan:
    emit TASK_FAILED("planner returned empty plan")

if unknown capability referenced:
    reject plan, retry with filtered manifest

if all retries exhausted:
    emit TASK_FAILED("planner unavailable")
```

---

## 13. Security Model

* **Identity**: per-agent credentials (mTLS or signed tokens)
* **Scopes**: per-capability invocation rights
* **Rate limits**: per-agent + per-capability
* **Network policy**: sandboxed egress (deny by default)

---

## 14. Client Interface

### 14.1 Submit Goal

```http
POST /tasks
```

```json
{
  "goal": {...}
}
```

---

### 14.2 Get Status

```http
GET /tasks/{task_id}
```

---

### 14.3 Cancel Task

```http
POST /tasks/{task_id}/cancel
```

---

### 14.4 Streaming

* WebSocket / SSE for real-time task events (subscribes to Event Bus).

---

## 15. Local vs Remote Execution

### 15.1 Agent Classification

```json
{
  "agent_id": "...",
  "location": "local | remote",
  "transport": "ipc | http | grpc | websocket",
  "avg_latency_ms": 20,
  "compute_class": "light | heavy"
}
```

---

### 15.2 Routing Strategy

#### Prefer Local Agents When:

* low-latency required
* short execution
* high interaction frequency

#### Prefer Remote Agents When:

* long-running tasks
* heavy computation
* async workloads

---

### 15.3 Transport Abstraction

Execution must be uniform:

```text
invoke(agent, payload):
    route via transport layer
```

* local → direct call / IPC
* remote → HTTP/gRPC

---

### 15.4 Failure Characteristics

| Type               | Local   | Remote   |
| ------------------ | ------- | -------- |
| latency            | low     | higher   |
| scaling            | limited | scalable |
| failure isolation  | low     | high     |
| network dependency | none    | yes      |

---

### 15.5 Hybrid Model

```text
Orchestrator
    ├── Local Agents (fast path)
    └── Remote Agents (scalable path)
```

---

## 16. Authoritative Control Loop

This is the **single, canonical** execution loop for the kernel.

```text
on agent_connect(manifest):
    validate → sandbox → register

on agent_disconnect(agent_id):
    drain → reassign in-flight → deregister

while task not complete:

    if task.status == "cancelled":
        cancel in-flight agents (best-effort)
        break

    if no pending steps:
        plan = planner(goal, state, capabilities)

        if planner fails:
            apply planner recovery (Section 12)
            if unrecoverable: fail task, break

    ready_steps = get_ready_steps(plan)  # steps with all depends_on met

    for each step in ready_steps (parallel):

        resolve_input(step)              # interpolate {{...}} references

        if violates_policy(step):
            halt / require approval

        agent = route(step.capability)   # select by health, SLA, trust
        result = invoke(agent, step)

        if mode == async:
            mark step pending_async
            continue                     # proceed to next ready step

        if success:
            validate_output(result)      # check against output_schema
            update_state(result)
            emit_event("STEP_COMPLETED")
        else:
            classify_failure(result.error)
            apply_enforcement(agent)     # degrade / quarantine / blacklist
            apply_retry_policy(step)     # retry → fallback → replan → abort

    await any pending_async completions

    continue
```

---

## 17. Internal Architecture

```
Orchestrator Kernel
├── Goal Manager
├── Planner Interface
├── Scheduler (DAG resolver)
├── Interpolation Engine
├── Execution Engine
├── Transport Layer (local / HTTP / gRPC)
├── State Store (tiered)
├── Policy Engine
├── Capability Registry
├── Request Router
├── Event Bus
├── Retry Manager
└── Audit Log
```

---

## 18. Design Constraints

### Deterministic Kernel Execution

* Given the same plan and the same agent responses, the kernel produces the same state transitions.
* **Note**: Agent *routing* is non-deterministic (depends on health, latency). The kernel's *processing* of results is deterministic.

---

### LLM Isolation

* LLM only plans/replans
* Never executes actions
* Never receives raw agent outputs

---

### Strict Schemas

* Validate all inputs/outputs against JSON Schema

---

### Centralized Communication

* All interactions mediated
* No bypass paths

---

## 19. Non-Negotiables (for stability)

* Capability-first (never agent-first) planning
* Strict schema validation (in/out)
* Centralized mediation (no bypass)
* Tiered trust (active → degraded → quarantined → blacklisted)
* Bounded planner context (manifest is curated)

**Bottom line**: To make "plug-in agents" viable, the orchestrator must behave like a **capability router + policy firewall + execution kernel**. If the handshake is strict, validation is hard, and isolation is automatic, system integrity is maintained regardless of which agents attach.

---

## 20. Validation Criteria

System is correct if:

* tasks can be paused/resumed
* execution is replayable (given same plan + agent responses)
* every message is traceable
* planner is replaceable
* agents can crash without corrupting kernel state
* cancelled tasks release all resources

---

## 21. Protocol Interoperability (A2A + MCP)

GAIA does not replace A2A or MCP. It **consumes** them. Any A2A-compatible agent or MCP-compatible tool can attach to GAIA through protocol adapters in the Transport Layer.

---

### 21.1 Protocol Positioning

```
┌─────────────────────────────────────────────┐
│              GAIA Kernel                    │
│  (orchestration, planning, trust, state)    │
├─────────────┬───────────────┬───────────────┤
│  A2A        │  MCP          │  Native       │
│  Adapter    │  Adapter      │  Protocol     │
├─────────────┼───────────────┼───────────────┤
│  Remote     │  Tool         │  Local/Remote │
│  Agents     │  Servers      │  Agents       │
└─────────────┴───────────────┴───────────────┘
```

* **A2A** = agent-to-agent communication (horizontal). GAIA uses it to delegate tasks to remote agents.
* **MCP** = agent-to-tool connectivity (vertical). GAIA uses it to access external tools/data sources.
* **Native** = GAIA's own protocol for agents built specifically for this kernel.

---

### 21.2 A2A Integration (Google Agent-to-Agent Protocol)

#### What A2A provides

* **Agent Card** (`/.well-known/agent.json`): Discovery metadata — name, skills, auth, supported modalities.
* **Task lifecycle**: `submitted` → `working` → `input-required` → `completed` / `failed`.
* **Artifacts**: Structured outputs (text, files, data) produced by agents.
* **Transport**: JSON-RPC 2.0 over HTTPS. Streaming via SSE.

#### How GAIA consumes A2A

1. **Discovery**: GAIA fetches `/.well-known/agent.json` from the remote agent's URL.
2. **Manifest Translation**: The A2A Adapter converts the Agent Card into a GAIA Agent Manifest:

```text
A2A Agent Card              →  GAIA Manifest
─────────────────────────────────────────────
name + description          →  agent_id
skills[].id                 →  capabilities[].name
skills[].description        →  capabilities[].description
skills[].inputModes         →  capabilities[].input_schema
skills[].outputModes        →  capabilities[].output_schema
authentication              →  auth
url                         →  base_url
```

3. **Invocation**: When dispatching to an A2A agent, GAIA sends `message/send` (sync) or `message/stream` (async) via JSON-RPC.
4. **Task Mapping**:

| GAIA Step Status   | A2A Task Status  |
| :----------------- | :--------------- |
| pending            | submitted        |
| running            | working          |
| pending_async      | working          |
| done               | completed        |
| failed             | failed           |

5. **Output**: A2A Artifacts are converted to GAIA step outputs. Parts (TextPart, FilePart, DataPart) are normalized into the step's `output` JSON.

---

### 21.3 MCP Integration (Anthropic Model Context Protocol)

#### What MCP provides

* **Tools**: Functions that an AI model can invoke (with JSON Schema for inputs/outputs).
* **Resources**: Contextual data (files, database rows) exposed to the model.
* **Prompts**: Templated messages and workflows.
* **Transport**: JSON-RPC 2.0 over stdio or HTTP+SSE. Stateful connections.

#### How GAIA consumes MCP

1. **Connection**: GAIA connects to MCP Servers as an MCP Client.
2. **Tool Discovery**: GAIA calls `tools/list` to discover available tools.
3. **Manifest Translation**: Each MCP Tool becomes a GAIA capability:

```text
MCP Tool                    →  GAIA Capability
─────────────────────────────────────────────
tool.name                   →  capability.name
tool.description            →  capability.description
tool.inputSchema            →  capability.input_schema
(inferred)                  →  capability.output_schema
tool.annotations.readOnly   →  constraint: read_only
tool.annotations.destructive→  constraint: mutates_state
```

4. **Invocation**: When dispatching to an MCP tool, GAIA sends `tools/call` via JSON-RPC.
5. **Output**: MCP tool results (content array with text/image/resource types) are normalized into GAIA step output JSON.
6. **Resources**: MCP Resources can be injected into step inputs via the Interpolation Engine using `{{mcp.resource_uri}}`.

---

### 21.4 Adapter Architecture

Adapters are internal Kernel components. They sit in the Transport Layer.

```
Execution Engine
    │
    ├── invoke(agent, payload)
    │
    └── Transport Layer
            ├── NativeAdapter   → direct call / IPC / HTTP
            ├── A2AAdapter      → JSON-RPC (message/send, tasks/get)
            └── MCPAdapter      → JSON-RPC (tools/call, resources/read)
```

#### Adapter responsibilities

* **Translate** GAIA payloads into protocol-specific wire formats.
* **Normalize** protocol-specific responses into GAIA's Response schema.
* **Handle** protocol-specific auth (OAuth for A2A, stdio/token for MCP).
* **Map** protocol-specific errors to GAIA's Error schema.

---

### 21.5 Registration Flow (per protocol)

| Protocol | Discovery                          | Registration                          |
| :------- | :--------------------------------- | :------------------------------------ |
| Native   | Agent submits manifest directly    | Standard handshake (Section 4.2)      |
| A2A      | Fetch `/.well-known/agent.json`    | Auto-translate Agent Card → manifest  |
| MCP      | Connect + call `tools/list`        | Auto-translate Tools → capabilities   |

All protocols converge into the same **Capability Registry**. The Planner never knows which protocol an agent uses.

---

### 21.6 What GAIA adds on top of A2A + MCP

Neither A2A nor MCP provides:

* **Autonomous planning**: Neither protocol has a planner. GAIA decomposes goals into steps automatically.
* **Trust & isolation**: A2A assumes agents are trusted peers. GAIA enforces tiered trust (active → quarantined → blacklisted).
* **Cross-protocol orchestration**: A2A agents and MCP tools can be used in the same plan, in the same task, seamlessly.
* **Centralized mediation**: A2A allows peer-to-peer. GAIA forbids it — all traffic is mediated.
* **State management**: Neither protocol manages global task state. GAIA provides tiered state with snapshotting.

---

## Summary

You are building:

> a deterministic orchestration kernel with a probabilistic planner, mediated communication, and bounded state — compatible with the A2A and MCP ecosystems

System is now:

* executable
* transport-agnostic
* scalable
* interruption-safe
* interoperable with industry-standard agent protocols

System stability depends on:

* strict contracts
* centralized control
* controlled planning
* observable execution
