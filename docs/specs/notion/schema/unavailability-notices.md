# Unavailability Notices Database Schema

## Overview
The Unavailability Notices database tracks contractor leave requests, absences, and time-off notifications. It supports approval workflows and integrates with contractor records for automated notifications.

**Database ID**: `2cc64b29-b84c-80ef-bb0e-000bf2c8bfcb`
**Parent Database ID**: `2cc64b29-b84c-80b2-95ea-e2c134f2ca20`
**Icon**: 🎆 (burst icon)

## Properties

### Core Fields

#### Leave Request (Title)
- **Type**: `title`
- **Description**: The name/title of the leave request
- **Required**: Yes
- **Example**: "Medical Leave - Jan 15-20"

#### Contractor (Relation)
- **Property ID**: `{iuX`
- **Type**: `relation`
- **Description**: Links to the contractor submitting the leave request
- **Related Database**: `ed2b9224-97d9-4dff-97f9-82598b61f65d`
- **Relation Type**: Single property
- **Required**: Yes

#### ID (Unique ID)
- **Property ID**: `wC|R`
- **Type**: `unique_id`
- **Description**: Auto-incrementing ID for each request
- **Prefix**: None
- **Auto-generated**: Yes

### Date Fields

#### Start Date
- **Property ID**: `HALl`
- **Type**: `date`
- **Description**: Leave start date
- **Required**: Yes

#### End Date
- **Property ID**: `TMW\``
- **Type**: `date`
- **Description**: Leave end date
- **Required**: No (if missing, treated as single-day leave)

#### Date Requested
- **Property ID**: `XxYa`
- **Type**: `date`
- **Description**: When the request was submitted
- **Auto-populated**: Via created_time

#### Date Approved
- **Property ID**: `nkKI`
- **Type**: `date`
- **Description**: When the request was approved
- **Required**: No

### Request Details

#### Unavailability Type (Select)
- **Property ID**: `dnv[`
- **Type**: `select`
- **Description**: Type of leave
- **Options**:
  - `Personal Time` (default)
  - `Health / Illness` (pink)
  - `Family / Emergency` (green)
  - `Travel / Vacation` (blue)
  - `Other` (brown)

#### Additional Context (Rich Text)
- **Property ID**: `e~iM`
- **Type**: `rich_text`
- **Description**: Reason for the leave request
- **Required**: No

#### Attachment (optional) (Files)
- **Property ID**: `JSLU`
- **Type**: `files`
- **Description**: Supporting documents for the request (medical certificates, etc.)
- **Required**: No

### Approval & Workflow

#### Status
- **Property ID**: `PcYS`
- **Type**: `status`
- **Description**: Approval status of the leave request
- **Groups**:
  - **To-do**: (empty)
  - **In progress**:
    - `New` (yellow) - Newly submitted request
  - **Complete**:
    - `Acknowledged` (green) - Request approved and acknowledged
    - `Not Applicable` (red) - Request rejected or not applicable
    - `Withdrawn` (gray) - Request withdrawn by contractor

#### Reviewed By (Relation)
- **Property ID**: `oM}>`
- **Type**: `relation`
- **Description**: Person who approved this request
- **Related Database**: `ed2b9224-97d9-4dff-97f9-82598b61f65d` (same as Contractor)
- **Relation Type**: Single property

#### Created by
- **Property ID**: `cqMO`
- **Type**: `created_by`
- **Description**: Who created this request
- **Auto-populated**: Yes

### Calculated Fields

#### Total Days (Formula)
- **Property ID**: `Qu{?`
- **Type**: `formula`
- **Description**: Calculates inclusive day count for the leave period
- **Formula Logic**:
  ```
  ifs(
    /* No start date -> can't calculate, return 0 */
    empty(Start Date), 0,
    /* End date missing -> treat as single-day leave */
    empty(End Date), 1,
    /* End date before start date -> invalid range, return 0 */
    End Date < Start Date, 0,
    /* Inclusive day count: difference in days + 1 */
    dateBetween(End Date, Start Date, "days") + 1
  )
  ```
- **Returns**: Number

#### Auto Name (Formula)
- **Property ID**: `X\Cs`
- **Type**: `formula`
- **Description**: Generates standardized leave request identifier
- **Format**: `LVR-YYYY-{Discord}-{ObfuscatedID}`
- **Example**: `LVR-2025-johndoe-A3K9`
- **Formula**: Uses character obfuscation for ID privacy

### Rollup Fields

#### Discord (Rollup)
- **Property ID**: `F\`qt`
- **Type**: `rollup`
- **Description**: Discord username from contractor record
- **Rollup From**: `Contractor` relation
- **Rollup Property**: `Discord` (`l\`p^`)
- **Function**: `show_original`

#### Person (Rollup)
- **Property ID**: `n=\`E`
- **Type**: `rollup`
- **Description**: Person/employee record from contractor
- **Rollup From**: `Contractor` relation
- **Rollup Property**: `Person` (`R\[I`)
- **Function**: `show_original`

#### Team Email (Rollup)
- **Property ID**: `QuMu`
- **Type**: `rollup`
- **Description**: Team email address for notifications
- **Rollup From**: `Contractor` relation
- **Rollup Property**: `Team Email` (`.|oa`)
- **Function**: `show_original`

## Workflow

### 1. Request Submission
- Contractor creates new leave request
- Sets Start Date, End Date (optional), Unavailability Type
- Adds Additional Context and Attachments if needed
- Status automatically set to `New`

### 2. Review Process
- Admin/Manager reviews request
- Sets `Reviewed By` to their user
- Updates Status to:
  - `Acknowledged` - Approved
  - `Not Applicable` - Rejected
  - `Withdrawn` - Cancelled by contractor

### 3. Notification
- Team Email (rollup) used for automated notifications
- Discord (rollup) used for Discord notifications
- Date Approved recorded when status changes to Acknowledged

## Integration Notes

### Fortress API Integration
- **Webhook URL**: `https://local.arcline.app/webhooks/notion/unavailability`
- **Purpose**: Sync leave requests to internal HR systems
- **Fields Synced**: Contractor, Start/End dates, Status, Type

### Data Validation
- Total Days must be >= 0
- End Date must be >= Start Date (or null)
- Status transitions follow workflow rules

## Related Databases
- **Contractors**: `ed2b9224-97d9-4dff-97f9-82598b61f65d` (for Contractor and Reviewed By relations)

## Business Rules

1. **Single-Day Leave**: If End Date is empty, Total Days = 1
2. **Invalid Ranges**: If End Date < Start Date, Total Days = 0
3. **Auto Name Format**: `LVR-{Year}-{Discord}-{ObfuscatedID}` for tracking
4. **Approval Required**: Status must progress from New → Acknowledged/Not Applicable
5. **Withdrawal**: Contractors can withdraw requests at any time (Status → Withdrawn)

## Last Updated
- **Created**: 2025-12-17
- **Last Modified**: 2026-01-14
- **Schema Version**: 1.0
