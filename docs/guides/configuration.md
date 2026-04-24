# Configuration Reference

> **Status**: 🔲 Not Started — will be written when implementation begins.

---

## Purpose

This document will provide the **complete configuration reference** for the GAIA kernel — every configurable parameter with its type, default value, and description.

---

## Planned Sections

### 1. Kernel Configuration
* Control loop timing
* Maximum concurrent steps
* Planner timeout
* Default retry policy

### 2. State Management Configuration
* Snapshot threshold (step count)
* State size limit
* Storage backend connection strings
* Archive retention period

### 3. Security Configuration
* Default auth method
* TLS settings
* Rate limit defaults
* Sandbox resource limits

### 4. Transport Configuration
* Default timeouts per transport type
* gRPC settings
* HTTP connection pool settings

### 5. Observability Configuration
* Log level and format
* Metrics export endpoint
* Trace sampling rate

### 6. Planner Configuration
* LLM provider settings
* API key
* Model selection
* Prompt templates path
* Max replan attempts

---

## TODO

- [ ] Write after implementation defines configuration schema
