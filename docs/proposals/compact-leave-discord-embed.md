# Proposal: Compact Discord Leave Notification Embeds

## Current State

The current Discord embeds for leave notifications contain multiple fields that make the messages quite long:

### Pending Approval
```
📋 New Leave Request - Pending Approval
[View in Notion](link)

Employee: John Doe (john@example.com)
Type: Annual Leave
Dates: 2025-01-15 to 2025-01-20
Reason: Family vacation
```

### Approved
```
✅ Leave Request Approved

Employee: John Doe
Type: Annual Leave  |  Shift: Full Day
Dates: 2025-01-15 to 2025-01-20
```

### Rejected
```
❌ Leave Request Rejected

Employee: John Doe
Type: Annual Leave
Dates: 2025-01-15 to 2025-01-20
Reason: Overlapping project deadline
```

## Proposed Compact Version

### Option 1: Single-Line Description (Most Compact)

#### Pending Approval
```
📋 Leave Request

**John Doe** • Annual Leave • Jan 15-20, 2025 • [Notion](link)
Reason: Family vacation
```

#### Approved
```
✅ Approved

**John Doe** • Annual Leave • Full Day • Jan 15-20, 2025
```

#### Rejected
```
❌ Rejected

**John Doe** • Annual Leave • Jan 15-20, 2025
Reason: Overlapping project deadline
```

### Option 2: Compact Fields (Balanced) - SELECTED

#### Pending Approval
```
📋 Leave Request

Employee: John Doe
Dates: Jan 15-20, 2025
Reason: Family vacation
```

#### Approved
```
✅ Approved

Employee: John Doe
Dates: Full Day • Jan 15-20, 2025
Reason: Family vacation
```

#### Rejected
```
❌ Rejected

Employee: John Doe
Dates: Jan 15-20, 2025
```

### Option 3: Description + Single Field (Most Readable)

#### Pending Approval
```
📋 Leave Request

**John Doe** requests Annual Leave from Jan 15-20, 2025 • [View in Notion](link)

Reason: Family vacation
```

#### Approved
```
✅ Leave Approved

**John Doe** • Annual Leave • Full Day • Jan 15-20, 2025
```

#### Rejected
```
❌ Leave Rejected

**John Doe** • Annual Leave • Jan 15-20, 2025

Reason: Overlapping project deadline
```

## Final Changes (Option 2 - Selected)

1. **Shorter titles**: "📋 Leave Request", "✅ Approved", "❌ Rejected"
2. **Remove Notion link** from all messages
3. **Remove leave type** (Annual Leave, Sick Leave, etc.)
4. **Remove email** from employee display (show name only)
5. **Remove reason** from rejected messages only
6. **Shorter date format**: `Jan 15-20, 2025` instead of `2025-01-15 to 2025-01-20`
7. **Combine shift with dates** for approved: `Full Day • Jan 15-20, 2025`
8. **Keep reason** for pending approval and approved messages

### Benefits:
- **~60% less vertical space** per notification
- **Faster to scan** - only essential info displayed
- **Less notification fatigue** - minimal, clean messages
- **Mobile-friendly** - fits better on mobile Discord
- **Clearer structure** - field names match content ("Dates" instead of "Leave")

## Implementation Notes

### Files to Update:
- `pkg/handler/webhook/notion_leave.go`

### Specific Changes:

1. **Add date formatting helper function**:
   ```go
   func formatShortDateRange(start, end time.Time) string {
       // Returns: "Jan 15-20, 2025" or "Jan 15, 2025 - Feb 2, 2025"
   }
   ```

2. **Update pending approval embeds** (lines ~274, ~308):
   - Title: "📋 Leave Request"
   - Remove description/Notion link
   - Fields: Employee (name only), Dates (short format), Reason

3. **Update approved embeds** (line ~487):
   - Title: "✅ Approved"
   - Fields: Employee (name only), Dates (shift + short format), Reason

4. **Update rejected embeds** (line ~534):
   - Title: "❌ Rejected"
   - Fields: Employee (name only), Dates (short format)
   - Remove Reason field

5. **Update validation failed embeds** (lines ~208, ~227, ~241, ~256, ~351, ~376, ~392):
   - Keep current format but ensure employee name only (no email)
