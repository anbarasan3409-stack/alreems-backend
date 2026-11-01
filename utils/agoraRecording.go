package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Agora Cloud Recording Configuration
type AgoraRecordingConfig struct {
	AppID          string
	CustomerID     string
	CustomerSecret string
	Region         string // e.g., "ap" for Asia Pacific
}

// Recording Request/Response structures
type AcquireRequest struct {
	Cname         string                 `json:"cname"`
	UID           string                 `json:"uid"`
	ClientRequest map[string]interface{} `json:"clientRequest"`
}

type AcquireResponse struct {
	ResourceID string `json:"resourceId"`
}

type StartRecordingRequest struct {
	Cname         string                 `json:"cname"`
	UID           string                 `json:"uid"`
	ClientRequest map[string]interface{} `json:"clientRequest"`
}

type StartRecordingResponse struct {
	ResourceID string `json:"resourceId"`
	SID        string `json:"sid"`
}

type StopRecordingRequest struct {
	Cname         string                 `json:"cname"`
	UID           string                 `json:"uid"`
	ClientRequest map[string]interface{} `json:"clientRequest"`
}

type QueryRecordingResponse struct {
	ResourceID     string                 `json:"resourceId"`
	SID            string                 `json:"sid"`
	ServerResponse map[string]interface{} `json:"serverResponse"`
}

// GetAgoraRecordingConfig reads Agora recording configuration from environment
func GetAgoraRecordingConfig() AgoraRecordingConfig {
	return AgoraRecordingConfig{
		AppID:          os.Getenv("AGORA_APP_ID"),
		CustomerID:     os.Getenv("AGORA_CUSTOMER_ID"),
		CustomerSecret: os.Getenv("AGORA_CUSTOMER_SECRET"),
		Region:         getAgoraRegion(),
	}
}

func getAgoraRegion() string {
	region := os.Getenv("AGORA_REGION")
	if region == "" {
		region = "ap" // Default to Asia Pacific
	}
	return region
}

// generateBasicAuth creates Basic Auth header for Agora Cloud Recording API
func generateBasicAuth(customerID, customerSecret string) string {
	auth := customerID + ":" + customerSecret
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// AcquireRecordingResource acquires a resource for cloud recording
func AcquireRecordingResource(channelName, uid string) (string, error) {
	config := GetAgoraRecordingConfig()

	url := fmt.Sprintf("https://api.agora.io/v1/apps/%s/cloud_recording/acquire", config.AppID)

	requestBody := AcquireRequest{
		Cname: channelName,
		UID:   uid,
		ClientRequest: map[string]interface{}{
			"resourceExpiredHour": 24,
			"scene":               0, // 0 for real-time recording
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateBasicAuth(config.CustomerID, config.CustomerSecret))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var acquireResp AcquireResponse
	if err := json.Unmarshal(body, &acquireResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return acquireResp.ResourceID, nil
}

// StartCloudRecording starts cloud recording for a channel
func StartCloudRecording(channelName, uid, resourceID, token string) (string, error) {
	config := GetAgoraRecordingConfig()

	url := fmt.Sprintf("https://api.agora.io/v1/apps/%s/cloud_recording/resourceid/%s/mode/mix/start",
		config.AppID, resourceID)

	// Configure recording storage (you can customize this)
	storageConfig := map[string]interface{}{
		"vendor":         1, // 1 for Agora Cloud Storage
		"region":         getStorageRegion(),
		"bucket":         os.Getenv("AGORA_RECORDING_BUCKET"),
		"accessKey":      os.Getenv("AGORA_RECORDING_ACCESS_KEY"),
		"secretKey":      os.Getenv("AGORA_RECORDING_SECRET_KEY"),
		"fileNamePrefix": []string{"recordings", channelName},
	}

	requestBody := StartRecordingRequest{
		Cname: channelName,
		UID:   uid,
		ClientRequest: map[string]interface{}{
			"token": token,
			"recordingConfig": map[string]interface{}{
				"channelType":     0, // 0 for communication, 1 for live broadcast
				"streamTypes":     2, // 2 for audio and video
				"maxIdleTime":     30,
				"audioProfile":    1,
				"videoStreamType": 0,
				"transcodingConfig": map[string]interface{}{
					"width":            1280,
					"height":           720,
					"fps":              15,
					"bitrate":          2260,
					"mixedVideoLayout": 1, // 1 for floating layout
				},
			},
			"recordingFileConfig": map[string]interface{}{
				"avFileType": []string{"hls", "mp4"},
			},
			"storageConfig": storageConfig,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateBasicAuth(config.CustomerID, config.CustomerSecret))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var startResp StartRecordingResponse
	if err := json.Unmarshal(body, &startResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return startResp.SID, nil
}

// StopCloudRecording stops the cloud recording
func StopCloudRecording(channelName, uid, resourceID, sid string) error {
	config := GetAgoraRecordingConfig()

	url := fmt.Sprintf("https://api.agora.io/v1/apps/%s/cloud_recording/resourceid/%s/sid/%s/mode/mix/stop",
		config.AppID, resourceID, sid)

	requestBody := StopRecordingRequest{
		Cname:         channelName,
		UID:           uid,
		ClientRequest: map[string]interface{}{},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateBasicAuth(config.CustomerID, config.CustomerSecret))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// QueryRecordingStatus queries the status of a recording
func QueryRecordingStatus(resourceID, sid string) (*QueryRecordingResponse, error) {
	config := GetAgoraRecordingConfig()

	url := fmt.Sprintf("https://api.agora.io/v1/apps/%s/cloud_recording/resourceid/%s/sid/%s/mode/mix/query",
		config.AppID, resourceID, sid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", generateBasicAuth(config.CustomerID, config.CustomerSecret))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var queryResp QueryRecordingResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &queryResp, nil
}

func getStorageRegion() int {
	// Return region code based on your cloud storage location
	// 0: US, 1: AP, 2: EU, 3: CN
	region := os.Getenv("AGORA_STORAGE_REGION")
	switch region {
	case "us":
		return 0
	case "ap":
		return 1
	case "eu":
		return 2
	case "cn":
		return 3
	default:
		return 1 // Default to AP
	}
}
