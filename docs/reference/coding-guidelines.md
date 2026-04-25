# GAIA Coding Guidelines

This document establishes the authoritative coding standards for the GAIA Orchestration Kernel and its associated SDKs. All contributors must adhere to these patterns to ensure system stability, performance, and security.

---

## 1. Core Kernel (Go)

The Kernel is built in **Go** for high-concurrency and "Linux-standard" robustness.

### 1.1 Concurrency Pattern
* **Rule**: Favor channels and goroutines over mutexes.
* **Why**: The GAIA 10-phase control loop is inherently event-driven. Channels provide a cleaner, less error-prone way to coordinate task state transitions.
* **Pattern**: Each active `Task` should be managed by a dedicated worker goroutine that listens for events (completion, failure, timeout) via a central `TaskCoordinator`.

### 1.2 Error Management
* **Rule**: Errors must be wrapped with architectural context.
* **Format**: `return fmt.Errorf("dispatcher [step=%s]: %w", stepID, err)`
* **Reference**: Map all terminal errors to a canonical code from `docs/reference/error-codes.md`.

### 1.3 High-Performance Hot Path
The "Hot Path" includes: Phase 5 (Policy Evaluation), Phase 6 (Interpolation), and Phase 7 (State Update).
* **Zero Allocation**: Use `sync.Pool` for `Event` and `DeltaLog` objects to minimize Garbage Collection (GC) pauses.
* **JSON Traversal**: Do **not** unmarshal JSON for interpolation. Use `tidwall/gjson` to scan raw byte slices.
* **Lock Granularity**: Never hold a global state lock during an I/O operation (e.g., calling an agent). Release locks before making network calls.

### 1.4 Zero Magic Numbers
* **Rule**: Literal values (e.g., `60`, `15000`, `0.9`) are strictly forbidden in the business logic.
* **Why**: All architectural limits, timeouts, retry counts, and status strings must be centralized to ensure the system is tunable and self-documenting.
* **Implementation**:
    - Use `const` blocks for all internal limits and status strings.
    - Use the `reference/` directory as the source of truth for all Enums and Error Codes.
    - Any value that may need to be tuned by an operator must be moved to the Kernel Configuration manifest.

### 1.5 Radical Modularity & DRY
* **No Monoliths**: Large, single-file implementations are strictly forbidden. The kernel must be decomposed into small, single-responsibility files and packages.
* **DRY (Don't Repeat Yourself)**: Zero tolerance for code duplication. Common logic (e.g., JSON traversal, error wrapping, event emission) must be abstracted into internal utility packages.
* **Package Isolation**: Each of the 10 phases of the Control Loop should ideally reside in its own package or sub-package to enforce clean boundaries and prevent circular dependencies.

---

## 2. Agent SDKs (Polyglot)

SDKs must make it effortless for developers to build secure GAIA agents.

### 2.1 TypeScript Standards
* **Type Safety**: Interfaces must be generated from `docs/specs/schemas.md`. Manual type definitions for core schemas are forbidden.
* **Async Hygiene**: Use `Promise.allSettled` when executing multiple sub-tasks to ensure the agent doesn't hang on a single failure.

### 2.2 Python Standards
* **Concurrency**: Use `asyncio` for the transport adapter layer.
* **Type Hinting**: Mandatory for all public-facing methods in the SDK.

---

## 3. Documentation & Comments

* **Godoc/TSDoc**: Every exported function, interface, and struct must have a documentation comment.
* **Why**: Comments should explain the *why*, not the *how*. 
* **Spec References**: If a function implements a specific phase of the Control Loop, reference it: `// Implements Phase 6 (Interpolation) of docs/specs/control-loop.md`.

---

## 4. Testing Philosophy

* **Table-Driven Tests**: Use table-driven testing in Go for all logic-heavy components (Planner, Dispatcher, Policy Engine).
* **Mocking**: Use standard interfaces to mock the `Transport` layer. Do not rely on real network connections for unit tests.
* **Race Detection**: All Go tests must be run with the `-race` flag enabled.

---

## Related Documents

* [Tech Stack Spec](../specs/tech-stack.md) — The language and library selections.
* [Control Loop Spec](../specs/control-loop.md) — The logic being implemented.
* [Error Codes](../reference/error-codes.md) — The canonical error definitions.
