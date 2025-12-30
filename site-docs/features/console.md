# Console & Logs

The Console page provides live access to miner output with ANSI color support and interactive input.

![Console](../assets/screenshots/console.png)

## Features

### Live Output

- **Real-time streaming** - Miner output appears as it happens
- **ANSI color support** - Full color rendering for miner output
- **Auto-scroll** - Automatically scrolls to latest output
- **Base64 transport** - Preserves special characters and escape codes

### Worker Selection

Use the dropdown to select which miner's output to view:

- Lists all running miners
- Defaults to the first running miner
- Switching miners loads their log history

### Command Input

The input field at the bottom allows sending commands to the miner's stdin:

```
> h
```

#### XMRig Commands

| Key | Command |
|-----|---------|
| `h` | Print hashrate |
| `p` | Pause mining |
| `r` | Resume mining |
| `s` | Print results/shares |
| `c` | Print connection info |

Press **Enter** to send the command.

### Auto-scroll Toggle

Toggle auto-scroll to:
- **On** - Automatically scroll to new output
- **Off** - Stay at current position for reading history

### Clear Button

Click **Clear** to empty the console display. This only clears the UI; the backend still retains the full log.

## Color Support

The console renders ANSI escape codes as HTML colors:

| Color | Usage |
|-------|-------|
| **Green** | Success messages, accepted shares |
| **Red** | Errors, rejected shares |
| **Cyan** | Values, hashrates |
| **Yellow** | Warnings |
| **Magenta** | Special messages |
| **White** | Normal text |

### Example Output

```
[2024-01-15 10:30:45] speed 10s/60s/15m 1234.5 1230.2 1228.8 H/s max 1250.1 H/s
[2024-01-15 10:30:50] accepted (1/0) diff 10000 (238 ms)
```

## Log Buffer

Each miner maintains a circular log buffer:

- **Size**: Last 1000 lines
- **Persistence**: Cleared when miner stops
- **Encoding**: Base64 to preserve special characters

## Implementation Details

### Backend

Logs are captured via:
1. Miner stdout/stderr is piped to a `LogBuffer`
2. Lines are base64 encoded for JSON transport
3. Retrieved via GET `/miners/{name}/logs`

### Frontend

The Angular component:
1. Fetches logs via HTTP
2. Decodes base64 to text
3. Converts ANSI escape codes to HTML spans
4. Renders with appropriate CSS colors

### Stdin

Commands are sent via:
1. POST `/miners/{name}/stdin` with `{"input": "h"}`
2. Backend writes to miner's stdin pipe
3. Response appears in log output

## API Endpoints

```
GET  /api/v1/mining/miners/{name}/logs   # Get log output (base64 encoded)
POST /api/v1/mining/miners/{name}/stdin  # Send stdin input
```

### Example: Get Logs

```bash
curl http://localhost:9090/api/v1/mining/miners/xmrig-123/logs
```

Response:
```json
[
  "W1hNUmlnXSBzcGVlZCAxMHMvNjBzLzE1bSAxMjM0LjUgMTIzMC4yIDEyMjguOCBIL3M=",
  "W1hNUmlnXSBhY2NlcHRlZCAoMS8wKSBkaWZmIDEwMDAw"
]
```

### Example: Send Command

```bash
curl -X POST http://localhost:9090/api/v1/mining/miners/xmrig-123/stdin \
  -H "Content-Type: application/json" \
  -d '{"input": "h"}'
```

Response:
```json
{"status": "sent", "input": "h"}
```
