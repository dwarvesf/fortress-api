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
			subject: "Invoice CONTR-202501-A1B2 for January 2025",
			wantID:  "CONTR-202501-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID at beginning of subject",
			subject: "CONTR-202512-XYZ9 - Monthly Invoice",
			wantID:  "CONTR-202512-XYZ9",
			wantErr: false,
		},
		{
			name:    "Invoice ID at end of subject",
			subject: "Please process invoice CONTR-202506-1234",
			wantID:  "CONTR-202506-1234",
			wantErr: false,
		},
		{
			name:    "Invoice ID with long suffix",
			subject: "Invoice CONTR-202501-ABCD1234",
			wantID:  "CONTR-202501-ABCD1234",
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
			name:        "invalid format - missing CONTR prefix",
			subject:     "Invoice 202501-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - wrong date format",
			subject:     "Invoice CONTR-2025-A1B2",
			wantErr:     true,
			errContains: "invoice ID not found",
		},
		{
			name:        "invalid format - lowercase suffix",
			subject:     "Invoice CONTR-202501-a1b2",
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
			pdfText: "Invoice Number: CONTR-202501-A1B2\nAmount: $1000",
			wantID:  "CONTR-202501-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID embedded in text",
			pdfText: "Please pay invoice CONTR-202512-XYZ9 by end of month",
			wantID:  "CONTR-202512-XYZ9",
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
			subject: "Invoice CONTR-202501-A1B2",
			hasPDF:  false,
			wantID:  "CONTR-202501-A1B2",
			wantErr: false,
		},
		{
			name:    "Invoice ID in subject - PDF present but not used",
			subject: "Invoice CONTR-202501-A1B2",
			pdfText: "Different CONTR-202501-XYZ9",
			hasPDF:  true,
			wantID:  "CONTR-202501-A1B2", // Subject takes priority
			wantErr: false,
		},
		{
			name:    "No Invoice ID in subject - fall back to PDF",
			subject: "Invoice for January",
			pdfText: "Invoice: CONTR-202501-PDF1",
			hasPDF:  true,
			wantID:  "CONTR-202501-PDF1",
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
	tests := []struct {
		input   string
		matches bool
		match   string
	}{
		{"CONTR-202501-A1B2", true, "CONTR-202501-A1B2"},
		{"CONTR-202512-XYZ9", true, "CONTR-202512-XYZ9"},
		{"CONTR-202506-1234", true, "CONTR-202506-1234"},
		{"CONTR-202501-ABCD1234", true, "CONTR-202501-ABCD1234"},
		{"CONTR-202501-A", true, "CONTR-202501-A"},  // Single char suffix is valid
		{"prefix CONTR-202501-X1 suffix", true, "CONTR-202501-X1"},
		{"contr-202501-A1B2", false, ""},            // lowercase prefix not matched
		{"CONTR-2025-A1B2", false, ""},              // wrong date format
		{"CONTR-20251-A1B2", false, ""},             // 5-digit date
		{"CONTR-2025010-A1B2", false, ""},           // 7-digit date
		{"INV-202501-A1B2", false, ""},              // wrong prefix
		{"CONTR-202501-a1b2", false, ""},            // lowercase suffix
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
