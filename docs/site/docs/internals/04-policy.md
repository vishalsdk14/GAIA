# Chapter 4: The Policy Engine

The **Policy Engine** is the GAIA Kernel's "Internal Firewall." It is responsible for the **Phase 5 Policy Check** in the control loop, ensuring that no action is taken without explicit permission, schema validation, and safety clearance.

---

## 4.1 Common Expression Language (CEL)

GAIA uses Google's **Common Expression Language (CEL)** for its policy definitions. CEL was chosen because it is:
*   **Fast**: Compiles to an abstract syntax tree (AST) and evaluates in microseconds.
*   **Safe**: Non-Turing complete, meaning it cannot enter infinite loops or crash the Kernel.
*   **Declarative**: Easy to read and write for developers and security auditors.

---

## 4.2 The Mediation Pipeline

Every message that moves between an Agent and the Kernel (or between the Planner and the Kernel) must pass through the Policy Engine.

### 1. Inbound Validation (Agent -> Kernel)
When an agent returns an output, the Policy Engine checks:
*   **Schema Check**: Does the output match the `output_schema` registered in the manifest?
*   **Contract Check**: Did the agent attempt to return data it isn't authorized to access?

### 2. Outbound Validation (Kernel -> Agent)
Before a request is dispatched to an agent, the Policy Engine checks:
*   **Scope Check**: Does the agent have the necessary OAuth/mTLS scopes for the requested capability?
*   **Constraint Check**: Is the action allowed in the current environment (e.g., `external_io` might be blocked in `test` environments)?
*   **Resource Check**: Does the task have enough budget (time or cost) remaining for this step?

---

## 4.3 Policy Examples

Policies are stored in the Registry and can be updated without restarting the Kernel.

### Cost Control Rule
```cel
task.metrics.cost_estimate + step.metrics.cost_estimate < task.budget.max_usd
```

### Safety & Approval Rule
```cel
(capability.name == "send_email" && task.user.trust_level < 2) ? approval_required : true
```

### Sandbox Enforcement
```cel
agent.auth.scopes.contains("state:write") || !capability.mutates_state
```

---

## 4.4 The `APPROVAL_REQUIRED` State

If a policy evaluation returns `false` but the rule is marked as "soft," the step enters the `APPROVAL_REQUIRED` state. 

1.  The control loop **yields**. It does not block other parallel steps.
2.  The Kernel emits a `STEP_APPROVAL_REQUIRED` event via the WebSocket stream.
3.  The task pauses until an administrator or authorized user sends an approval signal via the **Admin API**.
4.  Once approved, the step re-enters the scheduling queue and proceeds to Phase 6 (Routing).

---

## 4.5 Performance Considerations

Policy evaluation happens on the "Hot Path" of the execution engine. To maintain extreme throughput:
*   **Pre-Compilation**: CEL rules are pre-compiled and cached in memory.
*   **Zero-Allocation**: The Kernel uses specialized pools to evaluate rules without generating garbage collection churn.
*   **Short-Circuiting**: If a critical "Deny" rule fails, subsequent checks are skipped for immediate rejection.

---

## 4.6 Related Specifications

*   [Security Spec](../../specs/security.md)
*   [Control Loop Spec (Phase 5)](../../specs/control-loop.md)
*   [Client API (Approval Endpoints)](../../specs/client-api.md)
