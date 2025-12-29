# Sync Task Order Logs from Timesheet

## Overview

This feature creates Task Order Log entries from approved Timesheet entries on a monthly basis per contractor. It automates the process of generating work orders for contractor invoicing.

## Endpoint

```
POST /api/v1/cronjobs/sync-task-order-logs?month=YYYY-MM
```

### Authentication
- Requires JWT authentication (`conditionalAuthMW`)
- Requires `CronjobExecute` permission (`conditionalPermMW`)

### Query Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `month` | string | Yes | Target month in `YYYY-MM` format (e.g., `2025-12`) |
| `contractor` | string | No | Discord username to filter by specific contractor (e.g., `chinhld`) |

### Response

```json
{
  "data": {
    "month": "2025-12",
    "orders_created": 3,
    "line_items_created": 7,
    "contractors_processed": 3,
    "details": [
      {
        "contractor_discord": "thanh.pham",
        "order_page_id": "2b964b29-b84c-800c-bb63-fe28a4546f23",
        "total_hours": 26,
        "projects": [
          {
            "project_code": "kafi",
            "line_item_page_id": "2c464b29-b84c-8070-bf53-f6706bffbe71",
            "hours": 18,
            "timesheets_aggregated": 3
          },
          {
            "project_code": "mgm",
            "line_item_page_id": "2c464b29-b84c-800a-8e58-e03bef27d44a",
            "hours": 8,
            "timesheets_aggregated": 2
          }
        ]
      }
    ]
  }
}
```

## Data Flow

```
┌─────────────────┐      ┌──────────────────────┐      ┌─────────────────────┐
│   Timesheet DB  │──────│  Sync Task Order     │──────│  Task Order Log DB  │
│  (Source)       │      │  Logs Cronjob        │      │  (Target)           │
└─────────────────┘      └──────────────────────┘      └─────────────────────┘
        │                         │                            │
        │ 1. Query Approved       │                            │
        │    Timesheets           │                            │
        │    (Month=YYYY-MM)      │                            │
        ├─────────────────────────►                            │
        │                         │                            │
        │ 2. Group by Contractor  │                            │
        │    then by Project      │                            │
        │                         │                            │
        │                         │ 3. Per Contractor:         │
        │                         │    Create Order            │
        │                         │    (Type=Order)            │
        │                         ├───────────────────────────►│
        │                         │                            │
        │                         │ 4. Per Project:            │
        │                         │    - Aggregate hours       │
        │                         │    - Collect all PoW       │
        │                         │                            │
        │                    ┌────┴────┐                       │
        │                    │OpenRouter│                      │
        │                    │LLM       │                      │
        │                    │(gpt-5-   │                      │
        │                    │ nano)    │                      │
        │                    └────┬────┘                       │
        │                         │ Summarize to               │
        │                         │ bullet points              │
        │                         │                            │
        │                         │ 5. Create 1 Timesheet      │
        │                         │    per Project             │
        │                         │    (Type=Timesheet)        │
        │                         ├───────────────────────────►│
        │                         │                            │
        │                         │ 6. Link Sub-items          │
        │                         │    to Order                │
        │                         ├───────────────────────────►│
        │                         │                            │
```

## Processing Logic

### Step 1: Query Timesheets

Query the Timesheet database for entries matching:
- **Status**: `Approved`
- **Month**: Matches the `month` parameter (via formula property)

```json
{
  "filter": {
    "and": [
      {
        "property": "Status",
        "status": {
          "equals": "Approved"
        }
      },
      {
        "property": "Month",
        "formula": {
          "string": {
            "equals": "2025-12"
          }
        }
      }
    ]
  }
}
```

### Step 2: Group by Contractor, then by Project

Group timesheet entries in two levels:
1. **Level 1 - Contractor**: Group all timesheets by `Contractor` relation page ID
2. **Level 2 - Project**: Within each contractor, group timesheets by `Project` relation page ID

Each Contractor generates one **Order** entry. Each Project within that contractor generates one **Timesheet sub-item**.

### Step 3: Get Deployment for Each Contractor

For each contractor, query the Deployment Tracker database to find the active deployment:
- Filter by Contractor relation
- Filter by Status = Active (or similar)

This deployment ID is required for the Task Order Log `Deployment` relation.

### Step 4: Create Order Entry

For each contractor group, create one Order entry in Task Order Log:

```json
{
  "parent": {
    "database_id": "2b964b29-b84c-801c-accb-dc8ca1e38a5f"
  },
  "properties": {
    "Type": {
      "select": {
        "name": "Order"
      }
    },
    "Status": {
      "select": {
        "name": "Draft"
      }
    },
    "Date": {
      "date": {
        "start": "2025-12-01"
      }
    },
    "Deployment": {
      "relation": [
        {
          "id": "deployment_page_id"
        }
      ]
    }
  }
}
```

