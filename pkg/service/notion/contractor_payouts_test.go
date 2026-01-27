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
		require.Equal(t, 1, query.PageSize)

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
							Date: &nt.Date{Start: latestDt},
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
							Date: &nt.Date{Start: olderDt},
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
