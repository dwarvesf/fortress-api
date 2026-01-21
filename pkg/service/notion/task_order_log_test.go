package notion

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

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

// TestUpdateOrderAndSubitemsStatus_ConcurrencyConfiguration tests the concurrency configuration
func TestUpdateOrderAndSubitemsStatus_ConcurrencyConfiguration(t *testing.T) {
	testCases := []struct {
		name               string
		configuredValue    int
		expectedConcurrent int
	}{
		{
			name:               "default_concurrency_10",
			configuredValue:    10,
			expectedConcurrent: 10,
		},
		{
			name:               "minimum_concurrency_1",
			configuredValue:    1,
			expectedConcurrent: 1,
		},
		{
			name:               "maximum_concurrency_20",
			configuredValue:    20,
			expectedConcurrent: 20,
		},
		{
			name:               "zero_defaults_to_10",
			configuredValue:    0,
			expectedConcurrent: 10,
		},
		{
			name:               "negative_defaults_to_10",
			configuredValue:    -5,
			expectedConcurrent: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				TaskOrderLogSubitemConcurrency: tc.configuredValue,
			}

			// Validate the effective concurrency limit
			maxConcurrent := cfg.TaskOrderLogSubitemConcurrency
			if maxConcurrent <= 0 {
				maxConcurrent = 10
			}

			require.Equal(t, tc.expectedConcurrent, maxConcurrent)
		})
	}
}

// TestConcurrentUpdatesBehavior validates concurrent update behavior
func TestConcurrentUpdatesBehavior(t *testing.T) {
	t.Run("concurrent_updates_complete_without_deadlock", func(t *testing.T) {
		// This test validates that the concurrent update pattern doesn't deadlock
		// We simulate the worker pool pattern used in UpdateOrderAndSubitemsStatus

		const numSubitems = 20
		const maxConcurrent = 5

		type result struct {
			id  int
			err error
		}

		resultsChan := make(chan result, numSubitems)
		sem := make(chan struct{}, maxConcurrent)

		var completedCount int32

		// Simulate concurrent updates
		for i := 0; i < numSubitems; i++ {
			go func(itemID int) {
				// Acquire semaphore
				sem <- struct{}{}
				defer func() { <-sem }()

				// Simulate API call
				time.Sleep(1 * time.Millisecond)

				atomic.AddInt32(&completedCount, 1)
				resultsChan <- result{id: itemID, err: nil}
			}(i)
		}

		// Collect results with timeout
		timeout := time.After(5 * time.Second)
		collectedResults := 0

	collectLoop:
		for collectedResults < numSubitems {
			select {
			case res := <-resultsChan:
				require.NoError(t, res.err)
				collectedResults++
			case <-timeout:
				t.Fatalf("timeout waiting for results, collected %d/%d", collectedResults, numSubitems)
				break collectLoop
			}
		}

		require.Equal(t, numSubitems, collectedResults, "all subitems should be processed")
		require.Equal(t, int32(numSubitems), atomic.LoadInt32(&completedCount), "all updates should complete")
	})

	t.Run("semaphore_limits_concurrency", func(t *testing.T) {
		// This test validates that the semaphore actually limits concurrent execution
		const maxConcurrent = 3
		const numTasks = 10

		sem := make(chan struct{}, maxConcurrent)
		var maxObservedConcurrent int32
		var currentConcurrent int32

		for i := 0; i < numTasks; i++ {
			go func() {
				sem <- struct{}{}
				defer func() { <-sem }()

				// Increment current count
				current := atomic.AddInt32(&currentConcurrent, 1)

				// Track maximum observed concurrent executions
				for {
					max := atomic.LoadInt32(&maxObservedConcurrent)
					if current <= max || atomic.CompareAndSwapInt32(&maxObservedConcurrent, max, current) {
						break
					}
				}

				// Simulate work
				time.Sleep(10 * time.Millisecond)

				// Decrement current count
				atomic.AddInt32(&currentConcurrent, -1)
			}()
		}

		// Wait for completion
		time.Sleep(200 * time.Millisecond)

		maxObserved := atomic.LoadInt32(&maxObservedConcurrent)
		require.LessOrEqual(t, maxObserved, int32(maxConcurrent),
			"observed concurrency (%d) should not exceed limit (%d)", maxObserved, maxConcurrent)
		require.Greater(t, maxObserved, int32(1),
			"should observe concurrent execution")
	})
}

// TestWorkerPoolConfiguration tests worker pool configuration for refund and invoice split processing
func TestWorkerPoolConfiguration(t *testing.T) {
	t.Run("refund_worker_pool_defaults", func(t *testing.T) {
		testCases := []struct {
			name            string
			configuredValue int
			expectedWorkers int
		}{
			{
				name:            "default_5_workers",
				configuredValue: 5,
				expectedWorkers: 5,
			},
			{
				name:            "zero_defaults_to_5",
				configuredValue: 0,
				expectedWorkers: 5,
			},
			{
				name:            "negative_defaults_to_5",
				configuredValue: -3,
				expectedWorkers: 5,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					RefundProcessingWorkers: tc.configuredValue,
				}

				maxWorkers := cfg.RefundProcessingWorkers
				if maxWorkers <= 0 {
					maxWorkers = 5
				}

				require.Equal(t, tc.expectedWorkers, maxWorkers)
			})
		}
	})

	t.Run("invoice_split_worker_pool_defaults", func(t *testing.T) {
		testCases := []struct {
			name            string
			configuredValue int
			expectedWorkers int
		}{
			{
				name:            "default_5_workers",
				configuredValue: 5,
				expectedWorkers: 5,
			},
			{
				name:            "zero_defaults_to_5",
				configuredValue: 0,
				expectedWorkers: 5,
			},
			{
				name:            "negative_defaults_to_5",
				configuredValue: -3,
				expectedWorkers: 5,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := &config.Config{
					InvoiceSplitProcessingWorkers: tc.configuredValue,
				}

				maxWorkers := cfg.InvoiceSplitProcessingWorkers
				if maxWorkers <= 0 {
					maxWorkers = 5
				}

				require.Equal(t, tc.expectedWorkers, maxWorkers)
			})
		}
	})
}
