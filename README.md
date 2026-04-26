<div align="center">
  <img src="https://raw.githubusercontent.com/vishalsdk14/GAIA/main/docs/assets/logo.png" width="120" height="120" alt="GAIA Logo" />
  <h1>GAIA</h1>
  <p><b>The Runtime for Building Reliable AI Agent Systems.</b></p>

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status: Production Ready](https://img.shields.io/badge/Status-Phase%2011%20Complete-green.svg)](#roadmap)
[![Protocol: A2A](https://img.shields.io/badge/Protocol-Google%20A2A-4285F4.svg)](https://github.com/google/A2A)
[![Protocol: MCP](https://img.shields.io/badge/Protocol-Anthropic%20MCP-D97706.svg)](https://modelcontextprotocol.io/)

---

### GAIA is to AI Agents what an OS Kernel is to Processes.

It provides the stable ground, safety boundaries, and execution reliability that LLMs lack natively.

[Quickstart](#-quickstart) · [Examples](#-minimal-example) · [Use Cases](#-use-cases) · [Design Spec](docs/design.md)

</div>

---

## ⚡ What GAIA Lets You Do

*   **Build Resilient Swarms**: Automatically recover from agent timeouts or failures with a 4-tier escalation path (Retry → Fallback → Replan → Abort).
*   **Run Untrusted Agents Safely**: Execute agents in isolated sandboxes with a "Deny-by-Default" policy firewall. No shared state chaos.
*   **Connect Anything**: Seamlessly orchestrate OpenAI, Anthropic, and Google agents in a single unified pipeline via A2A and MCP protocols.
*   **Debug with Precision**: Every action is cryptographically signed and tracked in a tamper-proof audit log.

---

## 🚀 Quickstart

Get the GAIA Kernel running in less than 60 seconds.

### 1. Start the Kernel
```bash
# Clone the repository
git clone https://github.com/vishalsdk14/GAIA.git && cd GAIA

# Create a .env file with your LLM configuration (Local or Cloud)
cat <<EOF > src/kernel/.env
GAIA_AUDIT_SECRET=$(openssl rand -hex 32)
GAIA_PLANNER_PROVIDER="local"
GAIA_PLANNER_MODEL="llama3"
EOF

# Start the kernel (requires Go 1.22+)
cd src/kernel && go run main.go
```

> [!TIP]
> To use OpenAI, set `GAIA_PLANNER_PROVIDER="cloud"`, `GAIA_PLANNER_MODEL="gpt-4-turbo"`, and provide your `GAIA_PLANNER_API_KEY`.

### 2. Submit Your First Goal
Open a new terminal and use the unified CLI to talk to the kernel:
```bash
./gaia submit "Research the impact of Llama 3 on the agentic ecosystem and save a summary to state."
```

---

## 💻 Minimal Example (TypeScript SDK)

```typescript
import { GaiaClient } from '@gaia/sdk';

const client = new GaiaClient('http://localhost:8080');

// Submit a goal and stream the execution DAG in real-time
const task = await client.tasks.submit("Summarize the GAIA technical specs");

client.tasks.subscribe(task.id, (event) => {
  console.log(`[${event.type}] - ${event.payload.message}`);
});
```

---

## 📂 Use Cases

*   **Autonomous Research Pipelines**: Multi-step workflows that require reliable tool usage and long-running execution.
*   **Resilient Task Automation**: Enterprise workflows where an API failure at Step 50 shouldn't lose 2 hours of work.
*   **Multi-Agent Coordination**: Orchestrating specialized agents (Coding, Writing, Searching) across different LLM providers.
*   **Secure Plugin Marketplaces**: Running third-party agent capabilities without giving them full access to your environment.

---

## ⚖️ Why GAIA? (vs. Frameworks)

| Concern | LangGraph | CrewAI | AutoGen | GAIA |
| :--- | :---: | :---: | :---: | :---: |
| Dynamic agent attachment at runtime | ✗ | ✗ | ✗ | ✓ |
| Capability-first routing (not agent-first) | ✗ | ✗ | ✗ | ✓ |
| Policy firewall (no peer-to-peer) | ✗ | ✗ | ✗ | ✓ |
| Tiered trust & agent quarantine | ✗ | ✗ | ✗ | ✓ |
| A2A + MCP protocol support | ✗ | ✗ | ✗ | ✓ |
| Deterministic kernel / probabilistic planner | partial | ✗ | ✗ | ✓ |
| DAG-based parallel execution | ✓ | ✗ | ✗ | ✓ |
| State snapshotting & bounded context | ✗ | ✗ | ✗ | ✓ |

**GAIA is not a framework. It is a kernel.** Frameworks help you wire agents together. GAIA *is* the infrastructure that manages, secures, and orchestrates them.

---

## 🏗 Architecture & Principles

## Architecture

```text
                    ┌──────────────────────┐
                    │    User / Client     │
                    │   POST /tasks {goal} │
                    └──────────┬───────────┘
                               │
          ┌────────────────────▼────────────────────┐
          │              GAIA KERNEL                 │
          │                                          │
          │  ┌─────────────┐     ┌────────────────┐  │
          │  │ Goal Manager│────►│    Planner     │  │
          │  └─────────────┘     │   (LLM-based)  │  │
          │                      └───────┬────────┘  │
          │                              │ (steps)   │
          │  ┌──────────────┐    ┌───────▼────────┐  │
          │  │ Policy Engine│◄──►│   Scheduler    │  │
          │  │  (Firewall)  │    │  (DAG resolver)│  │
          │  └──────────────┘    └───────┬────────┘  │
          │                              │           │
          │  ┌──────────────┐    ┌───────▼────────┐  │
          │  │ State Store  │◄──►│   Execution    │  │
          │  │  (Tiered)    │    │    Engine      │  │
          │  └──────────────┘    └───────┬────────┘  │
          │                              │           │
          │  ┌──────────────┐    ┌───────▼────────┐  │
          │  │  Capability  │◄──►│   Request      │  │
          │  │  Registry    │    │    Router      │  │
          │  └──────────────┘    └───────┬────────┘  │
          │                              │           │
          │  ┌──────────────┐    ┌───────▼────────┐  │
          │  │  Audit Log   │◄───│   Event Bus   │  │
          │  └──────────────┘    └───────┬────────┘  │
          │                              │           │
          │         ┌────────────────────┼─────┐     │
          │         │   Transport Layer  │     │     │
          │         ├────────┬───────────┼─────┤     │
          │         │ Native │   A2A     │ MCP │     │
          │         └────────┴───────────┴─────┘     │
          └──────────────────┬──────────────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
     ┌────▼────┐       ┌────▼────┐        ┌────▼────┐
     │  Local  │       │  A2A    │        │  MCP    │
     │  Agents │       │  Agents │        │  Tools  │
     └─────────┘       └─────────┘        └─────────┘
```

---

## Core Principles

### 1. Deterministic Kernel, Probabilistic Planner

The LLM plans. The Kernel executes. The Kernel never "hallucinates" the status of a task, never skips a step, and never loses state. Given the same plan and the same agent responses, the Kernel produces the same result every time.

### 2. Capability-First Routing

The Planner never sees agent IDs. It only sees capabilities: `"read_pdf"`, `"send_email"`, `"translate_text"`. The Kernel dynamically selects the best available agent for each capability based on health, latency, and trust score.

### 3. Deny-by-Default Mediation

No agent talks to another agent. All data flows through the Kernel's Policy Engine, where it is:
* Schema-validated (input and output)
* Permission-checked (scopes and constraints)
* Audited (every message is logged with `task_id` and `step_id`)

### 4. Protocol Interoperability

GAIA natively consumes **Google A2A** (agent-to-agent) and **Anthropic MCP** (agent-to-tool) through protocol adapters. An A2A agent, an MCP tool, and a native GAIA agent can all participate in the same task, in the same plan, seamlessly.

---

## Current Status

> **⚠️ GAIA is in the Implementation phase.**

### What exists today

| Artifact | Status | Description |
| :--- | :---: | :--- |
| [Technical Specification](docs/design.md) | ✅ Complete | 1200+ line design document covering the full kernel architecture |
| [Kernel Internals Guide](docs/site/docs/internals/index.md) | ✅ Complete | A 10-chapter deep dive into the GAIA architecture and modules |
| Component Specifications | ✅ Complete | 12 detailed documents covering schemas, control loops, and security |
| Core Implementation | ✅ Complete | Go Kernel with 10-phase control loop, SQLite persistence, and CEL Policy Engine |
| Governance & Audit | ✅ Complete | Cryptographic HMAC chaining, state restoration, and usage policies |
| Performance Engine | ✅ Complete | Zero-allocation interpolation, UDS/gRPC hybrid routing, and resource quotas |
| Ecosystem & SDKs | ✅ Complete | Type-safe TS/Python SDKs, Unified CLI, and Docusaurus site |

---

## 🛠 Requirements & Setup

GAIA is a polyglot project. To initialize the repository and begin development, you need the following installed:

* **Go 1.22+**: For the core kernel.
* **Node.js 20+ & NPM**: For the TypeScript SDK and documentation site.
* **Python 3.10+**: For the Python SDK and validation scripts.

### Quick Start (CLI)

The easiest way to interact with GAIA is via the unified CLI:

```bash
# Register an agent, submit a goal, and monitor the stream
./gaia --help
```

For detailed setup instructions, visit the [Documentation Site](docs/site/docs/intro.md).

---

## Roadmap

### Phase 1: Specification (Complete)
- [x] Core architecture design
- [x] A2A + MCP interoperability design
- [x] Data model & JSON Schema definitions
- [x] Lifecycle state machine specs
- [x] Transport adapter specs
- [x] Security & policy specs
- [x] Tech Stack & Polyglot strategy

### Phase 2: Foundation (Complete)
- [x] Project scaffolding & modular monorepo setup
- [x] Core kernel types (Go)
- [x] State Store (Tier 1/4 In-Memory)
- [x] Capability Registry (Go)
- [x] Control Loop Skeleton (10-phase state machine)
- [x] Dynamic LLM Planner Adapters (Local/Cloud support)
- [ ] SDK scaffolding (TS/Python)

### Phase 3: Runtime (Complete)
- [x] Migrate Tier 4 `AgentStateStore` from in-memory to SQLite
- [x] Actual LLM API implementations (Ollama, OpenAI)
- [x] Async execution & DAG scheduler
- [x] State snapshotting & recovery (Tier 2 persistence)
- [x] MCP Adapter
- [x] A2A Adapter

### Phase 4: Resiliency & Persistence (Complete)
- [x] Exponential backoff & jitter logic
- [x] 4-tier escalation path (Retry -> Fallback -> Replan -> Abort)
- [x] Tier 2 Task Persistence (Stateful re-entry)
- [x] Multi-tenant SQLite store refactor

### Phase 5: Security & Policy (Complete)
- [x] CEL-based Policy Engine implementation
- [x] JSON Schema contract enforcement
- [x] Tier 5 Audit Log (Tamper-proof SHA-256 chaining)
- [x] Environment-based policy injection

### Phase 6: Client API & Gateway (Complete)
- [x] RESTful Task Management API
- [x] WebSocket Event Streaming (Real-time observability)
- [x] Orchestrator (Goal Manager) implementation
- [x] Multi-protocol transport routing (A2A, MCP, Native)

### Phase 7: SDKs & Ecosystem (Complete)
- [x] TypeScript SDK (libs/sdk-ts) with full type-safety
- [x] Python SDK (libs/sdk-py) with async/await support
- [x] Docusaurus-based Documentation Site (docs/site)
- [x] Automated JSON Schema ➔ SDK Type generation
- [x] Stress testing & failure injection frameworks
- [x] GAIA Unified CLI (`gaia` script)

### Phase 8: Security Hardening & mTLS (Complete)
- [x] mTLS Handshake for Agent Identity
- [x] JWT-based task authorization
- [x] Policy-based data encryption at rest
- [x] Secret management integration

### Phase 9: Observability & Human-in-the-Loop (HITL) (Complete)
- [x] Real-time DAG visualization (Dashboard)
- [x] `STEP_APPROVAL_REQUIRED` flow implementation
- [x] Agent health & trust score monitoring dashboard
- [x] Manual override & plan modification interface

### Phase 10: Enterprise Governance & Auditing (Complete)
- [x] Cryptographic Audit Log chaining (HMAC-SHA256)
- [x] Admin API for state restoration & trace verification
- [x] Advanced CEL-based policy management (Cost & Usage)
- [x] Tamper-proof deletion tombstones for reliable rollback

### Phase 11: High-Performance & Hybrid Routing (Complete)
- [x] Zero-allocation JSON interpolation engine (Nested dot-notation support)
- [x] Hybrid routing (Local IPC/UDS path vs. Remote gRPC/HTTP path)
- [x] Multi-tenant resource quotas & memory pressure handling
- [x] Kernel-level performance profiling & optimizations

---

## Project Structure

```text
GAIA/
├── docs/
│   ├── site/                      # Docusaurus documentation site
│   ├── design.md                  # Master technical specification
│   ├── specs/                     # Component-level specifications
│   │   ├── schemas/               # Canonical JSON Schemas (Source of Truth)
│   │   ├── schemas.md             # Schema definitions & contracts
│   │   ├── lifecycles.md          # State machines (Task, Step, Agent)
│   │   ├── control-loop.md        # Authoritative control loop
│   │   └── ...                    # (Policy, Planning, Registry, Security)
│   └── protocols/                 # Protocol integration specs (A2A, MCP)
├── src/
│   └── kernel/                    # Go Orchestration Kernel
│       ├── cmd/
│       │   └── schema-gen/        # Type-sync tool (Go -> JSON Schema)
│       ├── pkg/
│       │   ├── api/               # REST Handlers & WebSocket Handshake
│       │   ├── core/              # Loop, Planner, Scheduler, Transports
│       │   ├── policy/            # CEL Policy Engine & Enforcement
│       │   ├── state/             # Tiered Persistence (SQLite)
│       │   ├── registry/          # Capability Registry & Agent Discovery
│       │   └── types/             # Canonical Kernel Structs
│       └── main.go                # Kernel Entry Point
├── libs/                          # SDKs & Ecosystem
│   ├── sdk-ts/                    # TypeScript SDK (axios + ws)
│   └── sdk-py/                    # Python SDK (httpx + websockets)
├── scripts/                       # DevOps & CLI
│   └── gaia/                      # Unified CLI & validation scripts
├── gaia                           # Unified CLI Entry Point (Symlink)
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
└── SECURITY.md
```

---

## Contributing

GAIA is designed to be a community-driven project. Whether you're an architect, a protocol expert, or someone who just wants to help write better docs — there's a place for you.

**Right now, the most valuable contributions are:**

1. **Review the [Technical Specification](docs/design.md)** and open Issues for gaps, contradictions, or missing edge cases.
2. **Propose protocol adapters** — especially for protocols beyond A2A and MCP.
3. **Help define schemas** — the JSON Schema definitions will be the foundation of the entire codebase.

Please read the [Contributing Guide](CONTRIBUTING.md) before submitting changes.

---

## Community

* **Issues**: [github.com/vishalsdk14/GAIA/issues](https://github.com/vishalsdk14/GAIA/issues) — Bug reports, design feedback, and feature requests.
* **Discussions**: [github.com/vishalsdk14/GAIA/discussions](https://github.com/vishalsdk14/GAIA/discussions) — Open-ended conversations about architecture and direction.

---

## Governance

* [Code of Conduct](CODE_OF_CONDUCT.md)
* [Security Policy](SECURITY.md)
* [Changelog](CHANGELOG.md)

---

## License

GAIA is distributed under the [MIT License](LICENSE).

---

<div align="center">

*"The goal of GAIA is to provide the stable ground upon which a billion autonomous agents can safely and reliably work."*

</div>
