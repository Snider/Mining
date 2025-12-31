# Web Dashboard User Guide

The Mining Platform web dashboard provides a visual interface for monitoring and managing your mining operations.

## Overview

The web dashboard is an Angular-based web component that can be:
- Accessed via the built-in server
- Embedded in any web application
- Used standalone in a browser

## Accessing the Dashboard

### Via Built-in Server

Start the Mining Platform server:

```bash
miner-ctrl serve --port 9090
```

Then open your browser to:
```
http://localhost:9090
```

### Embedding in Your Application

The dashboard is available as a standalone web component:

```html
<!DOCTYPE html>
<html>
<head>
  <title>Mining Dashboard</title>
  <script type="module" src="./mbe-mining-dashboard.js"></script>
</head>
<body>
  <snider-mining api-base-url="http://localhost:9090/api/v1/mining"></snider-mining>
</body>
</html>
```

### Component Properties

The web component accepts these attributes:

- `api-base-url`: Base URL for the API (required)
- `theme`: UI theme (`light` or `dark`)
- `auto-refresh`: Auto-refresh interval in seconds (default: 10)
- `locale`: Locale for number formatting (default: `en-US`)

Example with all properties:

```html
<snider-mining
  api-base-url="http://localhost:9090/api/v1/mining"
  theme="dark"
  auto-refresh="5"
  locale="en-US">
</snider-mining>
```

## Dashboard Features

### Home Page

The home page displays an overview of your mining operations:

- **Active Miners**: Number of currently running miners
- **Total Hashrate**: Combined hashrate across all miners
- **Total Shares**: Accepted and rejected shares
- **System Resources**: CPU and GPU usage

### Miners Page

View and manage all your miners:

- **Running Miners**: List of active mining operations
  - Real-time hashrate
  - Pool connection status
  - Accepted/rejected shares
  - Uptime
  - Quick stop button

- **Available Miners**: Software you can install
  - Installation status
  - Version information
  - Quick install button

### Profiles Page

Manage your mining configurations:

- **Saved Profiles**: Reusable mining configurations
  - Create, edit, and delete profiles
  - One-click start from profile
  - Import/export profiles

- **Profile Editor**: Configure mining parameters
  - Pool selection
  - Wallet address
  - Algorithm selection
  - CPU/GPU settings
  - Advanced options

### Setup Wizard

A guided setup process for beginners:

1. **Select Coin**: Choose which cryptocurrency to mine
2. **Choose Pool**: Select from recommended pools
3. **Enter Wallet**: Input your wallet address
4. **Configure Hardware**: Select CPU or GPU mining
5. **Review & Start**: Confirm settings and start mining

### Admin Page

Advanced configuration and system management:

- **System Information**
  - Operating system
  - Go version
  - Total RAM
  - Detected GPUs (OpenCL/CUDA)

- **Miner Management**
  - Install/uninstall miners
  - Update miners
  - View logs
  - Run diagnostics

- **Settings**
  - Auto-start configuration
  - Notification preferences
  - API endpoint configuration
  - Theme selection

### Statistics Dashboard

Detailed performance metrics:

- **Hashrate Charts**
  - Real-time hashrate graph
  - Historical data (5 min high-res, 24h low-res)
  - Per-miner breakdown

- **Shares Analysis**
  - Accepted vs rejected shares
  - Share submission rate
  - Pool response time

- **Earnings Estimates**
  - Estimated daily earnings
  - Estimated monthly earnings
  - Based on current hashrate and coin difficulty

## Mobile Interface

The dashboard is fully responsive and optimized for mobile devices:

- **Drawer Navigation**: Swipe from left edge to open menu
- **Touch-Optimized**: Large buttons and touch targets
- **Adaptive Layout**: Single-column layout on small screens
- **Pull to Refresh**: Pull down on any page to refresh data

### Mobile Navigation

On mobile devices (screens < 768px):

- Tap the hamburger menu (☰) to open navigation
- Swipe left to close the drawer
- Tap outside the drawer to close it

## Using the Dashboard

### Starting a Miner

1. Go to **Miners** page
2. Click **Add Miner** or select a profile
3. Fill in the configuration:
   - Pool URL
   - Wallet address
   - Algorithm
   - Number of threads (CPU) or devices (GPU)
4. Click **Start Mining**

### Monitoring Performance

1. Go to **Statistics** page
2. Select the miner from the dropdown
3. View real-time charts:
   - Hashrate over time
   - Share acceptance rate
   - Temperature (if supported)

### Creating a Profile

