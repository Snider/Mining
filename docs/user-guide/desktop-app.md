# Desktop Application User Guide

The Mining Platform desktop application provides a native cross-platform experience built with Wails v3 and Angular.

## Overview

The desktop app combines the power of the Mining Platform backend with a modern desktop interface, offering:

- Native application performance
- System tray integration
- Auto-start on boot
- Local-first operation
- Embedded web dashboard
- No browser required

## Installation

### Download Pre-built Application

Download the appropriate version for your platform from the [Releases page](https://github.com/Snider/Mining/releases):

**Linux:**
- `mining-dashboard-linux-amd64` (standalone binary)
- `mining-dashboard_amd64.deb` (Debian/Ubuntu)
- `mining-dashboard_amd64.rpm` (Fedora/RHEL)

**macOS:**
- `mining-dashboard.dmg` (drag and drop installer)
- `mining-dashboard.app.zip` (extract and run)

**Windows:**
- `mining-dashboard-setup.exe` (installer)
- `mining-dashboard-portable.exe` (no installation required)

### Linux Installation

#### Using DEB Package
```bash
sudo dpkg -i mining-dashboard_amd64.deb
```

#### Using RPM Package
```bash
sudo rpm -i mining-dashboard_amd64.rpm
```

#### Standalone Binary
```bash
chmod +x mining-dashboard-linux-amd64
./mining-dashboard-linux-amd64
```

### macOS Installation

1. Open the DMG file
2. Drag Mining Dashboard to Applications
3. Open from Applications folder
4. If blocked by Gatekeeper:
   - Go to System Preferences → Security & Privacy
   - Click "Open Anyway"

### Windows Installation

#### Using Installer
1. Run `mining-dashboard-setup.exe`
2. Follow installation wizard
3. Launch from Start Menu

#### Portable Version
1. Extract `mining-dashboard-portable.exe`
2. Run directly (no installation needed)
3. Settings saved in same folder

## First Launch

When you first launch the desktop app:

1. **Welcome Screen**: Brief introduction to features
2. **Setup Wizard**: Optional guided setup
   - Install mining software
   - Configure first miner
   - Select pools
3. **Main Dashboard**: Ready to use

## User Interface

### Main Window

The main window contains:

**Menu Bar** (top):
- File: New profile, Import/Export, Preferences, Exit
- Miners: Install, Start, Stop, Update
- View: Refresh, Toggle Fullscreen, Developer Tools
- Help: Documentation, Check for Updates, About

**Sidebar** (left):
- Dashboard: Overview and quick stats
- Miners: Running and available miners
- Profiles: Saved configurations
- Statistics: Charts and analytics
- Pools: Pool information and recommendations
- Admin: System settings and diagnostics
- Settings: Application preferences

**Content Area** (center):
- Page-specific content
- Real-time data updates
- Interactive controls

**Status Bar** (bottom):
- Connection status
- Active miners count
- Total hashrate
- Last update time

### System Tray

The app runs in the system tray when minimized:

**Tray Icon**: Shows mining status
- Green: Mining active
- Gray: Idle
- Red: Error/stopped

**Tray Menu** (right-click):
- Show/Hide window
- Quick start mining
- Pause all miners
- Exit application

### Dashboard Page

The dashboard provides an at-a-glance view:

- **Active Miners Card**: Count and quick actions
- **Hashrate Card**: Total and per-miner breakdown
- **Shares Card**: Accepted/rejected statistics
- **Earnings Card**: Estimated daily/monthly earnings
- **Recent Activity**: Timeline of events
- **Quick Actions**: Common tasks

### Miners Page

Manage your mining operations:

**Running Miners:**
- List of active miners with status
- Real-time hashrate and statistics
- Stop/pause controls
- View detailed logs

**Available Miners:**
- Installable mining software
- Version information
- Install/update buttons

**Actions:**
- Add new miner
- Import configuration
- Batch operations

### Profiles Page

Save and manage mining configurations:

**Profile Cards:**
- Profile name and description
- Coin and algorithm
- Pool information
- Quick start button

**Profile Actions:**
- Create new profile
- Edit existing profile
- Duplicate profile
- Delete profile
- Export/import profiles

**Profile Editor:**
- Basic settings (name, coin, pool)
- Advanced settings (threads, priority, etc.)
- Test configuration
- Save or cancel

### Statistics Page

Detailed performance analytics:

**Charts:**
- Hashrate over time (line chart)
- Share distribution (pie chart)
- Pool comparison (bar chart)
- Temperature/power (multi-line)

**Time Ranges:**
- Last hour
- Last 24 hours
- Last 7 days
- Last 30 days
- Custom range

**Export:**
- Export to CSV
- Export to JSON
- Generate report (PDF)

### Pools Page

Mining pool information and management:

**Recommended Pools:**
- Top pools by reliability
- Fee comparison
- Payout information
- One-click configuration

**Pool Testing:**
- Test pool connectivity
- Latency measurement
- Difficulty estimation

**Custom Pools:**
- Add custom pool
- Edit pool details
- Remove pool

### Admin Page

System administration and diagnostics:

**System Information:**
- OS and architecture
- CPU information
- GPU detection (OpenCL/CUDA)
- Memory usage

**Miner Management:**
- Install/uninstall miners
- Update all miners
- Clear miner data

**Diagnostics:**
- Run system check
- View logs
- Export diagnostic report

**Advanced:**
- API endpoint configuration
- Data directory location
- Debug mode toggle

### Settings Page

Application preferences:

**General:**
- Language selection
- Theme (light/dark/auto)
- Auto-start on boot
- Minimize to tray

**Notifications:**
- Enable desktop notifications
- Sound alerts
- Notification types (start, stop, errors)

**Updates:**
- Auto-check for updates
- Update channel (stable/beta)
- Auto-install updates

**Mining:**
- Auto-refresh interval
- Default miner type
- Default CPU threads
- Power saving mode

**Advanced:**
- Enable developer tools
- API base URL (for custom backend)
- Log level
- Data retention period

## Using the Desktop App

### Starting Mining

**Method 1: Quick Start**
1. Click "Quick Start" on Dashboard
2. Select a profile (or use default)
3. Click "Start"

**Method 2: From Profile**
1. Go to Profiles page
2. Click "Start" on desired profile
3. Miner starts immediately

**Method 3: Manual Configuration**
1. Go to Miners page
2. Click "Add Miner"
3. Configure settings
4. Click "Start Mining"

### Monitoring Performance

**Real-time View:**
1. Go to Dashboard
2. View live hashrate and shares
3. Click miner card for details

**Detailed Statistics:**
1. Go to Statistics page
2. Select miner from dropdown
3. Choose time range
4. View charts and metrics

**System Tray:**
- Hover over tray icon for quick stats
- Click icon to show/hide window

### Managing Profiles

**Create Profile:**
1. Go to Profiles page
2. Click "New Profile"
3. Enter details:
   - Profile name
   - Select coin
   - Choose pool
   - Enter wallet address
   - Configure hardware
4. Click "Save"

**Edit Profile:**
1. Click "Edit" on profile card
2. Modify settings
3. Click "Save Changes"

**Use Profile:**
1. Click "Start" on profile card
2. Monitor on Dashboard
3. Click "Stop" when done

### Auto-Start Configuration

To start mining automatically on boot:

1. Go to Settings
2. Enable "Auto-start on boot"
3. Select default profile
4. Configure delay (optional)

The app will:
- Launch on system startup
- Wait for configured delay
- Start mining with selected profile
- Minimize to tray

### Updating Mining Software

**Auto Update (Recommended):**
1. Go to Settings → Updates
2. Enable "Auto-check for updates"
3. Updates happen automatically

**Manual Update:**
1. Go to Admin page
2. Click "Check for Updates"
3. Click "Update All" or select specific miner
4. Wait for download and installation

### Installing Additional Miners

1. Go to Admin page
2. Find miner in "Available Miners"
3. Click "Install"
4. Wait for download
5. Miner appears in Miners page

## Keyboard Shortcuts

Global shortcuts:

- `Ctrl+N`: New profile
- `Ctrl+O`: Open profiles
- `Ctrl+S`: Save current form
- `Ctrl+W`: Close window
- `Ctrl+Q`: Quit application
- `Ctrl+R`: Refresh data
- `Ctrl+,`: Open settings
- `F5`: Refresh current page
- `F11`: Toggle fullscreen
- `F12`: Open developer tools

Mining shortcuts:

- `Ctrl+M`: Start mining (quick start)
- `Ctrl+Shift+M`: Stop all miners
- `Ctrl+P`: Pause/resume mining

Navigation:

- `Ctrl+1-7`: Switch between pages
- `Ctrl+Tab`: Next page
- `Ctrl+Shift+Tab`: Previous page

## Command Line Arguments

Launch the app with arguments:

```bash
# Start minimized to tray
mining-dashboard --minimized

# Start with specific profile
mining-dashboard --profile "XMR Mining"

# Start mining immediately
mining-dashboard --auto-start

# Use custom data directory
mining-dashboard --data-dir ~/my-mining-data

# Enable debug mode
mining-dashboard --debug
```

## Data Storage

Application data is stored in:

**Linux:**
- Config: `~/.config/lethean-desktop/`
- Data: `~/.local/share/lethean-desktop/`
- Logs: `~/.local/share/lethean-desktop/logs/`

**macOS:**
- Config: `~/Library/Application Support/lethean-desktop/`
- Data: `~/Library/Application Support/lethean-desktop/`
- Logs: `~/Library/Logs/lethean-desktop/`

**Windows:**
- Config: `%APPDATA%\lethean-desktop\`
- Data: `%LOCALAPPDATA%\lethean-desktop\`
- Logs: `%LOCALAPPDATA%\lethean-desktop\logs\`

## Backing Up Data

To back up your profiles and settings:

**Method 1: Export Profiles**
1. Go to Profiles page
2. Click "Export All"
3. Save JSON file
4. Store file safely

**Method 2: Copy Data Directory**
```bash
# Linux/macOS
cp -r ~/.config/lethean-desktop ~/backup/

# Windows (PowerShell)
Copy-Item -Path "$env:APPDATA\lethean-desktop" -Destination "C:\backup\" -Recurse
```

## Troubleshooting

### App Won't Start

**Linux:**
```bash
# Check for errors
./mining-dashboard --debug

# Check permissions
chmod +x mining-dashboard-linux-amd64
```

**macOS:**
- Remove from quarantine:
  ```bash
  xattr -cr /Applications/Mining\ Dashboard.app
  ```

**Windows:**
- Run as administrator
- Check antivirus isn't blocking
- Install Visual C++ Redistributable if needed

### Miner Won't Start

1. Go to Admin → Diagnostics
2. Click "Run System Check"
3. Review errors and warnings
4. Follow suggested fixes

### High CPU Usage

1. Go to Settings → Mining
2. Enable "Power Saving Mode"
3. Reduce auto-refresh interval
4. Close unnecessary pages

### Data Not Syncing

1. Check internet connection
2. Verify API endpoint in Settings
3. Restart application
4. Clear cache: Settings → Advanced → Clear Cache

### Window Won't Restore from Tray

1. Right-click tray icon
2. Select "Show Window"
3. If still hidden, restart application

## Uninstalling

### Linux

**DEB Package:**
```bash
sudo dpkg -r mining-dashboard
```

**RPM Package:**
```bash
sudo rpm -e mining-dashboard
```

**Standalone:**
```bash
rm ~/mining-dashboard
rm -rf ~/.config/lethean-desktop
rm -rf ~/.local/share/lethean-desktop
```

### macOS

1. Quit the application
2. Move to Trash from Applications
3. Delete data (optional):
   ```bash
   rm -rf ~/Library/Application\ Support/lethean-desktop
   ```

### Windows

1. Use Programs and Features (Control Panel)
2. Or run uninstaller from Start Menu
3. Delete data (optional):
   - Navigate to `%APPDATA%\lethean-desktop`
   - Delete folder

## Next Steps

- Read the [CLI Guide](cli.md) for command-line usage
- Explore the [Web Dashboard](web-dashboard.md) features
- Review [Pool Recommendations](../reference/pools.md)
- Check the [API Documentation](../api/endpoints.md)
