# Invoice Email Processor Unit Test Specification

## Overview

This document specifies unit tests for the Invoice Email Processor service in `pkg/service/invoiceemail/`. The processor orchestrates the end-to-end flow of extracting Invoice IDs and updating Notion payables.

## Test Files

**Location**:
- `pkg/service/invoiceemail/processor_test.go` - Main processor tests
- `pkg/service/invoiceemail/extractor_test.go` - Invoice ID extraction tests (separate file)

## Test Setup

### Test Dependencies

```go
import (
    "context"
    "errors"
    "testing"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "google.golang.org/api/gmail/v1"
)
```

### Mock Interfaces

```go
// Mock Gmail Service
type MockGmailService struct {
    mock.Mock
}

func (m *MockGmailService) ListInboxMessages(query string, maxResults int64) ([]*gmail.Message, error) {
    args := m.Called(query, maxResults)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).([]*gmail.Message), args.Error(1)
}

func (m *MockGmailService) GetMessage(messageID string) (*gmail.Message, error) {
    args := m.Called(messageID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*gmail.Message), args.Error(1)
}

func (m *MockGmailService) GetAttachment(messageID, attachmentID string) ([]byte, error) {
    args := m.Called(messageID, attachmentID)
    return args.Get(0).([]byte), args.Error(1)
}

func (m *MockGmailService) AddLabel(messageID, labelName string) error {
    args := m.Called(messageID, labelName)
    return args.Error(0)
}

// Mock Notion Service
type MockNotionService struct {
    mock.Mock
}

func (m *MockNotionService) FindPayableByInvoiceID(ctx context.Context, invoiceID string) (*PayableRecord, error) {
    args := m.Called(ctx, invoiceID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*PayableRecord), args.Error(1)
}

func (m *MockNotionService) UpdatePayableStatus(ctx context.Context, pageID, status, note string) error {
    args := m.Called(ctx, pageID, status, note)
    return args.Error(0)
}
```

### Test Service Factory

```go
func createTestProcessor(gmailService *MockGmailService, notionService *MockNotionService) *ProcessorService {
    cfg := &config.Config{
        InvoiceEmailListener: config.InvoiceEmailListener{
            Enabled:        true,
            EmailAddress:   "test@example.com",
            ProcessedLabel: "fortress-api-test/processed",
            MaxMessages:    50,
        },
    }

    return &ProcessorService{
        cfg:           cfg,
        logger:        logger.NewLogrusLogger("error"),
        gmailService:  gmailService,
        notionService: notionService,
    }
}
```

## Unit Test Cases - processor_test.go

### 1. TestProcessSingleEmail - Success Path

#### Test Case 1.1: Complete Success Flow

**Test Name**: `TestProcessSingleEmail_Success`

**Description**: Verify full flow when Invoice ID found, payable exists with status "New", and update succeeds

**Input/Setup**:
```go
messageID := "msg_123"
invoiceID := "CONTR-202501-A1B2"

mockGmail := new(MockGmailService)
mockNotion := new(MockNotionService)

// Mock GetMessage - returns message with Invoice ID in subject
mockGmail.On("GetMessage", messageID).Return(&gmail.Message{
    Id: messageID,
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
            {Name: "From", Value: "contractor@example.com"},
        },
    },
}, nil)

// Mock FindPayableByInvoiceID - returns payable with status "New"
mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(&PayableRecord{
    PageID:    "page_abc",
    InvoiceID: invoiceID,
    Status:    "New",
}, nil)

// Mock UpdatePayableStatus - succeeds
mockNotion.On("UpdatePayableStatus", mock.Anything, "page_abc", "Pending", "").Return(nil)

// Mock AddLabel - succeeds
mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(nil)
```

**Expected Output**:
- No error returned
- UpdatePayableStatus called with status "Pending"
- AddLabel called with correct label

