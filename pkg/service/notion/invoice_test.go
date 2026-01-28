package notion

import (
	"testing"

	nt "github.com/dstotijn/go-notion"
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

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}
