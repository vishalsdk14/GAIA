# Technology Stack Architecture

> **Status**: 🟢 Complete
>
> **Source**: Architecture Decision Record (ADR)

---

## Purpose

This document formally specifies the Technology Stack for the GAIA Orchestration Kernel. Given the system's transport-agnostic design, GAIA adopts a **Polyglot Architecture**, selecting the optimal language for each layer of the system to balance performance, concurrency, safety, and ecosystem compatibility.

---

## 1. The Core Kernel (Go)

The central Orchestrator (Control Loop, Event Bus, Dispatcher, and Registry) is implemented in **Go (Golang)**.

### Why Go?
* **High Concurrency**: The 10-phase control loop manages thousands of parallel tasks and step executions. Go’s native `goroutines` and `channels` are mathematically proven models for managing this level of concurrency with minimal memory overhead.
* **Orchestration Standard**: Go is the industry standard for infrastructure orchestration (e.g., Kubernetes, Docker, Terraform).
* **Transport Agnostic**: Go's standard library provides robust, production-grade HTTP/2 and gRPC servers out of the box, perfectly aligning with GAIA's multi-transport requirements.
* **Deployment**: Compiles to a single, statically linked binary, making distribution and containerization frictionless.

---

## 2. Agent SDKs & Adapters (TypeScript & Python)

While the Core is in Go, the agents that connect to GAIA can be written in any language. To accelerate development, the official GAIA Agent SDKs are provided in **TypeScript** and **Python**.

### 2.1 TypeScript (Node.js)
* **Target Audience**: Web developers, MCP integration builders, and full-stack AI teams.
* **Why**: TypeScript natively speaks JSON, making JSON Schema validation and manipulation seamless. It has the largest ecosystem of API integrations and frontend frameworks.

### 2.2 Python
* **Target Audience**: Data scientists, ML engineers, and core LLM researchers.
* **Why**: Python is the lingua franca of AI (PyTorch, LangChain, LlamaIndex). Providing a Python SDK ensures that the most powerful, cutting-edge AI models can connect natively to the GAIA Kernel.

---

## 3. High-Performance Internal Modules

To maintain extreme throughput, the kernel avoids cross-language FFI (Foreign Function Interfaces) or WebAssembly boundaries on the hot path, relying instead on highly optimized native Go libraries.

### 3.1 The Policy Engine (`cel-go`)
* **Implementation**: The CEL policies (Phase 5 of the Control Loop) are evaluated using Google's official `cel-go` library.
* **Why**: Policy evaluations block every inbound and outbound request. Using native `cel-go` compiles rules to an AST and evaluates them in nanoseconds with zero serialization overhead (avoiding the millisecond penalty of crossing a Wasm memory boundary).

### 3.2 Zero-Allocation JSON Interpolation
* **Implementation**: Dynamic step interpolation (resolving `{{step.output}}`) is handled via specialized zero-allocation JSON parsers (e.g., `tidwall/gjson`).
* **Why**: Using Go's standard `encoding/json` to unmarshal arbitrary dynamic output into `map[string]interface{}` generates massive Garbage Collection (GC) churn. Direct byte-slice traversal prevents heap fragmentation and keeps dispatcher latency flat.

---

### 4.1 Monorepo Strategy
GAIA uses a **Modular Monorepo** approach. To maintain clean boundaries between different technology stacks and prevent dependency leakage:
* **Isolated Modules**: Each sub-project (Kernel, TS SDK, Python SDK) is its own isolated language module (e.g., its own `go.mod`, `package.json`, or `pyproject.toml`).
* **Root Cleanliness**: The root directory is reserved for documentation, high-level project metadata, and repository-wide automation scripts. No source code or language-specific configuration resides in the root.

### 4.2 Directory Layout
```text
/gaia
├── src/kernel/           # Core Orchestrator (Go Module: gaia/kernel)
├── libs/sdk-ts/          # Agent SDK (NPM Package: @gaia/sdk)
├── libs/sdk-py/          # Agent SDK (Python Package: gaia-sdk)
├── scripts/              # Repository automation and init scripts
└── docs/specs/           # The Source of Truth (Markdown/JSON Schema)
```

---

## Related Documents

* [Transport Spec](transport.md) — Explains how these different languages communicate over the wire.
* [Security Spec](security.md) — The CEL Policy Engine implementation details.
