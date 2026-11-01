package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// SendWhatsAppMessage sends a WhatsApp message using a third-party API (like Twilio or Meta WhatsApp Business API)
func SendWhatsAppMessage(toPhone, message string) error {
	whatsappApiUrl := os.Getenv("WHATSAPP_API_URL")
	whatsappApiToken := os.Getenv("WHATSAPP_API_TOKEN")

	if whatsappApiUrl == "" || whatsappApiToken == "" {
		log.Println("WhatsApp API credentials not configured. Skipping WhatsApp send.")
		// Return nil to not block execution in dev mode
		return nil
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                toPhone,
		"type":              "text",
		"text": map[string]string{
			"body": message,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal whatsapp payload: %v", err)
	}

	req, err := http.NewRequest("POST", whatsappApiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create whatsapp request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+whatsappApiToken)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Sending WhatsApp message to %s", toPhone)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send WhatsApp message to %s: %v", toPhone, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("WhatsApp API returned error %d: %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("whatsapp API error: %d", resp.StatusCode)
	}

	log.Println("WhatsApp message sent successfully!")
	return nil
}
