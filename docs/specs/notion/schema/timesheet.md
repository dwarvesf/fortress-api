# Timesheet Database Schema

## Overview

- **Database ID**: `2c664b29-b84c-8089-b304-e9c5b5c70ac3`
- **Title**: Timesheet
- **Created**: 2025-12-11
- **Last Edited**: 2025-12-22
- **Icon**: Clock (gray)
- **URL**: https://www.notion.so/2c664b29b84c8089b304e9c5b5c70ac3

## Purpose

The Timesheet database tracks work hours logged by contractors for various projects. It supports time tracking, task categorization, approval workflows, and proof of work documentation.

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `(auto) Timesheet Entry` | Title | `title` | Auto-generated title combining project, contractor discord, date, and hours (e.g., "kafi :: cor3.co :: Dec 15 – 6h") |
| `Date` | Date | `joDR` | The date when the work was performed |
| `Hours` | Number | `RUz\` | Number of hours worked (numeric format) |
| `Task Type` | Select | `ZA~q` | Category of work performed |
| `Proof of Works` | Rich Text | `SJfi` | Detailed description of work completed, often including links to tickets or design files |
| `Status` | Status | `QeGl` | Current approval status of the timesheet entry |

### Task Type Options

- Development (blue)
- Design (orange)
- Meeting (green)
- Overtime (orange)
- Documentation (purple)
- Research (pink)
- Planning (yellow)
- Training (red)

### Status Options

The status property uses grouped statuses:

**To-do Group** (gray):
- Draft

**In Progress Group** (blue):
- Pending Approval

**Complete Group** (green):
- Approved
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
        + if(empty({{Hours}}, "", " – " + {{Hours}} + "h")
)
```
Creates a formatted title with project code, contractor's discord username (in blue), formatted date, and hours.

## Sample Data Structure

```json
{
  "object": "page",
  "properties": {
    "Status": {
      "type": "status",
      "status": {
        "name": "Pending Approval",
        "color": "blue"
      }
    },
    "Hours": {
      "type": "number",
      "number": 6
    },
    "Proof of Works": {
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
    "Task Type": {
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

1. **Creation**: Contractor creates a timesheet entry with date, hours, task type, project, and proof of work
2. **Draft**: Entry starts in "Draft" status
3. **Submission**: Entry is moved to "Pending Approval" status
4. **Review**: Assigned reviewers validate the proof of work and hours
5. **Approval**: Reviewers approve the entry (status: "Approved")
6. **Completion**: Entry is marked as "Completed" after processing

## Integration Notes

- The `Discord` rollup property provides quick access to contractor identification
- The `Month` formula enables easy filtering and grouping by month for payroll processing
- The auto-generated title makes entries easily scannable in list views
- Task Type categorization allows for reporting on different types of work
- Proof of Works supports rich text with links to tickets, designs, and documentation

## Related Databases

- **Contractor Database** (`9d468753-ebb4-4977-a8dc-156428398a6b`): Contains contractor information including Discord usernames
- **Project Database** (`0ddadba5-bbf2-440c-a286-9f607eca88db`): Contains project information and codes

## Notes

- Second database ID `2b964b29-b84c-801c-accb-dc8ca1e38a5f` was not accessible (404 error - either not found or not shared with the integration)
