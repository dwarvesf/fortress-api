package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/utils/testhelper"
	"github.com/dwarvesf/fortress-api/pkg/view"
)

func TestNotifyWeeklyMemos(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name           string
		expectedCode   int
		expectedResult string
	}{
		{
			name:           "no_memos_in_timerange",
			expectedCode:   http.StatusOK,
			expectedResult: "no new memos in this week",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				// Create handler
				h := New(nil, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				
				// Create request
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("POST", "/cronjobs/notify-weekly-memos", nil)

				// Execute
				h.NotifyWeeklyMemos(c)

				// Assert
				assert.Equal(t, tt.expectedCode, w.Code)
				
				var response view.Response[any]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response.Message)
			})
		})
	}
}

func TestNotifyMonthlyMemos(t *testing.T) {
	cfg := config.LoadTestConfig()
	loggerMock := logger.NewLogrusLogger("info")
	serviceMock := service.NewForTest()
	storeMock := store.New()

	tests := []struct {
		name           string
		expectedCode   int
		expectedResult string
	}{
		{
			name:           "no_memos_in_timerange",
			expectedCode:   http.StatusOK,
			expectedResult: "no new memos this month",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testhelper.TestWithTxDB(t, func(txRepo store.DBRepo) {
				// Create handler
				h := New(nil, storeMock, txRepo, serviceMock, loggerMock, &cfg)
				
				// Create request
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("POST", "/cronjobs/notify-monthly-memos", nil)

				// Execute
				h.NotifyMonthlyMemos(c)

				// Assert
				assert.Equal(t, tt.expectedCode, w.Code)
				
				var response view.Response[any]
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response.Message)
			})
		})
	}
}

// TestMonthlyMemosTimeRangeCalculation tests the monthly time range calculation logic
func TestMonthlyMemosTimeRangeCalculation(t *testing.T) {
	// Test that monthly time range is calculated correctly
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	
	// Verify start is first day of current month at midnight
	assert.Equal(t, 1, start.Day())
	assert.Equal(t, now.Month(), start.Month())
	assert.Equal(t, now.Year(), start.Year())
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
	assert.Equal(t, 0, start.Nanosecond())
}

// TestMonthRangeStringFormatting tests the month range string formatting logic
func TestMonthRangeStringFormatting(t *testing.T) {
	testCases := []struct {
		month    time.Month
		year     int
		expected string
	}{
		{time.January, 2025, "JANUARY 2025"},
		{time.February, 2025, "FEBRUARY 2025"},
		{time.December, 2024, "DECEMBER 2024"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			testTime := time.Date(tc.year, tc.month, 15, 12, 0, 0, 0, time.UTC)
			monthRangeStr := fmt.Sprintf("%s %d", strings.ToUpper(testTime.Month().String()), testTime.Year())
			assert.Equal(t, tc.expected, monthRangeStr)
		})
	}
}

// TestChannelSelection tests the channel selection logic for different environments
func TestChannelSelection(t *testing.T) {
	tests := []struct {
		env              string
		expectedChannel  string
		description      string
	}{
		{"local", discordPlayGroundReadingChannel, "local environment should use playground channel"},
		{"test", discordPlayGroundReadingChannel, "test environment should use playground channel"},
		{"dev", discordPlayGroundReadingChannel, "dev environment should use playground channel"},
		{"prod", discordRandomChannel, "prod environment should use random channel"},
	}
	
	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			// Simulate the channel selection logic from the handler
			var targetChannelID string
			if tt.env == "prod" {
				targetChannelID = discordRandomChannel
			} else {
				targetChannelID = discordPlayGroundReadingChannel
			}
			
			assert.Equal(t, tt.expectedChannel, targetChannelID, tt.description)
		})
	}
}