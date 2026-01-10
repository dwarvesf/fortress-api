# PDF Parser and Invoice ID Extractor Unit Test Specification

## Overview

This document specifies unit tests for Invoice ID extraction logic in `pkg/service/invoiceemail/extractor.go`. This includes subject line parsing, PDF text extraction, and regex pattern matching.

## Test File

**Location**: `pkg/service/invoiceemail/extractor_test.go`

## Test Setup

### Test Dependencies

```go
import (
    "context"
    "os"
    "strings"
    "testing"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "google.golang.org/api/gmail/v1"
)
```

### Test Data Directory

**Location**: `pkg/service/invoiceemail/testdata/`

**Files**:
- `invoice_with_id.pdf` - Valid PDF with Invoice ID "CONTR-202501-A1B2"
- `invoice_no_id.pdf` - Valid PDF without Invoice ID
- `malformed.pdf` - Corrupted/invalid PDF file
- `empty.pdf` - Empty PDF file
- `multipage.pdf` - Multi-page PDF with Invoice ID on first page

### Test Service Factory

```go
func createTestExtractor() *ProcessorService {
    cfg := &config.Config{
        InvoiceEmailListener: config.InvoiceEmailListener{
            Enabled:      true,
            EmailAddress: "test@example.com",
        },
    }

    return &ProcessorService{
        cfg:    cfg,
        logger: logger.NewLogrusLogger("error"),
        // gmailService and notionService not needed for extraction tests
    }
}
```

## Unit Test Cases - extractor_test.go

### 1. TestExtractInvoiceID - Subject Line Extraction

#### Test Case 1.1: Invoice ID in Subject (Primary Path)

**Test Name**: `TestExtractInvoiceID_FromSubject_Success`

**Description**: Verify extraction when Invoice ID is in email subject

**Input/Setup**:
```go
message := &gmail.Message{
    Id: "msg_123",
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice CONTR-202501-A1B2 for January"},
            {Name: "From", Value: "contractor@example.com"},
        },
        Parts: nil, // No attachments needed
    },
}
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`
- error: `nil`

**Implementation**:
```go
func TestExtractInvoiceID_FromSubject_Success(t *testing.T) {
    extractor := createTestExtractor()

    message := &gmail.Message{
        Id: "msg_123",
        Payload: &gmail.MessagePart{
            Headers: []*gmail.MessagePartHeader{
                {Name: "Subject", Value: "Invoice CONTR-202501-A1B2 for January"},
            },
        },
    }

    invoiceID, err := extractor.ExtractInvoiceID(context.Background(), message)

    assert.NoError(t, err)
    assert.Equal(t, "CONTR-202501-A1B2", invoiceID)
}
```

#### Test Case 1.2: Invoice ID at Start of Subject

**Test Name**: `TestExtractInvoiceID_FromSubject_StartOfLine`

**Description**: Verify extraction when Invoice ID is at beginning of subject

**Input/Setup**:
```go
subject := "CONTR-202501-A1B2 - Monthly Invoice"
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`

#### Test Case 1.3: Invoice ID at End of Subject

**Test Name**: `TestExtractInvoiceID_FromSubject_EndOfLine`

**Description**: Verify extraction when Invoice ID is at end of subject

**Input/Setup**:
```go
subject := "Monthly Invoice - CONTR-202501-A1B2"
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`

#### Test Case 1.4: Invoice ID Surrounded by Text

**Test Name**: `TestExtractInvoiceID_FromSubject_SurroundedByText`

**Description**: Verify extraction when Invoice ID is in middle of subject

**Input/Setup**:
```go
subject := "RE: Invoice CONTR-202501-A1B2 - Please review"
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`

#### Test Case 1.5: Multiple Invoice IDs in Subject (Edge Case)

**Test Name**: `TestExtractInvoiceID_FromSubject_MultipleIDs`

**Description**: Verify returns first match when multiple Invoice IDs present

**Input/Setup**:
```go
subject := "Invoices CONTR-202501-A1B2 and CONTR-202501-C3D4"
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"` (first match)

#### Test Case 1.6: No Invoice ID in Subject (Fallback to PDF)

**Test Name**: `TestExtractInvoiceID_FromSubject_NotFound`

