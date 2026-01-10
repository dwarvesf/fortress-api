# Gmail Inbox Service Unit Test Specification

## Overview

This document specifies unit tests for Gmail Inbox Service operations in `pkg/service/googlemail/`. The tests follow the existing table-driven test pattern used in the codebase.

## Test File

**Location**: `pkg/service/googlemail/google_mail_inbox_test.go`

## Test Setup

### Test Dependencies

```go
import (
    "testing"

    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/stretchr/testify/assert"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
)
```

### Mock Service Structure

```go
// Use the existing googleService struct with nil/mock service
func createTestService(service *gmail.Service) *googleService {
    return &googleService{
        config: &oauth2.Config{
            ClientID:     "test-client-id",
            ClientSecret: "test-client-secret",
            Endpoint:     google.Endpoint,
            Scopes:       []string{gmail.MailGoogleComScope},
        },
        appConfig: &config.Config{},
        service:   service,
        token:     &oauth2.Token{AccessToken: "test-token"},
    }
}
```

## Unit Test Cases

### 1. TestListInboxMessages

#### Test Case 1.1: Service Not Initialized

**Test Name**: `TestListInboxMessages_ServiceNotInitialized`

**Description**: Verify error handling when Gmail service is not initialized

**Input/Setup**:
- googleService with nil service
- query: `"is:unread"`
- maxResults: `50`

**Expected Output**:
- Error returned
- Error message contains "service not initialized" or similar

**Implementation**:
```go
func TestListInboxMessages_ServiceNotInitialized(t *testing.T) {
    g := createTestService(nil)

    messages, err := g.ListInboxMessages("is:unread", 50)

    assert.Error(t, err)
    assert.Nil(t, messages)
}
```

#### Test Case 1.2: Empty Query String

**Test Name**: `TestListInboxMessages_EmptyQuery`

**Description**: Verify handling of empty query (should list all messages)

**Input/Setup**:
- googleService with nil service (will fail at service check)
- query: `""`
- maxResults: `50`

**Expected Output**:
- Error returned (service not initialized)

**Note**: With real service, empty query is valid and lists all messages

#### Test Case 1.3: Method Signature Verification

**Test Name**: `TestListInboxMessagesMethodExists`

**Description**: Verify method exists with correct signature

**Implementation**:
```go
func TestListInboxMessagesMethodExists(t *testing.T) {
    var _ interface {
        ListInboxMessages(query string, maxResults int64) ([]*gmail.Message, error)
    } = (*googleService)(nil)
}
```

### 2. TestGetMessage

#### Test Case 2.1: Service Not Initialized

**Test Name**: `TestGetMessage_ServiceNotInitialized`

**Description**: Verify error handling when Gmail service is not initialized

**Input/Setup**:
- googleService with nil service
- messageID: `"msg_abc123"`

**Expected Output**:
- Error returned
- Error message indicates service not initialized

**Implementation**:
```go
func TestGetMessage_ServiceNotInitialized(t *testing.T) {
    g := createTestService(nil)

    message, err := g.GetMessage("msg_abc123")

    assert.Error(t, err)
    assert.Nil(t, message)
}
```

#### Test Case 2.2: Empty Message ID

**Test Name**: `TestGetMessage_EmptyMessageID`

**Description**: Verify validation of empty message ID

**Input/Setup**:
- googleService with nil service
- messageID: `""`

**Expected Output**:
- Error returned
- Error message indicates invalid message ID

**Implementation**:
```go
func TestGetMessage_EmptyMessageID(t *testing.T) {
    g := createTestService(nil)

    message, err := g.GetMessage("")

    assert.Error(t, err)
    assert.Nil(t, message)
}
```

#### Test Case 2.3: Method Signature Verification

**Test Name**: `TestGetMessageMethodExists`

**Description**: Verify method exists with correct signature

**Implementation**:
```go
func TestGetMessageMethodExists(t *testing.T) {
    var _ interface {
        GetMessage(messageID string) (*gmail.Message, error)
    } = (*googleService)(nil)
}
```

### 3. TestGetAttachment

#### Test Case 3.1: Service Not Initialized

**Test Name**: `TestGetAttachment_ServiceNotInitialized`

**Description**: Verify error handling when Gmail service is not initialized

