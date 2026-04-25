<div align="center">

# GAIA

### The Orchestration Kernel for Autonomous Agents

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status: Implementation Phase](https://img.shields.io/badge/Status-Implementation%20Phase-green.svg)](#current-status)
[![Protocol: A2A](https://img.shields.io/badge/Protocol-Google%20A2A-4285F4.svg)](https://github.com/google/A2A)
[![Protocol: MCP](https://img.shields.io/badge/Protocol-Anthropic%20MCP-D97706.svg)](https://modelcontextprotocol.io/)

**GAIA is a deterministic execution kernel that turns a probabilistic planner and a swarm of untrusted, plug-in agents into a reliable, goal-completing system.**

[Design Spec](docs/design.md) · [Lifecycle Handbook](docs/guides/task-lifecycle-handbook.md) · [Kernel Internals](docs/site/docs/internals/index.md) · [Contributing](CONTRIBUTING.md)

</div>

---

## The Problem

AI agents today are powerful but **fragile**. They work in demos but fail in production because:

* **No recovery**: One API timeout at Step 10 of 50 loses all progress.
* **No security**: Agents get unlimited access to tools, data, and each other.
* **No interoperability**: OpenAI, Anthropic, and Google agents can't work together.
* **No separation of concerns**: The same LLM that plans also executes — and it forgets Step 34.

Every team building "AI agents" is independently solving the same infrastructure problems. GAIA solves them once, at the kernel level.

---

## Why GAIA? (vs. Existing Frameworks)

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
| [Lifecycle Handbook](docs/guides/task-lifecycle-handbook.md) | ✅ Complete | A narrative guide to the journey of a goal through the kernel |
| [Kernel Internals Guide](docs/site/docs/internals/index.md) | ✅ Complete | A 10-chapter deep dive into the GAIA architecture and modules |
| Component Specifications | ✅ Complete | 12 detailed documents covering schemas, control loops, and security |
| Core Implementation | ✅ Complete | Go Kernel with 10-phase control loop, SQLite persistence, and CEL Policy Engine |
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

### Phase 8: Security Hardening & mTLS (Next)
- [ ] mTLS Handshake for Agent Identity
- [ ] JWT-based task authorization
- [ ] Policy-based data encryption at rest
- [ ] Secret management integration

### Phase 9: Observability & Human-in-the-Loop (HITL)
- [ ] Real-time DAG visualization (Dashboard)
- [ ] `STEP_APPROVAL_REQUIRED` flow implementation
- [ ] Agent health & trust score monitoring dashboard
- [ ] Manual override & plan modification interface

### Phase 10: Enterprise Governance & Auditing
- [ ] Cryptographic Audit Log chaining (SHA-256)
- [ ] Admin API for log querying & agent restoration
- [ ] Advanced CEL-based policy management (Cost control, regional routing)
- [ ] Tamper-proof event persistence

### Phase 11: High-Performance & Hybrid Routing
- [ ] Zero-allocation JSON interpolation engine
- [ ] Hybrid routing (Local IPC path vs. Remote gRPC/HTTP path)
- [ ] Multi-tenant resource quotas & memory pressure handling
- [ ] Kernel-level performance profiling & optimizations

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
