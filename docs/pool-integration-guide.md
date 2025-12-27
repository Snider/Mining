# Pool Integration Guide for Mining UI

## Quick Start: Adding Pool Support to Your Miner

This guide provides code examples for integrating XMR pool data into your mining application.

---

## Part 1: TypeScript/JavaScript Implementation

### Loading Pool Database

```typescript
import poolDatabase from './xmr-pools-database.json';

interface MiningPool {
  id: string;
  name: string;
  website: string;
  fee_percent: number;
  minimum_payout_xmr: number;
  stratum_servers: StratumServer[];
  authentication: AuthConfig;
}

interface StratumServer {
  region_id: string;
  region_name: string;
  hostname: string;
  ports: PoolPort[];
}

interface PoolPort {
  port: number;
  difficulty: string;
  protocol: "stratum+tcp" | "stratum+ssl";
  description: string;
}

interface AuthConfig {
  username_format: string;
  password_default: string;
  registration_required: boolean;
}

interface ConnectionConfig {
  url: string;
  username: string;
  password: string;
  pool_name: string;
  pool_fee: number;
}

// Load pools
const pools: MiningPool[] = poolDatabase.pools;

// Find a specific pool
function getPool(poolId: string): MiningPool | undefined {
  return pools.find(p => p.id === poolId);
}

// Get all pools sorted by recommendation
function getRecommendedPools(userType: 'beginner' | 'advanced' | 'solo'): MiningPool[] {
  const recommendedIds = poolDatabase.recommended_pools[userType + 's'];
  return recommendedIds.map(id => getPool(id)).filter(p => p !== undefined);
}
```

### Connection String Generator

```typescript
class PoolConnector {
  /**
   * Generate complete connection configuration for a mining pool
   */
  static generateConnectionConfig(
    poolId: string,
    walletAddress: string,
    workerName: string = "default",
    preferTls: boolean = false,
    difficulty: 'standard' | 'medium' | 'high' = 'standard'
  ): ConnectionConfig {
    const pool = getPool(poolId);
    if (!pool) throw new Error(`Pool ${poolId} not found`);

    // Select primary stratum server (usually first region)
    const stratumServer = pool.stratum_servers[0];

    // Find port matching difficulty preference
    let selectedPort = stratumServer.ports[0]; // Default to first (usually standard)

    if (difficulty === 'medium' && stratumServer.ports.length > 1) {
      selectedPort = stratumServer.ports.find(p => p.difficulty === 'medium') || stratumServer.ports[0];
    } else if (difficulty === 'high' && stratumServer.ports.length > 2) {
      selectedPort = stratumServer.ports.find(p => p.difficulty === 'high') || stratumServer.ports[0];
    }

    // Use TLS if preferred and available
    if (preferTls) {
      const tlsPort = stratumServer.ports.find(p => p.protocol === 'stratum+ssl');
      if (tlsPort) selectedPort = tlsPort;
    }

    // Build connection URL
    const url = `${selectedPort.protocol}://${stratumServer.hostname}:${selectedPort.port}`;

    // Build username (most pools use wallet.worker format)
    const username = `${walletAddress}.${workerName}`;

    return {
      url,
      username,
      password: pool.authentication.password_default,
      pool_name: pool.name,
      pool_fee: pool.fee_percent
    };
  }

  /**
   * Test connection to a pool
   */
  static async testConnection(config: ConnectionConfig, timeoutMs: number = 5000): Promise<boolean> {
    try {
      const urlObj = new URL(config.url);
      const hostname = urlObj.hostname;
      const port = parseInt(urlObj.port);

      return new Promise((resolve) => {
        const socket = new net.Socket();
        const timeout = setTimeout(() => {
          socket.destroy();
          resolve(false);
        }, timeoutMs);

        socket.connect(port, hostname, () => {
          clearTimeout(timeout);
          socket.destroy();
          resolve(true);
        });

        socket.on('error', () => {
          clearTimeout(timeout);
          resolve(false);
        });
      });
    } catch (error) {
      return false;
    }
  }

  /**
   * Get fallback pool if primary is unavailable
   */
  static async findWorkingPool(
    poolIds: string[],
    walletAddress: string
  ): Promise<ConnectionConfig | null> {
    for (const poolId of poolIds) {
      const config = this.generateConnectionConfig(poolId, walletAddress);
      if (await this.testConnection(config)) {
        return config;
      }
    }
    return null;
  }
}

