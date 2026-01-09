# Task Order Confirmation Email Update - Requirements

## Overview
Update the existing task order confirmation email to use a new invoice reminder format with signature from Han Ngo (CTO & Managing Director) instead of Team Dwarves (People Operations).

## Business Requirements

### Email Purpose
- **Current**: Work order confirmation requiring contractor acknowledgment
- **New**: Invoice reminder + client milestones awareness + positive reinforcement

### Target Recipients
- All active contractors receiving monthly work orders
- Sent via existing cronjob endpoint: `/cronjobs/send-task-order-confirmation`

## Functional Requirements

### 1. Email Header Changes
- **Subject**: "Quick update for [Month Year] – Invoice reminder & client milestones"
- **From**: "Spawn @ Dwarves LLC" <spawn@d.foundation> (NO CHANGE)
- **Signature**: "Han Ngo, CTO & Managing Director, Dwarves LLC" (CHANGED from "Team Dwarves, People Operations")

### 2. Email Body Content
New structure with the following sections:

#### a. Greeting
- Use contractor's last name (existing behavior)
- Format: "Hi [LastName],"

#### b. Opening
- "Hope you're having a great start to [Month Year]!"

#### c. Invoice Reminder
- Mention monthly invoice for [Month Year] services
- **Dynamic Due Date**: Based on contractor's Payday schedule
  - Payday 1 → Due by 10th of month
  - Payday 15 → Due by 25th of month
- Template submission destination: billing@dwarves.llc

#### d. Client Milestones (for awareness)
- Bullet list of upcoming client milestones
- **Phase 1**: Hardcoded mock data
- **Phase 2** (future): Dynamic data from real source (TBD)
- Structure must be easily replaceable

#### e. Encouragement
- Acknowledge contractor's excellent work on embedded team
- Note client satisfaction

#### f. Support Offer
- Open door for questions or support needs

#### g. Signature
- Best regards from Han Ngo
- Title: CTO & Managing Director
- Company: Dwarves LLC

### 3. Invoice Due Date Logic

#### Data Source
- **Field**: Payday (Select field with values "01" or "15")
- **Database**: Service Rate (Notion database)
- **Filter**: Contractor relation + Status="Active"

#### Calculation Rules
```
IF Payday = "01" OR Payday = null/invalid
  THEN Due Date = "10th"
ELSE IF Payday = "15"
  THEN Due Date = "25th"
```

#### Error Handling
- Graceful fallback to "10th" for any missing/invalid data
- Email must always send successfully (no blocking on Payday issues)
- Log issues for monitoring but continue processing

### 4. Client Milestones

#### Phase 1 (Current Implementation)
- Hardcoded array of milestone strings
- Located in handler code
- Clearly marked with TODO comment for future replacement

#### Example Mock Data
```
- "Project Alpha: Feature X delivery by Jan 20"
- "Project Beta: Code review session on Jan 15"
```

#### Phase 2 (Future - Out of Scope)
Replace with dynamic data from:
- Option A: Notion Project database properties
- Option B: External project management API
- Option C: Configuration file

Must be structured to make replacement straightforward.

## Non-Functional Requirements

### Compatibility
- Maintain existing endpoint and query parameters
- Preserve MIME multipart email format
- Keep Gmail API integration (spawn@d.foundation account)
- Support existing test_email override for testing

### Performance
- No significant performance degradation
- Payday fetch should not add more than 100ms per contractor

### Observability
- Debug logs for Payday fetch results
- Debug logs for due date calculation
- Track default fallback usage (payday=0 cases)
- Monitor email delivery success rate (target: >95%)

### Data Quality
- Expected: >90% of contractors should have configured Payday
- Monitor and alert on missing Payday data

## Technical Constraints

### Email Format
- HTML with quoted-printable encoding
- MIME 1.0 multipart/mixed boundary format
- Reuse existing signature.tpl template
- Maintain responsive design

### Data Access
- Use existing Notion SDK patterns
- Follow repository pattern from contractor_payables service
- No direct database queries (use Notion API)

### Configuration
- Service Rate database ID must be in config
- No new environment variables required
- No schema changes to Notion databases

## Success Criteria

### Must Have
1. ✅ Email sends successfully to all contractors
2. ✅ Subject line includes "Invoice reminder & client milestones"
3. ✅ Due date shows "10th" or "25th" based on Payday
4. ✅ Signature shows "Han Ngo, CTO & Managing Director"
5. ✅ Email tone is friendly and encouraging
6. ✅ Milestones display as bullet list

### Should Have
1. ✅ >90% of contractors have correct Payday data
2. ✅ Email delivery rate >95%
3. ✅ Default fallback usage <10%
4. ✅ Debug logs available for troubleshooting

### Nice to Have
1. ⭕ Metrics dashboard for Payday coverage
2. ⭕ Automated alerts for missing Payday data
3. ⭕ A/B testing framework for email variations

## Out of Scope
- Creating new email types (this replaces existing)
- Implementing real milestone data source (Phase 2)
- Changing email sending frequency
- Adding attachments or images to email
- Multi-language support
- SMS notifications

## Rollback Plan

### Quick Rollback (Template Only)
- Revert `pkg/templates/taskOrderConfirmation.tpl` only
- Result: Old content, new signature

### Full Rollback
- Revert entire commit
- Result: Complete restoration to previous state

## User Stories

### As a Contractor
- I want to receive a clear invoice due date
- I want to know upcoming client milestones
- I want encouragement and support from leadership

### As Operations Team
- I want automated invoice reminders
- I want visibility into contractor engagement
- I want easy troubleshooting when issues occur

### As CTO (Han Ngo)
- I want personal connection with contractors
- I want consistent branding in communications
- I want to acknowledge contractor contributions