**Implementation**:
```go
func TestProcessSingleEmail_Success(t *testing.T) {
    messageID := "msg_123"
    invoiceID := "CONTR-202501-A1B2"

    mockGmail := new(MockGmailService)
    mockNotion := new(MockNotionService)

    mockGmail.On("GetMessage", messageID).Return(&gmail.Message{
        Id: messageID,
        Payload: &gmail.MessagePart{
            Headers: []*gmail.MessagePartHeader{
                {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
                {Name: "From", Value: "contractor@example.com"},
            },
        },
    }, nil)

    mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(&PayableRecord{
        PageID:    "page_abc",
        InvoiceID: invoiceID,
        Status:    "New",
    }, nil)

    mockNotion.On("UpdatePayableStatus", mock.Anything, "page_abc", "Pending", "").Return(nil)
    mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(nil)

    processor := createTestProcessor(mockGmail, mockNotion)
    err := processor.ProcessSingleEmail(context.Background(), messageID)

    assert.NoError(t, err)
    mockGmail.AssertExpectations(t)
    mockNotion.AssertExpectations(t)
}
```

### 2. TestProcessSingleEmail - Error Paths

#### Test Case 2.1: GetMessage Fails

**Test Name**: `TestProcessSingleEmail_GetMessageFails`

**Description**: Verify error handling when Gmail API fails to fetch message

**Input/Setup**:
```go
mockGmail := new(MockGmailService)
mockGmail.On("GetMessage", "msg_123").Return(nil, errors.New("gmail api error"))
```

**Expected Output**:
- Error returned
- No Notion API calls made
- No label added

#### Test Case 2.2: No Invoice ID Found

**Test Name**: `TestProcessSingleEmail_NoInvoiceID`

**Description**: Verify handling when Invoice ID cannot be extracted

**Input/Setup**:
```go
// Message with no Invoice ID in subject, no PDF attachment
mockGmail.On("GetMessage", messageID).Return(&gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice for January"},
        },
        Parts: nil, // No attachments
    },
}, nil)

// Label should still be added to mark as processed
mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(nil)
```

**Expected Output**:
- Error returned
- No Notion API calls made
- Email labeled as processed (to prevent reprocessing)

#### Test Case 2.3: Payable Not Found in Notion

**Test Name**: `TestProcessSingleEmail_PayableNotFound`

**Description**: Verify handling when Invoice ID doesn't match any Notion record

**Input/Setup**:
```go
mockGmail.On("GetMessage", messageID).Return(&gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
        },
    },
}, nil)

// Return nil (not found)
mockNotion.On("FindPayableByInvoiceID", mock.Anything, "CONTR-202501-A1B2").Return(nil, nil)

mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(nil)
```

**Expected Output**:
- No error returned
- No UpdatePayableStatus call
- Email labeled as processed

#### Test Case 2.4: Payable Status Not "New"

**Test Name**: `TestProcessSingleEmail_PayableAlreadyProcessed`

**Description**: Verify handling when payable status is not "New" (already processed)

**Input/Setup**:
```go
mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(&PayableRecord{
    PageID:    "page_abc",
    InvoiceID: invoiceID,
    Status:    "Pending", // Already processed
}, nil)

mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(nil)
```

**Expected Output**:
- No error returned
- No UpdatePayableStatus call
- Email labeled as processed

#### Test Case 2.5: Notion Update Fails

**Test Name**: `TestProcessSingleEmail_NotionUpdateFails`

**Description**: Verify error handling when Notion API fails to update status

**Input/Setup**:
```go
mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(&PayableRecord{
    PageID:    "page_abc",
    InvoiceID: invoiceID,
    Status:    "New",
}, nil)

mockNotion.On("UpdatePayableStatus", mock.Anything, "page_abc", "Pending", "").Return(errors.New("notion api error"))
```

**Expected Output**:
- Error returned
- Email NOT labeled (should be retried next run)

#### Test Case 2.6: Label Addition Fails After Successful Update

**Test Name**: `TestProcessSingleEmail_LabelFailsAfterUpdate`

**Description**: Verify handling when label addition fails after successful Notion update (partial failure)

**Input/Setup**:
```go
mockNotion.On("UpdatePayableStatus", mock.Anything, "page_abc", "Pending", "").Return(nil)
mockGmail.On("AddLabel", messageID, "fortress-api-test/processed").Return(errors.New("label error"))
```

**Expected Output**:
- No error returned (Notion update succeeded - acceptable)
- Error logged but not propagated

### 3. TestProcessInvoiceEmails - Batch Processing

#### Test Case 3.1: No Unread Messages

