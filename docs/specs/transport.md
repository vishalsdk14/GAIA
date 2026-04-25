# Transport Layer Specification

> **Status**: 🟢 Complete
>
> **Source**: [design.md — Sections 3, 11](../design.md)

---

## Purpose

The Transport Layer acts as the physical bridge between the GAIA Kernel and external Agents. This specification defines how abstract `Request` and `Response` objects (defined in schemas.md) are serialized and transmitted over various network and inter-process protocols.

---

## 1. Supported Transport Protocols

The GAIA Kernel is transport-agnostic, supporting multiple protocols to accommodate different agent architectures ranging from lightweight local scripts to massive remote clusters.

### 1.1 HTTP/REST (`transport: "http"`)
* **Wire Format**: JSON payloads over HTTP/1.1 or HTTP/2.
* **Request Mapping**: Kernel sends an `HTTP POST` to the agent's `endpoint`. The `Request` JSON is the request body.
* **Response Mapping**: Agent responds with HTTP 200 OK. The `Response` JSON is the response body.
* **Timeouts**: Enforced via standard HTTP client timeouts. If the connection drops before HTTP 200, the step fails with `TIMEOUT`.

### 1.2 gRPC (`transport: "grpc"`)
* **Wire Format**: Protobuf over HTTP/2.
* **Request/Response Mapping**: The kernel dynamically translates the JSON schema into a generic Protobuf payload:
  ```protobuf
  message InvokeRequest {
    string request_id = 1;
    string task_id = 2;
    string step_id = 3;
    string capability = 4;
    google.protobuf.Struct input = 5;
    string mode = 6;  // "sync" or "async"
    int32 timeout_ms = 7;
  }

  message InvokeResponse {
    string request_id = 1;
    bool success = 2;
    google.protobuf.Struct output = 3;
    Error error = 4;
    string job_id = 5;
  }

  message Error {
    string code = 1;
    string message = 2;
    bool retryable = 3;
  }

  service Agent {
    rpc Invoke (InvokeRequest) returns (InvokeResponse);
  }
  ```
* **Best For**: High-throughput, low-latency remote capabilities.

### 1.3 WebSocket (`transport: "websocket"`)
* **Wire Format**: JSON strings or binary MessagePack over WSS.
* **Mapping**: A persistent bi-directional connection. The Kernel pushes the `Request` frame. The Agent pushes the `Response` frame asynchronously.
* **Connection Lifecycle**: 
  * Requires a Ping/Pong keep-alive every 30 seconds.
  * If the socket closes abruptly, all pending steps assigned to the agent are marked `failed`.

### 1.4 Inter-Process Communication (`transport: "ipc"`)
* **Wire Format**: JSON over Unix Domain Sockets (`unix://...`) or Windows Named Pipes (`\\.\pipe\...`).
* **Usage**: Used strictly for locally deployed agents running on the same host as the kernel.
* **Benefit**: Zero network latency. The Dispatcher applies a `+10%` transport bonus for IPC agents.

---

## 2. Integration Protocols (Dialects)

While the Transport determines *how* bits are moved, the Integration Protocol (`protocol` field in AgentManifest) determines the semantic wrapping.

### 2.1 Native Protocol (`protocol: "native"`)
Agents explicitly designed for GAIA. They speak the exact JSON schemas defined in `schemas.md` natively without translation.

### 2.2 Agent-to-Agent (A2A) (`protocol: "a2a"`)
Standardized external agents. The kernel injects an adapter layer that translates GAIA `Request` schemas into the standard A2A Agent Card invocation format, and translates A2A replies back to GAIA `Response` schemas.

| GAIA | → | A2A |
|:---|:---|:---|
| `Request.input` | → | `parameters` |
| `Response.output` | ← | `result.artifact` |
| `Error.code` (TIMEOUT) | ← | `error_type`: "timeout" |
| `Error.code` (POLICY_DENIED) | ← | `error_type`: "unauthorized" |

*(See external A2A specs for full Agent Card format).*

### 2.3 Model Context Protocol (MCP) (`protocol: "mcp"`)
Used to expose standard LLM tools (e.g., local file system readers, DB executors) to the GAIA kernel.
* The kernel acts as an MCP Client.
* When the kernel routes a step to an MCP agent, it translates the GAIA `Request` into an MCP `CallToolRequest`.
* MCP `CallToolResult` is mapped back to the GAIA `Response.output`.

| GAIA | → | MCP |
|:---|:---|:---|
| `Request.capability` | → | `method`: "tools/call", `name` |
| `Request.input` | → | `arguments` |
| `Response.output` | ← | `content[0].text` |
| `Error.message` | ← | `isError`: true |

---

## 3. Security at the Transport Layer

* **mTLS Requirement**: For `grpc` and `http` transports, Mutual TLS is strongly recommended for remote agents to ensure zero-trust verification.
* **Local Sandboxing**: `ipc` sockets must be secured via standard POSIX file permissions, ensuring only the Kernel process and the Agent process can read/write to the socket.

---

## Related Documents

* [Schemas Spec](schemas.md) — The payloads sent over these transports.
* [Communication Spec](communication.md) — Async delivery flows riding on top of these transports.
* [Registry Spec](registry.md) — Dispatcher scoring bonuses based on transport type.
