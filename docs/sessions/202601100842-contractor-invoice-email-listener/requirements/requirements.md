# Contractor Invoice Email Listener - Requirements

## Overview
Implement a Gmail inbox listener that monitors a configurable email address (e.g., `bill@d.foundation`) for contractor invoice submissions. When an invoice email with PDF attachment is received, the system extracts the Invoice ID and updates the corresponding Contractor Payable status from "New" to "Pending".

## Business Context
- Contractors submit invoices via email to a designated billing address
- This creates a verification step before payables become eligible for payment
- Status flow: `New → Pending → Paid`
- "Pending" status indicates the contractor has submitted their invoice

## Functional Requirements

### FR-1: Email Monitoring
- Monitor a configurable Gmail inbox for incoming emails
- Use Gmail API Watch/Polling mechanism
- Process only unread emails

### FR-2: Invoice ID Extraction
- Extract Invoice ID from email subject line first
- Fallback: Parse PDF attachment to find Invoice ID
- Invoice ID pattern: `CONTR-YYYYMM-XXXX` (e.g., `CONTR-202501-A1B2`)

### FR-3: Payable Status Update
- Match extracted Invoice ID with Notion Contractor Payables database
- Update payable status from "New" to "Pending"
- Only update if current status is "New"

### FR-4: Email Processing Tracking
- Mark processed emails (add Gmail label)
- Prevent duplicate processing

## Non-Functional Requirements

### NFR-1: Configuration
- Email address fully configurable via environment variable
- Gmail refresh token configurable (can use different account)
- Poll interval configurable
- Gmail labels configurable

### NFR-2: Reliability
- Handle Gmail API rate limits gracefully
- Log all processing steps for debugging
- Continue processing on individual email failures

### NFR-3: Security
- Use existing OAuth2 infrastructure
- No new credentials required (reuse Google OAuth config)

## User Decisions (from planning session)
1. **Email mechanism**: Gmail API Watch/Polling (not SendGrid Inbound Parse)
2. **Invoice ID extraction**: Subject first, fallback to PDF parsing
3. **PDF handling**: No upload to Notion, just status update
4. **Gmail account**: Configurable via dedicated environment variable

## Out of Scope
- Webhook-based email receiving (requires Pub/Sub setup)
- PDF upload to Notion attachments
- Email reply/notification to contractor
