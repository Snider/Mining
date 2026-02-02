# Mining Protocol Security Audit: AUDIT-PROTOCOL.md

## 1. Stratum Protocol Security

**Findings:**

- **Insecure Default Connections:** The miner defaults to `stratum+tcp`, transmitting data in plaintext. This exposes sensitive information, such as wallet addresses and passwords, to interception. An attacker with network access could easily capture and exploit this data.

- **Lack of Certificate Pinning:** Although TLS is an option, there is no mechanism for certificate pinning. Without it, the client cannot verify the authenticity of the pool's certificate, leaving it vulnerable to man-in-the-middle attacks where a malicious actor could impersonate the mining pool.

- **Vulnerability to Protocol-Level Attacks:** The Stratum protocol implementation does not adequately protect against attacks like share hijacking or difficulty manipulation. An attacker could potentially modify Stratum messages to redirect shares or disrupt the mining process.

**Recommendations:**

- **Enforce TLS by Default:** Mandate the use of `stratum+ssl` to ensure all communication between the miner and the pool is encrypted.
- **Implement Certificate Pinning:** Add support for certificate pinning to allow users to specify the expected certificate, preventing man-in-the-middle attacks.
- **Add Protocol-Level Integrity Checks:** Implement checksums or signatures for Stratum messages to ensure their integrity and prevent tampering.

## 2. Pool Authentication

**Findings:**

- **Credentials in Plaintext:** Authentication credentials, including the worker's username and password, are sent in plaintext over unencrypted connections. This makes them highly susceptible to theft.

- **Weak Password Hashing:** The `config.json` file stores the password as `"x"`, which is a weak default. While users can change this, there is no enforcement of strong password policies.

- **Risk of Brute-Force Attacks:** The absence of rate limiting or account lockout mechanisms on the pool side exposes the authentication process to brute-force attacks, where an attacker could repeatedly guess passwords until they gain access.

**Recommendations:**

- **Mandate Encrypted Authentication:** Require all authentication attempts to be transmitted over a TLS-encrypted connection.
- **Enforce Strong Password Policies:** Encourage the use of strong, unique passwords and consider implementing a password strength meter.
- **Implement Secure Authentication Mechanisms:** Support more secure authentication methods, such as token-based authentication, to reduce the reliance on passwords.

## 3. Share Validation

**Findings:**

- **Lack of Share Signatures:** Shares submitted by the miner are not cryptographically signed, making it possible for an attacker to intercept and modify them. This could lead to share stealing, where an attacker redirects a legitimate miner's work to their own account.

- **Vulnerability to Replay Attacks:** There is no protection against replay attacks, where an attacker could resubmit old shares. While pools may have some defenses, the client-side implementation lacks measures to prevent this.

**Recommendations:**

- **Implement Share Signing:** Introduce a mechanism for miners to sign each share with a unique key, allowing the pool to verify its authenticity.
- **Add Nonces to Shares:** Include a unique, single-use nonce in each share submission to prevent replay attacks.

## 4. Block Template Handling

**Findings:**

- **Centralized Block Template Distribution:** The miner relies on a centralized pool for block templates, creating a single point of failure. If the pool is compromised, an attacker could distribute malicious or inefficient templates.

- **No Template Validation:** The miner does not independently validate the block templates received from the pool. This makes it vulnerable to block withholding attacks, where a malicious pool sends invalid templates, causing the miner to waste resources on unsolvable blocks.

**Recommendations:**

- **Support Decentralized Template Distribution:** Explore decentralized alternatives for block template distribution to reduce reliance on a single pool.
- **Implement Independent Template Validation:** Add a mechanism for the miner to validate block templates against the network's consensus rules before starting to mine.

## 5. Network Message Validation

**Findings:**

- **Insufficient Input Sanitization:** Network messages from the pool are not consistently sanitized, creating a risk of denial-of-service attacks. An attacker could send malformed messages to crash the miner.

- **Lack of Rate Limiting:** The client does not implement rate limiting for incoming messages, making it vulnerable to flooding attacks that could overwhelm its resources.

**Recommendations:**

- **Implement Robust Message Sanitization:** Sanitize all incoming network messages to ensure they conform to the expected format and do not contain malicious payloads.
- **Add Rate Limiting:** Introduce rate limiting for incoming messages to prevent a single source from overwhelming the miner.