// Usage examples:
const config = PoolConnector.generateConnectionConfig(
  'supportxmr',
  '4ABC1234567890ABCDEF...',
  'miner1',
  false,
  'standard'
);

console.log(`Pool URL: ${config.url}`);
console.log(`Username: ${config.username}`);
console.log(`Password: ${config.password}`);

// Test connection
const isConnected = await PoolConnector.testConnection(config);
console.log(`Pool online: ${isConnected}`);

// Find working pool from list
const workingConfig = await PoolConnector.findWorkingPool(
  ['supportxmr', 'nanopool', 'moneroocean'],
  walletAddress
);
```

### React Component: Pool Selector

```typescript
// PoolSelector.tsx
import React, { useState, useEffect } from 'react';
import poolDatabase from './xmr-pools-database.json';

interface PoolSelectorProps {
  onPoolSelect: (config: ConnectionConfig) => void;
  walletAddress: string;
  userType?: 'beginner' | 'advanced' | 'solo';
}

export const PoolSelector: React.FC<PoolSelectorProps> = ({
  onPoolSelect,
  walletAddress,
  userType = 'beginner'
}) => {
  const [selectedPoolId, setSelectedPoolId] = useState('supportxmr');
  const [selectedDifficulty, setSelectedDifficulty] = useState('standard');
  const [useTls, setUseTls] = useState(false);
  const [connectionConfig, setConnectionConfig] = useState<ConnectionConfig | null>(null);

  const recommendedPools = poolDatabase.recommended_pools[userType + 's'];
  const availablePools = poolDatabase.pools.filter(p =>
    recommendedPools.includes(p.id)
  );

  useEffect(() => {
    const config = PoolConnector.generateConnectionConfig(
      selectedPoolId,
      walletAddress,
      'default',
      useTls,
      selectedDifficulty as any
    );
    setConnectionConfig(config);
  }, [selectedPoolId, useTls, selectedDifficulty]);

  const handleConnect = () => {
    if (connectionConfig) {
      onPoolSelect(connectionConfig);
    }
  };

  return (
    <div className="pool-selector">
      <h2>Mining Pool Configuration</h2>

      <div className="form-group">
        <label>Select Pool:</label>
        <select value={selectedPoolId} onChange={(e) => setSelectedPoolId(e.target.value)}>
          {availablePools.map(pool => (
            <option key={pool.id} value={pool.id}>
              {pool.name} - {pool.fee_percent}% fee (Min payout: {pool.minimum_payout_xmr} XMR)
            </option>
          ))}
        </select>
      </div>

      <div className="form-group">
        <label>Difficulty Level:</label>
        <select value={selectedDifficulty} onChange={(e) => setSelectedDifficulty(e.target.value)}>
          <option value="standard">Standard (Auto-adjust)</option>
          <option value="medium">Medium</option>
          <option value="high">High (Powerful miners)</option>
        </select>
      </div>

      <div className="form-group">
        <label>
          <input
            type="checkbox"
            checked={useTls}
            onChange={(e) => setUseTls(e.target.checked)}
          />
          Use TLS/SSL Encryption
        </label>
      </div>

      {connectionConfig && (
        <div className="connection-preview">
          <h3>Connection Details:</h3>
          <code>
            <div>URL: {connectionConfig.url}</div>
            <div>Username: {connectionConfig.username}</div>
            <div>Password: {connectionConfig.password}</div>
          </code>
          <button onClick={handleConnect} className="btn-primary">
            Connect to {connectionConfig.pool_name}
          </button>
        </div>
      )}
    </div>
  );
};
```

---

## Part 2: Go Implementation

### Go Structs and Functions

```go
package mining

