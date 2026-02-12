package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	nt "github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newNotionTestClient(t *testing.T, handler func(*http.Request) (*http.Response, error)) *nt.Client {
	t.Helper()

	httpClient := &http.Client{
		Transport: roundTripFunc(handler),
	}

	return nt.NewClient("test-secret", nt.WithHTTPClient(httpClient))
}

func TestGetLatestPayoutDateByDiscord_QueryAndParse(t *testing.T) {
	dbID := "db-id"
	discord := "adeki_"
	wantDate := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	latestDt, err := nt.ParseDateTime("2026-01-10")
	require.NoError(t, err)
	olderDt, err := nt.ParseDateTime("2025-09-30")
	require.NoError(t, err)
	oldestDate := time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC)

	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/databases/"+dbID+"/query", r.URL.Path)

		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var query nt.DatabaseQuery
		require.NoError(t, json.Unmarshal(bodyBytes, &query))
		require.NotNil(t, query.Filter)
		require.Len(t, query.Filter.And, 2)

		discordFilter := query.Filter.And[0]
		require.Equal(t, "Discord", discordFilter.Property)
		require.NotNil(t, discordFilter.Rollup)
		require.NotNil(t, discordFilter.Rollup.Any)
		require.NotNil(t, discordFilter.Rollup.Any.RichText)
		require.Equal(t, discord, discordFilter.Rollup.Any.RichText.Contains)

		dateFilter := query.Filter.And[1]
		require.Equal(t, "Date", dateFilter.Property)
		require.NotNil(t, dateFilter.Date)
		require.True(t, dateFilter.Date.IsNotEmpty)

		require.Len(t, query.Sorts, 1)
		require.Equal(t, "Date", query.Sorts[0].Property)
		require.Equal(t, nt.SortDirDesc, query.Sorts[0].Direction)
		require.Equal(t, 5, query.PageSize)

		resp := nt.DatabaseQueryResponse{
			Results: []nt.Page{
				{
					ID: "page-1",
					Parent: nt.Parent{
						Type:       nt.ParentTypeDatabase,
						DatabaseID: dbID,
					},
					Properties: nt.DatabasePageProperties{
						"Date": nt.DatabasePageProperty{
							Date: &nt.Date{Start: olderDt},
						},
					},
				},
				{
					ID: "page-2",
					Parent: nt.Parent{
						Type:       nt.ParentTypeDatabase,
						DatabaseID: dbID,
					},
					Properties: nt.DatabasePageProperties{
						"Date": nt.DatabasePageProperty{
							Date: &nt.Date{Start: latestDt},
						},
					},
				},
			},
		}

		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	service := &ContractorPayoutsService{
		client: client,
		cfg: &config.Config{
			Notion: config.Notion{
				Databases: config.NotionDatabase{
					ContractorPayouts: dbID,
				},
			},
		},
		logger: logger.NewLogrusLogger("debug"),
	}

	got, err := service.GetLatestPayoutDateByDiscord(context.Background(), discord)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, got.Equal(wantDate))
	require.NotEqual(t, oldestDate, *got)
}

