package webhook

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNocoInvoiceRowPayloadRecordID(t *testing.T) {
	tests := []struct {
		name string
		id   interface{}
		want string
	}{
		{"string", "abc", "abc"},
		{"float", float64(12), "12"},
		{"int", 5, "5"},
		{"other", struct{}{}, "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := nocoInvoiceRowPayload{ID: tt.id}
			require.Equal(t, tt.want, payload.recordID())
		})
	}
}

func TestExtractRowPayload(t *testing.T) {
	row := map[string]any{
		"Id":                  float64(7),
		"invoice_number":      "INV-1",
		"fortress_invoice_id": "uuid-1",
		"status":              "paid",
	}

	payload := extractRowPayload(row)

	require.Equal(t, "INV-1", payload.InvoiceNumber)
	require.Equal(t, "uuid-1", payload.FortressInvoiceID)
	require.Equal(t, "paid", payload.Status)
	require.Equal(t, "7", payload.recordID())

	payload = extractRowPayload(map[string]any{"id": 3})
	require.Equal(t, "3", payload.recordID())
}
