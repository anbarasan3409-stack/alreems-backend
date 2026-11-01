package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	rtctokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
)

var (
	appID          = getEnvOrDefault("AGORA_APP_ID", "YOUR_AGORA_APP_ID")
	appCertificate = getEnvOrDefault("AGORA_APP_CERTIFICATE", "YOUR_AGORA_APP_CERTIFICATE")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateChannelName creates a unique channel name for Agora sessions
func GenerateChannelName() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("session_%s_%d", hex.EncodeToString(bytes), time.Now().Unix())
}

// GetAgoraAppID returns the Agora App ID for frontend usage
func GetAgoraAppID() string {
	if id := os.Getenv("AGORA_APP_ID"); id != "" {
		return id
	}
	return appID
}

// GenerateStudentAgoraToken generates a token for students to join sessions
func GenerateStudentAgoraToken(channelName, uidStr, roleStr string) (string, error) {
	currentAppID := GetAgoraAppID()
	currentAppCert := os.Getenv("AGORA_APP_CERTIFICATE")
	if currentAppCert == "" {
		currentAppCert = appCertificate
	}

	if currentAppID == "YOUR_AGORA_APP_ID" || currentAppCert == "YOUR_AGORA_APP_CERTIFICATE" || currentAppCert == "" {
		return "", nil // Return empty token for development without certificate
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid uid: %s", uidStr)
	}

	var role rtctokenbuilder.Role
	if roleStr == "teacher" || roleStr == "host" || roleStr == "publisher" || roleStr == "student" || roleStr == "audience" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime := uint32(time.Now().Unix() + 3600) // 1 hour token
	token, err := rtctokenbuilder.BuildTokenWithUid(currentAppID, currentAppCert, channelName, uint32(uid), role, expireTime)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}

	return token, nil
}

func GenerateAgoraToken(w http.ResponseWriter, r *http.Request) {
	channelName := r.URL.Query().Get("channelName")
	uidStr := r.URL.Query().Get("uid")
	roleStr := r.URL.Query().Get("role") // "teacher" or "student"

	if channelName == "" || uidStr == "" || roleStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing parameters"))
		return
	}

	currentAppID := GetAgoraAppID()
	currentAppCert := os.Getenv("AGORA_APP_CERTIFICATE")
	if currentAppCert == "" {
		currentAppCert = appCertificate
	}

	if currentAppCert == "" || currentAppCert == "YOUR_AGORA_APP_CERTIFICATE" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token":null}`))
		return
	}

	uid, _ := strconv.ParseUint(uidStr, 10, 32)
	var role rtctokenbuilder.Role
	if roleStr == "teacher" || roleStr == "host" || roleStr == "publisher" || roleStr == "student" || roleStr == "audience" {
		role = rtctokenbuilder.RolePublisher
	} else {
		role = rtctokenbuilder.RoleSubscriber
	}

	expireTime := uint32(time.Now().Unix() + 3600) // 1 hour token
	token, err := rtctokenbuilder.BuildTokenWithUid(currentAppID, currentAppCert, channelName, uint32(uid), role, expireTime)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error generating token"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"token":"%s"}`, token)))
}

// GenerateAgoraTokenForRecording generates a token for the recording bot
func GenerateAgoraTokenForRecording(channelName, uidStr string) (string, error) {
	if appID == "YOUR_AGORA_APP_ID" || appCertificate == "YOUR_AGORA_APP_CERTIFICATE" {
		return "mock_recording_token_for_development", nil
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		return "", fmt.Errorf("invalid uid: %s", uidStr)
	}

	// Recording bots need publisher role
	role := rtctokenbuilder.RolePublisher
	expireTime := uint32(time.Now().Unix() + 86400) // 24 hour token for recordings

	token, err := rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, channelName, uint32(uid), rtctokenbuilder.Role(role), expireTime)
	if err != nil {
		return "", fmt.Errorf("error generating recording token: %v", err)
	}

	return token, nil
}