import (
  "encoding/json"
  "fmt"
  "net"
  "time"
)

type PoolPort struct {
  Port        int    `json:"port"`
  Difficulty  string `json:"difficulty"`
  Protocol    string `json:"protocol"`
  Description string `json:"description"`
}

type StratumServer struct {
  RegionID   string     `json:"region_id"`
  RegionName string     `json:"region_name"`
  Hostname   string     `json:"hostname"`
  Ports      []PoolPort `json:"ports"`
}

type AuthConfig struct {
  UsernameFormat      string `json:"username_format"`
  PasswordDefault     string `json:"password_default"`
  RegistrationRequired bool   `json:"registration_required"`
}

type MiningPool struct {
  ID                string           `json:"id"`
  Name              string           `json:"name"`
  Website           string           `json:"website"`
  Description       string           `json:"description"`
  FeePercent        float64          `json:"fee_percent"`
  MinimumPayoutXMR  float64          `json:"minimum_payout_xmr"`
  StratumServers    []StratumServer  `json:"stratum_servers"`
  Authentication    AuthConfig       `json:"authentication"`
  LastVerified      string           `json:"last_verified"`
  ReliabilityScore  float64          `json:"reliability_score"`
  Recommended       bool             `json:"recommended"`
}

type ConnectionConfig struct {
  URL      string
  Username string
  Password string
  PoolName string
  PoolFee  float64
}

type PoolDatabase struct {
  Pools            []MiningPool `json:"pools"`
  RecommendedPools map[string][]string `json:"recommended_pools"`
}

// LoadPoolDatabase loads pools from JSON file
func LoadPoolDatabase(filePath string) (*PoolDatabase, error) {
  data, err := ioutil.ReadFile(filePath)
  if err != nil {
    return nil, err
  }

  var db PoolDatabase
  if err := json.Unmarshal(data, &db); err != nil {
    return nil, err
  }

  return &db, nil
}

// GetPool retrieves a pool by ID
func (db *PoolDatabase) GetPool(poolID string) *MiningPool {
  for i := range db.Pools {
    if db.Pools[i].ID == poolID {
      return &db.Pools[i]
    }
  }
  return nil
}

// GenerateConnectionConfig creates a connection configuration
func GenerateConnectionConfig(
  pool *MiningPool,
  walletAddress string,
  workerName string,
  useTLS bool,
  difficulty string,
) *ConnectionConfig {
  if pool == nil || len(pool.StratumServers) == 0 {
    return nil
  }

  server := pool.StratumServers[0]
  if len(server.Ports) == 0 {
    return nil
  }

  // Select port based on difficulty
  selectedPort := server.Ports[0]

  for _, port := range server.Ports {
    if port.Difficulty == difficulty {
      if !useTLS && port.Protocol == "stratum+tcp" {
        selectedPort = port
        break
      } else if useTLS && port.Protocol == "stratum+ssl" {
        selectedPort = port
        break
      }
    }
  }

  // If TLS requested but not found, look for TLS port
  if useTLS {
    for _, port := range server.Ports {
      if port.Protocol == "stratum+ssl" {
        selectedPort = port
        break
      }
    }
  }

  url := fmt.Sprintf("%s://%s:%d",
    selectedPort.Protocol,
    server.Hostname,
    selectedPort.Port,
  )

  username := fmt.Sprintf("%s.%s", walletAddress, workerName)

  return &ConnectionConfig{
    URL:      url,
    Username: username,
    Password: pool.Authentication.PasswordDefault,
    PoolName: pool.Name,
    PoolFee:  pool.FeePercent,
  }
}

