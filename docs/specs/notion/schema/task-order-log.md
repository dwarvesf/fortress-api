# Task Order Log Database Schema

## Overview

- **Database ID**: `2b964b29-b84c-801c-accb-dc8ca1e38a5f`
- **Title**: Task Order Log
- **Created**: 2025-11-28
- **Last Edited**: 2025-12-23
- **Icon**: Checklist (gray)
- **URL**: https://www.notion.so/2b964b29b84c801caccbdc8ca1e38a5f

## Purpose

The Task Order Log database manages work orders and associated timesheet line items for contractor deployments. It provides a hierarchical structure where Orders contain multiple Timesheet line items, enabling detailed tracking of work performed and hours logged.

**Description**:
```
company ~> contractor

1. order log
2. confirm service fee (mail)
3. artifact: invoice from contract → company
```

## Properties

### Core Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Name` | Title | `title` | Entry name (auto-generated or manual) |
| `Type` | Select | `pXKA` | Entry type: Order or Timesheet |
| `Date` | Date | `Ri:O` | Date of the work or order |
| `Status` | Select | `LnUY` | Current status of the entry |
| `Line Item Hours` | Number | `NZcl` | Hours for individual timesheet line items |
| `Proof of Works` | Rich Text | `hlty` | Detailed description of work completed |

### Type Options

| Type | Color | Description |
|------|-------|-------------|
| Order | Blue | Parent entry representing a work order for a contractor |
| Timesheet | Green | Child entry representing individual timesheet line items |

### Status Options

| Status | Color | Description |
|--------|-------|-------------|
| Draft | Default (gray) | Initial draft state |
| Pending Approval | Yellow | Submitted for approval |
| Approved | Green | Approved by reviewer |
| Completed | Purple | Finalized and processed |

### Relations

| Property | Type | ID | Related Database | Cardinality | Description |
|----------|------|-----|------------------|-------------|-------------|
| `Deployment` | Relation | `;udS` | `2b864b29-b84c-8079-9568-dc17685f4f33` | Single | Links to contractor deployment |
| `Timesheet` | Relation | `Hz;@` | `2c664b29-b84c-8089-b304-e9c5b5c70ac3` | Multiple | Links to timesheet entries |
| `Sub-item` | Relation | `xFUg` | `2b964b29-b84c-801c-accb-dc8ca1e38a5f` (self) | Multiple | Child line items under an Order |
| `Parent item` | Relation | `{_e{` | `2b964b29-b84c-801c-accb-dc8ca1e38a5f` (self) | Single | Parent Order for Timesheet line items |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Contractor` | Rollup | `q?kW` | From `Deployment` → `Contractor` | Contractor associated with the deployment |
| `Project` | Rollup | `VXv@` | From `Deployment` → `Project` | Project associated with the deployment |
| `Deployment Status` | Rollup | `h>~Q` | From `Deployment` → `Deployment Status` | Current status of the deployment |
| `Subtotal Hours` | Rollup | `u<HB` | From `Sub-item` → `Line Item Hours` (sum) | Total hours from all sub-items |
| `Timesheet Hours (sum)` | Rollup | `Le:i` | From `Timesheet` → Hours (sum) | Total hours from linked timesheet entries |
| `Timesheet Statuses` | Rollup | `gOLD` | From `Timesheet` → Status | Statuses of linked timesheet entries |

### Formula Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Auto Name` | Formula | `:&#124;Xa` | Generates entry name based on type, contractor, and date |
| `Final Hours Worked` | Formula | `;J>Y` | Calculates total hours (Subtotal Hours for Orders, Line Item Hours for Timesheet) |
| `Month` | Formula | `P;xT` | Extracts YYYY-MM from the Date field |
| `Person` | Formula | `gu[o` | Extracts person name from contractor profile |

### System Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `ID` | Unique ID | `tBeu` | Auto-incrementing unique identifier (no prefix) |

## Formula Details

### Auto Name Formula

```javascript
let(
  /* ───── Common variables for line items ───── */
  profile, {{Contractor}},
  discord, profile.first().{{discord_field}},
  monthYear, if(empty({{Date}}), "", formatDate({{Date}}, "Y MMM")),

  if ({{Type}} == "Order",
    "Order " + style(discord, "blue")
      + " :: "
      + if(empty(monthYear), "", monthYear),
    "Time " + discord
      + " :: "
      + formatDate({{Date}}, "MMM D")
      + if(empty({{Line Item Hours}}), "", " – " + {{Line Item Hours}} + "h")
  )
)
```

**Output Examples**:
- **Order**: `Order thanh.pham :: 2025 Nov`
- **Timesheet**: `Time thanh.pham :: Nov 3 – 8h`

### Final Hours Worked Formula

```javascript
if (
  {{Type}} == "Order",
  {{Subtotal Hours}},
  {{Line Item Hours}}
)
```

Returns the appropriate hours based on entry type:
- **Order**: Sum of all sub-item hours
- **Timesheet**: Direct line item hours

### Month Formula

```javascript
if(
  empty({{Date}}),
  empty(),
  /* Show year-month from the date, ignoring time or range end */
  formatDate(
    dateStart({{Date}}),
    "YYYY-MM"
  )
)
```

Returns YYYY-MM format (e.g., "2025-11") from the date field.

## Hierarchical Structure

The database supports a parent-child relationship:

```
Order Entry (Type: Order)
├── Timesheet Line Item 1 (Type: Timesheet)
├── Timesheet Line Item 2 (Type: Timesheet)
└── Timesheet Line Item 3 (Type: Timesheet)
```

- **Order entries** have `Sub-item` relations but no `Parent item`
- **Timesheet entries** have a `Parent item` relation but no `Sub-item`
- Hours roll up from Timesheet entries to their parent Order via `Subtotal Hours`