**Description**: Verify fallback to PDF when subject doesn't contain Invoice ID

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Monthly Invoice for January"},
        },
        Parts: []*gmail.MessagePart{
            // PDF with Invoice ID
        },
    },
}
```

**Expected Behavior**:
- Attempts PDF extraction (tested separately)

### 2. TestExtractInvoiceID - PDF Extraction (Fallback Path)

#### Test Case 2.1: Invoice ID in PDF (Fallback Success)

**Test Name**: `TestExtractInvoiceID_FromPDF_Success`

**Description**: Verify extraction from PDF when not in subject

**Input/Setup**:
```go
// Requires mock Gmail service to return PDF bytes
message := &gmail.Message{
    Id: "msg_123",
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice for January"}, // No Invoice ID
        },
        Parts: []*gmail.MessagePart{
            {
                Filename: "invoice.pdf",
                MimeType: "application/pdf",
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_123",
                },
            },
        },
    },
}

// Mock GetAttachment to return PDF bytes from testdata
pdfBytes, _ := os.ReadFile("testdata/invoice_with_id.pdf")
mockGmail.On("GetAttachment", "msg_123", "attach_123").Return(pdfBytes, nil)
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`
- error: `nil`

**Note**: This test requires mock Gmail service or test PDF bytes.

#### Test Case 2.2: No PDF Attachment

**Test Name**: `TestExtractInvoiceID_FromPDF_NoPDFAttachment`

**Description**: Verify error when no PDF attachment found

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice for January"},
        },
        Parts: nil, // No attachments
    },
}
```

**Expected Output**:
- error: "no PDF attachment found"

#### Test Case 2.3: PDF Attachment but No Invoice ID

**Test Name**: `TestExtractInvoiceID_FromPDF_NoIDInPDF`

**Description**: Verify error when PDF doesn't contain Invoice ID

**Input/Setup**:
```go
// PDF without Invoice ID pattern
pdfBytes, _ := os.ReadFile("testdata/invoice_no_id.pdf")
mockGmail.On("GetAttachment", messageID, attachmentID).Return(pdfBytes, nil)
```

**Expected Output**:
- error: "Invoice ID not found in subject or PDF"

#### Test Case 2.4: Malformed PDF

**Test Name**: `TestExtractInvoiceID_FromPDF_MalformedPDF`

**Description**: Verify error handling when PDF is corrupted

**Input/Setup**:
```go
pdfBytes := []byte("not a valid pdf")
mockGmail.On("GetAttachment", messageID, attachmentID).Return(pdfBytes, nil)
```

**Expected Output**:
- error: "failed to extract text from PDF"

#### Test Case 2.5: Empty PDF

**Test Name**: `TestExtractInvoiceID_FromPDF_EmptyPDF`

**Description**: Verify handling of empty PDF file

**Input/Setup**:
```go
pdfBytes, _ := os.ReadFile("testdata/empty.pdf")
```

**Expected Output**:
- error: "Invoice ID not found in subject or PDF"

#### Test Case 2.6: PDF Download Fails

**Test Name**: `TestExtractInvoiceID_FromPDF_DownloadFails`

**Description**: Verify error handling when PDF download fails

**Input/Setup**:
```go
mockGmail.On("GetAttachment", messageID, attachmentID).Return(nil, errors.New("download failed"))
```

**Expected Output**:
- error: "failed to download PDF"

### 3. TestExtractFromText - Regex Pattern Matching

#### Test Case 3.1: Valid Invoice ID Pattern

**Test Name**: `TestExtractFromText_ValidPattern`

**Description**: Verify regex matches valid Invoice ID patterns

**Input/Setup**:
```go
tests := []struct {
    name     string
    text     string
    expected string
}{
    {
        name:     "standard format",
        text:     "Invoice ID: CONTR-202501-A1B2",
        expected: "CONTR-202501-A1B2",
    },
    {
        name:     "numeric suffix",
        text:     "Invoice CONTR-202501-1234",
        expected: "CONTR-202501-1234",
    },
    {
        name:     "alphanumeric suffix",
        text:     "CONTR-202501-X9Y8Z7",
        expected: "CONTR-202501-X9Y8Z7",
    },
    {
        name:     "single character suffix",
        text:     "CONTR-202501-A",
        expected: "CONTR-202501-A",
    },
}
```