**Input/Setup**:
- googleService with nil service
- messageID: `"msg_abc123"`
- attachmentID: `"attach_xyz789"`

**Expected Output**:
- Error returned
- Empty byte array

**Implementation**:
```go
func TestGetAttachment_ServiceNotInitialized(t *testing.T) {
    g := createTestService(nil)

    data, err := g.GetAttachment("msg_abc123", "attach_xyz789")

    assert.Error(t, err)
    assert.Nil(t, data)
}
```

#### Test Case 3.2: Empty Message ID

**Test Name**: `TestGetAttachment_EmptyMessageID`

**Description**: Verify validation of empty message ID

**Input/Setup**:
- googleService with nil service
- messageID: `""`
- attachmentID: `"attach_xyz789"`

**Expected Output**:
- Error returned
- Error message indicates invalid message ID

#### Test Case 3.3: Empty Attachment ID

**Test Name**: `TestGetAttachment_EmptyAttachmentID`

**Description**: Verify validation of empty attachment ID

**Input/Setup**:
- googleService with nil service
- messageID: `"msg_abc123"`
- attachmentID: `""`

**Expected Output**:
- Error returned
- Error message indicates invalid attachment ID

#### Test Case 3.4: Method Signature Verification

**Test Name**: `TestGetAttachmentMethodExists`

**Description**: Verify method exists with correct signature

**Implementation**:
```go
func TestGetAttachmentMethodExists(t *testing.T) {
    var _ interface {
        GetAttachment(messageID, attachmentID string) ([]byte, error)
    } = (*googleService)(nil)
}
```

### 4. TestAddLabel

#### Test Case 4.1: Service Not Initialized

**Test Name**: `TestAddLabel_ServiceNotInitialized`

**Description**: Verify error handling when Gmail service is not initialized

**Input/Setup**:
- googleService with nil service
- messageID: `"msg_abc123"`
- labelName: `"fortress-api/processed"`

**Expected Output**:
- Error returned

**Implementation**:
```go
func TestAddLabel_ServiceNotInitialized(t *testing.T) {
    g := createTestService(nil)

    err := g.AddLabel("msg_abc123", "fortress-api/processed")

    assert.Error(t, err)
}
```

#### Test Case 4.2: Empty Message ID

**Test Name**: `TestAddLabel_EmptyMessageID`

**Description**: Verify validation of empty message ID

**Input/Setup**:
- googleService with nil service
- messageID: `""`
- labelName: `"fortress-api/processed"`

**Expected Output**:
- Error returned
- Error message indicates invalid message ID

#### Test Case 4.3: Empty Label Name

**Test Name**: `TestAddLabel_EmptyLabelName`

**Description**: Verify validation of empty label name

**Input/Setup**:
- googleService with nil service
- messageID: `"msg_abc123"`
- labelName: `""`

**Expected Output**:
- Error returned
- Error message indicates invalid label name

#### Test Case 4.4: Method Signature Verification

**Test Name**: `TestAddLabelMethodExists`

**Description**: Verify method exists with correct signature

**Implementation**:
```go
func TestAddLabelMethodExists(t *testing.T) {
    var _ interface {
        AddLabel(messageID, labelName string) error
    } = (*googleService)(nil)
}
```

### 5. TestGetOrCreateLabel

#### Test Case 5.1: Service Not Initialized

**Test Name**: `TestGetOrCreateLabel_ServiceNotInitialized`

**Description**: Verify error handling when Gmail service is not initialized

**Input/Setup**:
- googleService with nil service
- labelName: `"fortress-api/processed"`

**Expected Output**:
- Error returned
- Empty label ID

**Implementation**:
```go
func TestGetOrCreateLabel_ServiceNotInitialized(t *testing.T) {
    g := createTestService(nil)

    labelID, err := g.GetOrCreateLabel("fortress-api/processed")

    assert.Error(t, err)
    assert.Empty(t, labelID)
}
```

#### Test Case 5.2: Empty Label Name

**Test Name**: `TestGetOrCreateLabel_EmptyLabelName`

**Description**: Verify validation of empty label name

**Input/Setup**:
- googleService with nil service
- labelName: `""`

**Expected Output**:
- Error returned
- Error message indicates invalid label name

#### Test Case 5.3: Method Signature Verification

