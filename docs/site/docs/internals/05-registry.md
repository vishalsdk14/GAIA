# Chapter 5: The Capability Registry

The **Capability Registry** is GAIA's "Yellow Pages." It is the authoritative source for discovering which agents are connected, what they can do (capabilities), and how "trustworthy" they are.

---

## 5.1 The Agent Manifest

When an agent connects to the Kernel, it must submit an **Agent Manifest** (see [schemas.md](../../specs/schemas.md)). This document is the formal contract between the agent and the Kernel.

### Key Fields:
*   **`agent_id`**: A unique identifier for the agent instance.
*   **`capabilities`**: A list of actions the agent can perform, each with its own `input_schema` and `output_schema`.
*   **`endpoint`**: The URL where the agent can be reached.
*   **`protocol`**: The communication protocol (e.g., `native`, `a2a`, `mcp`).

---

## 5.2 The Handshake Flow

The registration process is a strict 4-step handshake:

1.  **CONNECT**: The agent sends its manifest to the `/api/v1/registry/register` endpoint.
2.  **VALIDATE**: The Registry validates the manifest against GAIA's strict JSON Schemas. It ensures no duplicate capability names exist across different trust levels.
3.  **SANDBOX**: The Kernel assigns a security profile to the agent based on its `auth` credentials.
4.  **ACTIVATE**: The agent is marked as `active` in the registry and becomes eligible for dispatch.

---

## 5.3 Dynamic Capability Routing

In GAIA, the Planner never references an AgentID. It only says: *"I need to run the `get_weather` capability."*

The **Dispatcher** then queries the Registry:
1.  **Filter**: Find all `active` agents that support `get_weather`.
2.  **Score**: Rank agents by their **Trust Score**, **Success Rate**, and **P95 Latency**.
3.  **Select**: Choose the best agent (or use a round-robin strategy for load balancing).

This abstraction allows for **Hot-Swapping**. If an agent crashes mid-task, the Kernel can immediately select a different provider for the same capability for the next retry attempt.

---

## 5.4 The Trust Model

Every agent in the registry has a dynamic **Trust Score** (0.0 to 1.0).

*   **Positive Feedback**: Successful step completions increase the score.
*   **Negative Feedback**: Timeouts, schema violations, or policy denials penalize the score.

### State Transitions:
*   **Active**: Healthy and available for work.
*   **Degraded**: Performance is failing; priority is lowered.
*   **Quarantined**: Agent has returned invalid data; blocked from new work until inspected.
*   **Blacklisted**: Permanent removal due to security or policy violations.

---

## 5.5 Protocol Adapters (Interoperability)

The Registry is protocol-agnostic. It uses **Adapters** to translate external manifests into GAIA's native format:

*   **A2A Adapter**: Fetches the Google `agent.json` and maps it to a GAIA manifest.
*   **MCP Adapter**: Calls `tools/list` on an MCP server and turns each tool into a GAIA capability.

This ensures that GAIA can orchestrate a swarm of agents speaking entirely different languages seamlessly.

---

## 5.6 Related Specifications

*   [Registry Spec](../../specs/registry.md)
*   [Lifecycles Spec (Agent Lifecycle)](../../specs/lifecycles.md)
*   [Transport Spec](../../specs/transport.md)
