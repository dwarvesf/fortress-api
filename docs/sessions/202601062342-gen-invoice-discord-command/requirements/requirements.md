# Requirements: Discord Command `?gen invoice` for Contractor Invoice Generation

## Overview
Create a Discord command that allows active contractors to generate their own invoices via async processing.

## Functional Requirements

### Command Syntax
- **Command**: `?gen invoice` or `?gen inv` (short form)
- **Target**: Self only - generates invoice for the caller's Discord username
- **Optional argument**: Month in YYYY-MM format (defaults to current month)

### Examples
```
?gen invoice           # Generate for current month
?gen inv               # Short form
?gen invoice 2025-01   # Generate for specific month
```

### Permission & Validation
- API validates contractor status - returns error if not an active contractor
- User must exist in Notion Contractor page with status "Active"

### Async Processing (avoid Discord 3s timeout)
1. User submits command
2. **DM user immediately** with "Processing..." embed message
3. POST webhook to fortress-api with `dmChannelID` and `dmMessageID`
4. Return immediately (avoid timeout)
5. fortress-api processes async in background
6. **Update the same DM message** with result (success/error)

### Notifications
- **Discord DM**: Update the processing message with result
- **Google Drive sharing**: Share PDF file to contractor's Personal Email (from Notion Contractor table)
- Google sends automatic notification email when file is shared

### Rate Limiting
- Limit: 3 times per day per user
- Storage: In-memory (no database migration needed)
- Resets at midnight / server restart

## Technical Requirements

### fortress-discord Changes
- New command: `?gen` with subcommand `invoice/inv`
- New model, adapter, service, view, command layers
- Send DM with processing embed, return `dmChannelID` and `dmMessageID`
- POST webhook to fortress-api

### fortress-api Changes
- New webhook endpoint: `POST /webhooks/discord/gen-invoice`
- In-memory rate limiter
- Google Drive file sharing method (`ShareFileWithEmail`)
- Use existing `UpdateChannelMessage` to edit DM
- Use existing `GenerateContractorInvoice` controller
- Use existing `GetContractorPersonalEmail` from Notion service

## Non-Functional Requirements

- No database migration needed (in-memory rate limiting)
- No SendGrid email notification (use Google Drive sharing notification)
- Reuse existing invoice generation logic
- Thread-safe rate limiter with mutex