1. Go to **Profiles** page
2. Click **New Profile**
3. Enter profile details:
   - Name (e.g., "XMR - SupportXMR")
   - Miner type (xmrig, etc.)
   - Pool configuration
4. Click **Save**

To use the profile:
1. Click **Start** on the profile card
2. The miner will start with the saved configuration

### Stopping a Miner

1. Go to **Miners** page
2. Find the running miner
3. Click the **Stop** button
4. Confirm if prompted

## Keyboard Shortcuts

The dashboard supports keyboard shortcuts for common actions:

- `Ctrl+N`: New miner/profile
- `Ctrl+S`: Save current form
- `Ctrl+R`: Refresh data
- `Esc`: Close modal/drawer
- `?`: Show keyboard shortcuts help

## Theme Customization

Switch between light and dark themes:

1. Go to **Settings** (Admin page)
2. Select **Theme**
3. Choose **Light** or **Dark**
4. Theme is saved to localStorage

Or set via HTML attribute:

```html
<snider-mining theme="dark"></snider-mining>
```

## Auto-Refresh

The dashboard automatically refreshes data every 10 seconds by default.

To customize:

```html
<snider-mining auto-refresh="5"></snider-mining>
```

To disable auto-refresh:

```html
<snider-mining auto-refresh="0"></snider-mining>
```

## Notifications

The dashboard can display browser notifications for important events:

- Miner started
- Miner stopped
- Miner crashed
- Low hashrate warning
- Share rejection spike

Enable notifications:
1. Go to **Settings**
2. Enable **Desktop Notifications**
3. Grant permission when prompted by browser

## Exporting Data

Export your profiles or statistics:

### Export Profiles

1. Go to **Profiles** page
2. Click **Export**
3. Choose format (JSON or CSV)
4. Save file

### Import Profiles

1. Go to **Profiles** page
2. Click **Import**
3. Select your exported file
4. Profiles will be added to your list

### Export Statistics

1. Go to **Statistics** page
2. Select date range
3. Click **Export**
4. Choose format (CSV or JSON)
5. Save file

## Troubleshooting

### Dashboard Won't Load

Check that the server is running:

```bash
curl http://localhost:9090/api/v1/mining/info
```

If no response, start the server:

```bash
miner-ctrl serve --port 9090
```

### "Connection Refused" Error

Ensure the `api-base-url` matches your server:

```html
<snider-mining api-base-url="http://localhost:9090/api/v1/mining"></snider-mining>
```

### Data Not Updating

1. Check auto-refresh is enabled
2. Verify API endpoint is reachable
3. Check browser console for errors (F12)
4. Try manual refresh (Ctrl+R)

### Profile Won't Start

1. Verify miner software is installed
2. Check wallet address is valid
3. Test pool connectivity
4. Review error message in notification

### Charts Not Showing

1. Ensure miner has been running for at least 1 minute
2. Check that statistics are being collected
3. Verify browser supports Canvas/SVG
4. Try clearing browser cache

## Performance Tips

### Reduce CPU Usage

If the dashboard is using too much CPU:

1. Increase auto-refresh interval:
   ```html
   <snider-mining auto-refresh="30"></snider-mining>
   ```

2. Disable charts on Statistics page when not needed

3. Close unnecessary browser tabs

### Optimize for Low-End Devices

For Raspberry Pi or low-power devices:

1. Use light theme (uses less GPU)
2. Set auto-refresh to 60 seconds
3. Limit number of active miners shown
4. Disable desktop notifications

## Advanced Usage

### Custom Styling

Override dashboard styles with CSS:

```html
<style>
  snider-mining {
    --primary-color: #00ff00;
    --background-color: #1a1a1a;
    --text-color: #ffffff;
  }
</style>
```

### JavaScript API

Interact with the component via JavaScript:

```javascript
const dashboard = document.querySelector('snider-mining');

// Listen for events
dashboard.addEventListener('miner-started', (e) => {
  console.log('Miner started:', e.detail);
});

dashboard.addEventListener('miner-stopped', (e) => {
  console.log('Miner stopped:', e.detail);
});

// Programmatic control
dashboard.startMiner({
  type: 'xmrig',
  config: { /* ... */ }
});

dashboard.stopMiner('xmrig');
```

## Next Steps

- Try the [Desktop Application](desktop-app.md) for a native experience
- Learn about [Pool Selection](../reference/pools.md)
- Explore the [REST API](../api/endpoints.md) for automation
- Read the [Development Guide](../development/index.md) to contribute
