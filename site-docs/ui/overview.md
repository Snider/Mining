# UI Overview

The Mining Dashboard features a modern, responsive web interface built with Angular.

## Layout

```
┌────────────────────────────────────────────────────────┐
│  Logo    Stats Bar (Hashrate, Shares, etc.)    Workers │
├────────┬───────────────────────────────────────────────┤
│        │                                               │
│ Sidebar│              Main Content Area               │
│        │                                               │
│ ○ Dash │                                               │
│ ○ Work │                                               │
│ ○ Cons │                                               │
│ ○ Pool │                                               │
│ ○ Prof │                                               │
│ ○ Mine │                                               │
│ ○ Node │                                               │
│        │                                               │
├────────┴───────────────────────────────────────────────┤
│  Status: Mining Active                                 │
└────────────────────────────────────────────────────────┘
```

## Navigation

### Sidebar

| Page | Icon | Description |
|------|------|-------------|
| **Dashboard** | Chart | Main monitoring view |
| **Workers** | Server | Running miner instances |
| **Console** | Terminal | Live miner output |
| **Pools** | Globe | Connected pools |
| **Profiles** | Bookmark | Saved configurations |
| **Miners** | CPU | Install/manage miners |
| **Nodes** | Network | P2P peer management |

### Stats Bar

Always visible at the top:
- Hashrate
- Shares (accepted/rejected)
- Uptime
- Pool name
- Average difficulty
- Worker count

### Worker Selector

Dropdown to filter stats by:
- All Workers
- Individual miner

## Pages

### Dashboard
Real-time monitoring with hashrate charts and key metrics.

### Workers
Start/stop miners, view running instances.

### Console
Live terminal output with ANSI colors and command input.

### Pools
View connected mining pools (from running miners).

### Profiles
Create, edit, delete mining configurations.

### Miners
Install/uninstall miner software (XMRig, TT-Miner).

### Nodes
P2P peer management for multi-node setups.

## Design System

### Colors

| Color | Usage |
|-------|-------|
| **Cyan** (#06b6d4) | Primary/accent |
| **Lime** (#a3e635) | Success, active |
| **Red** (#ef4444) | Errors, rejected |
| **Purple** (#a855f7) | Difficulty stats |
| **Slate** (#1e293b) | Backgrounds |

### Typography

- **Sans-serif** - UI text (system fonts)
- **Monospace** - Stats, values, code

### Components

- **Cards** - Profile cards, stat cards
- **Tables** - Peer list, pool list
- **Forms** - Profile creation, settings
- **Charts** - Hashrate over time
- **Badges** - Miner type, status

## Responsive Design

The UI adapts to different screen sizes:

| Breakpoint | Layout |
|------------|--------|
| Desktop | Full sidebar + content |
| Tablet | Collapsible sidebar |
| Mobile | Hidden sidebar, hamburger menu |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `1-7` | Navigate to page |
| `/` | Focus search |
| `Esc` | Close modal |

## Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+
