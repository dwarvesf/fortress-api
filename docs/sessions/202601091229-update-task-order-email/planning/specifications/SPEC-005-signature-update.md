# SPEC-005: Signature Update

## Overview
Update email signature from "Team Dwarves, People Operations" to "Han Ngo, CTO & Managing Director".

## Implementation Locations

### Location 1: Template Composition
**File**: `pkg/service/googlemail/utils.go`  
**Function**: `composeTaskOrderConfirmationContent`  
**Lines**: ~130-138

### Location 2: Raw Email Generation
**File**: `pkg/service/notion/task_order_log.go`  
**Function**: `loadTaskOrderSignature`  
**Lines**: ~1821-1830

## Changes Required

### Current State
```go
"signatureName": func() string {
    return "Team Dwarves"
},
"signatureTitle": func() string {
    return "People Operations"
},
"signatureNameSuffix": func() string {
    return "" // No dot for task order emails
},
```

### New State
```go
"signatureName": func() string {
    return "Han Ngo"
},
"signatureTitle": func() string {
    return "CTO & Managing Director"
},
"signatureNameSuffix": func() string {
    return "" // No dot for task order emails
},
```

## Template Function Specifications

### signatureName
- **Returns**: `string`
- **Old Value**: `"Team Dwarves"`
- **New Value**: `"Han Ngo"`
- **Usage**: Displayed as primary name in signature block

### signatureTitle
- **Returns**: `string`
- **Old Value**: `"People Operations"`
- **New Value**: `"CTO & Managing Director"`
- **Usage**: Displayed as title/role under name

### signatureNameSuffix
- **Returns**: `string`
- **Value**: `""` (empty string - NO CHANGE)
- **Purpose**: Controls dot/period after name (none for task order emails)
- **Note**: Other email types may use "." suffix

## Signature Template Structure

### Template File
**File**: `pkg/templates/signature.tpl`  
**Changes**: NONE - structure unchanged

### Template Usage
```html
{{if (ne signatureName "")}}
<div dir=3D"ltr" class=3D"gmail_signature">
    <!-- Signature content -->
    <h4>{{signatureName}}{{signatureNameSuffix}}</h4>
    <h6>{{signatureTitle}}</h6>
    <!-- Contact info, logo, social links -->
</div>
{{end}}
```

## Consistency Requirements

### Both Locations Must Match
1. `composeTaskOrderConfirmationContent` in utils.go
2. `loadTaskOrderSignature` in task_order_log.go

**Why Two Locations**:
- `utils.go`: Used for standard email sending via cronjob
- `task_order_log.go`: Used for raw email generation (Order page body, webhooks)

### Verification
After implementation, verify both functions return identical signature HTML.

## Other Signature Usages

### DO NOT CHANGE
These files use signature.tpl for OTHER email types:

**File**: `pkg/service/googlemail/utils.go`  
**Function**: `composeMailContent`  
**Lines**: ~39-63

This function handles signatures for:
- `accountingUser` → "Eddie Ng, Accountant"
- `teamEmail` → "Dwarves Foundation, Team Dwarves"
- `spawnEmail` → "Team Dwarves, Hiring"

**Action**: Leave unchanged - only modify task order email signatures

## Visual Output

### Before
```
Team Dwarves
People Operations
```

### After
```
Han Ngo
CTO & Managing Director
```

### Full Signature Block
```
Han Ngo
CTO & Managing Director

WA: +1 (818) 408 6969 | W: dwarves.foundation
A: 131 Continental Drive, Suite 305, Newark, DE 19713, US

[Social media icons: GitHub, Dribbble, LinkedIn, Facebook]
[Dwarves Foundation logo]
```

## Testing Requirements

### Unit Tests
1. Test `signatureName()` returns "Han Ngo"
2. Test `signatureTitle()` returns "CTO & Managing Director"
3. Test `signatureNameSuffix()` returns ""

### Integration Tests
1. Render full email and verify signature
2. Check both code paths (utils.go and task_order_log.go)
3. Verify HTML structure unchanged

### Visual Testing
1. Send test email
2. Verify signature displays correctly in Gmail
3. Check mobile rendering
4. Verify no formatting issues

## Rollback

### Quick Rollback
```go
// Revert to old values
"signatureName": func() string {
    return "Team Dwarves"
},
"signatureTitle": func() string {
    return "People Operations"
},
```

### Impact
- Signature reverts to old
- Email template and content remain updated
- Can revert independently of template changes

## Implementation Notes

### Search Pattern
To find all instances to update:
```bash
grep -r "signatureName.*Team Dwarves" pkg/service/
```

Expected results:
- 2 matches in task order email functions ✓
- Additional matches in other email types (ignore)

### Validation
After changes, verify:
```bash
grep -A2 "signatureName.*Han Ngo" pkg/service/
```

Should return 2 matches.
