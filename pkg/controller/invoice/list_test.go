package invoice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dwarvesf/fortress-api/pkg/model"
)

func TestSanitizeListInputForLog(t *testing.T) {
	in := GetListInvoiceInput{
		Pagination: model.Pagination{Page: 2, Size: 25, Sort: "-created_at"},
		ProjectIDs: []string{"project-1"},
		Statuses:   []string{"sent", "credited"},
		OnProgress: func(completed, total int) {},
	}

	fields := sanitizeListInputForLog(in)

	assert.Equal(t, int64(2), fields["page"])
	assert.Equal(t, int64(25), fields["size"])
	assert.Equal(t, "-created_at", fields["sort"])
	assert.Equal(t, []string{"project-1"}, fields["projectIDs"])
	assert.Equal(t, []string{"sent", "credited"}, fields["statuses"])
	assert.Equal(t, true, fields["hasProgressCallback"])
	_, exists := fields["OnProgress"]
	assert.False(t, exists)
	_, exists = fields["onProgress"]
	assert.False(t, exists)
}

func TestMapNotionStatusToAPI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.InvoiceStatus
	}{
		{name: "credited", input: "Credited", expected: model.InvoiceStatusCredited},
		{name: "cancelled", input: "Cancelled", expected: model.InvoiceStatusUncollectible},
		{name: "sent", input: "Sent", expected: model.InvoiceStatusSent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, mapNotionStatusToAPI(tt.input))
		})
	}
}
