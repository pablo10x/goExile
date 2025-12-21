# Game Client API Documentation

This document outlines the API endpoints and WebSocket protocol for the game client to interact with the Master Server.

## Authentication

### 1. Authenticate Player (`POST /api/game/auth`)

Authenticates a player using a Firebase ID Token. This is the entry point for the game client.

**Request:**
```json
{
  "id_token": "FIREBASE_ID_TOKEN_FROM_CLIENT_SDK",
  "name": "PlayerName",
  "device_id": "UNIQUE_DEVICE_ID"
}
```

**Response:**
```json
{
  "player": {
    "id": 123,
    "uid": "firebase_uid_xyz",
    "name": "PlayerName",
    "xp": 1500,
    "last_joined_server": "US-East-1",
    "friends": [],
    "incoming_friend_requests": [],
    "outgoing_friend_requests": []
  },
  "ws_auth_key": "TEMPORARY_SESSION_KEY_FOR_WEBSOCKET",
  "ws_endpoint": "/api/game/ws"
}
```

## WebSocket Connection

After authentication, connect to the WebSocket endpoint using the provided session key.

**Endpoint:** `ws://<MASTER_SERVER_HOST>/api/game/ws?key=<WS_AUTH_KEY>`

**Flow:**
1.  Call `/api/game/auth` to get `ws_auth_key`.
2.  Connect to WebSocket immediately (key expires in 1 minute).
3.  Listen for messages.

### Message Format

Messages are standard JSON. Structure to be defined as features are added (e.g., chat, lobby invites).

**Server -> Client (Example):**
```json
{
  "type": "FRIEND_REQUEST",
  "payload": {
    "sender_id": 456,
    "sender_name": "FriendName"
  }
}
```

**Client -> Server (Example):**
```json
{
  "type": "CHAT",
  "payload": {
    "message": "Hello World"
  }
}
```

## Player System

### Get Player Details (`GET /api/game/players/{id}`)
Returns public details of a player.

### Friend Requests

#### Send Request (`POST /api/game/friends/request`)
```json
{
  "sender_id": 123,
  "receiver_id": 456
}
```

#### Accept Request (`POST /api/game/friends/accept`)
```json
{
  "sender_id": 456,
  "receiver_id": 123
}
```
