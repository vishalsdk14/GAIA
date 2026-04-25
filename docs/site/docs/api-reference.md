# API Reference

This document provides a detailed reference for all GAIA Kernel API endpoints.

## Public API (`/api/v1`)

The Public API is used by orchestrators, dashboards, and end-users to interact with the GAIA ecosystem.

### Task Management

#### `POST /tasks`
Submits a high-level natural language goal to the Kernel.
- **Input**: `{ "goal": string }`
- **Output**: Returns the created `Task` object.
- **Status**: `201 Created`

#### `GET /tasks/{taskID}`
Retrieves the current status and execution plan for a task.
- **Output**: Returns the `Task` object.

### Registry

#### `GET /registry/agents`
Lists all currently active agents in the ecosystem.
- **Output**: `Array<AgentRecord>`

#### `GET /registry/capabilities`
Lists all available capabilities (tools) across the network.
- **Output**: `Array<Capability>`

#### `POST /registry/register`
Registers a new agent or reconnects an existing one.
- **Input**: `AgentManifest`
- **Status**: `201 Created`

#### `DELETE /registry/agents/{agentID}`
Deregisters an agent and unbinds its capabilities.
- **Status**: `204 No Content`

### Streaming

#### `GET /stream` (WebSocket)
Connects to the real-time event stream.
- **Protocol**: WebSocket
- **Output**: Stream of `Event` objects.

---

## Agent API (`/internal/v1`)

The Agent API is specifically for agents to manage their own state.

### Managed State (Tier 4)

#### `GET /state`
Lists all keys stored by the calling agent.
- **Query Params**: `limit`, `offset`
- **Output**: `{ "keys": string[], "total_keys": int, "bytes_used": int }`

#### `GET /state/{key}`
Retrieves the JSON document for a key.

#### `PUT /state/{key}`
Stores or overwrites a JSON document.
- **Input**: Any valid JSON.
- **Status**: `200 OK` or `413 Payload Too Large`.

#### `DELETE /state/{key}`
Removes a key-value pair.
- **Status**: `204 No Content`
