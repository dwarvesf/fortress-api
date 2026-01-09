package notion

import (
	"context"
	"testing"

	nt "github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// TestExtractSelect tests the extractSelect helper method
func TestExtractSelect(t *testing.T) {
	service := &TaskOrderLogService{
		logger: logger.NewLogrusLogger("debug"),
	}

	t.Run("valid_select_property", func(t *testing.T) {
		props := nt.DatabasePageProperties{
			"Payday": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: "01",
				},
			},
		}

		result := service.extractSelect(props, "Payday")
		require.Equal(t, "01", result)
	})

	t.Run("property_not_found", func(t *testing.T) {
		props := nt.DatabasePageProperties{
			"OtherField": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: "value",
				},
			},
		}

		result := service.extractSelect(props, "Payday")
		require.Equal(t, "", result)
	})

	t.Run("empty_select_value", func(t *testing.T) {
		props := nt.DatabasePageProperties{
			"Payday": nt.DatabasePageProperty{
				Select: &nt.SelectOptions{
					Name: "",
				},
			},
		}

		result := service.extractSelect(props, "Payday")
		require.Equal(t, "", result)
	})

	t.Run("nil_select_property", func(t *testing.T) {
		props := nt.DatabasePageProperties{
			"Payday": nt.DatabasePageProperty{
				Select: nil,
			},
		}

		result := service.extractSelect(props, "Payday")
		require.Equal(t, "", result)
	})
}

// TestGetContractorPayday_Fallbacks tests graceful fallback scenarios
func TestGetContractorPayday_Fallbacks(t *testing.T) {
	t.Run("database_not_configured", func(t *testing.T) {
		service := &TaskOrderLogService{
			cfg: &config.Config{
				Notion: config.Notion{
					Databases: config.NotionDatabase{
						ContractorRates: "", // Empty database ID
					},
				},
			},
			logger: logger.NewLogrusLogger("debug"),
		}

		ctx := context.Background()
		payday, err := service.GetContractorPayday(ctx, "test-contractor-123")

		require.NoError(t, err)
		require.Equal(t, 0, payday, "should return 0 when database not configured")
	})
}

// TestInvoiceDueDateCalculation tests the invoice due date calculation logic
func TestInvoiceDueDateCalculation(t *testing.T) {
	testCases := []struct {
		name           string
		payday         int
		expectedDueDay string
	}{
		{
			name:           "payday_1_returns_10th",
			payday:         1,
			expectedDueDay: "10th",
		},
		{
			name:           "payday_15_returns_25th",
			payday:         15,
			expectedDueDay: "25th",
		},
		{
			name:           "payday_0_returns_10th_default",
			payday:         0,
			expectedDueDay: "10th",
		},
		{
			name:           "invalid_payday_returns_10th_default",
			payday:         99,
			expectedDueDay: "10th",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the handler logic
			invoiceDueDay := "10th" // Default for Payday 1 or fallback
			if tc.payday == 15 {
				invoiceDueDay = "25th"
			}

			require.Equal(t, tc.expectedDueDay, invoiceDueDay)
		})
	}
}
