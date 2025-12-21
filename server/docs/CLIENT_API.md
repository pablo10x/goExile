# Game Client API Documentation

This document outlines the API endpoints and WebSocket protocol for the game client to interact with the Master Server.

## Security

All API requests must include the Game API Key in the header:
`X-Game-API-Key: <YOUR_GAME_API_KEY>`

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

All WebSocket messages follow this envelope:
```json
{
  "type": "MESSAGE_TYPE",
  "payload": { ... }
}
```

### 1. Friend Management (Client -> Server)

#### Send Friend Request
```json
{
  "type": "FRIEND_REQUEST_SEND",
  "payload": {
    "receiver_id": 456
  }
}
```

#### Accept Friend Request
```json
{
  "type": "FRIEND_REQUEST_ACCEPT",
  "payload": {
    "sender_id": 456
  }
}
```

#### Reject Friend Request
```json
{
  "type": "FRIEND_REQUEST_REJECT",
  "payload": {
    "sender_id": 456
  }
}
```

#### Remove Friend
```json
{
  "type": "FRIEND_REMOVE",
  "payload": {
    "friend_id": 456
  }
}
```

### 2. Notifications (Server -> Client)

#### Incoming Friend Request
```json
{
  "type": "NOTIFY_FRIEND_REQUEST",
  "payload": {
    "sender_id": 123,
    "sender_name": "PlayerName"
  }
}
```

#### Friendship Established
```json
{
  "type": "NOTIFY_FRIEND_ACCEPTED",
  "payload": {
    "friend_id": 123,
    "friend_name": "PlayerName"
  }
}
```

#### Error Message
```json
{
  "type": "ERROR",
  "payload": {
    "message": "Friend request failed: already friends"
  }
}
```
