package webhook

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseInvoiceNumberFromTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		want    string
		wantErr bool
	}{
		{
			name:  "valid title",
			title: "Accounting | 11/2025 - #2025105-TRADI-002",
			want:  "2025105-TRADI-002",
		},
		{
			name:    "invalid format",
			title:   "Invoice",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseInvoiceNumberFromTitle(tt.title)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
