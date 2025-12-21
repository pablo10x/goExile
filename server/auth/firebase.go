package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2/google"

	"exile/server/utils"
)

// FirebaseManager handles Firebase Remote Config operations via REST API
type FirebaseManager struct {
	mu sync.RWMutex

	ProjectID     string
	Connected     bool
	lastError     string
	lastSync      time.Time
	configCache   map[string]*RemoteConfigParameter
	httpClient    *http.Client
	accessToken   string
	tokenExpiry   time.Time
	credsFilePath string
}

// RemoteConfigParameter represents a Firebase Remote Config parameter
type RemoteConfigParameter struct {
	Key          string    `json:"key"`
	DefaultValue string    `json:"value"`
	ValueType    string    `json:"valueType"` // string, number, boolean, json
	Description  string    `json:"description"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FirebaseStatusResponse is the API response for firebase status
type FirebaseStatusResponse struct {
	Connected   bool                     `json:"connected"`
	ProjectID   string                   `json:"project_id,omitempty"`
	LastSync    *time.Time               `json:"last_sync,omitempty"`
	LastError   string                   `json:"last_error,omitempty"`
	Configs     []*RemoteConfigParameter `json:"configs,omitempty"`
	ConfigCount int                      `json:"config_count"`
}

// Firebase Remote Config REST API structures
type remoteConfigTemplate struct {
	Parameters map[string]remoteConfigParam `json:"parameters,omitempty"`
	Version    *templateVersion             `json:"version,omitempty"`
	Etag       string                       `json:"-"`
}

type remoteConfigParam struct {
	DefaultValue      *paramValue            `json:"defaultValue,omitempty"`
	Description       string                 `json:"description,omitempty"`
	ConditionalValues map[string]*paramValue `json:"conditionalValues,omitempty"`
	ValueType         string                 `json:"valueType,omitempty"`
}

type paramValue struct {
	Value           string `json:"value,omitempty"`
	UseInAppDefault bool   `json:"useInAppDefault,omitempty"`
}

type templateVersion struct {
	VersionNumber string `json:"versionNumber,omitempty"`
	UpdateTime    string `json:"updateTime,omitempty"`
	UpdateUser    *struct {
		Email string `json:"email,omitempty"`
	} `json:"updateUser,omitempty"`
	UpdateOrigin string `json:"updateOrigin,omitempty"`
	UpdateType   string `json:"updateType,omitempty"`
}

var FirebaseMgr *FirebaseManager

// InitFirebase initializes the Firebase connection
func InitFirebase() error {
	ProjectID := os.Getenv("FIREBASE_PROJECT_ID")
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	// Remove quotes if present (common issue with env vars)
	if len(credsPath) >= 2 && credsPath[0] == '\'' && credsPath[len(credsPath)-1] == '\'' {
		credsPath = credsPath[1 : len(credsPath)-1]
	}

	FirebaseMgr = &FirebaseManager{
		ProjectID:     ProjectID,
		Connected:     false,
		configCache:   make(map[string]*RemoteConfigParameter),
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		credsFilePath: credsPath,
	}

	if ProjectID == "" {
		return nil
	}

	if credsPath == "" {
		return nil
	}

	// Check if credentials file exists
	if _, err := os.Stat(credsPath); os.IsNotExist(err) {
		FirebaseMgr.lastError = fmt.Sprintf("Credentials file not found: %s", credsPath)
		return nil
	}

	// Try to get initial access token
	if err := FirebaseMgr.refreshAccessToken(); err != nil {
		FirebaseMgr.lastError = fmt.Sprintf("Failed to get access token: %v", err)
		return nil
	}

	FirebaseMgr.Connected = true
	FirebaseMgr.lastError = ""

	// Initial sync of remote config (run synchronously)
	_ = FirebaseMgr.syncRemoteConfig()

	return nil
}

// refreshAccessToken gets a new OAuth2 access token using service account credentials
func (fm *FirebaseManager) refreshAccessToken() error {
	ctx := context.Background()

	data, err := os.ReadFile(fm.credsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	conf, err := google.JWTConfigFromJSON(data,
		"https://www.googleapis.com/auth/firebase.remoteconfig",
	)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	token, err := conf.TokenSource(ctx).Token()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	fm.accessToken = token.AccessToken
	fm.tokenExpiry = token.Expiry

	return nil
}

// getAccessToken returns a valid access token, refreshing if necessary
func (fm *FirebaseManager) getAccessToken() (string, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Refresh if token expires in less than 5 minutes
	if time.Until(fm.tokenExpiry) < 5*time.Minute {
		if err := fm.refreshAccessToken(); err != nil {
			return "", err
		}
	}

	return fm.accessToken, nil
}

// getRemoteConfigURL returns the Remote Config REST API URL
func (fm *FirebaseManager) getRemoteConfigURL() string {
	return fmt.Sprintf("https://firebaseremoteconfig.googleapis.com/v1/projects/%s/remoteConfig", fm.ProjectID)
}

// syncRemoteConfig fetches the current remote config template
// Returns error if sync failed
func (fm *FirebaseManager) syncRemoteConfig() error {
	token, err := fm.getAccessToken()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get access token: %v", err)
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	url := fm.getRemoteConfigURL()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create request: %v", err)
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := fm.httpClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to fetch remote config: %v", err)
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to read response body: %v", err)
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Remote Config API error: %s - %s", resp.Status, string(body))
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	var template remoteConfigTemplate
	if err := json.Unmarshal(body, &template); err != nil {
		errMsg := fmt.Sprintf("Failed to decode template: %v", err)
		fm.mu.Lock()
		fm.lastError = errMsg
		fm.mu.Unlock()
		return fmt.Errorf("%s", errMsg)
	}

	// Store ETag for later updates
	template.Etag = resp.Header.Get("ETag")

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// Clear cache and rebuild from template
	fm.configCache = make(map[string]*RemoteConfigParameter)

	for key, param := range template.Parameters {
		rcParam := &RemoteConfigParameter{
			Key:         key,
			Description: param.Description,
			ValueType:   param.ValueType,
			UpdatedAt:   time.Now(),
		}

		// Extract default value
		if param.DefaultValue != nil {
			rcParam.DefaultValue = param.DefaultValue.Value
		}

		// Determine value type if not set
		if rcParam.ValueType == "" {
			rcParam.ValueType = detectValueType(rcParam.DefaultValue)
		}

		fm.configCache[key] = rcParam
	}

	fm.lastSync = time.Now()
	fm.lastError = ""
	return nil
}

// getTemplate fetches the current template with ETag
func (fm *FirebaseManager) getTemplate() (*remoteConfigTemplate, error) {
	token, err := fm.getAccessToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	req, err := http.NewRequest("GET", fm.getRemoteConfigURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := fm.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var template remoteConfigTemplate
	if err := json.NewDecoder(resp.Body).Decode(&template); err != nil {
		return nil, fmt.Errorf("failed to decode template: %w", err)
	}

	template.Etag = resp.Header.Get("ETag")

	return &template, nil
}

// publishTemplate publishes an updated template
func (fm *FirebaseManager) publishTemplate(template *remoteConfigTemplate, validateOnly bool) error {
	token, err := fm.getAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	body, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	url := fm.getRemoteConfigURL()
	if validateOnly {
		url += "?validateOnly=true"
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; UTF-8")
	req.Header.Set("If-Match", template.Etag)

	resp, err := fm.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish template: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return nil
}

// VerifyIDToken verifies a Firebase ID token and returns the UID
func (fm *FirebaseManager) VerifyIDToken(idToken string) (string, error) {
	// This is a simplified verification using Google's public keys endpoint.
	// For production, consider using the official Firebase Admin SDK for Go.
	// However, since we are avoiding extra dependencies, we can verify against Google's token info endpoint.
	// Endpoint: https://oauth2.googleapis.com/tokeninfo?id_token=XYZ

	resp, err := http.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("invalid token")
	}

	var claims struct {
		Sub string `json:"sub"` // Subject (UID)
		Aud string `json:"aud"` // Audience (Project ID)
	}

	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return "", fmt.Errorf("failed to decode token claims: %w", err)
	}

	if claims.Aud != fm.ProjectID {
		return "", fmt.Errorf("token audience mismatch: expected %s, got %s", fm.ProjectID, claims.Aud)
	}

	return claims.Sub, nil
}

// detectValueType attempts to determine the type of a value
func detectValueType(value string) string {
	if value == "" {
		return "STRING"
	}

	// Check for boolean
	if value == "true" || value == "false" {
		return "BOOLEAN"
	}

	// Check for JSON
	if len(value) > 0 && ((value[0] == '{' && value[len(value)-1] == '}') ||
		(value[0] == '[' && value[len(value)-1] == ']')) {
		var js json.RawMessage
		if json.Unmarshal([]byte(value), &js) == nil {
			return "JSON"
		}
	}

	// Check for number
	var num float64
	if _, err := fmt.Sscanf(value, "%f", &num); err == nil {
		return "NUMBER"
	}

	return "STRING"
}

// GetFirebaseStatusHandler returns the current Firebase connection status
func GetFirebaseStatusHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil {
		utils.WriteJSON(w, http.StatusOK, FirebaseStatusResponse{
			Connected: false,
			LastError: "Firebase not initialized",
		})
		return
	}

	FirebaseMgr.mu.RLock()
	defer FirebaseMgr.mu.RUnlock()

	response := FirebaseStatusResponse{
		Connected:   FirebaseMgr.Connected,
		ProjectID:   FirebaseMgr.ProjectID,
		LastError:   FirebaseMgr.lastError,
		ConfigCount: len(FirebaseMgr.configCache),
	}

	if !FirebaseMgr.lastSync.IsZero() {
		response.LastSync = &FirebaseMgr.lastSync
	}

	// Include configs if Connected
	if FirebaseMgr.Connected {
		configs := make([]*RemoteConfigParameter, 0, len(FirebaseMgr.configCache))
		for _, cfg := range FirebaseMgr.configCache {
			configs = append(configs, cfg)
		}
		response.Configs = configs
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// GetFirebaseConfigsHandler returns all Firebase Remote Config parameters
func GetFirebaseConfigsHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil || !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected")
		return
	}

	FirebaseMgr.mu.RLock()
	defer FirebaseMgr.mu.RUnlock()

	configs := make([]*RemoteConfigParameter, 0, len(FirebaseMgr.configCache))
	for _, cfg := range FirebaseMgr.configCache {
		configs = append(configs, cfg)
	}

	utils.WriteJSON(w, http.StatusOK, configs)
}

// SyncFirebaseConfigHandler triggers a manual sync of remote config
func SyncFirebaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not initialized")
		return
	}

	if !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected - check FIREBASE_PROJECT_ID and GOOGLE_APPLICATION_CREDENTIALS")
		return
	}

	err := FirebaseMgr.syncRemoteConfig()

	FirebaseMgr.mu.RLock()
	lastSync := FirebaseMgr.lastSync
	lastError := FirebaseMgr.lastError
	configCount := len(FirebaseMgr.configCache)
	FirebaseMgr.mu.RUnlock()

	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Sync failed: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Sync completed successfully",
		"synced_at":    lastSync,
		"config_count": configCount,
		"last_error":   lastError,
	})
}

// UpdateFirebaseConfigHandler updates a Firebase Remote Config parameter
func UpdateFirebaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil || !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected")
		return
	}

	var req struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		ValueType   string `json:"valueType"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key is required")
		return
	}

	// Get current template
	template, err := FirebaseMgr.getTemplate()
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get template: %v", err))
		return
	}

	// Ensure parameters map exists
	if template.Parameters == nil {
		template.Parameters = make(map[string]remoteConfigParam)
	}

	// Update or add the parameter
	template.Parameters[req.Key] = remoteConfigParam{
		DefaultValue: &paramValue{Value: req.Value},
		Description:  req.Description,
		ValueType:    req.ValueType,
	}

	// Validate first
	if err := FirebaseMgr.publishTemplate(template, true); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, fmt.Sprintf("Template validation failed: %v", err))
		return
	}

	// Publish
	if err := FirebaseMgr.publishTemplate(template, false); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to publish template: %v", err))
		return
	}

	// Update local cache
	FirebaseMgr.mu.Lock()
	FirebaseMgr.configCache[req.Key] = &RemoteConfigParameter{
		Key:          req.Key,
		DefaultValue: req.Value,
		ValueType:    req.ValueType,
		Description:  req.Description,
		UpdatedAt:    time.Now(),
	}
	FirebaseMgr.mu.Unlock()

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Parameter updated and published",
	})
}

