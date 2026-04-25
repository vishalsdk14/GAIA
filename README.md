<div align="center">

# GAIA

### The Orchestration Kernel for Autonomous Agents

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Status: Implementation Phase](https://img.shields.io/badge/Status-Implementation%20Phase-green.svg)](#current-status)
[![Protocol: A2A](https://img.shields.io/badge/Protocol-Google%20A2A-4285F4.svg)](https://github.com/google/A2A)
[![Protocol: MCP](https://img.shields.io/badge/Protocol-Anthropic%20MCP-D97706.svg)](https://modelcontextprotocol.io/)

**GAIA is a deterministic execution kernel that turns a probabilistic planner and a swarm of untrusted, plug-in agents into a reliable, goal-completing system.**

[Design Spec](docs/design.md) · [Contributing](CONTRIBUTING.md) · [Issues](https://github.com/vishalsdk14/GAIA/issues) · [Discussions](https://github.com/vishalsdk14/GAIA/discussions)

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
| Component Specifications | ✅ Complete | 12 detailed documents covering schemas, control loops, and security |
| Tech Stack Decision | ✅ Complete | Polyglot architecture (Go Core, TS/Python SDKs, Rust modules) |
| Repository Scaffolding | ✅ Complete | Modular monorepo with production Go kernel structure |
| Core Implementation (Code) | ✅ Complete | Phases 1-4 implemented; Kernel core is feature-complete |

---

## 🛠 Requirements & Setup

GAIA is a polyglot project. To initialize the repository and begin development, you need the following installed:

* **Go 1.22+**: For the core kernel.
* **Node.js 20+ & NPM**: For the TypeScript SDK and protocol adapters.
* **Python 3.10+**: For the Python SDK and AI agent integrations.
* **Git**: For version control.

### Quick Start (Scaffolding)

To initialize the repository structure and language modules, run:

```bash
./scripts/init.sh
```

This will create the following isolated modules:
* `src/kernel/` (Go)
* `libs/sdk-ts/` (TypeScript)
* `libs/sdk-py/` (Python)

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

### Phase 4: Production Readiness (Complete)
- [x] Policy Engine & sandbox enforcement (CEL-based)
- [x] Observability (structured logs, Event Bus)
- [x] Failure Handling (retries, escalation, circuit breakers)
- [ ] Stress testing & failure injection
- [ ] Documentation site

---

## Project Structure

```text
GAIA/
├── docs/
│   ├── design.md                  # Master technical specification
│   ├── specs/                     # Component-level specifications
│   │   ├── schemas.md             # JSON Schema definitions
│   │   ├── lifecycles.md          # State machines (Task, Step, Agent)
│   │   ├── control-loop.md        # Authoritative control loop
│   │   ├── communication.md       # Messages, events, routing
│   │   ├── registry.md            # Capability Registry
│   │   ├── planning.md            # Planner & interpolation engine
│   │   ├── state-management.md    # Tiered state & snapshotting
│   │   ├── failure-handling.md    # Retries, escalation, circuit breakers
│   │   ├── security.md            # Policy engine & sandbox
│   │   ├── transport.md           # Transport layer & adapters
│   │   └── client-api.md          # REST API & streaming
│   ├── protocols/                 # Protocol integrations
│   │   ├── a2a-integration.md     # Google A2A
│   │   ├── mcp-integration.md     # Anthropic MCP
│   │   └── native-protocol.md     # GAIA native protocol
│   ├── guides/                    # User & developer guides
│   │   ├── getting-started.md
│   │   ├── building-agents.md
│   │   ├── building-adapters.md
│   │   ├── deployment.md
│   │   └── configuration.md
│   ├── reference/                 # Reference materials
│   │   ├── glossary.md
│   │   ├── error-codes.md
│   │   └── event-catalog.md
│   └── rfcs/                      # Design proposals
│       └── 000-template.md
├── src/                           # Implementation (coming soon)
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