// TestConnection tests if a pool is reachable
func TestConnection(hostname string, port int, timeoutSecs int) bool {
  address := fmt.Sprintf("%s:%d", hostname, port)
  conn, err := net.DialTimeout("tcp", address, time.Duration(timeoutSecs)*time.Second)
  if err != nil {
    return false
  }
  defer conn.Close()
  return true
}

// FindWorkingPool attempts to connect to multiple pools and returns first working one
func (db *PoolDatabase) FindWorkingPool(
  poolIDs []string,
  walletAddress string,
  timeoutSecs int,
) *ConnectionConfig {
  for _, poolID := range poolIDs {
    pool := db.GetPool(poolID)
    if pool == nil {
      continue
    }

    if len(pool.StratumServers) == 0 || len(pool.StratumServers[0].Ports) == 0 {
      continue
    }

    server := pool.StratumServers[0]
    port := server.Ports[0]

    if TestConnection(server.Hostname, port.Port, timeoutSecs) {
      return GenerateConnectionConfig(pool, walletAddress, "default", false, "standard")
    }
  }

  return nil
}

// GetRecommendedPools returns pools recommended for user type
func (db *PoolDatabase) GetRecommendedPools(userType string) []*MiningPool {
  poolIDs := db.RecommendedPools[userType+"s"]
  var pools []*MiningPool

  for _, id := range poolIDs {
    if pool := db.GetPool(id); pool != nil {
      pools = append(pools, pool)
    }
  }

  return pools
}
```

### Go Usage Examples

```go
package main

import (
  "fmt"
  "log"
)

func main() {
  // Load pool database
  db, err := LoadPoolDatabase("xmr-pools-database.json")
  if err != nil {
    log.Fatal("Failed to load pool database:", err)
  }

  walletAddress := "4ABC1234567890ABCDEF..."

  // Example 1: Get recommended pools for beginners
  recommendedPools := db.GetRecommendedPools("beginner")
  fmt.Println("Recommended pools for beginners:")
  for _, pool := range recommendedPools {
    fmt.Printf("  - %s (%.1f%% fee)\n", pool.Name, pool.FeePercent)
  }

  // Example 2: Generate connection config
  pool := db.GetPool("supportxmr")
  config := GenerateConnectionConfig(pool, walletAddress, "miner1", false, "standard")
  fmt.Printf("\nConnection Config:\n")
  fmt.Printf("  URL: %s\n", config.URL)
  fmt.Printf("  Username: %s\n", config.Username)
  fmt.Printf("  Password: %s\n", config.Password)

  // Example 3: Test connection
  isOnline := TestConnection("pool.supportxmr.com", 3333, 5)
  fmt.Printf("Pool online: %v\n", isOnline)

  // Example 4: Find first working pool
  poolIDs := []string{"supportxmr", "nanopool", "moneroocean"}
  workingConfig := db.FindWorkingPool(poolIDs, walletAddress, 5)
  if workingConfig != nil {
    fmt.Printf("\nWorking pool found: %s\n", workingConfig.PoolName)
    fmt.Printf("URL: %s\n", workingConfig.URL)
  }
}
```

---

## Part 3: Configuration Storage

### Saving User Pool Selection

```typescript
// Save to localStorage
function savePoolPreference(poolId: string, walletAddress: string) {
  localStorage.setItem('preferred_pool', poolId);
  localStorage.setItem('wallet_address', walletAddress);
}

// Load from localStorage
function loadPoolPreference(): { poolId: string; walletAddress: string } | null {
  const poolId = localStorage.getItem('preferred_pool');
  const walletAddress = localStorage.getItem('wallet_address');

  if (poolId && walletAddress) {
    return { poolId, walletAddress };
  }
  return null;
}
```

### Persisting to Config File (Go)

```go
type UserConfig struct {
  PreferredPoolID string `json:"preferred_pool_id"`
  WalletAddress   string `json:"wallet_address"`
  WorkerName      string `json:"worker_name"`
  UseTLS          bool   `json:"use_tls"`
  Difficulty      string `json:"difficulty"`
  LastUpdated     string `json:"last_updated"`
}

