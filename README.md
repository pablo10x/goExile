# 🎮 goExile - Game Server Management System

A comprehensive, production-ready system for managing and deploying game server instances across multiple spawner nodes. The system consists of a Master Server (registry), Spawner services, and a modern web dashboard.

---

## 📋 Table of Contents

- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Components](#components)
- [How to Add a Button](#how-to-add-a-button)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)

---

## 🏗️ System Overview

goExile is a distributed game server management platform that allows you to:

- **Register and manage multiple Spawner nodes** across different regions
- **Dynamically spawn game server instances** on demand
- **Monitor resource usage** (CPU, memory, disk) in real-time
- **Manage game server versions** and deploy updates
- **Backup and restore** game server instances
- **View logs and metrics** through a web dashboard

### Key Features

- ✅ Centralized Master Server for spawner coordination
- ✅ Automatic port assignment for game instances
- ✅ Real-time monitoring via Server-Sent Events (SSE)
- ✅ Version management and automatic updates
- ✅ Backup/restore functionality
- ✅ Authentication and API key security
- ✅ SQLite persistence for spawner registry
- ✅ Graceful shutdown and process management

---

## 🏛️ Architecture

The system is composed of three main components:

### 1. **Master Server** (`server/`)

The central registry that:
- Tracks all registered Spawner nodes
- Receives heartbeats from Spawners
- Proxies spawn requests to appropriate Spawners
- Serves game server build files
- Provides a REST API for the web dashboard
- Manages authentication and sessions

**Key Files:**
- `server/main.go` - Main server entry point
- `server/handlers.go` - API request handlers
- `server/registry.go` - Spawner registry management
- `server/sse.go` - Server-Sent Events for real-time updates

### 2. **Spawner Service** (`spawner/`)

Individual spawner nodes that:
- Register with the Master Server on startup
- Send periodic heartbeats with status updates
- Manage game server instances (spawn, stop, restart, update)
- Monitor resource usage (CPU, memory, disk)
- Handle instance lifecycle and logging

**Key Files:**
- `spawner/main.go` - Spawner service entry point
- `spawner/api/handlers.go` - API endpoints for instance management
- `spawner/internal/game/manager.go` - Game instance management logic
- `spawner/internal/updater/updater.go` - Game server update handling

### 3. **Web Dashboard** (`web-dashboard/`)

A modern SvelteKit-based frontend that:
- Displays spawner status and metrics
- Allows spawning and managing game instances
- Shows real-time resource usage charts
- Provides log viewing and instance management
- Handles authentication and user sessions

**Key Files:**
- `web-dashboard/src/routes/+page.svelte` - Main dashboard page
- `web-dashboard/src/routes/spawners/[id]/+page.svelte` - Spawner detail page
- `web-dashboard/src/lib/components/` - Reusable UI components

---

## 🧩 Components

### Master Server Components

```
server/
├── main.go              # Server initialization and routing
├── handlers.go          # HTTP request handlers
├── registry.go          # Spawner registry (in-memory + SQLite)
├── auth.go              # Authentication middleware
├── sse.go               # Server-Sent Events hub
├── dashboard.go         # Dashboard API endpoints
└── database/
    └── registry.db      # SQLite database (optional)
```

### Spawner Components

```
spawner/
├── main.go              # Spawner service entry point
├── api/
│   └── handlers.go      # REST API handlers
├── internal/
│   ├── game/
│   │   └── manager.go    # Instance lifecycle management
│   ├── config/
│   │   └── config.go     # Configuration loading
│   └── updater/
│       └── updater.go    # Game server update logic
└── instances/            # Individual game server instances
```

### Web Dashboard Components

```
web-dashboard/src/
├── routes/
│   ├── +page.svelte           # Main dashboard
│   ├── spawners/[id]/
│   │   └── +page.svelte       # Spawner detail page
│   └── login/
│       └── +page.svelte        # Login page
└── lib/
    ├── components/
    │   ├── InstanceRow.svelte      # Instance list item
    │   ├── SpawnerTable.svelte     # Spawner list table
    │   ├── StatsCard.svelte        # Metric display card
    │   ├── ConfirmDialog.svelte   # Confirmation modal
    │   └── ...                     # Other components
    └── composables/
        └── useResourceMetrics.ts   # Resource metrics logic
```

---

## 🔘 How to Add a Button

The web dashboard uses **Svelte 5** with TypeScript. Here's how to add a button to any component:

### Basic Button Example

```svelte
<script lang="ts">
    function handleClick() {
        console.log('Button clicked!');
        // Your action here
    }
</script>

<button 
    onclick={handleClick}
    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-sm font-semibold transition-colors"
>
    Click Me
</button>
```

### Button with Icon (using lucide-svelte)

```svelte
<script lang="ts">
    import { Play, Settings } from 'lucide-svelte';
    
    function startInstance() {
        // Your logic here
    }
</script>

<button 
    onclick={startInstance}
    class="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-sm font-semibold transition-colors"
>
    <Play class="w-4 h-4" />
    Start Instance
</button>
```

### Button with API Call

```svelte
<script lang="ts">
    let isLoading = $state(false);
    
    async function performAction() {
        isLoading = true;
        try {
            const res = await fetch('/api/spawners/1/instances/instance-1/start', {
                method: 'POST'
            });
            if (!res.ok) {
                throw new Error('Action failed');
            }
            const data = await res.json();
            console.log('Success:', data);
        } catch (error) {
            alert(error.message);
        } finally {
            isLoading = false;
        }
    }
</script>

<button 
    onclick={performAction}
    disabled={isLoading}
    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
>
    {isLoading ? 'Loading...' : 'Start Instance'}
</button>
```

### Button with Confirmation Dialog

```svelte
<script lang="ts">
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    
    let isDialogOpen = $state(false);
    
    function openDialog() {
        isDialogOpen = true;
    }
    
    async function confirmAction() {
        try {
            const res = await fetch('/api/spawners/1/instances/instance-1/delete', {
                method: 'DELETE'
            });
            if (!res.ok) throw new Error('Delete failed');
            isDialogOpen = false;
        } catch (error) {
            alert(error.message);
        }
    }
</script>

<button 
    onclick={openDialog}
    class="px-4 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg"
>
    Delete Instance
</button>

<ConfirmDialog
    bind:isOpen={isDialogOpen}
    title="Delete Instance"
    message="Are you sure you want to delete this instance? This action cannot be undone."
    confirmText="Delete"
    onConfirm={confirmAction}
/>
```

### Adding a Button to InstanceRow Component

To add a button to the `InstanceRow` component (which displays each game instance):

1. **Open** `web-dashboard/src/lib/components/InstanceRow.svelte`

2. **Add the button** in the quick actions section (around line 100):

```svelte
<button 
    onclick={() => dispatch('customAction', { spawnerId, instanceId: instance.id })}
    class="p-1.5 text-purple-400 hover:text-purple-300 hover:bg-purple-400/10 rounded transition-colors"
    title="Custom Action"
>
    <CustomIcon class="w-4 h-4" />
</button>
```

3. **Handle the event** in the parent component (`spawners/[id]/+page.svelte`):

```svelte
<InstanceRow 
    spawnerId={spawnerId}
    instance={instance}
    on:customAction={(e) => handleCustomAction(e.detail.instanceId)}
/>
```

4. **Implement the handler**:

```svelte
async function handleCustomAction(instanceId: string) {
    try {
        const res = await fetch(`/api/spawners/${spawnerId}/instances/${instanceId}/custom`, {
            method: 'POST'
        });
        if (!res.ok) throw new Error('Action failed');
        fetchSpawnerData(); // Refresh data
    } catch (error) {
        alert(error.message);
    }
}
```

### Button Styling Guide

The dashboard uses **Tailwind CSS**. Common button styles:

```svelte
<!-- Primary Action (Blue) -->
<button class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg">Action</button>

<!-- Success (Green) -->
<button class="px-4 py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg">Start</button>

<!-- Warning (Yellow) -->
<button class="px-4 py-2 bg-yellow-600 hover:bg-yellow-500 text-white rounded-lg">Stop</button>

<!-- Danger (Red) -->
<button class="px-4 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg">Delete</button>

<!-- Secondary (Gray) -->
<button class="px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg">Cancel</button>

<!-- Disabled State -->
<button 
    disabled
    class="px-4 py-2 bg-slate-800/50 text-slate-500 rounded-lg cursor-not-allowed"
>
    Disabled
</button>
```

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.20 or higher
- **Node.js** 18+ and npm
- **Git**

### 1. Clone the Repository

```bash
git clone <repository-url>
cd goExile
```

### 2. Start the Master Server

```bash
cd server
go mod download
go run main.go
```

The Master Server will start on `http://localhost:8081`

### 3. Configure a Spawner

Create `spawner/.env`:

```ini
REGION=US-East
PORT=8080
MASTER_URL=http://localhost:8081
MASTER_API_KEY=your_very_secret_master_api_key_here
GAME_INSTALL_DIR=game_server
INSTANCES_DIR=instances
PORT_RANGE_START=8000
PORT_RANGE_END=9000
MAX_INSTANCES=10
```

### 4. Start a Spawner

```bash
cd spawner
go mod download
go run main.go
```

### 5. Start the Web Dashboard

```bash
cd web-dashboard
npm install
npm run dev
```

The dashboard will be available at `http://localhost:5173`

### 6. Login

Default credentials (if authentication is enabled):
- Username: `admin`
- Password: `admin` (change this in production!)

---

## 📡 API Reference

### Master Server API

**Base URL:** `http://localhost:8081/api`

#### Spawner Management

- `POST /api/spawners` - Register a new spawner
- `GET /api/spawners` - List all spawners
- `GET /api/spawners/{id}` - Get spawner details
- `POST /api/spawners/{id}/heartbeat` - Send heartbeat
- `DELETE /api/spawners/{id}` - Remove spawner
- `POST /api/spawners/{id}/spawn` - Spawn a new instance

#### Instance Management

- `GET /api/spawners/{id}/instances` - List instances
- `POST /api/spawners/{id}/instances/{instance_id}/start` - Start instance
- `POST /api/spawners/{id}/instances/{instance_id}/stop` - Stop instance
- `POST /api/spawners/{id}/instances/{instance_id}/restart` - Restart instance
- `POST /api/spawners/{id}/instances/{instance_id}/update` - Update instance
- `DELETE /api/spawners/{id}/instances/{instance_id}` - Delete instance
- `GET /api/spawners/{id}/instances/{instance_id}/stats` - Get instance stats
- `GET /api/spawners/{id}/instances/{instance_id}/logs` - Get instance logs

### Spawner API

**Base URL:** `http://localhost:8080` (configurable)

- `POST /spawn` - Spawn a new instance
- `GET /instances` - List instances
- `POST /instance/{id}/start` - Start instance
- `POST /instance/{id}/stop` - Stop instance
- `POST /instance/{id}/restart` - Restart instance
- `POST /instance/{id}/update` - Update instance
- `DELETE /instance/{id}` - Remove instance
- `GET /instance/{id}/stats` - Get instance statistics
- `GET /health` - Health check

---

## 🔧 Development

### Running Tests

```bash
# Server tests
cd server
go test ./...

# Spawner tests
cd spawner
go test ./...

# Dashboard tests
cd web-dashboard
npm test
```

### Building for Production

```bash
# Master Server
cd server
go build -o server.exe

# Spawner
cd spawner
go build -o spawner.exe

# Dashboard
cd web-dashboard
npm run build
```

---

## 📝 Configuration

### Master Server Environment Variables

```bash
DB_PATH=database/registry.db          # SQLite database path
MASTER_API_KEY=your_secret_key        # API key for spawner authentication
AUTH_ENABLED=true                     # Enable dashboard authentication
AUTH_USERNAME=admin                   # Dashboard username
AUTH_PASSWORD=admin                   # Dashboard password
```

### Spawner Environment Variables

```bash
REGION=US-East                        # Spawner region identifier
PORT=8080                            # Spawner API port
MASTER_URL=http://localhost:8081     # Master server URL
MASTER_API_KEY=your_secret_key        # API key for master authentication
GAME_INSTALL_DIR=game_server         # Game server template directory
INSTANCES_DIR=instances              # Instances directory
PORT_RANGE_START=8000                # Starting port for instances
PORT_RANGE_END=9000                  # Ending port for instances
MAX_INSTANCES=10                     # Maximum concurrent instances
```

---

## 🎨 UI Component Examples

### Using Existing Components

```svelte
<script lang="ts">
    import StatsCard from '$lib/components/StatsCard.svelte';
    import ConfirmDialog from '$lib/components/ConfirmDialog.svelte';
    import Dropdown from '$lib/components/Dropdown.svelte';
    import { Server, Cpu } from 'lucide-svelte';
</script>

<!-- Stats Card -->
<StatsCard 
    title="Active Instances" 
    value="5 / 10" 
    Icon={Server}
    color="blue"
/>

<!-- Dropdown Menu -->
<Dropdown label="Actions" Icon={Server}>
    <div slot="default" let:close>
        <button onclick={() => { doAction(); close(); }}>Action 1</button>
        <button onclick={() => { doAction2(); close(); }}>Action 2</button>
    </div>
</Dropdown>
```

---

## 🐛 Troubleshooting

### Spawner not registering with Master

- Check `MASTER_URL` and `MASTER_API_KEY` in spawner `.env`
- Verify Master Server is running and accessible
- Check spawner logs in `spawner.log`

### Instances not starting

- Verify game server files exist in `spawner/game_server/`
- Check port range availability
- Review instance logs in `spawner/instances/{id}/`

### Dashboard not loading

- Ensure Master Server is running on port 8081
- Check browser console for API errors
- Verify authentication credentials

---

## 📚 Additional Resources

- [SvelteKit Documentation](https://kit.svelte.dev/)
- [Tailwind CSS Documentation](https://tailwindcss.com/)
- [Go Documentation](https://go.dev/doc/)

---

## 📄 License

[Your License Here]

---

## 🤝 Contributing

[Contributing Guidelines]

---

**Happy Coding! 🚀**

