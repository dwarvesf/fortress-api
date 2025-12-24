# Sync Task Order Logs - Requirements

## Overview

Create a cronjob endpoint that automatically generates Task Order Log entries from approved Timesheet entries on a monthly basis per contractor.

## Functional Requirements

### FR-1: Endpoint Definition
- **Endpoint**: `POST /api/v1/cronjobs/sync-task-order-logs`
- **Query Parameter**: `month` (required, format: `YYYY-MM`)
- **Authentication**: JWT with `CronjobExecute` permission

### FR-2: Data Grouping
- Group timesheets by **Contractor** (Level 1)
- Within each contractor, group by **Project** (Level 2)
- One **Order** entry per Contractor per Month
- One **Timesheet sub-item** per Project (aggregated)

### FR-3: Timesheet Filtering
- Only include timesheets with **Status = Approved**
- Filter by **Month** matching the query parameter

### FR-4: Order Entry Creation
- Type: `Order`
- Status: `Draft`
- Date: First day of the month
- Deployment: Linked to contractor's active deployment

### FR-5: Timesheet Sub-item Creation
- Type: `Timesheet`
- Status: `Approved`
- Date: First day of month
- Line Item Hours: Sum of all hours from timesheets for that project
- Proof of Works: LLM-summarized from all timesheets (bullet points)
- Parent item: Linked to the Order entry
- Timesheet: Linked to original timesheet entries

### FR-6: LLM Summarization
- Use OpenRouter API with model `openai/gpt-5-nano`
- Summarize all Proof of Works for a project into major work bullet points
- Focus on key deliverables and outcomes
- Combine related items

### FR-7: Response Format
```json
{
  "data": {
    "month": "2025-12",
    "orders_created": 3,
    "line_items_created": 7,
    "contractors_processed": 3,
    "details": [...]
  }
}
```

## Non-Functional Requirements

### NFR-1: Idempotency
- Check for existing Order before creating
- Skip creation if Order already exists for Contractor + Month

### NFR-2: Error Handling
- OpenRouter API errors: Use original text (no summarization)
- OpenRouter rate limit: Retry with exponential backoff (max 3)
- Empty Proof of Works: Skip summarization
- Partial failures: Continue processing, report in response

### NFR-3: Logging
- DEBUG level logs for all operations
- Log each step of processing

## Database References

| Database | ID | Purpose |
|----------|-----|---------|
| Timesheet | `2c664b29-b84c-8089-b304-e9c5b5c70ac3` | Source |
| Task Order Log | `2b964b29-b84c-801c-accb-dc8ca1e38a5f` | Target |
| Deployment Tracker | `2b864b29-b84c-8079-9568-dc17685f4f33` | Lookup |

## Environment Variables

```bash
NOTION_TIMESHEET_DB_ID=2c664b29-b84c-8089-b304-e9c5b5c70ac3
NOTION_TASK_ORDER_LOG_DB_ID=2b964b29-b84c-801c-accb-dc8ca1e38a5f
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29-b84c-8079-9568-dc17685f4f33
OPENROUTER_API_KEY=sk-or-xxx
OPENROUTER_MODEL=openai/gpt-5-nano
```

## Existing Spec Reference

See: `docs/specs/notion/sync-task-order-logs.md`