**Test Name**: `TestProcessInvoiceEmails_NoMessages`

**Description**: Verify handling when no unread messages found

**Input/Setup**:
```go
mockGmail := new(MockGmailService)
mockGmail.On("ListInboxMessages", mock.Anything, int64(50)).Return([]*gmail.Message{}, nil)
```

**Expected Output**:
```go
result := &ProcessingResult{
    TotalEmails:       0,
    SuccessfulUpdates: 0,
    SkippedEmails:     0,
    FailedUpdates:     0,
    Errors:            []EmailProcessingError{},
}
```

#### Test Case 3.2: Single Message Success

**Test Name**: `TestProcessInvoiceEmails_SingleMessageSuccess`

**Description**: Verify batch processing with one successful message

**Input/Setup**:
```go
messages := []*gmail.Message{
    {Id: "msg_123"},
}
mockGmail.On("ListInboxMessages", mock.Anything, int64(50)).Return(messages, nil)
// Mock successful ProcessSingleEmail flow
```

**Expected Output**:
```go
result := &ProcessingResult{
    TotalEmails:       1,
    SuccessfulUpdates: 1,
    SkippedEmails:     0,
    FailedUpdates:     0,
}
```

#### Test Case 3.3: Multiple Messages Mixed Results

**Test Name**: `TestProcessInvoiceEmails_MixedResults`

**Description**: Verify batch processing with mix of success, skip, and failure

**Input/Setup**:
```go
messages := []*gmail.Message{
    {Id: "msg_success"},  // Will succeed
    {Id: "msg_skip"},     // Will skip (no Invoice ID)
    {Id: "msg_fail"},     // Will fail (Notion error)
}
```

**Expected Output**:
```go
result := &ProcessingResult{
    TotalEmails:       3,
    SuccessfulUpdates: 1,
    SkippedEmails:     1,
    FailedUpdates:     1,
}
```

**Verification**:
- Processing continues after individual failures
- All messages attempted

#### Test Case 3.4: ListInboxMessages Fails

**Test Name**: `TestProcessInvoiceEmails_ListFails`

**Description**: Verify handling when Gmail API fails to list messages

**Input/Setup**:
```go
mockGmail.On("ListInboxMessages", mock.Anything, int64(50)).Return(nil, errors.New("gmail api error"))
```

**Expected Output**:
- Error returned
- Result is nil
- No messages processed

### 4. TestFindPayableByInvoiceID

#### Test Case 4.1: Payable Found

**Test Name**: `TestFindPayableByInvoiceID_Found`

**Description**: Verify successful lookup of payable

**Input/Setup**:
```go
invoiceID := "CONTR-202501-A1B2"
mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(&PayableRecord{
    PageID:    "page_abc",
    InvoiceID: invoiceID,
    Status:    "New",
}, nil)
```

**Expected Output**:
- PayableRecord returned
- No error

#### Test Case 4.2: Payable Not Found

**Test Name**: `TestFindPayableByInvoiceID_NotFound`

**Description**: Verify handling when no matching record exists

**Input/Setup**:
```go
mockNotion.On("FindPayableByInvoiceID", mock.Anything, "CONTR-999999-XXXX").Return(nil, nil)
```

**Expected Output**:
- nil PayableRecord
- No error

#### Test Case 4.3: Notion Query Error

**Test Name**: `TestFindPayableByInvoiceID_QueryError`

**Description**: Verify error handling when Notion API fails

**Input/Setup**:
```go
mockNotion.On("FindPayableByInvoiceID", mock.Anything, invoiceID).Return(nil, errors.New("notion api error"))
```

**Expected Output**:
- nil PayableRecord
- Error returned

## Edge Cases and Error Scenarios

### Edge Case 1: Empty Message ID

**Test Name**: `TestProcessSingleEmail_EmptyMessageID`

**Description**: Verify validation of empty message ID

**Input/Setup**:
```go
processor := createTestProcessor(mockGmail, mockNotion)
err := processor.ProcessSingleEmail(context.Background(), "")
```

**Expected Output**:
- Error returned
- Error message indicates invalid message ID

### Edge Case 2: Context Cancellation

**Test Name**: `TestProcessSingleEmail_ContextCancelled`

**Description**: Verify handling of cancelled context

