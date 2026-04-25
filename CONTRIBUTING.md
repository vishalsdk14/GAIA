# Contributing to GAIA

Thank you for your interest in contributing to GAIA. This document provides guidelines and standards for contributing to the project.

---

## Table of Contents

- [Current Phase](#current-phase)
- [How to Contribute](#how-to-contribute)
- [Contribution Types](#contribution-types)
- [Submitting Changes](#submitting-changes)
- [Design Document Standards](#design-document-standards)
- [Code Standards](#code-standards-future)
- [Communication](#communication)

---

## Current Phase

GAIA is in the **Implementation** phase. This means:

* ✅ We are actively accepting **code contributions** for the Go Kernel and SDKs.
* ✅ We continue to accept **design reviews and specification feedback**.
* ✅ We welcome **protocol adapter** and **agent SDK** implementations.

All contributions must strictly adhere to the finalized [Technical Specifications](docs/specs/).

---

## How to Contribute

### 1. Report a Design Gap or Issue

If you find a contradiction, missing edge case, or architectural flaw in the [Technical Specification](docs/design.md):

1. Open an [Issue](https://github.com/vishalsdk14/GAIA/issues/new) with the label `design`.
2. Reference the specific section number (e.g., "Section 6.3 — Parallel Execution").
3. Describe the gap and, if possible, propose a fix.

### 2. Propose a New Feature or Adapter

To propose a new protocol adapter (beyond A2A and MCP) or a new kernel feature:

1. Open a [Discussion](https://github.com/vishalsdk14/GAIA/discussions/new) in the "Ideas" category.
2. Include:
   * **Problem**: What limitation does this address?
   * **Proposal**: How would it work within GAIA's architecture?
   * **Compatibility**: Does it affect existing schemas or the Control Loop?
3. If the proposal is accepted, it will be converted into a tracked Issue.

### 3. Improve Documentation

Documentation contributions are always welcome. This includes:

* Fixing typos, improving clarity, or adding examples to existing docs.
* Writing guides for specific use cases.
* Translating documentation into other languages.

---

## Contribution Types

| Type | How to Submit | Label |
| :--- | :--- | :--- |
| Design review / feedback | Issue | `design` |
| Bug in specification | Issue | `bug`, `design` |
| New feature proposal | Discussion → Issue | `enhancement` |
| New protocol adapter | Discussion → Issue | `adapter` |
| Documentation improvement | Pull Request | `docs` |
| Schema definition | Pull Request | `schema` |
| Code contribution (future) | Pull Request | `implementation` |

---

## Submitting Changes

### For Documentation and Specifications

1. **Fork** the repository.
2. **Create a branch** from `main`:
   ```bash
   git checkout -b docs/your-change-description
   ```
3. **Make your changes** following the [Design Document Standards](#design-document-standards) below.
4. **Commit** with a clear, prefixed message:
   ```bash
   git commit -m "docs: add retry policy examples to schemas.md"
   ```
5. **Push** and open a Pull Request against `main`.
6. **Describe** what you changed and why in the PR description.

### Commit Message Format

All commits must follow this format:

```
<type>: <short description>
```

| Type | Use For |
| :--- | :--- |
| `docs` | Documentation and specification changes |
| `schema` | JSON Schema definitions |
| `feat` | New features (future) |
| `fix` | Bug fixes (future) |
| `refactor` | Code restructuring (future) |
| `test` | Test additions (future) |

---

## Design Document Standards

All specification documents in `docs/` must follow these rules:

### Structure

* Use **numbered sections** (e.g., `## 4. Capability & Agent Registration`).
* Use **horizontal rules** (`---`) between major sections.
* Use **JSON code blocks** for all schema definitions.
* Use **tables** for mappings, comparisons, and status fields.

### Schemas

* All schemas must be valid **JSON Schema** (or clearly labeled as illustrative pseudoschema).
* Every schema field must have a defined type.
* Enum values must be explicitly listed (e.g., `"status": "pending | running | done | failed"`).

### Naming Conventions

* Section titles: **Title Case** (e.g., "Agent Manifest").
* JSON field names: **snake_case** (e.g., `agent_id`, `task_id`).
* Status values: **lowercase** (e.g., `pending`, `running`, `done`).
* Event names: **UPPER_SNAKE_CASE** (e.g., `STEP_COMPLETED`, `TASK_FAILED`).

### Cross-References

* When referencing another section, use the format: `(see Section X)`.
* Never duplicate a schema in two places. Define it once and reference it.

---

## Code Standards

GAIA uses a **Polyglot Architecture**. The Core Kernel is implemented in Go, while Agent SDKs are provided in TypeScript and Python.

### 1. Documentation & Comments (Strict Policy)
* **Mandatory Comments**: Code without comprehensive comments will **not be accepted**. This is a strict rule for all PRs.
* **Copyright**: Every source file must begin with the standard GAIA Contributors MIT License header.
* **Godoc/TSDoc**: Every exported function, interface, and struct must have a documentation comment explaining *why* it exists, not just *what* it does.

### 2. Core Kernel (Go)
* **Language**: Go 1.22+
* **Concurrency**: Use `goroutines` and `channels` for the control loop. Avoid shared state/mutexes unless performance-critical.
* **Error Handling**: Use the standard `if err != nil` pattern. Wrap errors with context using `fmt.Errorf("context: %w", err)`. Use GAIA error codes from `docs/reference/error-codes.md`.
* **Performance**: 
    - Use `sync.Pool` for high-frequency state objects (Delta Log entries).
    - Use `tidwall/gjson` for zero-allocation JSON traversal during interpolation.
* **Formatting**: All code must be formatted with `gofmt` and linted with `golangci-lint`.

### 3. Agent SDKs (TypeScript / Python)
* **TypeScript**: 
    - No `any`. 
    - All types must be generated from the `docs/specs/schemas.md`.
    - Use `Async/Await` exclusively for I/O.
* **Python**:
    - Use Type Hints for all function signatures.
    - Follow PEP 8 style guidelines.
    - Use `asyncio` for non-blocking agent communication.

### 4. Testing Requirements
* **Unit Tests**: Mandatory for all kernel phases (Phase 1-10) and interpolation logic.
* **Integration Tests**: Required for all transport adapters (HTTP, gRPC, WebSocket).
* **Fuzzing**: Used for the CEL Policy Engine and JSON schema validation logic.

---

## Communication

* **Issues**: For specific, actionable problems or tasks.
* **Discussions**: For open-ended architectural conversations, proposals, and questions.
* **Pull Requests**: For submitting concrete changes.

Please be respectful, constructive, and specific in all communications.

All participants are expected to follow our [Code of Conduct](CODE_OF_CONDUCT.md).

For security-related issues, please follow the [Security Policy](SECURITY.md) — do **not** open a public Issue.

---

## Recognition

All contributors will be recognized in the project. Significant design contributions will be credited in the relevant specification documents.

---

Thank you for helping build the foundation for the agentic era.