func SaveUserConfig(filePath string, config *UserConfig) error {
  config.LastUpdated = time.Now().Format(time.RFC3339)

  data, err := json.MarshalIndent(config, "", "  ")
  if err != nil {
    return err
  }

  return ioutil.WriteFile(filePath, data, 0644)
}

func LoadUserConfig(filePath string) (*UserConfig, error) {
  data, err := ioutil.ReadFile(filePath)
  if err != nil {
    return nil, err
  }

  var config UserConfig
  if err := json.Unmarshal(data, &config); err != nil {
    return nil, err
  }

  return &config, nil
}
```

---

## Part 4: UI Components

### Pool List Display

```typescript
// Display pool information with comparison
function PoolComparison() {
  const pools = poolDatabase.pools;

  return (
    <table className="pool-comparison">
      <thead>
        <tr>
          <th>Pool Name</th>
          <th>Fee</th>
          <th>Min Payout</th>
          <th>Reliability</th>
          <th>Recommended</th>
        </tr>
      </thead>
      <tbody>
        {pools.map(pool => (
          <tr key={pool.id}>
            <td>
              <a href={pool.website} target="_blank">
                {pool.name}
              </a>
            </td>
            <td>{pool.fee_percent}%</td>
            <td>{pool.minimum_payout_xmr} XMR</td>
            <td>
              <ProgressBar value={pool.reliability_score * 100} max={100} />
            </td>
            <td>{pool.recommended ? '✓' : '-'}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
```

### Connection String Copy-to-Clipboard

```typescript
function ConnectionDisplay({ config }: { config: ConnectionConfig }) {
  const [copied, setCopied] = useState(false);

  const connectionString = `${config.url}\n${config.username}\n${config.password}`;

  const handleCopy = () => {
    navigator.clipboard.writeText(connectionString);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="connection-display">
      <pre>{connectionString}</pre>
      <button onClick={handleCopy}>
        {copied ? 'Copied!' : 'Copy to Clipboard'}
      </button>
    </div>
  );
}
```

---

## Part 5: Validation & Error Handling

### Wallet Address Validation

```typescript
// XMR address validation
function validateXMRAddress(address: string): boolean {
  // Standard Monero address
  // - 95 characters long
  // - Starts with 4 (mainnet) or 8 (stagenet/testnet)
  // - Base58 characters only

  const base58Regex = /^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$/;

  return (
    (address.startsWith('4') || address.startsWith('8')) &&
    address.length === 95 &&
    base58Regex.test(address)
  );
}

function validatePoolConfiguration(config: ConnectionConfig): ValidationResult {
  const errors: string[] = [];

  if (!config.url) errors.push('Pool URL required');
  if (!config.username) errors.push('Username required');
  if (!config.url.includes('://')) errors.push('Invalid protocol format');

  return {
    isValid: errors.length === 0,
    errors
  };
}
```

---

## Part 6: Migration Guide

If you have existing hardcoded pool configs:

```typescript
// OLD CODE (hardcoded):
const poolConfig = {
  url: 'stratum+tcp://pool.supportxmr.com:3333',
  username: 'wallet.worker',
  password: 'x'
};

// NEW CODE (from database):
const poolId = 'supportxmr';
const pool = poolDatabase.pools.find(p => p.id === poolId);
const config = PoolConnector.generateConnectionConfig(
  poolId,
  'wallet_address',
  'worker',
  false,
  'standard'
);
```

---

## Summary

Your mining UI can now:

1. Load pools from the JSON database
2. Display pool selection interface
3. Generate connection strings dynamically
4. Validate pool connectivity
5. Save user preferences
6. Suggest fallback pools
7. Support both TCP and TLS connections
8. Auto-detect optimal difficulty levels

This approach makes it easy to:
- Update pools without code changes
- Add new pools instantly
- Validate connection details
- Scale to other cryptocurrencies
