# Chapter 9: Security & Trust Models

Security in GAIA is not a bolt-on feature—it is baked into the Kernel's core execution loop. This chapter explores how GAIA manages identity, authorization, and the "Trust Score" system.

---

## 9.1 Zero-Trust Architecture

GAIA operates on a **Zero-Trust** model. Every component—even internal ones—must prove its identity and authorization for every single action.

### The Identity Stack:
1.  **Agent Identity**: Verified via **mTLS certificates** or **Signed JWTs** during the handshake.
2.  **Capability Scopes**: Fine-grained permissions that define which specific actions an agent can perform (e.g., `calendar:read` vs `calendar:write`).
3.  **Task Context**: Every request is bound to a `TaskID`. An agent cannot access data or state belonging to a different task.

---

## 9.2 Tiered Trust Scores

The **Capability Registry** maintains a dynamic **Trust Score** (0.0 to 1.0) for every registered agent.

### How it is calculated:
*   **Base Score**: Starts at 0.5 for new agents.
*   **Performance Factor**: Calculated based on the agent's success rate and P95 latency.
*   **Correctness Factor**: Penalized heavily for **Schema Violations** or **Policy Denials**.
*   **Audit Factor**: Bonus for agents that provide high-fidelity internal logs (A2A artifacts).

### Ejection Thresholds:
*   **Score > 0.8**: Preferred for critical steps.
*   **Score < 0.3**: Agent is moved to `degraded` status.
*   **Score < 0.1**: Agent is automatically `quarantined`.

---

## 9.3 Cryptographic Audit Chaining (Tier 5)

The **Audit Log** is the definitive record of everything that happened in the system. To make it tamper-proof, GAIA uses **SHA-256 Chaining**:

1.  Each log entry contains a JSON blob of the event.
2.  The entry includes a `hash` of its own contents.
3.  The entry includes a `prev_hash` of the previous entry in the log.

This creates a linear chain of events. If an attacker modifies an old log entry to hide a malicious action, the `prev_hash` of the next entry will no longer match, instantly alerting the Kernel's security monitoring subsystem.

---

## 9.4 Sandbox Enforcement

GAIA acts as the "Sandbox" for all attached agents. Through the **Transport Layer**, the Kernel can enforce:
*   **Network Isolation**: Preventing agents from calling external APIs unless whitelisted.
*   **Memory Isolation**: Using **Tier 4 Managed State** to ensure agents cannot "scrape" each other's data.
*   **Timeouts**: Hard execution limits to prevent "Resource Exhaustion" attacks.

---

## 9.5 The Policy Firewall

All security rules are enforced by the **CEL Policy Engine** (see Chapter 4). This allows security teams to inject global safety rules—such as *"No agent can access user_billing_info"*—without modifying a single line of agent code.

---

## 9.6 Related Specifications

*   [Security Spec](../../specs/security.md)
*   [Registry Spec](../../specs/registry.md)
*   [Transport Spec](../../specs/transport.md)
