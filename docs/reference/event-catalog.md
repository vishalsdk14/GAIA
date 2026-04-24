# Event Catalog Reference

> **Status**: 🔲 Not Started
>
> **Source**: [design.md — Section 3.6: Event Catalog](../design.md)

---

## Purpose

This document is the **authoritative reference** for every event type emitted by the GAIA kernel Event Bus. It defines the event name, payload schema, when it fires, and who can subscribe to it.

---

## System Events

### Task Events

| Event | Payload | When Emitted |
| :--- | :--- | :--- |
| `TASK_CREATED` | | |
| `TASK_PLANNING` | | |
| `TASK_EXECUTING` | | |
| `TASK_COMPLETED` | | |
| `TASK_FAILED` | | |
| `TASK_CANCELLED` | | |

---

### Step Events

| Event | Payload | When Emitted |
| :--- | :--- | :--- |
| `STEP_STARTED` | | |
| `STEP_COMPLETED` | | |
| `STEP_FAILED` | | |

---

### Plan Events

| Event | Payload | When Emitted |
| :--- | :--- | :--- |
| `PLAN_GENERATED` | | |
| `PLAN_REJECTED` | | |
| `REPLAN_TRIGGERED` | | |

---

### Agent Events

| Event | Payload | When Emitted |
| :--- | :--- | :--- |
| `AGENT_REGISTERED` | | |
| `AGENT_DEGRADED` | | |
| `AGENT_QUARANTINED` | | |
| `AGENT_BLACKLISTED` | | |
| `AGENT_EJECTED` | | |
| `AGENT_DISCONNECTED` | | |

---

## Subscription Rules

| Subscriber | Allowed Events | Transport |
| :--- | :--- | :--- |
| Client | Task + Step events | WebSocket / SSE |
| Audit Log | All events | Internal (mandatory) |
| Agents | None | N/A |

---

## TODO

- [ ] Define payload schemas for each event
- [ ] Document event ordering guarantees
- [ ] Specify event filtering syntax for clients
