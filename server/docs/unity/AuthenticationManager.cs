using System;
using System.Collections;
using System.Text;
using UnityEngine;
using UnityEngine.Networking;

// Data models for JSON serialization/deserialization
[Serializable]
public class AuthRequest
{
    public string id_token;
    public string name;
    public string device_id;
}

[Serializable]
public class AuthResponse
{
    public Player player;
    public string ws_auth_key;
    public string ws_endpoint;
}

[Serializable]
public class Player
{
    public long id;
    public string uid;
    public string name;
    public int xp;
    // Add other fields matching server/models/players.go
}

public class AuthenticationManager : MonoBehaviour
{
    [Header("Server Configuration")]
    [SerializeField] private string masterServerUrl = "http://localhost:8081";
    [SerializeField] private string gameApiKey = "YOUR_GAME_API_KEY"; // Set this in Inspector

    // Events for UI or other systems to listen to
    public event Action<AuthResponse> OnAuthenticated;
    public event Action<string> OnAuthError;

    // Singleton Instance
    public static AuthenticationManager Instance { get; private set; }

    private void Awake()
    {
        if (Instance == null)
        {
            Instance = this;
            DontDestroyOnLoad(gameObject);
        }
        else
        {
            Destroy(gameObject);
        }
    }

    /// <summary>
    /// Authenticates with the Master Server using Form Data (WWWForm).
    /// Recommended for simple integrations or if using standard Unity fields.
    /// </summary>
    public void AuthenticateWithFormData(string firebaseIdToken, string playerName)
    {
        StartCoroutine(AuthRoutineFormData(firebaseIdToken, playerName));
    }

    /// <summary>
    /// Authenticates with the Master Server using JSON payload.
    /// Modern approach, useful for complex nested objects.
    /// </summary>
    public void AuthenticateWithJson(string firebaseIdToken, string playerName)
    {
        StartCoroutine(AuthRoutineJson(firebaseIdToken, playerName));
    }

    private IEnumerator AuthRoutineFormData(string idToken, string name)
    {
        string url = $"{masterServerUrl}/api/game/auth";
        string deviceId = SystemInfo.deviceUniqueIdentifier;

        Debug.Log($"[Auth] Connecting to {url} via Form Data...");

        WWWForm form = new WWWForm();
        form.AddField("id_token", idToken);
        form.AddField("name", name);
        form.AddField("device_id", deviceId);

        using (UnityWebRequest request = UnityWebRequest.Post(url, form))
        {
            // Set Security Header
            request.SetRequestHeader("X-Game-API-Key", gameApiKey);

            yield return request.SendWebRequest();

            HandleResponse(request);
        }
    }

    private IEnumerator AuthRoutineJson(string idToken, string name)
    {
        string url = $"{masterServerUrl}/api/game/auth";
        string deviceId = SystemInfo.deviceUniqueIdentifier;

        Debug.Log($"[Auth] Connecting to {url} via JSON...");

        // JsonUtility requires a class/struct (anonymous objects NOT supported)
        AuthRequest payload = new AuthRequest
        {
            id_token = idToken,
            name = name,
            device_id = deviceId
        };

        string jsonBody = JsonUtility.ToJson(payload);
        byte[] bodyRaw = Encoding.UTF8.GetBytes(jsonBody);

        using (UnityWebRequest request = new UnityWebRequest(url, "POST"))
        {
            request.uploadHandler = new UploadHandlerRaw(bodyRaw);
            request.downloadHandler = new DownloadHandlerBuffer();

            // Set Headers
            request.SetRequestHeader("Content-Type", "application/json");
            request.SetRequestHeader("X-Game-API-Key", gameApiKey);

            yield return request.SendWebRequest();

            HandleResponse(request);
        }
    }

    private void HandleResponse(UnityWebRequest request)
    {
        if (request.result != UnityWebRequest.Result.Success)
        {
            string errorMsg = $"Auth Failed: {request.error} (Status: {request.responseCode})\nResponse: {request.downloadHandler.text}";
            Debug.LogError(errorMsg);
            OnAuthError?.Invoke(errorMsg);
        }
        else
        {
            string jsonResponse = request.downloadHandler.text;
            Debug.Log($"[Auth] Success: {jsonResponse}");

            try
            {
                AuthResponse response = JsonUtility.FromJson<AuthResponse>(jsonResponse);
                OnAuthenticated?.Invoke(response);
            }
            catch (Exception e)
            {
                string parseError = $"Failed to parse response: {e.Message}";
                Debug.LogError(parseError);
                OnAuthError?.Invoke(parseError);
            }
        }
    }
}
