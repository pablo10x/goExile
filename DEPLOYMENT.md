# 🚀 Exile Deployment & Production Guide

This document contains critical information for deploying the **Exile Master Server** to a VPS and connecting to it via the **Tauri Command HUD**.

## 1. Server-Side Setup (VPS)

### Environment Configuration (`.env`)
Copy `.env.example` to `.env` and configure the following critical variables:

```bash
# Force strict security checks (2FA, Secure Cookies)
PRODUCTION_MODE=true

# Port the Go server listens on
SERVER_PORT=8081

# Crucial: Add your Tauri origins and your VPS domain/IP
ALLOWED_ORIGINS=https://tauri.localhost,tauri://localhost,https://your-domain.com

# Use a strong random key for Node authentication
MASTER_API_KEY=your_very_strong_random_key_here
GAME_API_KEY=your_very_strong_game_key_here
```

### Reverse Proxy (Recommended)
Always run the server behind a reverse proxy like **Nginx** or **Caddy** to handle SSL (HTTPS/WSS).

**Nginx Example:**
```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8081;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
    proxy_set_header Host $host;
}
```

### Database
For production, switch from SQLite to **PostgreSQL** by setting:
`DB_DRIVER=pgx`
`DB_DSN=postgres://user:pass@localhost:5432/exile?sslmode=disable`

---

## 2. Desktop Client Setup (Tauri)

### Connecting to your VPS
The Tauri app is configured to be "Server Agnostic". By default, it looks at `localhost:8081`. To point it to your VPS:

1.  Open the Tauri app on your PC.
2.  Open the Web Inspector (`Right Click` -> `Inspect Element`).
3.  In the **Console**, run the following command:
    ```javascript
    localStorage.setItem('server_url', 'https://your-vps-domain.com');
    window.location.reload();
    ```
4.  The app will now use that URL for all API and WebSocket traffic.

### Security
*   The app uses **Bearer Token** authentication automatically. 
*   On first login, the session is stored in `localStorage` as `exile_session`.
*   The CSP (`tauri.conf.json`) is configured to allow `https:` and `wss:` connections to any destination.

---

## 3. Node Setup (Game Hosts)

1.  Deploy the `node` binary to your game server machines.
2.  Configure `MASTER_URL` to point to your VPS (e.g., `https://your-vps-domain.com`).
3.  Use the **Enrollment Flow** from the Dashboard to link the node securely.

---

## 🔒 Security Reminders
- [ ] Rotate `MASTER_API_KEY` before going live.
- [ ] Ensure `PRODUCTION_MODE=true` is set on the VPS.
- [ ] Never expose the raw `:8081` port to the internet; always use a firewall (UFW) and only allow your Reverse Proxy.
