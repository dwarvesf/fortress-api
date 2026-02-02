package invoiceemail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// mockPDFParser is a mock implementation of pdfparser.IService
type mockPDFParser struct {
	extractTextFn func(pdfBytes []byte) (string, error)
}

func (m *mockPDFParser) ExtractText(pdfBytes []byte) (string, error) {
	if m.extractTextFn != nil {
		return m.extractTextFn(pdfBytes)
	}
	return "", nil
}

func TestExtractInvoiceIDFromSubject(t *testing.T) {
	l := logger.NewLogrusLogger("debug")
	pdfParser := &mockPDFParser{}
	extractor := NewExtractor(pdfParser, l)

	tests := []struct {
		name        string
		subject     string
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid Invoice ID in subject",
			subject: "Invoice INVC-202501-JOHN-A1B2 for January 2025",
			wantID:  "INVC-202501-JOHN-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID at beginning of subject",
			subject: "INVC-202512-JANE-XYZ9 - Monthly Invoice",
			wantID:  "INVC-202512-JANE-XYZ9",
			wantErr: false,
		},
		{
			name:    "Invoice ID at end of subject",
			subject: "Please process invoice INVC-202506-BOB-1234",
			wantID:  "INVC-202506-BOB-1234",
			wantErr: false,
		},
		{
			name:    "Invoice ID with long name",
			subject: "Invoice INVC-202501-QUANG-ABCD",
			wantID:  "INVC-202501-QUANG-ABCD",
			wantErr: false,
		},
		{
			name:        "no Invoice ID in subject",
			subject:     "Invoice for January 2025",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "empty subject",
			subject:     "",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - missing INVC prefix",
			subject:     "Invoice 202501-JOHN-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - wrong date format",
			subject:     "Invoice INVC-2025-JOHN-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - lowercase name",
			subject:     "Invoice INVC-202501-john-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - old 3-part format",
			subject:     "Invoice INVC-202501-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractor.ExtractInvoiceIDFromSubject(tt.subject)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestExtractInvoiceIDFromPDF(t *testing.T) {
	l := logger.NewLogrusLogger("debug")

	tests := []struct {
		name        string
		pdfText     string
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Invoice ID in PDF text",
			pdfText: "Invoice Number: INVC-202501-JOHN-A1B2\nAmount: $1000",
			wantID:  "INVC-202501-JOHN-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID embedded in text",
			pdfText: "Please pay invoice INVC-202512-JANE-XYZ9 by end of month",
			wantID:  "INVC-202512-JANE-XYZ9",
			wantErr: false,
		},
		{
			name:        "no Invoice ID in PDF",
			pdfText:     "Invoice for services rendered\nTotal: $500",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "empty PDF text",
			pdfText:     "",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfParser := &mockPDFParser{
				extractTextFn: func(pdfBytes []byte) (string, error) {
					return tt.pdfText, nil
				},
			}
			extractor := NewExtractor(pdfParser, l)

			got, err := extractor.ExtractInvoiceIDFromPDF([]byte("fake pdf content"))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestExtractInvoiceID(t *testing.T) {
	l := logger.NewLogrusLogger("debug")

	tests := []struct {
		name        string
		subject     string
		pdfText     string
		hasPDF      bool
		wantID      string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Invoice ID in subject - no PDF needed",
			subject: "Invoice INVC-202501-JOHN-A1B2",
			hasPDF:  false,
			wantID:  "INVC-202501-JOHN-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID in subject - PDF present but not used",
			subject: "Invoice INVC-202501-JOHN-A1B2",
			pdfText: "Different INVC-202501-JANE-XYZ9",
			hasPDF:  true,
			wantID:  "INVC-202501-JOHN-A1B2", // Subject takes priority
			wantErr: false,
		},
		{
			name:    "No Invoice ID in subject - fall back to PDF",
			subject: "Invoice for January",
			pdfText: "Invoice: INVC-202501-BOB-PDF1",
			hasPDF:  true,
			wantID:  "INVC-202501-BOB-PDF1",
			wantErr: false,
		},
		{
			name:        "No Invoice ID in subject - no PDF",
			subject:     "Invoice for January",
			hasPDF:      false,
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "No Invoice ID in subject or PDF",
			subject:     "Invoice for January",
			pdfText:     "Payment due soon",
			hasPDF:      true,
			wantErr:     true,
			errContains: "invoice ID not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pdfParser := &mockPDFParser{
				extractTextFn: func(pdfBytes []byte) (string, error) {
					return tt.pdfText, nil
				},
			}
			extractor := NewExtractor(pdfParser, l)

			var pdfBytes []byte
			if tt.hasPDF {
				pdfBytes = []byte("fake pdf content")
			}

			got, err := extractor.ExtractInvoiceID(tt.subject, pdfBytes)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestInvoiceIDPattern(t *testing.T) {
	// Test the regex pattern directly
	// Format: INVC-YYYYMM-NAME-XXXX (e.g., INVC-202601-QUANG-4DRE)
	tests := []struct {
		input   string
		matches bool
		match   string
	}{
		{"INVC-202501-JOHN-A1B2", true, "INVC-202501-JOHN-A1B2"},
		{"INVC-202512-JANE-XYZ9", true, "INVC-202512-JANE-XYZ9"},
		{"INVC-202506-BOB-1234", true, "INVC-202506-BOB-1234"},
		{"INVC-202601-QUANG-4DRE", true, "INVC-202601-QUANG-4DRE"},
		{"INVC-202501-A-B", true, "INVC-202501-A-B"}, // Single char name and suffix is valid
		{"prefix INVC-202501-TEST-X1 suffix", true, "INVC-202501-TEST-X1"},
		{"invc-202501-JOHN-A1B2", false, ""},         // lowercase prefix not matched
		{"INVC-2025-JOHN-A1B2", false, ""},           // wrong date format (4 digits)
		{"INVC-20251-JOHN-A1B2", false, ""},          // 5-digit date
		{"INVC-2025010-JOHN-A1B2", false, ""},        // 7-digit date
		{"INV-202501-JOHN-A1B2", false, ""},          // wrong prefix
		{"INVC-202501-john-A1B2", false, ""},         // lowercase name
		{"INVC-202501-JOHN-a1b2", false, ""},         // lowercase suffix
		{"INVC-202501-A1B2", false, ""},              // old 3-part format (missing name)
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			match := invoiceIDPattern.FindString(tt.input)
			if tt.matches {
				assert.Equal(t, tt.match, match)
			} else {
				assert.Empty(t, match)
			}
		})
	}
}
