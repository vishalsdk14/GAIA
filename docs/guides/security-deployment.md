# Security Deployment Guide: Local vs. Production

This guide outlines the recommended security postures for GAIA based on your deployment scenario. GAIA follows a **"Deny-by-Default"** philosophy, but it provides flexible modes to support rapid local development.

---

## 1. Role Definitions

Understanding the distinction between the **Operator** and the **Agent Developer** is key to securing GAIA.

| Role | Responsibility | Key Assets |
| :--- | :--- | :--- |
| **Kernel Operator** | Manages the GAIA infrastructure and Root CA. | `GAIA_ENCRYPTION_KEY`, Root CA Cert/Key |
| **Agent Developer** | Builds and deploys agents that connect to the Kernel. | Agent Cert/Key, JWT Token |

---

## 2. Deployment Scenarios

### Scenario A: Local Development (The "Laptop" Use Case)
**Posture**: `legacy` mode.
**Description**: You are both the Operator and the Developer. Everything runs on `localhost`.

* **Auth Mode**: Set `GAIA_AUTH_MODE=legacy`.
* **Verification**: Disabled. The Kernel trusts the `X-Agent-ID` header.
* **Encryption**: **Recommended**. Set `GAIA_ENCRYPTION_KEY` to protect your agent's sensitive state data in the local SQLite file.
* **Why?**: Maximizes speed and minimizes the friction of managing certificates.

### Scenario B: Private Team / Shared VPC
**Posture**: `standard` mode.
**Description**: GAIA is hosted on a private server accessible to your team.

* **Auth Mode**: Set `GAIA_AUTH_MODE=standard`.
* **Verification**: Enabled via **JWT**.
* **Setup**: 
  1. Operator provides a `GAIA_JWT_SECRET`.
  2. Operator issues JWT tokens to Agent Developers.
  3. Developers include the token in the `Authorization: Bearer <token>` header.
* **Encryption**: **Mandatory**. Protects against unauthorized access if the server or database is compromised.

### Scenario C: Public / Untrusted Network (Production)
**Posture**: `strict` mode.
**Description**: GAIA is open to the internet or an environment where agent identities must be cryptographically guaranteed.

* **Auth Mode**: Set `GAIA_AUTH_MODE=strict`.
* **Verification**: **Mandatory mTLS**.
* **Setup**:
  1. Operator generates unique certificates for each agent using `generate_certs.sh`.
  2. Developers must provide their `.crt` and `.key` to their SDK client.
  3. The Kernel verifies the certificate Common Name (CN) against the `agent_id`.
* **Encryption**: **Mandatory**. Uses the Master Key to secure all data at rest.

---

## 3. Quick Reference: Security Env Vars

| Variable | Description | Default |
| :--- | :--- | :--- |
| `GAIA_AUTH_MODE` | `legacy`, `standard`, or `strict`. | `legacy` |
| `GAIA_ENCRYPTION_KEY` | 32-byte master key for AES-GCM. | None |
| `GAIA_AUTH_JWT_ENABLED` | Enables JWT enforcement in `standard` mode. | `false` |
| `GAIA_JWT_SECRET` | Secret key used to sign/verify JWT tokens. | None |

---

## 4. Best Practices

1. **Rotate Keys**: If your `GAIA_ENCRYPTION_KEY` is compromised, all Tier 4 data must be re-encrypted.
2. **Protect the Root CA**: In `strict` mode, anyone with access to the Root CA key can issue themselves "Valid" agent certificates.
3. **Use Env Secrets**: Never hardcode keys or secrets in source code. Use a secure Secret Manager in production.