**Test Name**: `TestGetOrCreateLabelMethodExists`

**Description**: Verify method exists with correct signature

**Implementation**:
```go
func TestGetOrCreateLabelMethodExists(t *testing.T) {
    var _ interface {
        GetOrCreateLabel(labelName string) (string, error)
    } = (*googleService)(nil)
}
```

### 6. TestParseEmailHeaders

#### Test Case 6.1: Parse All Standard Headers

**Test Name**: `TestParseEmailHeaders_StandardHeaders`

**Description**: Verify extraction of standard email headers

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
            {Name: "From", Value: "contractor@example.com"},
            {Name: "To", Value: "bill@d.foundation"},
            {Name: "Date", Value: "Mon, 10 Jan 2026 10:00:00 +0000"},
            {Name: "Message-ID", Value: "<msg123@example.com>"},
        },
    },
}
```

**Expected Output**:
```go
map[string]string{
    "subject": "Invoice CONTR-202501-A1B2",
    "from": "contractor@example.com",
    "to": "bill@d.foundation",
    "date": "Mon, 10 Jan 2026 10:00:00 +0000",
    "message-id": "<msg123@example.com>",
}
```

**Implementation**:
```go
func TestParseEmailHeaders_StandardHeaders(t *testing.T) {
    message := &gmail.Message{
        Payload: &gmail.MessagePart{
            Headers: []*gmail.MessagePartHeader{
                {Name: "Subject", Value: "Invoice CONTR-202501-A1B2"},
                {Name: "From", Value: "contractor@example.com"},
                {Name: "To", Value: "bill@d.foundation"},
            },
        },
    }

    headers := ParseEmailHeaders(message)

    assert.Equal(t, "Invoice CONTR-202501-A1B2", headers["subject"])
    assert.Equal(t, "contractor@example.com", headers["from"])
    assert.Equal(t, "bill@d.foundation", headers["to"])
}
```

#### Test Case 6.2: Missing Headers

**Test Name**: `TestParseEmailHeaders_MissingHeaders`

**Description**: Verify handling of messages with missing headers

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{
            {Name: "From", Value: "contractor@example.com"},
        },
    },
}
```

**Expected Output**:
- Map contains "from" key
- Missing headers return empty string

#### Test Case 6.3: Nil Payload

**Test Name**: `TestParseEmailHeaders_NilPayload`

**Description**: Verify handling of message with nil payload

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: nil,
}
```

**Expected Output**:
- Empty map or error (implementation dependent)

#### Test Case 6.4: Empty Headers Array

**Test Name**: `TestParseEmailHeaders_EmptyHeaders`

**Description**: Verify handling of empty headers array

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Headers: []*gmail.MessagePartHeader{},
    },
}
```

**Expected Output**:
- Empty map

### 7. TestFindPDFAttachment

#### Test Case 7.1: PDF Found in Parts

**Test Name**: `TestFindPDFAttachment_PDFFound`

**Description**: Verify detection of PDF attachment

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
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

**Expected Output**:
- attachmentID: `"attach_123"`
- filename: `"invoice.pdf"`
- found: `true`

**Implementation**:
```go
func TestFindPDFAttachment_PDFFound(t *testing.T) {
    message := &gmail.Message{
        Payload: &gmail.MessagePart{
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

    attachmentID, filename, found := FindPDFAttachment(message)

    assert.True(t, found)
    assert.Equal(t, "attach_123", attachmentID)
    assert.Equal(t, "invoice.pdf", filename)
}
```

#### Test Case 7.2: No PDF Attachment

**Test Name**: `TestFindPDFAttachment_NoPDF`

**Description**: Verify handling when no PDF is present

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Parts: []*gmail.MessagePart{
            {
                Filename: "document.docx",
                MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_456",
                },
            },
        },
    },
}
```

**Expected Output**:
- attachmentID: `""`
- filename: `""`
- found: `false`

#### Test Case 7.3: Multiple Attachments (PDF Second)

**Test Name**: `TestFindPDFAttachment_MultipleAttachments`

**Description**: Verify returns first PDF when multiple files present

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Parts: []*gmail.MessagePart{
            {
                Filename: "cover.txt",
                MimeType: "text/plain",
            },
            {
                Filename: "invoice.pdf",
                MimeType: "application/pdf",
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_789",
                },
            },
            {
                Filename: "receipt.pdf",
                MimeType: "application/pdf",
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_999",
                },
            },
        },
    },
}
```

