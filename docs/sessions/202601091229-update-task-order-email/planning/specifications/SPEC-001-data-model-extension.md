# SPEC-001: Data Model Extension

## Overview
Extend the `TaskOrderConfirmationEmail` struct to support invoice due dates and client milestones.

## Current State

**File**: `pkg/model/email.go`

```go
type TaskOrderConfirmationEmail struct {
    ContractorName string
    TeamEmail      string
    Month          string // YYYY-MM format
    Clients        []TaskOrderClient
}
```

## Proposed Changes

### New Fields

```go
type TaskOrderConfirmationEmail struct {
    ContractorName string
    TeamEmail      string
    Month          string // YYYY-MM format
    Clients        []TaskOrderClient
    InvoiceDueDay  string   // NEW: "10th" or "25th"
    Milestones     []string // NEW: Array of milestone descriptions
}
```

### Field Specifications

#### InvoiceDueDay
- **Type**: `string`
- **Format**: Ordinal day number with suffix (e.g., "10th", "25th")
- **Valid Values**: "10th" | "25th"
- **Default**: "10th" (when Payday is missing or invalid)
- **Validation**: None required (populated by handler logic)
- **Usage**: Displayed in email template as invoice due date

#### Milestones
- **Type**: `[]string`
- **Format**: Array of human-readable milestone descriptions
- **Example**: `["Project Alpha: Feature X delivery by Jan 20", "Project Beta: Code review on Jan 15"]`
- **Default**: Empty array `[]` or mock data array
- **Validation**: None required
- **Usage**: Rendered as bullet list in email template

## Backward Compatibility

### Breaking Changes
**None** - New fields are additive only.

### Migration Path
- Existing code that creates `TaskOrderConfirmationEmail` will continue to work
- New fields will be zero-valued (empty string, nil slice) if not populated
- Template must handle empty Milestones array gracefully

## Template Impact

### Template Functions Required
```go
// In composeTaskOrderConfirmationContent
"invoiceDueDay": func() string {
    return data.InvoiceDueDay
}
```

### Template Usage
```html
<p>Your invoice is due by <b>{{invoiceDueDay}}</b>.</p>

{{if .Milestones}}
<p>Upcoming milestones:</p>
<ul>
    {{range .Milestones}}
    <li>{{.}}</li>
    {{end}}
</ul>
{{end}}
```

## Error Handling

### InvoiceDueDay
- Handler must always populate this field
- Never null/empty in production
- Default to "10th" on any error

### Milestones
- Can be empty array (valid state)
- Template handles empty array by not rendering section
- No error if array is nil

## Testing Requirements

### Unit Tests
1. Test struct creation with new fields
2. Test JSON marshaling/unmarshaling
3. Test zero-value behavior

### Integration Tests
1. Test template rendering with populated fields
2. Test template rendering with empty Milestones
3. Test template rendering with default InvoiceDueDay

## Implementation Notes

### Location
- File: `pkg/model/email.go`
- Lines: ~15-20 (add fields to existing struct)

### Dependencies
- No new imports required
- No changes to other structs

### Rollback
- Remove the two new fields
- Update handler to not populate them
- Revert template changes
