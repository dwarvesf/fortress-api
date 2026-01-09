# SPEC-001: Data Model Update

**Status**: Ready for Implementation
**Priority**: High
**Estimated Effort**: 0.5 hours
**Dependencies**: None

## Overview

Add two new fields to the `TaskOrderConfirmationEmail` struct to support dynamic invoice due dates and client milestones display in the email template.

## Objectives

1. Extend the email data model to include invoice due day
2. Add support for milestone data in email content
3. Maintain backward compatibility with existing email sending logic

## Technical Details

### File to Modify

`pkg/model/email.go`

### Current Structure

```go
type TaskOrderConfirmationEmail struct {
    ContractorName      string
    ContractorLastName  string
    PeriodStartDate     string
    PeriodEndDate       string
    ConfirmationLink    string
    // ... other existing fields
}
```

### Required Changes

Add the following fields to `TaskOrderConfirmationEmail`:

```go
type TaskOrderConfirmationEmail struct {
    ContractorName      string
    ContractorLastName  string
    PeriodStartDate     string
    PeriodEndDate       string
    ConfirmationLink    string
    // ... other existing fields

    // New fields for invoice reminder email
    InvoiceDueDay       string   // "10th" or "25th"
    Milestones          []string // Array of client milestone strings
}
```

### Field Specifications

#### InvoiceDueDay
- **Type**: `string`
- **Valid Values**: "10th" or "25th"
- **Default**: "10th" (when Payday data is unavailable)
- **Purpose**: Display personalized invoice due date in email
- **Example**: "10th", "25th"

#### Milestones
- **Type**: `[]string`
- **Valid Values**: Array of non-empty strings
- **Default**: Empty array `[]string{}` (template will handle gracefully)
- **Purpose**: Display client milestones in bullet list format
- **Example**: `[]string{"Launched new feature X", "Completed Q1 deliverables"}`

### Validation Rules

No validation is required at the model level:
- Empty/nil `Milestones` array is acceptable (template handles it)
- Any string value for `InvoiceDueDay` is acceptable (defaults to "10th" in handler)

### Backward Compatibility

This change is **backward compatible**:
- Existing code that creates `TaskOrderConfirmationEmail` will continue to work
- New fields will have Go zero values (empty string, nil slice) if not populated
- Template will be updated to handle missing data gracefully

## Testing Requirements

### Unit Tests

No new unit tests required for this change:
- Struct field additions don't require testing
- Validation logic is handled in handler layer
- Template rendering tests will verify correct usage

### Integration Tests

Verify in handler integration tests:
1. Email struct can be created with new fields populated
2. Email struct can be created without new fields (backward compatibility)
3. Template renders correctly with both populated and empty new fields

## Acceptance Criteria

- [ ] `InvoiceDueDay` field added to `TaskOrderConfirmationEmail` struct
- [ ] `Milestones` field added to `TaskOrderConfirmationEmail` struct
- [ ] Code compiles without errors
- [ ] Existing email sending functionality remains unchanged
- [ ] Handler can populate new fields (verified in SPEC-003)
- [ ] Template can access new fields (verified in SPEC-004)

## Implementation Notes

### Code Location
- **File**: `pkg/model/email.go`
- **Struct**: `TaskOrderConfirmationEmail`
- **Line Number**: Approximately line 10-20 (after existing fields)

### Implementation Steps
1. Open `pkg/model/email.go`
2. Locate `TaskOrderConfirmationEmail` struct
3. Add `InvoiceDueDay string` field with comment
4. Add `Milestones []string` field with comment
5. Save and verify compilation

### Example Implementation

```go
// TaskOrderConfirmationEmail represents the data for task order confirmation email
type TaskOrderConfirmationEmail struct {
    ContractorName      string
    ContractorLastName  string
    PeriodStartDate     string
    PeriodEndDate       string
    ConfirmationLink    string
    // ... other existing fields

    // Invoice reminder fields
    InvoiceDueDay       string   // Due date for invoice: "10th" or "25th"
    Milestones          []string // Client milestones for the period
}
```

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing code | High | All existing code continues to work (zero values) |
| Template rendering errors | Medium | Template updated to handle missing data (SPEC-004) |
| Memory overhead | Low | Minimal overhead (one string, one slice pointer) |

## Documentation Updates

No documentation updates required:
- Internal model change
- Usage documented in handler and template specs
- Swagger annotations unaffected (email sending is internal operation)

## Related Specifications

- **SPEC-002**: Payday Fetching Service (provides data for `InvoiceDueDay`)
- **SPEC-003**: Handler Logic Update (populates new fields)
- **SPEC-004**: Email Template Update (uses new fields)
- **SPEC-005**: Template Function Update (provides template functions)

## Sign-off

- [ ] Implementation completed
- [ ] Code review passed
- [ ] Integration with handler verified
- [ ] Integration with template verified