**Note**: The `Name` (title) is auto-generated by Notion formula based on Type, Contractor, and Date.

### Step 5: Create Timesheet Line Items (One per Project)

For each **Project** within a contractor's Order:
1. Collect all timesheets for that project
2. Sum total hours across all timesheets
3. Concatenate all `Proof of Works` texts
4. Call LLM to summarize into major work bullet points
5. Create ONE Timesheet sub-item per project

#### LLM Summarization

Use OpenRouter API with model `openai/gpt-5-nano` to summarize all proof of works for a project into major work bullet points.

**Request:**
```bash
POST https://openrouter.ai/api/v1/chat/completions
Authorization: Bearer <OPENROUTER_API_KEY>
```

```json
{
  "model": "openai/gpt-5-nano",
  "messages": [
    {
      "role": "system",
      "content": "You are a technical writer. Summarize the following work logs into a concise list of major works as bullet points. Focus on key deliverables and outcomes. Combine related items. Keep it professional and brief. Output format: bullet points only, no introduction."
    },
    {
      "role": "user",
      "content": "Summarize these work logs into major works:\n\n--- Entry 1 (Dec 10, 6h) ---\n- Setup Airflow port-forward - 1h\n- Support MGM release - 1h\n- KBI-165: [Data Application] BI Web Refactor - 4h\n\n--- Entry 2 (Dec 15, 8h) ---\n- KBI-165: Continue BI Web Refactor - 6h\n- Code review for team - 2h\n\n--- Entry 3 (Dec 20, 4h) ---\n- KBI-165: BI Web Refactor testing and deployment - 4h"
    }
  ],
  "max_tokens": 200,
  "temperature": 0.3
}
```

**Response:**
```json
{
  "choices": [
    {
      "message": {
        "content": "• Completed BI Web data application refactoring (KBI-165) including development, testing, and deployment\n• Configured Airflow infrastructure setup\n• Supported MGM production release\n• Performed code reviews for team"
      }
    }
  ]
}
```

#### Create Line Item

For each project, create ONE Timesheet sub-item with aggregated data:

```json
{
  "parent": {
    "database_id": "2b964b29-b84c-801c-accb-dc8ca1e38a5f"
  },
  "properties": {
    "Type": {
      "select": {
        "name": "Timesheet"
      }
    },
    "Status": {
      "select": {
        "name": "Approved"
      }
    },
    "Date": {
      "date": {
        "start": "2025-12-15"
      }
    },
    "Line Item Hours": {
      "number": 6
    },
    "Proof of Works": {
      "rich_text": [
        {
          "text": {
            "content": "- Setup Airflow port-forward - 1h\n- Support MGM release - 1h"
          }
        }
      ]
    },
    "Parent item": {
      "relation": [
        {
          "id": "order_page_id"
        }
      ]
    },
    "Timesheet": {
      "relation": [
        {
          "id": "original_timesheet_page_id"
        }
      ]
    }
  }
}
```

### Step 6: Update Order with Sub-items

After creating all line items, update the Order entry to add Sub-item relations:

```json
{
  "properties": {
    "Sub-item": {
      "relation": [
        {"id": "line_item_1_id"},
        {"id": "line_item_2_id"},
        {"id": "line_item_3_id"}
      ]
    }
  }
}
```

## Database References

| Database | ID | Purpose |
|----------|-----|---------|
| Timesheet | `2c664b29-b84c-8089-b304-e9c5b5c70ac3` | Source of approved work entries |
| Task Order Log | `2b964b29-b84c-801c-accb-dc8ca1e38a5f` | Target for Order and Line Items |
| Deployment Tracker | `2b864b29-b84c-8079-9568-dc17685f4f33` | Lookup for Deployment relation |
| Contractor | `9d468753-ebb4-4977-a8dc-156428398a6b` | Referenced via Timesheet.Contractor |

## Property Mappings

### Timesheet → Task Order Log Line Item

| Timesheet Property | Task Order Log Property | Notes |
|--------------------|------------------------|-------|
| `Date` | `Date` | Direct copy |
| `Hours` | `Line Item Hours` | Direct copy |
| `Proof of Works` | `Proof of Works` | Direct copy |
| `Status` | `Status` | Copy as-is (Approved) |
| (via Contractor) | `Parent item` | Relation to Order |
| Page ID | `Timesheet` | Relation to source |

### Order Entry Properties

| Property | Value | Source |
|----------|-------|--------|
| `Type` | `Order` | Static |
| `Status` | `Draft` | Static (initial state) |
| `Date` | First day of month | Computed from `month` param |
| `Deployment` | Deployment page ID | Lookup from Contractor |

## Idempotency

To prevent duplicate entries on re-runs:
1. Before creating an Order, check if one already exists for the same Contractor + Month
2. Query Task Order Log: `Type=Order`, `Month=YYYY-MM`, `Deployment` matches contractor's deployment
3. If exists, skip creation and optionally update line items

