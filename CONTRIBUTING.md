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

GAIA is in the **Design & Specification** phase. This means:

* ✅ We are actively accepting **design reviews, schema proposals, and specification feedback**.
* ✅ We welcome **documentation improvements** and **protocol adapter proposals**.
* 🔲 We are **not yet accepting code contributions** (the implementation has not started).

When the project transitions to the Implementation phase, this guide will be updated with code-specific standards.

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

## Code Standards (Future)

> These standards will be finalized when the project enters the Implementation phase.

Preliminary decisions:

* **Language**: TypeScript (Node.js runtime).
* **Style**: Strict typing. No `any`. All interfaces defined in `src/types/`.
* **Testing**: Unit tests for all kernel components. Integration tests for protocol adapters.
* **Linting**: ESLint with strict configuration.

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
