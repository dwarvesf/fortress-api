# Gmail Inbox Service Specification

## Overview

Extend the existing `pkg/service/googlemail/` service to support inbox reading operations. Currently, the service only supports sending emails. This specification adds methods for listing messages, fetching message details, downloading attachments, and managing labels.

## Service Extension

### File: `pkg/service/googlemail/google_mail.go`

### New Methods

#### 1. ListInboxMessages

**Purpose**: List unread messages matching a query filter

**Signature**:
```go
func (g *googleService) ListInboxMessages(query string, maxResults int64) ([]*gmail.Message, error)
```

**Parameters**:
- `query` (string): Gmail search query (e.g., `"is:unread has:attachment"`)
- `maxResults` (int64): Maximum number of messages to return (default: 100)

**Returns**:
- `[]*gmail.Message`: List of message metadata (ID, ThreadID, LabelIDs, Snippet)
- `error`: Error if API call fails

**Behavior**:
- Ensures OAuth token is valid before calling API
- Prepares Gmail service if not already initialized
- Uses `Users.Messages.List` API with query filter
- Returns minimal message metadata (IDs only)
- Logs API call at DEBUG level

**Error Handling**:
- Returns error if token refresh fails
- Returns error if service initialization fails
- Returns error if Gmail API call fails (network, auth, rate limit)

**Example Usage**:
```go
query := "is:unread -label:fortress-api/processed has:attachment"
messages, err := gmailService.ListInboxMessages(query, 50)
if err != nil {
    return fmt.Errorf("failed to list messages: %w", err)
}
```

#### 2. GetMessage

**Purpose**: Fetch full message details including headers, body, and attachments

**Signature**:
```go
func (g *googleService) GetMessage(messageID string) (*gmail.Message, error)
```

**Parameters**:
- `messageID` (string): Gmail message ID from ListInboxMessages

**Returns**:
- `*gmail.Message`: Full message with payload, headers, parts, attachments
- `error`: Error if API call fails

**Behavior**:
- Ensures OAuth token is valid
- Prepares Gmail service if needed
- Uses `Users.Messages.Get` with `format=full` to get all details
- Returns complete message structure including attachments
- Logs message ID at DEBUG level

**Error Handling**:
- Returns error if message ID is empty
- Returns error if message not found (404)
- Returns error if Gmail API call fails

**Example Usage**:
```go
message, err := gmailService.GetMessage("msg_abc123")
if err != nil {
    return fmt.Errorf("failed to get message: %w", err)
}
```

#### 3. GetAttachment

**Purpose**: Download attachment content from a message

**Signature**:
```go
func (g *googleService) GetAttachment(messageID, attachmentID string) ([]byte, error)
```

**Parameters**:
- `messageID` (string): Gmail message ID
- `attachmentID` (string): Attachment ID from message parts

**Returns**:
- `[]byte`: Decoded attachment content (binary data)
- `error`: Error if download fails

**Behavior**:
- Ensures OAuth token is valid
- Uses `Users.Messages.Attachments.Get` API
- Decodes base64-encoded attachment data
- Returns raw bytes for further processing
- Logs attachment ID and size at DEBUG level

**Error Handling**:
- Returns error if messageID or attachmentID is empty
- Returns error if attachment not found
- Returns error if base64 decoding fails
- Returns error if Gmail API call fails

**Example Usage**:
```go
pdfBytes, err := gmailService.GetAttachment(messageID, attachmentID)
if err != nil {
    return fmt.Errorf("failed to get attachment: %w", err)
}
```

#### 4. AddLabel

**Purpose**: Add a label to a message (mark as processed)

**Signature**:
```go
func (g *googleService) AddLabel(messageID, labelName string) error
```

**Parameters**:
- `messageID` (string): Gmail message ID
- `labelName` (string): Label name (e.g., `"fortress-api/processed"`)

**Returns**:
- `error`: Error if label application fails

**Behavior**:
- Ensures OAuth token is valid
- Gets or creates label by name (calls `GetOrCreateLabel`)
- Uses `Users.Messages.Modify` to add label ID to message
- Logs label application at DEBUG level

**Error Handling**:
- Returns error if messageID or labelName is empty
- Returns error if label creation fails
- Returns error if label application fails

**Example Usage**:
```go
err := gmailService.AddLabel("msg_abc123", "fortress-api/processed")
if err != nil {
    return fmt.Errorf("failed to add label: %w", err)
}
```

