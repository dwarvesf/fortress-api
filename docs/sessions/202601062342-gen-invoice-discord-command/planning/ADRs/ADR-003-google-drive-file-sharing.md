# ADR-003: Google Drive File Sharing for Invoice Delivery

## Status
Proposed

## Context
After generating an invoice PDF, we need to deliver it to the contractor. The contractor should:
- Receive access to the invoice file
- Be notified that the invoice is ready
- Be able to access the file from their personal email
- Have the file stored centrally (not just emailed)

Currently, fortress-api:
- Uploads invoice PDFs to Google Drive (existing functionality)
- Has Google Drive service at `pkg/service/googledrive/google_drive.go`
- Has SendGrid integration for email notifications
- Can retrieve contractor's Personal Email from Notion

### Constraints
- Invoice already uploaded to Google Drive (existing process)
- Want to avoid duplicate storage (Drive + email attachment)
- Contractor's Personal Email is available from Notion
- Must provide notification to contractor
- Should leverage existing infrastructure

### Options Considered

#### Option 1: SendGrid Email with Embedded Link (Rejected)
Send email via SendGrid with link to Google Drive file.

**Pros:**
- Explicit notification control
- Custom email template/branding
- Can include additional context
- Existing SendGrid integration

**Cons:**
- Requires SendGrid API call + Drive API call
- Duplicate notification mechanism (Google also notifies on share)
- Email template maintenance overhead
- Contractor needs to click link to access file
- Link might break if permissions change
- More code to maintain

#### Option 2: SendGrid Email with PDF Attachment (Rejected)
Download PDF from Drive and email as attachment via SendGrid.

**Pros:**
- Direct file access in email
- Works offline once downloaded
- Familiar email attachment UX

**Cons:**
- Duplicate storage (Drive + email + SendGrid CDN)
- Large email size (PDFs can be several MB)
- Email deliverability issues with attachments
- No central file management (scattered across mailboxes)
- SendGrid attachment size limits
- Wastes Google Drive upload (file already there)
- More bandwidth and storage costs

#### Option 3: Google Drive API File Sharing (Selected)
Use Google Drive Share API to grant contractor's email access to the file.

**Pros:**
- Single source of truth (file in Drive)
- **Google automatically sends notification email** when file is shared
- Native Google Workspace integration
- Contractor can access file directly from Gmail
- Central file management (all invoices in organized Drive structure)
- No duplicate storage
- Better permission control (can revoke access)
- Simpler code (one API call)
- Built-in notification system

**Cons:**
- Relies on Google's notification system (less customization)
- Requires contractor email to be Google-compatible
- Notification email template is Google's default (not branded)

## Decision
We will use **Option 3: Google Drive API File Sharing**.

### Implementation Pattern

#### New Method in Google Drive Service
```go
// pkg/service/googledrive/google_drive.go

// ShareFileWithEmail grants access to a file and sends notification
func (s *service) ShareFileWithEmail(fileID, email string) error {
    permission := &drive.Permission{
        Type:         "user",
        Role:         "reader",
        EmailAddress: email,
    }

    _, err := s.client.Permissions.Create(fileID, permission).
        SendNotificationEmail(true).  // Google sends email notification
        EmailMessage("Your invoice has been generated and is ready for review.").
        Do()

    return err
}
```

#### Usage in Invoice Generation Flow
```go
// After uploading invoice to Drive
fileID := uploadedFile.Id

// Get contractor's personal email from Notion
personalEmail, err := s.notion.GetContractorPersonalEmail(discordUsername)
if err != nil {
    return fmt.Errorf("failed to get contractor email: %w", err)
}

// Share file with contractor
err = s.googleDrive.ShareFileWithEmail(fileID, personalEmail)
if err != nil {
    return fmt.Errorf("failed to share file: %w", err)
}

// Google automatically sends notification email to contractor
```

### Google's Notification Email
Google automatically sends an email like:
```
Subject: [User] shared "Invoice_January_2025.pdf" with you
Body:
  [User] has shared a file with you:

  Invoice_January_2025.pdf
  [Open in Drive]

  Custom message: Your invoice has been generated and is ready for review.
```

Contractor can:
- Click "Open in Drive" to view in browser
- Download the file
- Add to their own Drive with "Add to My Drive"

### Permission Model
- Type: `user` (specific email address)
- Role: `reader` (view and download only)
- Notification: `true` (Google sends email)
- Custom message: Optional context message

### File Organization
Invoices stored in Drive with structure:
```
/Invoices/
  /Contractors/
    /2025/
      /01-January/
        invoice_contractor1_2025-01.pdf
        invoice_contractor2_2025-01.pdf
```

Each contractor gets `reader` permission on their own invoice only (not the entire folder).

### Error Handling

#### Email Not Found
```go
if personalEmail == "" {
    return fmt.Errorf("contractor email not found in Notion")
}
```

#### Invalid Email
```go
// Google Drive API returns error for invalid email
if err != nil && strings.Contains(err.Error(), "invalid email") {
    return fmt.Errorf("invalid contractor email: %s", personalEmail)
}
```

#### Permission Already Exists
```go
// Google Drive API handles this gracefully (idempotent)
// Still sends notification email even if permission exists
```

## Consequences

### Positive
- Leverages existing Google infrastructure (already using Drive)
- **Zero additional notification infrastructure** (Google handles it)
- Single source of truth for files
- Better file management (central storage)
- Permission control (can revoke access)
- Automatic notification without custom email code
- Smaller attack surface (fewer systems involved)
- Less code to maintain (no email templates)
- Native integration with Google Workspace

### Negative
- Cannot customize notification email template
- Relies on contractor having email access (acceptable - payment info is delivered via email anyway)
- Notification email is in English (Google's default)
- Cannot add custom branding to notification

### Acceptable Trade-offs

#### Google's Default Email Template
This is acceptable because:
- Email clearly states file is shared
- Contains direct link to file
- Professional appearance (Google's template)
- Includes custom message field (can add context)
- Contractor is already familiar with Google Drive notifications

#### English-Only Notification
This is acceptable because:
- Current contractor base primarily works in English
- Invoice documents themselves are in English
- If internationalization is needed later, we can:
  - Add SendGrid notification alongside Drive sharing
  - Use Google Workspace custom templates (if available)

### Security Considerations
- File permissions are scoped to specific email (not public link)
- Permission can be revoked if needed
- Audit trail in Google Drive Activity logs
- No file data leaves Google infrastructure
- HTTPS encryption for all file access

### Testing Strategy
```go
func TestShareFileWithEmail(t *testing.T) {
    // Test cases:
    // 1. Successful share with valid email
    // 2. Error on invalid email format
    // 3. Idempotent - sharing twice doesn't error
    // 4. Notification email is sent (check mock calls)
}
```

### Future Enhancements
If we need custom notifications later, we can:
1. Keep Drive sharing (for file access)
2. Add SendGrid email (for custom branding)
3. Disable Google's notification (`SendNotificationEmail(false)`)
4. Send our own branded email with Drive link

This gives us flexibility without architectural changes.

## References
- Google Drive API Permissions: https://developers.google.com/drive/api/v3/reference/permissions
- SendNotificationEmail: https://developers.google.com/drive/api/v3/reference/permissions/create
- fortress-api Google Drive service: `pkg/service/googledrive/google_drive.go`
- fortress-api Notion service: `pkg/service/notion/task_order_log.go:1595`
