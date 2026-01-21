package notion

import (
	"testing"
)

func TestHasOverlap(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "both slices have overlapping elements",
			a:        []string{"a", "b", "c"},
			b:        []string{"c", "d", "e"},
			expected: true,
		},
		{
			name:     "no overlapping elements",
			a:        []string{"a", "b", "c"},
			b:        []string{"d", "e", "f"},
			expected: false,
		},
		{
			name:     "first slice empty",
			a:        []string{},
			b:        []string{"a", "b", "c"},
			expected: false,
		},
		{
			name:     "second slice empty",
			a:        []string{"a", "b", "c"},
			b:        []string{},
			expected: false,
		},
		{
			name:     "both slices empty",
			a:        []string{},
			b:        []string{},
			expected: false,
		},
		{
			name:     "single element overlap",
			a:        []string{"x"},
			b:        []string{"x"},
			expected: true,
		},
		{
			name:     "multiple overlaps",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "d"},
			expected: true,
		},
		{
			name:     "first slice nil",
			a:        nil,
			b:        []string{"a", "b"},
			expected: false,
		},
		{
			name:     "second slice nil",
			a:        []string{"a", "b"},
			b:        nil,
			expected: false,
		},
		{
			name:     "both slices nil",
			a:        nil,
			b:        nil,
			expected: false,
		},
		{
			name:     "exact same slices",
			a:        []string{"page-1", "page-2", "page-3"},
			b:        []string{"page-1", "page-2", "page-3"},
			expected: true,
		},
		{
			name:     "partial overlap at end",
			a:        []string{"page-1", "page-2", "page-3"},
			b:        []string{"page-3", "page-4", "page-5"},
			expected: true,
		},
		{
			name:     "realistic payout IDs - no overlap (different invoice types)",
			a:        []string{"uuid-service-1", "uuid-service-2", "uuid-refund-1"},
			b:        []string{"uuid-extra-1", "uuid-extra-2"},
			expected: false,
		},
		{
			name:     "realistic payout IDs - overlap (same invoice type regenerated)",
			a:        []string{"uuid-service-1", "uuid-service-2", "uuid-service-3"},
			b:        []string{"uuid-service-1", "uuid-service-2"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasOverlap(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("hasOverlap(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
