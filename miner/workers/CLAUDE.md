# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

xmrig-workers is the source code for the http://workers.xmrig.info/ dashboard - a browser-based interface for monitoring and managing XMRig cryptocurrency miners. It's a pure frontend application (no backend) that connects directly to XMRig miners via their HTTP API.

**Note**: This project is not actively developed. See https://github.com/ludufre/xmworkers for an alternative.

## Build & Development Commands

```bash
npm install              # Install dependencies
npm run start            # Dev server with hot reload on http://localhost:8080
npm run build            # Production build to public/ directory
npm run dev              # One-time development build
npm run watch            # Watch mode (rebuilds on file changes)
```

## Architecture Overview

### State Management (Redux)

Store structure in `src/store/`:
- `workers`: `{ keys: [], values: {} }` - managed miners
- `settings`: interval, pagination state
- `modal`: current modal type and data
- `router`: React Router state

Action types defined in `src/constants/ActionTypes.js`. Reducers in `src/reducers/`.

### Data Flow

1. **No backend server** - workers list and settings persist to localStorage (`xmrig.workers`, `xmrig.settings`)
2. **Polling architecture** - Worker model (`src/app/models/Worker.js`) polls XMRig's `/1/summary` endpoint at configurable intervals (~10s default)
3. **HTTP client** (`src/app/Net.js`) handles Bearer token authentication for XMRig API

### Component Structure

```
src/
├── components/         # Presentational React components
│   ├── worker/        # Worker detail subcomponents (backends, config, etc.)
│   ├── modals/        # Modal dialogs (add/delete worker, export, etc.)
│   └── forms/         # Form components
├── containers/        # Redux-connected components
├── app/
│   ├── Workers.js     # Worker management (add/remove/sync with localStorage)
│   ├── models/Worker.js  # Worker model with polling logic
│   └── Net.js         # HTTP client with auth
└── reducers/          # Redux reducers
```

### Key Patterns

- **Container/Presentational separation**: Containers (`src/containers/`) handle Redux wiring, components (`src/components/`) handle UI
- **Immutable updates**: Uses `immutability-helper` for state mutations
- **Event system** (`src/app/events.js`): Simple EventEmitter for cross-module communication (e.g., settings changes trigger worker refresh)

### Routes

Defined in `src/routes.js`:
- `/` - Workers list
- `/worker/:id` - Worker detail (5 tabs: summary, backends, config)
- `/settings` - User settings
- `/import/:data` - Import configuration

## Build System

- **Webpack 5** with Babel transpilation
- Entry: `src/index.js` + `src/index.scss`
- Output: `public/assets/` with content hashing
- Production builds include minification and subresource integrity (SRI)

## Deployment

Copy `public/` directory to a web server. Example nginx config provided in `config/xmrig-workers.conf` with SPA routing support.