func TestGetLatestPayoutDateByDiscord_NoResults(t *testing.T) {
	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		resp := nt.DatabaseQueryResponse{
			Results: []nt.Page{},
		}

		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	service := &ContractorPayoutsService{
		client: client,
		cfg: &config.Config{
			Notion: config.Notion{
				Databases: config.NotionDatabase{
					ContractorPayouts: "db-id",
				},
			},
		},
		logger: logger.NewLogrusLogger("debug"),
	}

	got, err := service.GetLatestPayoutDateByDiscord(context.Background(), "adeki_")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestGetLatestPayoutDateByDiscord_EmptyDate(t *testing.T) {
	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		resp := nt.DatabaseQueryResponse{
			Results: []nt.Page{
				{
					ID: "page-1",
					Parent: nt.Parent{
						Type:       nt.ParentTypeDatabase,
						DatabaseID: "db-id",
					},
					Properties: nt.DatabasePageProperties{
						"Date": nt.DatabasePageProperty{
							Date: nil,
						},
					},
				},
			},
		}

		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	service := &ContractorPayoutsService{
		client: client,
		cfg: &config.Config{
			Notion: config.Notion{
				Databases: config.NotionDatabase{
					ContractorPayouts: "db-id",
				},
			},
		},
		logger: logger.NewLogrusLogger("debug"),
	}

	got, err := service.GetLatestPayoutDateByDiscord(context.Background(), "adeki_")
	require.NoError(t, err)
	require.Nil(t, got)
}

func TestDetermineSourceType(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	service := &ContractorPayoutsService{logger: l}

	tests := []struct {
		name     string
		entry    PayoutEntry
		expected PayoutSourceType
	}{
		// TaskOrder takes priority
		{
			name:     "TaskOrder present - always ServiceFee",
			entry:    PayoutEntry{TaskOrderID: "task-001"},
			expected: PayoutSourceTypeServiceFee,
		},
		// InvoiceSplit: role-only descriptions → ServiceFee
		{
			name:     "InvoiceSplit with Delivery Lead role → ServiceFee",
			entry:    PayoutEntry{InvoiceSplitID: "split-001", Description: "Delivery Lead"},
			expected: PayoutSourceTypeServiceFee,
		},
		{
			name:     "InvoiceSplit with Account Management role → ServiceFee",
			entry:    PayoutEntry{InvoiceSplitID: "split-002", Description: "Account Management"},
			expected: PayoutSourceTypeServiceFee,
		},
		{
			name:     "InvoiceSplit with Delivery Lead and amount suffix → ServiceFee",
			entry:    PayoutEntry{InvoiceSplitID: "split-003", Description: "[RENAISS :: INV-DO5S8] Delivery Lead - $43.64 USD"},
			expected: PayoutSourceTypeServiceFee,
		},
		{
			name:     "InvoiceSplit with Account Management and period → ServiceFee",
			entry:    PayoutEntry{InvoiceSplitID: "split-004", Description: "Account Management (Jan 2026)"},
			expected: PayoutSourceTypeServiceFee,
		},
		// InvoiceSplit: incentive/commission descriptions → Commission
		{
			name:     "InvoiceSplit with Account Management Incentive → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-005", Description: "Account Management Incentive for Invoice INV-DO5S8"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with Account Management Incentive and prefix → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-006", Description: "[RENAISS :: INV-DO5S8] Account Management Incentive for Invoice INV-DO5S8 - $43.64 USD"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with Sales Commission → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-007", Description: "[PLOT :: INV-OBI5D] Sales Commission for Invoice INV-OBI5D"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with Bonus keyword → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-008", Description: "Account Management Bonus"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with Delivery Lead supplemental fee → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-009", Description: "[NGHENHAN :: INV-J2JSB] Delivery Lead supplemental fee for Invoice INV-J2JSB (Dec 2025 Project Support) - $45 USD"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with no matching keywords → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-010", Description: "Some other description"},
			expected: PayoutSourceTypeCommission,
		},
		{
			name:     "InvoiceSplit with empty description → Commission",
			entry:    PayoutEntry{InvoiceSplitID: "split-011"},
			expected: PayoutSourceTypeCommission,
		},
		// RefundRequest → Refund
		{
			name:     "RefundRequest present → Refund",
			entry:    PayoutEntry{RefundRequestID: "refund-001"},
			expected: PayoutSourceTypeRefund,
		},
		// No relations → ExtraPayment
		{
			name:     "No relations → ExtraPayment",
			entry:    PayoutEntry{Description: "Standalone payment"},
			expected: PayoutSourceTypeExtraPayment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.determineSourceType(tt.entry)
			require.Equal(t, tt.expected, result)
		})
	}
}