#### 5. GetOrCreateLabel

**Purpose**: Get label ID by name, creating it if it doesn't exist

**Signature**:
```go
func (g *googleService) GetOrCreateLabel(labelName string) (string, error)
```

**Parameters**:
- `labelName` (string): Label name (e.g., `"fortress-api/processed"`)

**Returns**:
- `string`: Label ID
- `error`: Error if operation fails

**Behavior**:
- Lists all labels for user
- Searches for label by name (case-sensitive)
- If found, returns existing label ID
- If not found, creates new label with:
  - `name`: labelName
  - `messageListVisibility`: "show"
  - `labelListVisibility`: "labelShow"
- Returns label ID
- Logs label creation at INFO level

**Error Handling**:
- Returns error if labelName is empty
- Returns error if list labels API call fails
- Returns error if create label API call fails

**Example Usage**:
```go
labelID, err := gmailService.GetOrCreateLabel("fortress-api/processed")
if err != nil {
    return fmt.Errorf("failed to get or create label: %w", err)
}
```

### Helper Methods

#### ParseEmailHeaders

**Purpose**: Extract common headers from message

**Signature**:
```go
func ParseEmailHeaders(message *gmail.Message) map[string]string
```

**Parameters**:
- `message` (*gmail.Message): Full message from GetMessage

**Returns**:
- `map[string]string`: Map of header name → value (e.g., `"Subject"`, `"From"`, `"To"`, `"Date"`)

**Behavior**:
- Iterates through message.Payload.Headers
- Extracts key headers: Subject, From, To, Cc, Date, Message-ID
- Returns lowercase header names as keys
- Returns empty string for missing headers

**Example Usage**:
```go
headers := ParseEmailHeaders(message)
subject := headers["subject"]
from := headers["from"]
```

#### FindPDFAttachment

**Purpose**: Find first PDF attachment in message

**Signature**:
```go
func FindPDFAttachment(message *gmail.Message) (attachmentID, filename string, found bool)
```

**Parameters**:
- `message` (*gmail.Message): Full message from GetMessage

**Returns**:
- `attachmentID` (string): Attachment ID for downloading
- `filename` (string): Original filename
- `found` (bool): True if PDF found

**Behavior**:
- Iterates through message.Payload.Parts
- Looks for MIME type `application/pdf`
- Returns first PDF found
- Returns empty strings and false if no PDF found

**Example Usage**:
```go
attachmentID, filename, found := FindPDFAttachment(message)
if !found {
    return errors.New("no PDF attachment found")
}
```

## Interface Update

### File: `pkg/service/googlemail/interface.go`

Add new methods to `IService` interface:

```go
type IService interface {
    // ... existing methods ...

    // Inbox reading methods
    ListInboxMessages(query string, maxResults int64) ([]*gmail.Message, error)
    GetMessage(messageID string) (*gmail.Message, error)
    GetAttachment(messageID, attachmentID string) ([]byte, error)
    AddLabel(messageID, labelName string) error
    GetOrCreateLabel(labelName string) (string, error)
}
```

## Configuration

No new configuration required. Existing OAuth2 credentials and refresh tokens will be used.

**Refresh Token Selection**:
- Use `INVOICE_LISTENER_REFRESH_TOKEN` (new env var) for invoice listener operations
- Falls back to existing tokens if not configured
- Allows using dedicated Google account for billing operations

## Gmail API Scopes

**Required Scopes** (already configured in existing OAuth2 setup):
- `https://www.googleapis.com/auth/gmail.readonly` - Read messages and labels
- `https://www.googleapis.com/auth/gmail.modify` - Add labels to messages
- `https://www.googleapis.com/auth/gmail.labels` - Create and manage labels

These scopes are already included in the existing Gmail service configuration.

## Rate Limiting

**Gmail API Quotas**:
- 250 quota units per user per second
- 1,000,000,000 quota units per day

**Method Quota Costs**:
- `Users.Messages.List`: 5 quota units
- `Users.Messages.Get`: 5 quota units
- `Users.Messages.Attachments.Get`: 5 quota units
- `Users.Messages.Modify`: 5 quota units
- `Users.Labels.List`: 1 quota unit
- `Users.Labels.Create`: 5 quota units