**Expected Output**:
- Returns first PDF: `"invoice.pdf"`, `"attach_789"`

#### Test Case 7.4: Nil Parts

**Test Name**: `TestFindPDFAttachment_NilParts`

**Description**: Verify handling of message with no parts

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Parts: nil,
    },
}
```

**Expected Output**:
- found: `false`

#### Test Case 7.5: Empty Parts Array

**Test Name**: `TestFindPDFAttachment_EmptyParts`

**Description**: Verify handling of empty parts array

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Parts: []*gmail.MessagePart{},
    },
}
```

**Expected Output**:
- found: `false`

#### Test Case 7.6: Case Insensitive MIME Type

**Test Name**: `TestFindPDFAttachment_CaseInsensitiveMimeType`

**Description**: Verify MIME type matching is case insensitive (if implemented)

**Input/Setup**:
```go
message := &gmail.Message{
    Payload: &gmail.MessagePart{
        Parts: []*gmail.MessagePart{
            {
                Filename: "invoice.pdf",
                MimeType: "APPLICATION/PDF",  // Uppercase
                Body: &gmail.MessagePartBody{
                    AttachmentId: "attach_abc",
                },
            },
        },
    },
}
```

**Expected Output**:
- found: `true` (if case-insensitive matching implemented)

## Edge Cases and Error Scenarios

### Edge Case 1: Very Long Query String

**Test Name**: `TestListInboxMessages_LongQuery`

**Description**: Verify handling of very long query strings

**Input/Setup**:
- Query with 1000+ characters

**Expected Behavior**:
- Should pass query to Gmail API (API will handle validation)

### Edge Case 2: Special Characters in Label Name

**Test Name**: `TestAddLabel_SpecialCharacters`

**Description**: Verify handling of label names with special characters

**Input/Setup**:
- labelName: `"fortress-api/processed/2026-01"`

**Expected Behavior**:
- Gmail API handles validation

### Edge Case 3: Unicode in Headers

**Test Name**: `TestParseEmailHeaders_UnicodeCharacters`

**Description**: Verify handling of Unicode characters in headers

**Input/Setup**:
```go
headers := []*gmail.MessagePartHeader{
    {Name: "Subject", Value: "Invoice 中文 CONTR-202501-A1B2"},
    {Name: "From", Value: "contractor@example.com"},
}
```

**Expected Output**:
- Unicode characters preserved in parsed headers

## Test Coverage Goals

### Method Coverage
- All public methods tested
- All error paths covered
- All validation logic tested

### Error Handling Coverage
- Service not initialized
- Invalid parameters (empty strings)
- Nil pointer handling
- Missing data handling

### Signature Verification
- All new interface methods verified to exist

## Testing Patterns Used

### Table-Driven Tests
Follow existing pattern from `google_mail_sendas_test.go`:

```go
func TestMethodName(t *testing.T) {
    tests := []struct {
        name        string
        // input fields
        wantErr     bool
        expectedErr string
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Interface Verification Tests
Verify method signatures exist:

```go
func TestMethodExists(t *testing.T) {
    var _ interface {
        MethodName(params) (return, error)
    } = (*googleService)(nil)
}
```

### Assertions
Use `testify/assert` for cleaner assertions:

```go
assert.Error(t, err)
assert.NoError(t, err)
assert.Equal(t, expected, actual)
assert.Nil(t, value)
assert.NotNil(t, value)
assert.True(t, condition)
assert.False(t, condition)
```

## Notes for Implementer

1. **Service Initialization**: Tests use nil service to verify error handling. Integration tests will use real Gmail service.

2. **Token Handling**: Tests use dummy token. Real OAuth flow tested separately.

3. **Helper Functions**: `ParseEmailHeaders` and `FindPDFAttachment` are pure functions - easier to test thoroughly.

4. **Error Messages**: Don't assert exact error messages (may vary with token state) - just verify error occurred.

5. **Mock Limitations**: Without full Gmail API mocks, tests verify interface contracts and basic validation only.

6. **Future Enhancement**: Consider adding gomock-based tests for full Gmail API interaction testing.
