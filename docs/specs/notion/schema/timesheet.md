# Project Updates Database Schema

## Overview

- **Database ID**: `2c664b29-b84c-8048-b7e2-000bb8278044`
- **Title**: Project Updates
- **Created**: 2025-12-11
- **Last Edited**: 2026-01-15
- **Icon**: Bullseye/Target (gray)
- **URL**: https://www.notion.so/2c664b29b84c8048b7e2000bb8278044

## Purpose

The Project Updates database is for contractors to record key activities, deliverables, or updates for project coordination and billing verification. It supports activity tracking, task categorization, feedback workflows, and deliverable documentation.

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `(auto) Entry` | Title | `title` | Auto-generated title combining project, contractor discord, date, and approximate effort (e.g., "kafi :: cor3.co :: Dec 15 – 6h") |
| `Date` | Date | `joDR` | The date when the work was performed |
| `Appx. effort` | Number | `RUz\` | Approximate effort in hours (numeric format) |
| `Focus areas` | Select | `ZA~q` | Category of work performed |
| `Key deliverables` | Rich Text | `SJfi` | Detailed description of key deliverables and work completed, often including links to tickets or design files |
| `Status` | Status | `QeGl` | Current review status of the project update entry |

### Focus Areas Options

- Development (blue)
- Design (orange)
- Meeting (green)
- Documentation (purple)
- Research (pink)
- Planning (yellow)
- Other (brown)

### Status Options

The status property uses grouped statuses:

**To-do Group** (gray):
- Draft

**In Progress Group** (blue):
- Pending Feedback

**Complete Group** (green):
- Reviewed
- Completed

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Contractor` | Relation | `Z^dj` | `9d468753-ebb4-4977-a8dc-156428398a6b` | Single relation to the contractor who performed the work |
| `Project` | Relation | `cjkf` | `0ddadba5-bbf2-440c-a286-9f607eca88db` | Single relation to the project the work was performed for |
| `Task Order` | Relation | `Dq}d` | N/A | Relation to task orders (if applicable) |

### Rollup Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Discord` | Rollup | `Wu_R` | Rolls up the Discord username from the related Contractor |

### People Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Review` | People | `nWkE` | People assigned to review/approve the timesheet entry |
| `Created by` | Created By | `nlOG` | User who created the timesheet entry |

### Formula Properties

| Property | Type | ID | Formula Description |
|----------|------|-----|---------------------|
| `Month` | Formula | `t>_\|` | Extracts year-month from the Date field (format: YYYY-MM, e.g., "2025-12") |
| `Auto Name` | Formula | `yMJL` | Generates the title combining project code, discord username, formatted date, and hours |

### System Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `ID` | Unique ID | `_RTD` | Auto-incrementing unique identifier (no prefix) |

## Formula Details

### Month Formula
```javascript
if(empty({{Date}}), "",
  formatDate({{Date}}, "YYYY-MM")
)
```
Returns the year and month from the Date field, or empty string if no date is set.

### Auto Name Formula
```javascript
let (
    discord, {{Discord}},
    person, {{Created by}},
    project, {{Project}}.first().{{project_code}}.lower(),

    project + " :: " + style(discord, "blue")
        + " :: "
        + formatDate({{Date}}, "MMM D")
        + if(empty({{Appx. effort}}, "", " – " + {{Appx. effort}} + "h")
)
```
Creates a formatted title with project code, contractor's discord username (in blue), formatted date, and approximate effort hours.

## Sample Data Structure

```json
{
  "object": "page",
  "properties": {
    "Status": {
      "type": "status",
      "status": {
        "name": "Pending Feedback",
        "color": "blue"
      }
    },
    "Appx. effort": {
      "type": "number",
      "number": 6
    },
    "Key deliverables": {
      "type": "rich_text",
      "rich_text": [{
        "text": {
          "content": "- Setup Airflow port-forward - 1h\n- Support MGM release - 1h\n- KBI-165: [Data Application] BI Web Refactor - 4h"
        }
      }]
    },
    "Discord": {
      "type": "rollup",
      "rollup": {
        "array": [{
          "rich_text": [{
            "text": {
              "content": "cor3.co"
            }
          }]
        }]
      }
    },
    "Focus areas": {
      "type": "select",
      "select": {
        "name": "Development",
        "color": "blue"
      }
    },
    "Date": {
      "type": "date",
      "date": {
        "start": "2025-12-15"
      }
    },
    "Month": {
      "type": "formula",
      "formula": {
        "string": "2025-12"
      }
    },
    "Auto Name": {
      "type": "formula",
      "formula": {
        "string": "kafi :: cor3.co :: Dec 15 – 6h"
      }
    }
  }
}
```

## Workflow

1. **Creation**: Contractor creates a project update entry with date, approximate effort, focus area, project, and key deliverables
2. **Draft**: Entry starts in "Draft" status
3. **Submission**: Entry is moved to "Pending Feedback" status
4. **Review**: Assigned reviewers validate the key deliverables and effort
5. **Review Complete**: Reviewers complete review (status: "Reviewed")
6. **Completion**: Entry is marked as "Completed" after processing

## Integration Notes

- The `Discord` rollup property provides quick access to contractor identification
- The `Month` formula enables easy filtering and grouping by month for payroll processing
- The auto-generated title makes entries easily scannable in list views
- Focus areas categorization allows for reporting on different types of work
- Key deliverables supports rich text with links to tickets, designs, and documentation
- Webhook integration: `https://local.arcline.app/webhooks/notion/timesheet`

## Related Databases

- **Contractor Database** (`9d468753-ebb4-4977-a8dc-156428398a6b`): Contains contractor information including Discord usernames
- **Project Database** (`0ddadba5-bbf2-440c-a286-9f607eca88db`): Contains project information and codes

## Notes

- **Database Renamed**: This database was previously named "Timesheet" with ID `2c664b29-b84c-8089-b304-e9c5b5c70ac3`. It has been renamed to "Project Updates" with new ID `2c664b29-b84c-8048-b7e2-000bb8278044` (January 2026)
- **Property Changes**: `Hours` → `Appx. effort`, `Proof of Works` → `Key deliverables`, `Task Type` → `Focus areas`
- **Status Changes**: `Pending Approval` → `Pending Feedback`, `Approved` → `Reviewed`
- Second database ID `2b964b29-b84c-801c-accb-dc8ca1e38a5f` was not accessible (404 error - either not found or not shared with the integration)
