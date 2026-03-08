package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const headerSignature256 = "X-Payrune-Signature-256"

type paymentReceiptStatusChangedPayload struct {
	EventType      string `json:"eventType"`
	EventVersion   int    `json:"eventVersion"`
	NotificationID int64  `json:"notificationId"`
}

func newFakeWebhookHandler(logger *log.Logger, secret string) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Printf("fake webhook read body failed method=%s path=%s err=%v", r.Method, r.URL.Path, err)
			http.Error(w, "read body failed", http.StatusInternalServerError)
			return
		}

		result := verifyWebhookRequest(secret, r, body)
		logger.Printf(
			"fake webhook request:\n  method=%s\n  path=%s\n  headers:\n%s\n  raw_body:\n%s",
			r.Method,
			r.URL.Path,
			indentMultiline(formatHeadersForLog(r.Header), "    "),
			indentMultiline(formatBodyForLog(body), "    "),
		)

		if result.DecodeErr != nil {
			logger.Printf(
				"fake webhook payload invalid_json:\n  err=%v\n  raw_body=%s",
				result.DecodeErr,
				string(body),
			)
			http.Error(w, "invalid json payload", http.StatusBadRequest)
			return
		}

		if result.VerificationEnabled && !result.SignatureValid {
			logger.Printf(
				"fake webhook signature invalid:\n  method=hmac-sha256\n  event=%s\n  version=%s\n  notification_id_header=%s\n  notification_id_match=%t\n  provided_signature=%q\n  computed_signature=%q",
				result.Event,
				result.Version,
				result.NotificationIDHeader,
				result.NotificationIDMatch,
				result.ProvidedSignature,
				result.ComputedSignature,
			)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})
}

type webhookVerificationResult struct {
	Event                string
	Version              string
	NotificationIDHeader string
	ProvidedSignature    string
	ComputedSignature    string
	SignatureValid       bool
	VerificationEnabled  bool
	NotificationIDMatch  bool
	Payload              paymentReceiptStatusChangedPayload
	DecodeErr            error
}

func verifyWebhookRequest(secret string, r *http.Request, body []byte) webhookVerificationResult {
	result := webhookVerificationResult{
		Event:                strings.TrimSpace(r.Header.Get("X-Payrune-Event")),
		Version:              strings.TrimSpace(r.Header.Get("X-Payrune-Event-Version")),
		NotificationIDHeader: strings.TrimSpace(r.Header.Get("X-Payrune-Notification-ID")),
		ProvidedSignature:    strings.TrimSpace(r.Header.Get(headerSignature256)),
		VerificationEnabled:  strings.TrimSpace(secret) != "",
	}

	if result.VerificationEnabled {
		result.ComputedSignature = "sha256=" + computeWebhookSignature([]byte(secret), body)
		result.SignatureValid = hmac.Equal([]byte(result.ComputedSignature), []byte(result.ProvidedSignature))
	}

	var payload paymentReceiptStatusChangedPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		result.DecodeErr = err
		return result
	}

	result.Payload = payload
	result.NotificationIDMatch = result.NotificationIDHeader == strconv.FormatInt(payload.NotificationID, 10)
	return result
}

func computeWebhookSignature(secret []byte, body []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func marshalHeadersJSON(header http.Header) string {
	if header == nil {
		return "{}"
	}

	bytes, err := json.Marshal(header)
	if err != nil {
		return `{"_marshal_error":"header_json_failed"}`
	}
	return string(bytes)
}

func formatHeadersForLog(header http.Header) string {
	return prettyPrintJSON([]byte(marshalHeadersJSON(header)))
}

func formatBodyForLog(body []byte) string {
	return prettyPrintJSON(body)
}

func prettyPrintJSON(raw []byte) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return "{}"
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, []byte(trimmed), "", "  "); err != nil {
		return trimmed
	}
	return pretty.String()
}

func indentMultiline(text string, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
