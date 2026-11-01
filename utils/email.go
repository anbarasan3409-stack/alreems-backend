package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// ResendEmailPayload represents the JSON payload expected by Resend REST API
type ResendEmailPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendEmail sends an HTML or text email using Resend API (primary) or SMTP (fallback)
func SendEmail(to []string, subject, body string) error {
	if len(to) == 0 {
		return nil
	}

	resendAPIKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	resendFrom := strings.TrimSpace(os.Getenv("RESEND_FROM"))

	// Ignore placeholder keys
	if resendAPIKey == "your_resend_api_key_here" || strings.HasPrefix(resendAPIKey, "YOUR_") {
		resendAPIKey = ""
	}

	if resendFrom == "" || strings.Contains(resendFrom, "onboarding@resend.dev") {
		resendFrom = "Al Reems Academy <info@reems.co.in>"
	}

	// 1. Primary: Use Resend REST API if RESEND_API_KEY is configured
	if resendAPIKey != "" {
		err := sendViaResend(resendAPIKey, resendFrom, to, subject, body)
		if err == nil {
			return nil
		}
		log.Printf("[Email] Resend dispatch failed (%v). Checking SMTP fallback...", err)
	}

	// 2. Fallback: Use standard SMTP if SMTP credentials are configured
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("SMTP_FROM")
	if fromEmail == "" {
		fromEmail = resendFrom
	}

	if smtpHost != "" && smtpPort != "" && smtpUser != "" {
		return sendViaSMTP(smtpHost, smtpPort, smtpUser, smtpPass, fromEmail, to, subject, body)
	}

	log.Printf("[Email] Neither valid RESEND_API_KEY nor SMTP credentials configured. Skipping email send to %v.", to)
	return fmt.Errorf("email delivery unavailable: no valid email credentials configured")
}

// sendViaResend dispatches emails via the official Resend REST API (https://api.resend.com/emails)
func sendViaResend(apiKey, fromEmail string, to []string, subject, body string) error {
	if fromEmail == "" {
		fromEmail = "Al Reems Academy <onboarding@resend.dev>"
	}

	payload := ResendEmailPayload{
		From:    fromEmail,
		To:      to,
		Subject: subject,
		HTML:    body,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Resend email payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to construct Resend request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Dispatching email via Resend API to %v with subject: %s", to, subject)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Resend HTTP request failed: %v", err)
		return fmt.Errorf("resend request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("Resend API rejected email (status %d): %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("resend error (status %d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("Email successfully sent via Resend API to %v! Response: %s", to, string(respBody))
	return nil
}

// sendViaSMTP dispatches emails using standard SMTP as fallback
func sendViaSMTP(smtpHost, smtpPort, smtpUser, smtpPass, fromEmail string, to []string, subject, body string) error {
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	headers := make(map[string]string)
	headers["From"] = fromEmail
	headers["To"] = to[0]
	headers["Subject"] = subject
	headers["MIME-version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var message bytes.Buffer
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + body)

	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)

	log.Printf("Dispatching email via SMTP to %v with subject: %s", to, subject)
	err := smtp.SendMail(addr, auth, fromEmail, to, message.Bytes())
	if err != nil {
		log.Printf("Failed to send email via SMTP to %v: %v", to, err)
		return err
	}

	log.Println("Email sent successfully via SMTP!")
	return nil
}
