# GAIA Kernel Internals: The Authoritative Guide

Welcome to the **GAIA Kernel Internals** book. This guide is designed for developers, architects, and security auditors who need to understand the inner workings of the GAIA Orchestration Kernel.

Inspired by the systematic approach of "Linux Device Drivers," this guide breaks down GAIA into its component modules, explaining the theory, the code, and the specifications behind the system.

---

## Table of Contents

### Part I: The Foundation
*   **[Chapter 0: The Journey of a Goal](00-journey.md)**
    *   A narrative overview of the end-to-end task execution flow.
*   **[Chapter 1: The GAIA Paradigm](01-paradigm.md)**
    *   Deterministic Orchestration, Capability-First Routing, and Mediated Communication.
*   **[Chapter 2: Kernel Architecture](02-architecture.md)**
    *   The Process Model, Goroutines, and the 10-Phase Control Loop.
*   **[Chapter 3: Memory & State Management](03-state.md)**
    *   The 5-Tier State Model, SQLite persistence, and Agent Isolation.

### Part II: The Hot Path
*   **[Chapter 4: The Policy Engine](04-policy.md)**
    *   CEL-based Firewalling, Mediation, and Human-in-the-Loop flows.
*   **[Chapter 5: The Capability Registry](05-registry.md)**
    *   Handshakes, Agent Manifests, and dynamic routing logic.
*   **[Chapter 6: Planning & Interpolation](06-planning.md)**
    *   Incremental DAG generation and byte-level data binding.

### Part III: Connectivity & Ecosystem
*   **[Chapter 7: The Transport Layer](07-transport.md)**
    *   Protocol Adapters (Native, A2A, MCP) and Async Invocations.
*   **[Chapter 8: Resiliency & Escalation](08-resiliency.md)**
    *   The 4-tier failure path: Retry, Fallback, Replan, Abort.

### Part IV: Enterprise Hardening
*   **[Chapter 9: Security & Trust Models](09-security.md)**
    *   Zero-Trust architecture, Trust Scores, and Cryptographic Chaining.
*   **[Chapter 10: Performance & Scaling](10-scaling.md)**
    *   Hot-path optimizations, Hybrid routing, and Resource Quotas.

---

## How to use this guide

Each chapter is cross-referenced with the authoritative **Specifications** found in the `docs/specs/` directory of the repository. While the specifications define the "What," this guide explains the "Why" and the "How."

Whether you are building a new agent protocol, auditing the kernel's safety, or optimizing task performance, this guide is your Source of Truth.
