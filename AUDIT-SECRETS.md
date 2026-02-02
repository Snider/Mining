# Security Audit: Secrets & Configuration

This document outlines the findings of a security audit focused on exposed secrets and insecure configurations.

## 1. Secret Detection

### 1.1. Hardcoded Credentials & Sensitive Information

- **Placeholder Wallet Addresses:**
  - `miner/core/src/config.json`: Contains the placeholder `"YOUR_WALLET_ADDRESS"`.
  - `miner/proxy/src/config.json`: Contains the placeholder `"YOUR_WALLET"`.
  - `miner/core/doc/api/1/config.json`: Contains a hardcoded wallet address.

- **Default Passwords:**
  - `miner/core/src/config.json`: The `"pass"` field is set to `"x"`.
  - `miner/proxy/src/config.json`: The `"pass"` field is set to `"x"`.
  - `miner/core/doc/api/1/config.json`: The `"pass"` field is set to `"x"`.

- **Placeholder API Tokens:**
  - `miner/core/doc/api/1/config.json`: The `"access-token"` is set to the placeholder `"TOKEN"`.

## 2. Configuration Security

### 2.1. Insecure Default Configurations

- **`null` API Access Tokens:**
  - `miner/core/src/config.json`: The `http.access-token` is `null` by default. If the HTTP API is enabled without setting a token, it could allow unauthorized access.
  - `miner/proxy/src/config.json`: The `http.access-token` is `null` by default, posing a similar risk.

- **TLS Disabled by Default:**
  - `miner/core/src/config.json`: The `tls.enabled` flag is `false` by default. If services are exposed, communication would be unencrypted.
  - `miner/proxy/src/config.json`: While `tls.enabled` is `true`, the `cert` and `cert_key` fields are `null`, preventing a secure TLS connection from being established.

### 2.2. Verbose Error Messages

No instances of overly verbose error messages leaking sensitive information were identified during this audit.

### 2.3. CORS Policy

The CORS policy could not be audited as it was not explicitly defined in the scanned files.

### 2.4. Security Headers

No security headers (e.g., CSP, HSTS) were identified in the configuration files.