**Implementation**:
```go
func TestExtractFromText_ValidPattern(t *testing.T) {
    tests := []struct {
        name     string
        text     string
        expected string
    }{
        {"standard format", "Invoice ID: CONTR-202501-A1B2", "CONTR-202501-A1B2"},
        {"numeric suffix", "Invoice CONTR-202501-1234", "CONTR-202501-1234"},
        {"alphanumeric suffix", "CONTR-202501-X9Y8Z7", "CONTR-202501-X9Y8Z7"},
        {"single character", "CONTR-202501-A", "CONTR-202501-A"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractFromText(tt.text)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

#### Test Case 3.2: Invalid Patterns (Should Not Match)

**Test Name**: `TestExtractFromText_InvalidPattern`

**Description**: Verify regex rejects invalid patterns

**Input/Setup**:
```go
tests := []struct {
    name string
    text string
}{
    {
        name: "wrong prefix",
        text: "INV-202501-A1B2",
    },
    {
        name: "missing prefix",
        text: "202501-A1B2",
    },
    {
        name: "wrong year format (4 digits)",
        text: "CONTR-2025-A1B2",
    },
    {
        name: "wrong month format (8 digits)",
        text: "CONTR-20250101-A1B2",
    },
    {
        name: "lowercase suffix",
        text: "CONTR-202501-a1b2",
    },
    {
        name: "special characters in suffix",
        text: "CONTR-202501-A@B#",
    },
    {
        name: "missing suffix",
        text: "CONTR-202501-",
    },
    {
        name: "extra dash",
        text: "CONTR-202501-A1-B2",
    },
}
```

**Expected Output**:
- All tests return empty string (no match)

**Implementation**:
```go
func TestExtractFromText_InvalidPattern(t *testing.T) {
    tests := []struct {
        name string
        text string
    }{
        {"wrong prefix", "INV-202501-A1B2"},
        {"missing prefix", "202501-A1B2"},
        {"wrong year format", "CONTR-2025-A1B2"},
        {"lowercase suffix", "CONTR-202501-a1b2"},
        {"special characters", "CONTR-202501-A@B#"},
        {"missing suffix", "CONTR-202501-"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractFromText(tt.text)
            assert.Empty(t, result, "Expected no match for: %s", tt.text)
        })
    }
}
```

#### Test Case 3.3: Case Sensitivity

**Test Name**: `TestExtractFromText_CaseSensitivity`

**Description**: Verify regex is case-sensitive for prefix and suffix

**Input/Setup**:
```go
tests := []struct {
    name     string
    text     string
    expected string
}{
    {
        name:     "uppercase (valid)",
        text:     "CONTR-202501-A1B2",
        expected: "CONTR-202501-A1B2",
    },
    {
        name:     "lowercase prefix (invalid)",
        text:     "contr-202501-A1B2",
        expected: "",
    },
    {
        name:     "mixed case prefix (invalid)",
        text:     "Contr-202501-A1B2",
        expected: "",
    },
}
```

#### Test Case 3.4: Edge Cases

**Test Name**: `TestExtractFromText_EdgeCases`

**Description**: Verify handling of edge cases

**Input/Setup**:
```go
tests := []struct {
    name     string
    text     string
    expected string
}{
    {
        name:     "empty string",
        text:     "",
        expected: "",
    },
    {
        name:     "whitespace only",
        text:     "   \t\n   ",
        expected: "",
    },
    {
        name:     "Invoice ID with newlines",
        text:     "Invoice ID:\nCONTR-202501-A1B2\nTotal: $5000",
        expected: "CONTR-202501-A1B2",
    },
    {
        name:     "Invoice ID with tabs",
        text:     "Invoice ID:\tCONTR-202501-A1B2",
        expected: "CONTR-202501-A1B2",
    },
    {
        name:     "very long text",
        text:     strings.Repeat("Lorem ipsum ", 1000) + "CONTR-202501-A1B2",
        expected: "CONTR-202501-A1B2",
    },
}
```

### 4. TestExtractTextFromPDF - PDF Text Extraction

#### Test Case 4.1: Valid PDF with Text

**Test Name**: `TestExtractTextFromPDF_ValidPDF`

**Description**: Verify successful text extraction from valid PDF

**Input/Setup**:
```go
pdfBytes, err := os.ReadFile("testdata/invoice_with_id.pdf")
require.NoError(t, err)
```

**Expected Output**:
- Extracted text contains Invoice ID
- No error

**Implementation**:
```go
func TestExtractTextFromPDF_ValidPDF(t *testing.T) {
    pdfBytes, err := os.ReadFile("testdata/invoice_with_id.pdf")
    require.NoError(t, err)

    text, err := extractTextFromPDF(pdfBytes)

    assert.NoError(t, err)
    assert.Contains(t, text, "CONTR-202501-A1B2")
}
```

#### Test Case 4.2: Invalid PDF Data

**Test Name**: `TestExtractTextFromPDF_InvalidPDF`

**Description**: Verify error handling for invalid PDF data

**Input/Setup**:
```go
pdfBytes := []byte("This is not a PDF file")
```

**Expected Output**:
- error: "pdfcpu extraction failed" or similar

**Implementation**:
```go
func TestExtractTextFromPDF_InvalidPDF(t *testing.T) {
    pdfBytes := []byte("This is not a PDF file")

    text, err := extractTextFromPDF(pdfBytes)

    assert.Error(t, err)
    assert.Empty(t, text)
}
```

#### Test Case 4.3: Empty PDF Bytes

**Test Name**: `TestExtractTextFromPDF_EmptyBytes`

**Description**: Verify handling of empty byte array

**Input/Setup**:
```go
pdfBytes := []byte{}
```

**Expected Output**:
- error returned

#### Test Case 4.4: Multipage PDF (First Page Only)

**Test Name**: `TestExtractTextFromPDF_MultipagePDF`

**Description**: Verify only first page is extracted (per spec)

**Input/Setup**:
```go
// PDF with Invoice ID on page 1, different text on page 2
pdfBytes, err := os.ReadFile("testdata/multipage.pdf")
```

**Expected Output**:
- Text from first page only
- Invoice ID found

#### Test Case 4.5: PDF with No Extractable Text

**Test Name**: `TestExtractTextFromPDF_ImageOnlyPDF`

**Description**: Verify handling of PDF with images but no text

**Input/Setup**:
```go
// PDF containing only images (scanned invoice)
pdfBytes, err := os.ReadFile("testdata/image_only.pdf")
```

**Expected Output**:
- Empty text or error (pdfcpu can't extract from images)

### 5. TestExtractInvoiceID - Integration Tests

#### Test Case 5.1: Subject Priority Over PDF

**Test Name**: `TestExtractInvoiceID_SubjectPriorityOverPDF`

**Description**: Verify subject line extraction takes precedence over PDF

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
        },
        Parts: []*gmail.MessagePart{
            {
                Filename: "invoice.pdf",
                MimeType: "application/pdf",
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_123",
                },
            },
        },
    },
}
```

