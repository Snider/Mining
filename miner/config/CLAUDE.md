# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Install dependencies (required first time)
npm install -g bower
npm install
bower install

# Development server (http://127.0.0.1:8081 with hot reload)
npm start

# Production build (outputs to public/)
npm run build
```

## Architecture Overview

XMRig Config is a **pure client-side React/Redux SPA** for generating XMRig miner configurations. It runs entirely in the browser with no backend - all configuration stays local.

### Tech Stack
- **React 16.2** with **Redux 3.7** for state management
- **React Router 4** for client-side routing
- **Webpack 4** (bundling) + **Grunt** (CSS/asset processing)
- **LESS** for stylesheets, **Bootstrap 3** for UI

### Source Structure (`src/`)

```
src/
├── index.js          # App entry point
├── routes.js         # Main routing configuration
├── components/       # Presentational React components
│   ├── modals/       # Modal dialogs (add/edit/delete pools, threads, presets)
│   ├── misc/         # Misc settings sub-components
│   ├── network/      # Network/pool sub-components
│   └── start/        # Startup settings sub-components
├── containers/       # Redux-connected components
│   ├── xmrig/        # CPU miner containers
│   ├── amd/          # AMD miner containers (legacy)
│   ├── nvidia/       # NVIDIA miner containers (legacy)
│   └── proxy/        # Proxy containers
├── actions/          # Redux action creators
├── reducers/         # Redux reducers (config, modal, notification, presets)
├── store/            # Redux store setup (dev vs prod)
├── constants/        # Action types, modal types, product definitions
├── lib/              # Utilities (config generation, pool handling, serialization)
└── less/             # LESS stylesheets
```

### Redux State Shape

```javascript
{
  config: {
    xmrig: {...},        // CPU miner settings
    'xmrig-amd': {...},  // AMD miner (legacy)
    'xmrig-nvidia': {...}, // NVIDIA miner (legacy)
    proxy: {...}         // XMRig Proxy settings
  },
  notification: {...},   // Toast notifications
  modal: {...},          // Active modal state
  presets: {...},        // Saved configurations
  router: {...}          // React Router state
}
```

### Key Files

- **`src/lib/config.js`** (~12KB): Core config generation logic - serializes Redux state to XMRig JSON config and command-line args
- **`src/reducers/config.js`**: Largest reducer, handles all miner configuration state
- **`src/routes.js`**: Defines routes for each miner type (`/xmrig`, `/xmrig-amd`, `/xmrig-nvidia`, `/proxy`, `/presets`)

### Build Pipeline

1. **Development** (`npm start`): Grunt compiles LESS → Webpack dev server with HMR on :8081
2. **Production** (`npm run build`): Webpack production build → Grunt minifies CSS/JS → filerev hashes assets

### Deployment

Copy the `public/` directory to any static web server. Nginx config example in `config/xmrig-config.conf`.
