package controllers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	rtctokenbuilder "github.com/AgoraIO-Community/go-tokenbuilder/rtctokenbuilder"
)

// AgoraTokenRequest represents the request structure for token generation
type AgoraTokenRequest struct {
	ChannelName             string `json:"channelName" binding:"required"`
	UserRole                string `json:"userRole"`
	UID                     uint32 `json:"uid"`
	ExpirationTimeInSeconds uint32 `json:"expirationTimeInSeconds"`
}

// AgoraTokenResponse represents the response structure
type AgoraTokenResponse struct {
	Token     *string `json:"token"`
	AppID     string  `json:"appId"`
	Channel   string  `json:"channel"`
	UID       uint32  `json:"uid"`
	ExpiresAt int64   `json:"expiresAt"`
}

// Simple token generation using HMAC (for basic authentication)
func generateSimpleToken(appID, appCertificate, channelName string, uid uint32, role string, expireTime uint32) string {
	// Create a simple payload
	payload := map[string]interface{}{
		"appId":     appID,
		"channel":   channelName,
		"uid":       uid,
		"role":      role,
		"expireAt":  expireTime,
		"timestamp": time.Now().Unix(),
	}

	payloadBytes, _ := json.Marshal(payload)
	payloadBase64 := base64.StdEncoding.EncodeToString(payloadBytes)

	// Create HMAC signature
	h := hmac.New(sha256.New, []byte(appCertificate))
	h.Write(payloadBytes)
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return payloadBase64 + "." + signature
}

// GenerateAgoraToken generates an Agora RTC token
func GenerateAgoraToken(c *gin.Context) {
	var request AgoraTokenRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get Agora credentials from environment
	appID := os.Getenv("AGORA_APP_ID")
	appCertificate := os.Getenv("AGORA_APP_CERTIFICATE")

	if appID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Agora App ID not configured",
		})
		return
	}

	// Set default values
	if request.ExpirationTimeInSeconds == 0 {
		request.ExpirationTimeInSeconds = 3600 // 1 hour default
	}

	if request.UID == 0 {
		request.UID = 0 // Let Agora assign random UID
	}

	// Calculate expiration time
	expireTimeInSeconds := uint32(time.Now().Unix()) + request.ExpirationTimeInSeconds

	var token *string

	if appCertificate != "" && appCertificate != "YOUR_AGORA_APP_CERTIFICATE" {
		var role rtctokenbuilder.Role
		if request.UserRole == "teacher" || request.UserRole == "host" || request.UserRole == "publisher" || request.UserRole == "student" || request.UserRole == "audience" {
			role = rtctokenbuilder.RolePublisher
		} else {
			role = rtctokenbuilder.RoleSubscriber
		}

		tokenStr, err := rtctokenbuilder.BuildTokenWithUid(appID, appCertificate, request.ChannelName, request.UID, role, expireTimeInSeconds)
		if err == nil {
			token = &tokenStr
		}
	} else {
		// Return null token when no certificate (for development)
		token = nil
	}

	response := AgoraTokenResponse{
		Token:     token,
		AppID:     appID,
		Channel:   request.ChannelName,
		UID:       request.UID,
		ExpiresAt: int64(expireTimeInSeconds),
	}

	c.JSON(http.StatusOK, response)
}

// GetAgoraConfig returns Agora configuration for client
func GetAgoraConfig(c *gin.Context) {
	appID := os.Getenv("AGORA_APP_ID")
	appCertificate := os.Getenv("AGORA_APP_CERTIFICATE")

	c.JSON(http.StatusOK, gin.H{
		"appId":              appID,
		"certificateEnabled": appCertificate != "",
		"mode":               "production",
		"tokenRequired":      appCertificate != "",
	})
}
