# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in GAIA, please report it responsibly.

**Do NOT open a public GitHub Issue for security vulnerabilities.**

Instead, please email: **[security contact to be added]**

### What to include

* Description of the vulnerability
* Steps to reproduce
* Potential impact
* Suggested fix (if any)

### Response timeline

* **Acknowledgment**: Within 48 hours
* **Assessment**: Within 7 days
* **Fix/Disclosure**: Within 30 days (coordinated)

---

## Supported Versions

| Version | Supported |
| :--- | :---: |
| 0.x (pre-release) | ✅ |

---

## Security Design

GAIA's security architecture is documented in:
* [Security & Policy Spec](docs/specs/security.md)
* [design.md — Section 13: Security Model](docs/design.md)

### Core Security Principles

1. **Deny-by-default**: All agent egress is blocked unless explicitly allowed.
2. **Centralized mediation**: No agent-to-agent communication; all traffic passes through the kernel firewall.
3. **Schema validation**: Every input and output is validated against declared JSON Schemas.
4. **Tiered trust**: Agents are monitored and can be degraded, quarantined, or blacklisted automatically.
5. **Audit logging**: Every message and policy decision is logged.
