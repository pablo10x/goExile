using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;
using UnityEngine;
using NativeWebSocket; // Requires: https://github.com/endel/NativeWebSocket

[Serializable]
public class GameMessage<T>
{
    public string type;
    public T payload;
}

[Serializable]
public class FriendActionPayload
{
    public long receiver_id;
    public long sender_id;
    public long friend_id;
}

public class GameWebSocketManager : MonoBehaviour
{
    [SerializeField] private string serverUrl = "ws://localhost:8081";
    private WebSocket websocket;

    // Events
    public event Action OnConnected;
    public event Action<string> OnError;
    public event Action<string> OnMessageReceived;

    // Singleton
    public static GameWebSocketManager Instance { get; private set; }

    private void Awake()
    {
        if (Instance == null) { Instance = this; DontDestroyOnLoad(gameObject); }
        else { Destroy(gameObject); }
    }

    public void Connect(string authKey, string endpoint)
    {
        string wsUrl = $"{serverUrl}{endpoint}?key={authKey}";
        websocket = new WebSocket(wsUrl);

        websocket.OnOpen += () => {
            Debug.Log("[WS] Connected to Master Server");
            OnConnected?.Invoke();
        };

        websocket.OnError += (e) => {
            Debug.LogError($"[WS] Error: {e}");
            OnError?.Invoke(e);
        };

        websocket.OnClose += (e) => {
            Debug.Log("[WS] Connection closed");
        };

        websocket.OnMessage += (bytes) => {
            string message = Encoding.UTF8.GetString(bytes);
            Debug.Log($"[WS] Received: {message}");
            OnMessageReceived?.Invoke(message);
        };

        websocket.Connect();
    }

    private void Update()
    {
#if !UNITY_WEBGL || UNITY_EDITOR
        websocket?.DispatchMessageQueue();
#endif
    }

    public async void SendFriendRequest(long targetPlayerId)
    {
        var msg = new GameMessage<FriendActionPayload>
        {
            type = "FRIEND_REQUEST_SEND",
            payload = new FriendActionPayload { receiver_id = targetPlayerId }
        };
        await websocket.SendText(JsonUtility.ToJson(msg));
    }

    private async void OnApplicationQuit()
    {
        await websocket.Close();
    }
}