## Error Handling

| Error Case | Handling |
|------------|----------|
| Invalid month format | Return 400 Bad Request |
| No timesheets found | Return success with 0 counts |
| Contractor has no deployment | Skip contractor, log warning |
| Notion API error | Return 500 with error details |
| OpenRouter API error | Use original Proof of Works text (no summarization) |
| OpenRouter rate limit | Retry with exponential backoff (max 3 retries) |
| Empty Proof of Works | Skip summarization, use empty string |
| Partial failure | Continue processing, report failures in response |

## Environment Variables

```bash
# Existing
NOTION_SECRET=secret_xxx
NOTION_CONTRACTOR_DB_ID=9d468753-ebb4-4977-a8dc-156428398a6b

# New - Notion
NOTION_TIMESHEET_DB_ID=2c664b29-b84c-8089-b304-e9c5b5c70ac3
NOTION_TASK_ORDER_LOG_DB_ID=2b964b29-b84c-801c-accb-dc8ca1e38a5f
NOTION_DEPLOYMENT_TRACKER_DB_ID=2b864b29-b84c-8079-9568-dc17685f4f33

# New - OpenRouter LLM
OPENROUTER_API_KEY=sk-or-xxx
OPENROUTER_MODEL=openai/gpt-5-nano
```

## Example Result in Task Order Log

After syncing for month `2025-12`, contractor `thanh.pham` worked on 2 projects: `kafi` and `mgm`.

### Order Entry (Parent) - One per Contractor

| Property | Value |
|----------|-------|
| **Name** | `Order thanh.pham :: 2025 Dec` |
| **Type** | Order (blue) |
| **Status** | Draft (gray) |
| **Date** | 2025-12-01 |
| **Month** | 2025-12 |
| **Deployment** | → Deployment: thanh.pham |
| **Contractor** (rollup) | thanh.pham |
| **Sub-item** | → 2 line items (one per project) |
| **Subtotal Hours** (rollup) | 26 |
| **Final Hours Worked** | 26 |

### Timesheet Line Items (Sub-items) - One per Project

**Line Item 1 - Project: kafi** (aggregated from 3 timesheet entries):

| Property | Value |
|----------|-------|
| **Name** | `Time thanh.pham :: Dec 1 – 18h` |
| **Type** | Timesheet (green) |
| **Status** | Approved (green) |
| **Date** | 2025-12-01 (first day of month) |
| **Line Item Hours** | 18 (sum of 6h + 8h + 4h) |
| **Proof of Works** | `• Completed BI Web data application refactoring (KBI-165) including development, testing, and deployment`<br>`• Configured Airflow infrastructure setup`<br>`• Performed code reviews for team` |
| **Parent item** | → Order thanh.pham :: 2025 Dec |
| **Timesheet** | → [kafi timesheet 1], [kafi timesheet 2], [kafi timesheet 3] |

**Line Item 2 - Project: mgm** (aggregated from 2 timesheet entries):

| Property | Value |
|----------|-------|
| **Name** | `Time thanh.pham :: Dec 1 – 8h` |
| **Type** | Timesheet (green) |
| **Status** | Approved (green) |
| **Date** | 2025-12-01 (first day of month) |
| **Line Item Hours** | 8 (sum of 5h + 3h) |
| **Proof of Works** | `• Supported MGM production release and hotfix deployment`<br>`• Resolved critical authentication bug` |
| **Parent item** | → Order thanh.pham :: 2025 Dec |
| **Timesheet** | → [mgm timesheet 1], [mgm timesheet 2] |

### Hierarchy View

```
📋 Order thanh.pham :: 2025 Dec (26h total)
├── ⏱️ Time thanh.pham :: Dec 1 – 18h [kafi] - aggregated from 3 timesheets
│   • Completed BI Web refactoring (KBI-165)
│   • Configured Airflow infrastructure
│   • Performed code reviews
│
└── ⏱️ Time thanh.pham :: Dec 1 – 8h [mgm] - aggregated from 2 timesheets
    • Supported MGM release and hotfix
    • Resolved authentication bug
```

## Example Usage

```bash
# Sync task order logs for all contractors in December 2025
curl -X POST "https://api.fortress.d.foundation/api/v1/cronjobs/sync-task-order-logs?month=2025-12" \
  -H "Authorization: Bearer <token>"

# Sync task order logs for a specific contractor
curl -X POST "https://api.fortress.d.foundation/api/v1/cronjobs/sync-task-order-logs?month=2025-12&contractor=chinhld" \
  -H "Authorization: Bearer <token>"
```

## Related Documents

- [Timesheet Database Schema](./schema/timesheet.md)
- [Task Order Log Database Schema](./schema/task-order-log.md)
