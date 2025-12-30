# Mining Profiles

Profiles store your mining configurations for easy reuse.

![Profiles Page](../assets/screenshots/profiles.png)

## Creating a Profile

Click **New Profile** to create a new configuration:

![New Profile Form](../assets/screenshots/new-profile.png)

### Required Fields

| Field | Description |
|-------|-------------|
| **Profile Name** | A friendly name for this configuration |
| **Miner Type** | Select `xmrig` or `tt-miner` |
| **Pool Address** | Mining pool URL (e.g., `pool.supportxmr.com:3333`) |
| **Wallet Address** | Your cryptocurrency wallet address |

### Optional Settings

| Field | Default | Description |
|-------|---------|-------------|
| **TLS** | On | Encrypt pool connection |
| **Huge Pages** | On | Enable huge pages for XMRig (Linux) |
| **Threads** | Auto | Number of CPU threads |
| **Password** | x | Pool password (usually not needed) |

## Profile Cards

Each profile is displayed as a card showing:

- **Name** - Profile name
- **Miner type** - Badge showing xmrig/tt-miner
- **Pool** - Pool address
- **Wallet** - Truncated wallet address
- **Actions** - Start, Edit, Delete buttons

## Starting a Miner

Click **Start** on any profile card to launch the miner with that configuration.

!!! note "Multiple Instances"
    You can start the same miner type multiple times with different configurations. Each instance gets a unique name like `xmrig-123`.

## Editing Profiles

Click the **Edit** button (pencil icon) to modify a profile. Changes take effect on the next miner start.

## Deleting Profiles

Click the **Delete** button (trash icon) to remove a profile.

!!! warning
    This action cannot be undone. Running miners using this profile will continue running.

## Profile Storage

Profiles are stored in:
```
~/.config/lethean-desktop/mining_profiles.json
```

### JSON Format

```json
{
  "id": "uuid-here",
  "name": "My Mining Profile",
  "minerType": "xmrig",
  "config": {
    "pool": "pool.supportxmr.com:3333",
    "wallet": "4xxx...",
    "password": "x",
    "tls": true,
    "hugePages": true,
    "threads": 0,
    "algo": "",
    "devices": "",
    "intensity": 0,
    "cliArgs": ""
  }
}
```

## Advanced Configuration

### CPU Threads

Set `threads` to control CPU usage:
- `0` - Auto-detect (uses all available cores)
- `1-N` - Use exactly N threads

### GPU Devices (TT-Miner)

Set `devices` to specify which GPUs to use:
- `""` - Use all GPUs
- `"0,1"` - Use GPU 0 and 1
- `"0"` - Use only GPU 0

### Extra CLI Arguments

Use `cliArgs` to pass additional arguments directly to the miner:

```json
{
  "cliArgs": "--cpu-priority 2 --randomx-1gb-pages"
}
```

## API Endpoints

```
GET    /api/v1/mining/profiles           # List all profiles
POST   /api/v1/mining/profiles           # Create profile
PUT    /api/v1/mining/profiles/{id}      # Update profile
DELETE /api/v1/mining/profiles/{id}      # Delete profile
POST   /api/v1/mining/profiles/{id}/start # Start miner with profile
```