## Sample Data Structure

### Order Entry Example

```json
{
  "object": "page",
  "properties": {
    "Type": {
      "type": "select",
      "select": {
        "name": "Order",
        "color": "blue"
      }
    },
    "Status": {
      "type": "select",
      "select": {
        "name": "Completed",
        "color": "purple"
      }
    },
    "Date": {
      "type": "date",
      "date": {
        "start": "2025-11-05"
      }
    },
    "Month": {
      "type": "formula",
      "formula": {
        "string": "2025-11"
      }
    },
    "Subtotal Hours": {
      "type": "rollup",
      "rollup": {
        "number": 24,
        "function": "sum"
      }
    },
    "Final Hours Worked": {
      "type": "formula",
      "formula": {
        "number": 24
      }
    },
    "Auto Name": {
      "type": "formula",
      "formula": {
        "string": "Order thanh.pham :: 2025 Nov"
      }
    },
    "Sub-item": {
      "type": "relation",
      "relation": [
        {"id": "2c464b29-b84c-8070-bf53-f6706bffbe71"},
        {"id": "2c464b29-b84c-800a-8e58-e03bef27d44a"}
      ]
    }
  }
}
```

### Timesheet Line Item Example

```json
{
  "object": "page",
  "properties": {
    "Type": {
      "type": "select",
      "select": {
        "name": "Timesheet",
        "color": "green"
      }
    },
    "Status": {
      "type": "select",
      "select": {
        "name": "Pending Approval",
        "color": "yellow"
      }
    },
    "Date": {
      "type": "date",
      "date": {
        "start": "2025-11-03"
      }
    },
    "Line Item Hours": {
      "type": "number",
      "number": 8
    },
    "Final Hours Worked": {
      "type": "formula",
      "formula": {
        "number": 8
      }
    },
    "Auto Name": {
      "type": "formula",
      "formula": {
        "string": "Time thanh.pham :: Nov 3 – 8h"
      }
    },
    "Parent item": {
      "type": "relation",
      "relation": [
        {"id": "2b964b29-b84c-800c-bb63-fe28a4546f23"}
      ]
    },
    "Proof of Works": {
      "type": "rich_text",
      "rich_text": [
        {
          "text": {
            "content": "Implemented feature X and fixed bug Y"
          }
        }
      ]
    }
  }
}
```

## Workflow

### Order Workflow

1. **Order Creation**: Create Order entry with Type=Order, link to Deployment
2. **Line Item Addition**: Create Timesheet entries as Sub-items under the Order
3. **Hours Aggregation**: Subtotal Hours automatically sums all sub-item hours
4. **Status Tracking**: Progress through Draft → Pending Approval → Approved → Completed
5. **Invoice Generation**: Use completed Order as basis for contractor invoice

### Timesheet Line Item Workflow

1. **Creation**: Create Timesheet entry with Type=Timesheet, link to parent Order
2. **Details**: Add Date, Line Item Hours, and Proof of Works
3. **Review**: Submit for approval (Pending Approval status)
4. **Approval**: Reviewer approves (Approved status)
5. **Completion**: Mark as Completed after processing

## Integration Notes

### Hours Calculation Logic

- **Orders**: Display `Subtotal Hours` (sum of all Sub-item hours)
- **Timesheets**: Display `Line Item Hours` (individual entry hours)
- The `Final Hours Worked` formula handles this branching logic automatically

### Deployment Integration

- Each entry links to a Deployment record
- Deployment provides Contractor and Project context via rollups
- Deployment Status indicates if the contractor is currently Active

### Timesheet Database Integration

- Orders can link to Timesheet database entries directly
- Enables cross-referencing between Task Order Log and Timesheet systems
- `Timesheet Hours (sum)` rollup provides alternative hours calculation

### Data Model Pattern

This database implements a **Composite Pattern** where:
- Order entries act as **Composites** (containers)
- Timesheet entries act as **Leaf nodes** (individual items)
- Both share the same interface (properties) but different Type values
- Hours calculations aggregate from leaves to composites

## Use Cases

1. **Monthly Order Tracking**: Create one Order per contractor per month with multiple Timesheet line items
2. **Invoice Preparation**: Query Orders with Status=Completed to generate invoices
3. **Hours Validation**: Compare Subtotal Hours with Timesheet Hours (sum) to verify data consistency
4. **Approval Workflow**: Filter by Status=Pending Approval to review pending work
5. **Project Reporting**: Roll up hours by Project via Deployment relation

## Related Databases

- **Deployment Database** (`2b864b29-b84c-8079-9568-dc17685f4f33`): Contains contractor deployment information
- **Timesheet Database** (`2c664b29-b84c-8089-b304-e9c5b5c70ac3`): Contains detailed timesheet entries
- **Contractor Database** (via Deployment): Contains contractor profiles with Discord usernames
- **Project Database** (via Deployment): Contains project information and codes

## API Integration Considerations

### Querying Orders

```javascript
// Get all completed Orders for a specific month
{
  "filter": {
    "and": [
      {
        "property": "Type",
        "select": {
          "equals": "Order"
        }
      },
      {
        "property": "Status",
        "select": {
          "equals": "Completed"
        }
      },
      {
        "property": "Month",
        "formula": {
          "string": {
            "equals": "2025-11"
          }
        }
      }
    ]
  }
}
```

### Creating Hierarchical Entries

1. Create Order entry first
2. Create Timesheet entries with `Parent item` relation to Order
3. Hours automatically roll up to Order via `Subtotal Hours`

### Hours Validation Query

Retrieve Order with both `Subtotal Hours` and `Timesheet Hours (sum)` to compare and validate data integrity.