// DeleteFirebaseConfigHandler deletes a Firebase Remote Config parameter
func DeleteFirebaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil || !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected")
		return
	}

	var req struct {
		Key string `json:"key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key is required")
		return
	}

	// Get current template
	template, err := FirebaseMgr.getTemplate()
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get template: %v", err))
		return
	}

	// Check if parameter exists
	if _, exists := template.Parameters[req.Key]; !exists {
		utils.WriteError(w, r, http.StatusNotFound, "Parameter not found")
		return
	}

	// Delete the parameter
	delete(template.Parameters, req.Key)

	// Publish updated template
	if err := FirebaseMgr.publishTemplate(template, false); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to publish template: %v", err))
		return
	}

	// Update local cache
	FirebaseMgr.mu.Lock()
	delete(FirebaseMgr.configCache, req.Key)
	FirebaseMgr.mu.Unlock()

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Parameter deleted",
	})
}

// CreateFirebaseConfigHandler creates a new Firebase Remote Config parameter
func CreateFirebaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil || !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected")
		return
	}

	var req struct {
		Key         string `json:"key"`
		Value       string `json:"value"`
		ValueType   string `json:"valueType"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		utils.WriteError(w, r, http.StatusBadRequest, "Key is required")
		return
	}

	// Validate key format (alphanumeric, underscores only)
	for _, c := range req.Key {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			utils.WriteError(w, r, http.StatusBadRequest, "Key must contain only alphanumeric characters and underscores")
			return
		}
	}

	// Get current template
	template, err := FirebaseMgr.getTemplate()
	if err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to get template: %v", err))
		return
	}

	// Ensure parameters map exists
	if template.Parameters == nil {
		template.Parameters = make(map[string]remoteConfigParam)
	}

	// Check if parameter already exists
	if _, exists := template.Parameters[req.Key]; exists {
		utils.WriteError(w, r, http.StatusConflict, "Parameter already exists")
		return
	}

	// Set default value type if not provided
	valueType := req.ValueType
	if valueType == "" {
		valueType = detectValueType(req.Value)
	}

	// Add the new parameter
	template.Parameters[req.Key] = remoteConfigParam{
		DefaultValue: &paramValue{Value: req.Value},
		Description:  req.Description,
		ValueType:    valueType,
	}

	// Validate template
	if err := FirebaseMgr.publishTemplate(template, true); err != nil {
		utils.WriteError(w, r, http.StatusBadRequest, fmt.Sprintf("Template validation failed: %v", err))
		return
	}

	// Publish template
	if err := FirebaseMgr.publishTemplate(template, false); err != nil {
		utils.WriteError(w, r, http.StatusInternalServerError, fmt.Sprintf("Failed to publish template: %v", err))
		return
	}

	// Update local cache
	newParam := &RemoteConfigParameter{
		Key:          req.Key,
		DefaultValue: req.Value,
		ValueType:    valueType,
		Description:  req.Description,
		UpdatedAt:    time.Now(),
	}

	FirebaseMgr.mu.Lock()
	FirebaseMgr.configCache[req.Key] = newParam
	FirebaseMgr.mu.Unlock()

	utils.WriteJSON(w, http.StatusCreated, newParam)
}

// PublishFirebaseConfigHandler publishes all pending changes to Firebase
func PublishFirebaseConfigHandler(w http.ResponseWriter, r *http.Request) {
	if FirebaseMgr == nil || !FirebaseMgr.Connected {
		utils.WriteError(w, r, http.StatusServiceUnavailable, "Firebase not Connected")
		return
	}

	// Trigger a sync to ensure we have the latest
	FirebaseMgr.syncRemoteConfig()

	FirebaseMgr.mu.RLock()
	lastSync := FirebaseMgr.lastSync
	lastError := FirebaseMgr.lastError
	FirebaseMgr.mu.RUnlock()

	if lastError != "" {
		utils.WriteError(w, r, http.StatusInternalServerError, lastError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Configuration synced from Firebase",
		"synced_at": lastSync,
	})
}
