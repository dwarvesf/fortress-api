package notion

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	nt "github.com/dstotijn/go-notion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func TestExtractFormulaString(t *testing.T) {
	// Create a minimal notionService instance for testing
	ns := &notionService{}

	tests := []struct {
		name     string
		props    nt.DatabasePageProperties
		propName string
		expected string
	}{
		{
			name: "valid formula string property",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						String: stringPtr("John Doe"),
					},
				},
			},
			propName: "TestProp",
			expected: "John Doe",
		},
		{
			name:     "property does not exist",
			props:    nt.DatabasePageProperties{},
			propName: "NonExistent",
			expected: "",
		},
		{
			name: "property exists but formula is nil",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: nil,
				},
			},
			propName: "TestProp",
			expected: "",
		},
		{
			name: "formula exists but string is nil",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						String: nil,
					},
				},
			},
			propName: "TestProp",
			expected: "",
		},
		{
			name: "empty string value",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						String: stringPtr(""),
					},
				},
			},
			propName: "TestProp",
			expected: "",
		},
		{
			name: "formula with number instead of string",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						Number: floatPtr(123.45),
						String: nil,
					},
				},
			},
			propName: "TestProp",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ns.extractFormulaString(tt.props, tt.propName)
			if result != tt.expected {
				t.Errorf("extractFormulaString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractFormulaProp(t *testing.T) {
	// Create a minimal notionService instance for testing
	ns := &notionService{}

	tests := []struct {
		name     string
		props    nt.DatabasePageProperties
		propName string
		expected float64
	}{
		{
			name: "valid formula number property",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						Number: floatPtr(123.45),
					},
				},
			},
			propName: "TestProp",
			expected: 123.45,
		},
		{
			name:     "property does not exist",
			props:    nt.DatabasePageProperties{},
			propName: "NonExistent",
			expected: 0,
		},
		{
			name: "property exists but formula is nil",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: nil,
				},
			},
			propName: "TestProp",
			expected: 0,
		},
		{
			name: "formula exists but number is nil",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						Number: nil,
					},
				},
			},
			propName: "TestProp",
			expected: 0,
		},
		{
			name: "zero value",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						Number: floatPtr(0),
					},
				},
			},
			propName: "TestProp",
			expected: 0,
		},
		{
			name: "negative value",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Formula: &nt.FormulaResult{
						Number: floatPtr(-50.25),
					},
				},
			},
			propName: "TestProp",
			expected: -50.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ns.extractFormulaProp(tt.props, tt.propName)
			if result != tt.expected {
				t.Errorf("extractFormulaProp() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestExtractNumberProp(t *testing.T) {
	// Create a minimal notionService instance for testing
	ns := &notionService{}

	tests := []struct {
		name     string
		props    nt.DatabasePageProperties
		propName string
		expected float64
	}{
		{
			name: "valid number property",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Number: floatPtr(42.5),
				},
			},
			propName: "TestProp",
			expected: 42.5,
		},
		{
			name:     "property does not exist",
			props:    nt.DatabasePageProperties{},
			propName: "NonExistent",
			expected: 0,
		},
		{
			name: "property exists but number is nil",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Number: nil,
				},
			},
			propName: "TestProp",
			expected: 0,
		},
		{
			name: "zero value",
			props: nt.DatabasePageProperties{
				"TestProp": nt.DatabasePageProperty{
					Number: floatPtr(0),
				},
			},
			propName: "TestProp",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ns.extractNumberProp(tt.props, tt.propName)
			if result != tt.expected {
				t.Errorf("extractNumberProp() = %f, want %f", result, tt.expected)
			}
		})
	}
}

func TestNormalizeInvoiceStatusForNotion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "lowercase sent", input: "sent", expected: "Sent"},
		{name: "lowercase overdue", input: "overdue", expected: "Overdue"},
		{name: "lowercase credited", input: "credited", expected: "Credited"},
		{name: "already normalized", input: "Paid", expected: "Paid"},
		{name: "trim whitespace", input: "  uncollectible  ", expected: "Uncollectible"},
		{name: "unknown passthrough", input: "Custom Status", expected: "Custom Status"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, normalizeInvoiceStatusForNotion(tt.input))
		})
	}
}

func TestQueryInvoicesByMonth_SortsByIssueDateDescending(t *testing.T) {
	client := newNotionTestClient(t, func(r *http.Request) (*http.Response, error) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/v1/databases/"+ClientInvoicesDBID+"/query", r.URL.Path)

		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var query nt.DatabaseQuery
		require.NoError(t, json.Unmarshal(bodyBytes, &query))
		require.NotNil(t, query.Filter)
		require.Len(t, query.Sorts, 1)
		require.Equal(t, "Issue Date", query.Sorts[0].Property)
		require.Equal(t, nt.SortDirDesc, query.Sorts[0].Direction)

		resp := nt.DatabaseQueryResponse{Results: []nt.Page{}}
		payload, err := json.Marshal(resp)
		require.NoError(t, err)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	ns := &notionService{notionClient: client, l: logger.NewLogrusLogger("debug")}

	_, err := ns.QueryInvoicesByMonth(2026, 3, []string{"Paid", "Sent"}, "")
	require.NoError(t, err)
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
