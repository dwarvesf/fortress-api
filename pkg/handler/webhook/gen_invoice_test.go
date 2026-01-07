package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/ratelimit"
)

func TestIsValidMonthFormat(t *testing.T) {
	tests := []struct {
		name     string
		month    string
		expected bool
	}{
		{
			name:     "valid_format_2026-01",
			month:    "2026-01",
			expected: true,
		},
		{
			name:     "valid_format_2025-12",
			month:    "2025-12",
			expected: true,
		},
		{
			name:     "invalid_format_no_dash",
			month:    "202601",
			expected: false,
		},
		{
			name:     "invalid_format_single_digit_month",
			month:    "2026-1",
			expected: false,
		},
		{
			name:     "invalid_format_three_digit_year",
			month:    "202-01",
			expected: false,
		},
		{
			name:     "invalid_format_with_day",
			month:    "2026-01-01",
			expected: false,
		},
		{
			name:     "invalid_format_empty",
			month:    "",
			expected: false,
		},
		{
			name:     "invalid_format_letters",
			month:    "abcd-ef",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidMonthFormat(tt.month)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFileIDFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "valid_drive_url",
			url:      "https://drive.google.com/file/d/1ABC123xyz_-def/view",
			expected: "1ABC123xyz_-def",
		},
		{
			name:     "valid_drive_url_long_id",
			url:      "https://drive.google.com/file/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/view",
			expected: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
		},
		{
			name:     "url_without_view",
			url:      "https://drive.google.com/file/d/1ABC123/",
			expected: "1ABC123/",
		},
		{
			name:     "empty_url",
			url:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFileIDFromURL(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleGenInvoice_InvalidJSON(t *testing.T) {
	// Setup
	l := logger.NewLogrusLogger("debug")
	rl := ratelimit.NewInvoiceRateLimiter(l)
	defer rl.Stop()
	SetInvoiceRateLimiter(rl)

	h := &handler{
		logger: l,
	}

	// Create request with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/webhooks/discord/gen-invoice", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	h.HandleGenInvoice(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGenInvoice_MissingFields(t *testing.T) {
	// Setup
	l := logger.NewLogrusLogger("debug")
	rl := ratelimit.NewInvoiceRateLimiter(l)
	defer rl.Stop()
	SetInvoiceRateLimiter(rl)

	h := &handler{
		logger: l,
	}

	tests := []struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "missing_discord_username",
			payload: map[string]string{"month": "2026-01", "dm_channel_id": "123", "dm_message_id": "456"},
		},
		{
			name:    "missing_month",
			payload: map[string]string{"discord_username": "user1", "dm_channel_id": "123", "dm_message_id": "456"},
		},
		{
			name:    "missing_dm_channel_id",
			payload: map[string]string{"discord_username": "user1", "month": "2026-01", "dm_message_id": "456"},
		},
		{
			name:    "missing_dm_message_id",
			payload: map[string]string{"discord_username": "user1", "month": "2026-01", "dm_channel_id": "123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.payload)
			c.Request = httptest.NewRequest("POST", "/webhooks/discord/gen-invoice", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.HandleGenInvoice(c)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestHandleGenInvoice_InvalidMonthFormat(t *testing.T) {
	// Setup
	l := logger.NewLogrusLogger("debug")
	rl := ratelimit.NewInvoiceRateLimiter(l)
	defer rl.Stop()
	SetInvoiceRateLimiter(rl)

	h := &handler{
		logger: l,
	}

	// Create request with invalid month format
	payload := GenInvoiceRequest{
		DiscordUsername: "testuser",
		Month:           "2026-1", // Invalid format
		DMChannelID:     "123456",
		DMMessageID:     "789012",
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest("POST", "/webhooks/discord/gen-invoice", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	h.HandleGenInvoice(c)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response GenInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Invalid month format")
}

func TestHandleGenInvoice_RateLimitExceeded(t *testing.T) {
	// Setup
	l := logger.NewLogrusLogger("debug")
	rl := ratelimit.NewInvoiceRateLimiter(l)
	defer rl.Stop()
	SetInvoiceRateLimiter(rl)

	// Use mock service to avoid nil pointer in goroutine
	serviceMock := service.NewForTest()

	h := &handler{
		logger:  l,
		service: serviceMock,
	}

	username := "rate_limited_user"

	// Exhaust rate limit (3 requests)
	for i := 0; i < ratelimit.MaxInvoiceGenerationsPerDay; i++ {
		_ = rl.CheckLimit(username)
	}

	// Create request
	payload := GenInvoiceRequest{
		DiscordUsername: username,
		Month:           "2026-01",
		DMChannelID:     "123456",
		DMMessageID:     "789012",
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest("POST", "/webhooks/discord/gen-invoice", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	h.HandleGenInvoice(c)

	// Assert
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var response GenInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Rate limit exceeded")

	// Wait for goroutine to complete
	time.Sleep(100 * time.Millisecond)
}

func TestHandleGenInvoice_RateLimiterNotInitialized(t *testing.T) {
	// Setup - set rate limiter to nil
	SetInvoiceRateLimiter(nil)

	l := logger.NewLogrusLogger("debug")
	h := &handler{
		logger: l,
	}

	// Create request
	payload := GenInvoiceRequest{
		DiscordUsername: "testuser",
		Month:           "2026-01",
		DMChannelID:     "123456",
		DMMessageID:     "789012",
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(payload)
	c.Request = httptest.NewRequest("POST", "/webhooks/discord/gen-invoice", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	// Execute
	h.HandleGenInvoice(c)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response GenInvoiceResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "Rate limiter not configured")
}

func TestGenInvoiceRequest_JSONBinding(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected GenInvoiceRequest
	}{
		{
			name: "full_request",
			json: `{"discord_username":"testuser","month":"2026-01","dm_channel_id":"123","dm_message_id":"456"}`,
			expected: GenInvoiceRequest{
				DiscordUsername: "testuser",
				Month:           "2026-01",
				DMChannelID:     "123",
				DMMessageID:     "456",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req GenInvoiceRequest
			err := json.Unmarshal([]byte(tt.json), &req)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, req)
		})
	}
}

func TestGenInvoiceResponse_JSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response GenInvoiceResponse
		expected string
	}{
		{
			name: "success_response",
			response: GenInvoiceResponse{
				Success: true,
				Message: "Invoice generation started",
			},
			expected: `{"success":true,"message":"Invoice generation started"}`,
		},
		{
			name: "error_response",
			response: GenInvoiceResponse{
				Success: false,
				Message: "Rate limit exceeded",
			},
			expected: `{"success":false,"message":"Rate limit exceeded"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}