**Input/Setup**:
```go
ctx, cancel := context.WithCancel(context.Background())
cancel() // Cancel immediately

processor := createTestProcessor(mockGmail, mockNotion)
err := processor.ProcessSingleEmail(ctx, "msg_123")
```

**Expected Output**:
- Error returned (context.Canceled)

### Edge Case 3: Very Long Invoice ID

**Test Name**: `TestProcessSingleEmail_LongInvoiceID`

**Description**: Verify handling of unusually long Invoice ID

**Input/Setup**:
```go
// Invoice ID with 100 characters
longInvoiceID := "CONTR-202501-" + strings.Repeat("X", 87)
```

**Expected Behavior**:
- System attempts lookup with full Invoice ID
- Notion API handles validation

### Edge Case 4: Special Characters in Invoice ID

**Test Name**: `TestExtractInvoiceID_SpecialCharacters`

**Description**: Verify regex doesn't match invalid patterns

**Input/Setup**:
```go
subject := "Invoice CONTR-202501-A@B#"
```

**Expected Output**:
- Invoice ID not extracted (invalid pattern)

## Test Coverage Goals

### Method Coverage
- ProcessInvoiceEmails: 100% (all branches)
- ProcessSingleEmail: 100% (all branches)
- findPayableByInvoiceID: 100% (all branches)

### Error Path Coverage
- Gmail API failures
- Notion API failures
- Validation errors
- Partial failures (update succeeds, label fails)

### Business Logic Coverage
- Status check ("New" vs other statuses)
- Payable not found handling
- Invoice ID extraction fallback
- Label application idempotency

## Testing Patterns Used

### Mock-Based Testing

Use testify/mock for dependency injection:

```go
func TestExample(t *testing.T) {
    mockGmail := new(MockGmailService)
    mockNotion := new(MockNotionService)

    // Set expectations
    mockGmail.On("Method", args).Return(result, nil)

    // Execute test
    processor := createTestProcessor(mockGmail, mockNotion)
    err := processor.Method()

    // Verify
    assert.NoError(t, err)
    mockGmail.AssertExpectations(t)
}
```

### Table-Driven Tests for Multiple Scenarios

```go
func TestProcessSingleEmail_Scenarios(t *testing.T) {
    tests := []struct {
        name          string
        messageID     string
        setupMocks    func(*MockGmailService, *MockNotionService)
        expectedError bool
    }{
        {
            name:      "success",
            messageID: "msg_123",
            setupMocks: func(mg *MockGmailService, mn *MockNotionService) {
                // Setup success scenario
            },
            expectedError: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockGmail := new(MockGmailService)
            mockNotion := new(MockNotionService)
            tt.setupMocks(mockGmail, mockNotion)

            processor := createTestProcessor(mockGmail, mockNotion)
            err := processor.ProcessSingleEmail(context.Background(), tt.messageID)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Assertion Helpers

```go
// Verify result structure
func assertProcessingResult(t *testing.T, result *ProcessingResult, expected *ProcessingResult) {
    assert.Equal(t, expected.TotalEmails, result.TotalEmails)
    assert.Equal(t, expected.SuccessfulUpdates, result.SuccessfulUpdates)
    assert.Equal(t, expected.SkippedEmails, result.SkippedEmails)
    assert.Equal(t, expected.FailedUpdates, result.FailedUpdates)
}
```

## Notes for Implementer

1. **Mock Setup**: Create comprehensive mocks for Gmail and Notion services. Consider using gomock or testify/mock.

2. **Context Handling**: All tests should pass context.Background() unless testing cancellation.

3. **Error Scenarios**: Focus on error handling - this is critical for robustness.

4. **Partial Failures**: Test case where Notion update succeeds but label fails - this is non-fatal.

5. **Batch Processing**: Verify processing continues after individual email failures.

6. **Label Idempotency**: Label addition should be safe to retry (handled by Gmail API).

7. **Status Validation**: Only "New" status should trigger update - test all other statuses.

8. **Logging**: While tests don't verify logs, ensure logger is properly initialized to avoid nil pointer errors.

9. **Configuration**: Use test config with known values (e.g., test label name).

10. **Clean Mocks**: Reset mock expectations between tests using fresh mock instances.
