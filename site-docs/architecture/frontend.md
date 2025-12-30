# Frontend Architecture

The Angular frontend provides a modern, responsive dashboard for miner management.

## Technology Stack

- **Angular 20+** - Standalone components
- **Tailwind CSS** - Utility-first styling
- **Chart.js** - Hashrate visualization
- **xterm.js** - Terminal emulation for console

## Component Structure

```
ui/src/app/
├── app.ts                    # Root component
├── app.routes.ts             # Route definitions
├── app.config.ts             # App configuration
├── miner.service.ts          # API communication
├── node.service.ts           # P2P node service
│
├── components/
│   └── sidebar/              # Navigation sidebar
│
├── layouts/
│   └── main-layout/          # Page layout wrapper
│
├── pages/
│   ├── dashboard/            # Main monitoring view
│   ├── profiles/             # Profile management
│   ├── console/              # Terminal output
│   ├── workers/              # Running miners
│   ├── miners/               # Installation
│   ├── nodes/                # P2P management
│   └── pools/                # Pool information
│
├── dashboard.component.*     # Stats and charts
├── chart.component.*         # Hashrate chart
├── profile-*.component.*     # Profile CRUD
├── console.component.*       # Terminal
└── setup-wizard.component.*  # Initial setup
```

## Key Components

### Dashboard Component

Displays real-time mining statistics:

- **Stats Bar**: Hashrate, shares, uptime, pool, difficulty, workers
- **Chart**: Time-series hashrate visualization
- **Quick Stats**: Peak rate, efficiency, share time
- **Worker Selector**: Switch between multiple miners

### Chart Component

Chart.js-based hashrate visualization:

- Time range selector (5m, 15m, 1h, 6h, 24h)
- Real-time updates every 10 seconds
- Historical data from SQLite

### Console Component

Terminal emulator with:

- ANSI color support via ansi_up
- Auto-scroll toggle
- Stdin command input
- Worker selection dropdown

### Profile Components

- **List**: Card-based profile display with actions
- **Create**: Form for new profiles
- **Edit**: Modify existing profiles

## Services

### MinerService

Handles all API communication:

```typescript
@Injectable({ providedIn: 'root' })
export class MinerService {
    getMiners(): Observable<Miner[]>
    getSystemInfo(): Observable<SystemInfo>
    startMiner(profileId: string): Observable<any>
    stopMiner(name: string): Observable<any>
    getMinerStats(name: string): Observable<Stats>
    getMinerLogs(name: string): Observable<string[]>
    sendStdin(name: string, input: string): Observable<any>
    getHashrateHistory(name: string, range: string): Observable<Point[]>
    // ... profiles, installation, etc.
}
```

### NodeService

P2P node management:

```typescript
@Injectable({ providedIn: 'root' })
export class NodeService {
    getNodeInfo(): Observable<NodeInfo>
    getPeers(): Observable<Peer[]>
    addPeer(peer: PeerAdd): Observable<any>
    removePeer(id: string): Observable<any>
    pingPeer(id: string): Observable<PingResult>
}
```

## Routing

```typescript
export const routes: Routes = [
    { path: '', component: DashboardComponent },
    { path: 'profiles', component: ProfileListComponent },
    { path: 'profiles/new', component: ProfileCreateComponent },
    { path: 'profiles/:id/edit', component: ProfileEditComponent },
    { path: 'console', component: ConsoleComponent },
    { path: 'workers', component: WorkersComponent },
    { path: 'miners', component: MinersComponent },
    { path: 'nodes', component: NodesComponent },
    { path: 'pools', component: PoolsComponent },
    { path: 'admin', component: AdminComponent },
];
```

## Styling

### Design System

| Element | Style |
|---------|-------|
| Background | Dark slate (#0a0a12) |
| Cards | Slightly lighter slate |
| Accent | Cyan and lime |
| Text | White and gray variants |
| Success | Green |
| Error | Red |
| Warning | Yellow |

### Responsive Design

- Mobile-first approach
- Sidebar collapses on small screens
- Grid layouts adapt to viewport

## Build Output

The UI builds to a web component:

```bash
ng build
# Outputs: ui/dist/mbe-mining-dashboard.js
```

Can be embedded in any HTML page:

```html
<script src="mbe-mining-dashboard.js"></script>
<mbe-mining-dashboard></mbe-mining-dashboard>
```
