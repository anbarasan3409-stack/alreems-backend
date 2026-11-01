package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Netless (Agora Whiteboard) configuration
type NetlessConfig struct {
	AppID     string
	AccessKey string
	SecretKey string
	SDKToken  string
	Region    string
}

type CreateRoomRequest struct {
	IsRecord bool `json:"isRecord"`
}

type CreateRoomResponse struct {
	UUID string `json:"uuid"`
}

type RoomTokenRequest struct {
	Lifespan int64  `json:"lifespan"` // in milliseconds
	Role     string `json:"role"`     // admin, writer, reader
}

// GetNetlessConfig reads configuration from environment
func GetNetlessConfig() NetlessConfig {
	return NetlessConfig{
		AppID:     os.Getenv("NETLESS_APP_ID"),
		AccessKey: os.Getenv("NETLESS_ACCESS_KEY"),
		SecretKey: os.Getenv("NETLESS_SECRET_KEY"),
		SDKToken:  os.Getenv("NETLESS_SDK_TOKEN"),
		Region:    getNetlessRegion(),
	}
}

func getNetlessRegion() string {
	region := os.Getenv("NETLESS_REGION")
	if region == "" {
		region = "us-sv" // Default to US Silicon Valley
	}
	return region
}

// CreateWhiteboardRoom creates a new whiteboard room
func CreateWhiteboardRoom(isRecord bool) (string, error) {
	config := GetNetlessConfig()

	// API endpoint for creating room
	url := "https://api.netless.link/v5/rooms"

	requestBody := CreateRoomRequest{
		IsRecord: isRecord,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers - Use pre-generated SDK token from console
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", config.SDKToken)
	req.Header.Set("region", config.Region)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var roomResp CreateRoomResponse
	if err := json.Unmarshal(body, &roomResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return roomResp.UUID, nil
}

// GenerateRoomToken generates a token for accessing a whiteboard room
func GenerateRoomToken(roomUUID string, role string, lifespan int64) (string, error) {
	config := GetNetlessConfig()

	if lifespan == 0 {
		lifespan = 3600000 // Default 1 hour in milliseconds
	}

	// API endpoint for generating token
	url := fmt.Sprintf("https://api.netless.link/v5/tokens/rooms/%s", roomUUID)

	requestBody := RoomTokenRequest{
		Lifespan: lifespan,
		Role:     role,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers - Use pre-generated SDK token from console
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", config.SDKToken)
	req.Header.Set("region", config.Region)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Response is just the token string
	token := string(body)
	// Remove quotes if present
	if len(token) > 2 && token[0] == '"' && token[len(token)-1] == '"' {
		token = token[1 : len(token)-1]
	}

	return token, nil
}
