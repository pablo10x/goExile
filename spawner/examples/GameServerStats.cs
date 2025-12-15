using System;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using UnityEngine; // Assuming Unity, but works in plain C# too (remove MonoBehaviour)

// Attach this script to a GameObject in your scene
public class GameServerStats : MonoBehaviour
{
    private ClientWebSocket _ws;
    private string _wsUrl;
    private CancellationTokenSource _cts;

    // Stats to track
    public int CurrentPlayers = 0;
    public int MaxPlayers = 100;

    void Start()
    {
        // Parse command line arguments to find -ws URL
        string[] args = System.Environment.GetCommandLineArgs();
        for (int i = 0; i < args.Length; i++)
        {
            if (args[i] == "-ws" && i + 1 < args.Length)
            {
                _wsUrl = args[i + 1];
                Debug.Log($"Found WebSocket URL: {_wsUrl}");
                break;
            }
        }

        if (!string.IsNullOrEmpty(_wsUrl))
        {
            ConnectAndReport();
        }
        else
        {
            Debug.LogWarning("No WebSocket URL provided in command line args (-ws)");
        }
    }

    async void ConnectAndReport()
    {
        _ws = new ClientWebSocket();
        _cts = new CancellationTokenSource();

        try
        {
            await _ws.ConnectAsync(new Uri(_wsUrl), _cts.Token);
            Debug.Log("Connected to Spawner WebSocket");

            // Start reporting loop
            while (_ws.State == WebSocketState.Open && !_cts.IsCancellationRequested)
            {
                await SendStats();
                await Task.Delay(5000); // Report every 5 seconds
            }
        }
        catch (Exception e)
        {
            Debug.LogError($"WebSocket Error: {e.Message}");
        }
    }

    async Task SendStats()
    {
        if (_ws.State != WebSocketState.Open) return;

        // Simple JSON payload
        string json = $"{{"type": "stats", "player_count": {CurrentPlayers}, "max_players": {MaxPlayers}}}";
        byte[] bytes = Encoding.UTF8.GetBytes(json);

        await _ws.SendAsync(new ArraySegment<byte>(bytes), WebSocketMessageType.Text, true, _cts.Token);
    }

    void OnApplicationQuit()
    {
        _cts?.Cancel();
        _ws?.Dispose();
    }

    // Example method to update players (call this from your PlayerManager)
    public void PlayerJoined()
    {
        CurrentPlayers++;
        // Trigger immediate update or wait for loop
    }

    public void PlayerLeft()
    {
        CurrentPlayers--;
    }
}