**Mitigation**:
- Poll interval of 5 minutes = 288 polls per day
- With 50 messages per poll: ~7,200 API calls per day
- Total quota usage: ~36,000 units per day (0.0036% of daily limit)
- Implement exponential backoff if rate limit hit

## Error Handling

### Error Types

1. **Authentication Errors**:
   - Refresh token expired or invalid
   - OAuth scope insufficient
   - **Action**: Log error, send alert, halt processing

2. **Network Errors**:
   - Temporary API unavailability
   - Timeout
   - **Action**: Log warning, retry with exponential backoff

3. **Rate Limit Errors**:
   - HTTP 429 Too Many Requests
   - **Action**: Sleep and retry with exponential backoff

4. **Resource Not Found**:
   - Message or attachment deleted between list and get
   - **Action**: Log warning, skip message, continue processing

5. **Invalid Data**:
   - Malformed message structure
   - Missing expected fields
   - **Action**: Log error, skip message, continue processing

### Logging Standards

All methods must log:
- **Entry**: DEBUG level with parameters
- **Success**: DEBUG level with result summary
- **Error**: ERROR level with error details and context

**Example Logging**:
```go
logger.Debug(fmt.Sprintf("[DEBUG] gmail_inbox: listing messages query=%s maxResults=%d", query, maxResults))
logger.Debug(fmt.Sprintf("[DEBUG] gmail_inbox: found %d messages", len(messages)))
logger.Error(err, fmt.Sprintf("[ERROR] gmail_inbox: failed to list messages query=%s", query))
```

## Testing Strategy

### Unit Tests

**File**: `pkg/service/googlemail/google_mail_inbox_test.go`

**Test Cases**:
1. `TestListInboxMessages_Success` - Successfully list messages
2. `TestListInboxMessages_EmptyResult` - No messages found
3. `TestListInboxMessages_TokenRefreshFails` - Auth error handling
4. `TestGetMessage_Success` - Get full message
5. `TestGetMessage_NotFound` - Message deleted
6. `TestGetAttachment_Success` - Download PDF
7. `TestGetAttachment_Base64DecodeFails` - Malformed attachment
8. `TestAddLabel_Success` - Apply label
9. `TestGetOrCreateLabel_Exists` - Label already exists
10. `TestGetOrCreateLabel_Creates` - Create new label
11. `TestParseEmailHeaders` - Extract headers
12. `TestFindPDFAttachment_Found` - PDF in message
13. `TestFindPDFAttachment_NotFound` - No PDF

**Mocking**:
- Mock `gmail.Service` using interfaces
- Mock OAuth2 token source
- Use testdata for sample Gmail API responses

### Integration Tests

**Prerequisites**:
- Test Gmail account with OAuth2 credentials
- Test emails with PDF attachments

**Test Flow**:
1. Send test email with PDF attachment
2. List messages and verify test email found
3. Get message and verify headers
4. Download attachment and verify content
5. Add label and verify label applied
6. List again and verify labeled message excluded

## Security Considerations

1. **Token Storage**: Refresh tokens stored in environment variables (not in code)
2. **Scope Minimization**: Use only required scopes (readonly + modify for labels)
3. **Attachment Validation**: Verify MIME type before processing attachments
4. **Input Sanitization**: Validate message IDs and attachment IDs before API calls
5. **Error Messages**: Don't expose sensitive data in error logs (mask email addresses in non-prod)

## Migration Path

1. **Phase 1**: Add new methods to `googleService` struct (no breaking changes)
2. **Phase 2**: Update `IService` interface (compile error if interface consumers exist)
3. **Phase 3**: Add unit tests for new methods
4. **Phase 4**: Integration testing with test Gmail account
5. **Phase 5**: Deploy and monitor logs for errors

## Dependencies

**New Dependencies**: None (uses existing `google.golang.org/api/gmail/v1` package)

**Existing Dependencies**:
- `golang.org/x/oauth2` - OAuth2 token management
- `google.golang.org/api/gmail/v1` - Gmail API client
- `google.golang.org/api/option` - API client options

## References

- Gmail API Reference: https://developers.google.com/gmail/api/reference/rest
- Gmail API Guides: https://developers.google.com/gmail/api/guides
- Existing Gmail Service: `/pkg/service/googlemail/google_mail.go`
- OAuth2 Config: `/pkg/config/config.go` (Google struct)