**Expected Behavior**:
- Returns Invoice ID from subject
- PDF is NOT downloaded (optimization)

**Verification**:
- Mock GetAttachment should NOT be called

#### Test Case 5.2: Different Invoice IDs in Subject vs PDF (Edge Case)

**Test Name**: `TestExtractInvoiceID_DifferentIDsSubjectVsPDF`

**Description**: Verify subject takes precedence when IDs differ

**Input/Setup**:
```go
// Subject: CONTR-202501-A1B2
// PDF contains: CONTR-202501-C3D4
```

**Expected Output**:
- Returns Invoice ID from subject: `"CONTR-202501-A1B2"`

## Edge Cases and Error Scenarios

### Edge Case 1: Unicode Characters

**Test Name**: `TestExtractFromText_UnicodeCharacters`

**Description**: Verify handling of Unicode in text

**Input/Setup**:
```go
text := "发票 Invoice CONTR-202501-A1B2 中文"
```

**Expected Output**:
- invoiceID: `"CONTR-202501-A1B2"`

### Edge Case 2: Very Large PDF

**Test Name**: `TestExtractTextFromPDF_LargePDF`

**Description**: Verify handling of PDF larger than limit

**Input/Setup**:
```go
// PDF larger than PDFMaxSizeMB (5MB default)
```

**Expected Behavior**:
- Depends on implementation (may reject or process)

### Edge Case 3: PDF with Encrypted Content

**Test Name**: `TestExtractTextFromPDF_EncryptedPDF`

**Description**: Verify handling of password-protected PDF

**Input/Setup**:
```go
pdfBytes, _ := os.ReadFile("testdata/encrypted.pdf")
```

**Expected Output**:
- error: pdfcpu extraction failed

### Edge Case 4: Mixed Line Endings in Text

**Test Name**: `TestExtractFromText_MixedLineEndings`

**Description**: Verify handling of different line ending types

**Input/Setup**:
```go
tests := []struct {
    name string
    text string
}{
    {"unix", "Invoice\nCONTR-202501-A1B2"},
    {"windows", "Invoice\r\nCONTR-202501-A1B2"},
    {"mac", "Invoice\rCONTR-202501-A1B2"},
}
```

**Expected Output**:
- All extract Invoice ID correctly

## Test Data Files Required

### testdata/invoice_with_id.pdf

**Content**:
```
CONTRACTOR INVOICE

Invoice ID: CONTR-202501-A1B2
Date: January 15, 2025
Contractor: John Doe

Services rendered for January 2025

Total: $5,000.00
```

### testdata/invoice_no_id.pdf

**Content**:
```
CONTRACTOR INVOICE

Date: January 15, 2025
Contractor: John Doe

Services rendered for January 2025

Total: $5,000.00
```

### testdata/malformed.pdf

**Content**: Invalid binary data (not a real PDF)

### testdata/empty.pdf

**Content**: Valid but empty PDF file

### testdata/multipage.pdf

**Content**:
- Page 1: Contains "CONTR-202501-A1B2"
- Page 2: Different content without Invoice ID

## Test Coverage Goals

### Function Coverage
- ExtractInvoiceID: 100% (subject path, PDF path, error paths)
- extractFromText: 100% (valid patterns, invalid patterns)
- extractTextFromPDF: 100% (valid PDF, invalid PDF, edge cases)

### Regex Pattern Coverage
- All valid formats tested
- Common invalid formats tested
- Edge cases (empty, special chars) tested

### PDF Processing Coverage
- Valid PDF text extraction
- Invalid PDF error handling
- Edge cases (empty, multipage, encrypted)

## Testing Patterns Used

### Table-Driven Tests for Regex

```go
func TestExtractFromText_Patterns(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractFromText(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### File-Based Tests for PDF

```go
func TestExtractTextFromPDF_Files(t *testing.T) {
    tests := []struct {
        name        string
        filename    string
        shouldError bool
        contains    string
    }{
        {"valid PDF", "invoice_with_id.pdf", false, "CONTR-202501-A1B2"},
        {"invalid PDF", "malformed.pdf", true, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pdfBytes, err := os.ReadFile("testdata/" + tt.filename)
            require.NoError(t, err)

            text, err := extractTextFromPDF(pdfBytes)

            if tt.shouldError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                if tt.contains != "" {
                    assert.Contains(t, text, tt.contains)
                }
            }
        })
    }
}
```

## Notes for Implementer

1. **Test Data**: Create realistic PDF test files using a PDF generation tool or sample invoices.

2. **Regex Testing**: Thoroughly test regex pattern edge cases - this is the critical validation layer.

3. **PDF Library**: pdfcpu may have specific error types - verify error handling matches library behavior.

4. **Temp Files**: Ensure temp files are cleaned up after PDF processing tests.

5. **Performance**: PDF extraction tests may be slower - consider marking as integration tests if needed.

6. **Mock Gmail Service**: Some tests require mocking GetAttachment - use same mock pattern as processor tests.

7. **Case Sensitivity**: Verify regex requirements match business rules (uppercase only for suffix?).

8. **First Match**: When multiple Invoice IDs in text, verify only first is returned.

9. **Whitespace**: Test that regex handles various whitespace characters around Invoice ID.

10. **Error Messages**: Verify error messages are descriptive for debugging (especially PDF extraction failures).
