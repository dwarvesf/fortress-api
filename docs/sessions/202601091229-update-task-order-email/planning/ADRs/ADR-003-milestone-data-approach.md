# ADR-003: Milestone Data Approach for Phase 1

## Status
Accepted

## Context
The updated task order confirmation email includes a new "Client Milestones" section to keep contractors informed about upcoming project deliverables. This section should display a bulleted list of milestones relevant to the contractor's current projects.

### Business Requirements
From requirements document:
- Phase 1: Include milestone section in email template
- Phase 2 (Future): Replace with real data from actual data source (TBD)
- Must be structured for easy replacement when real data source is available
- Template must render milestones as bullet list

### Real Data Source Not Yet Defined
The business team has not yet determined:
- Where milestone data will come from (Notion Project database, external API, configuration file)
- What format the data will be in
- How milestones will be associated with contractors
- Update frequency for milestone information

### Phase 1 Timeline Constraint
Email update must be deployed without waiting for milestone data source definition. This is a time-boxed Phase 1 delivery focusing on:
1. Email content and tone update
2. Signature change
3. Dynamic invoice due date
4. Milestone section structure (placeholder data)

## Decision
**For Phase 1, we will hardcode mock milestone data in the handler code with a clear TODO marker for future replacement.**

### Implementation Approach
```go
// In pkg/handler/notion/task_order_log.go, around line 495
// TODO: Replace with real milestone data source when available
// Possible sources: Notion Project database properties, external PM API, or configuration
milestones := []string{
    "Q1 Feature Release: Backend API completion by Jan 20",
    "Code Review Session: Architecture review on Jan 15",
    "Client Demo: Progress showcase scheduled for Jan 25",
}

emailData := &model.TaskOrderConfirmationEmail{
    ContractorName: name,
    TeamEmail:      emailToSend,
    Month:          month,
    Clients:        clients,
    InvoiceDueDay:  invoiceDueDay,
    Milestones:     milestones, // Mock data
}
```

### Mock Data Characteristics
- Generic enough to be plausible for any contractor
- Demonstrates the format and structure
- 3-4 milestone items to show visual layout
- Includes dates to demonstrate time-sensitive nature

## Rationale

### Alternatives Considered

#### Option 1: Skip Milestone Section Entirely in Phase 1
**Rejected** because:
- Email template redesign would be incomplete
- Would require another template update in Phase 2
- Loses opportunity to communicate section purpose to contractors
- Business stakeholder (Han Ngo) wants full content in Phase 1

#### Option 2: Use Notion Project Database Properties
**Rejected** because:
- Not all projects have milestone properties configured
- Milestone data format is inconsistent across projects
- Requires data modeling work (out of scope for Phase 1)
- Would delay email update deployment

#### Option 3: Create New Notion Database for Milestones
**Rejected** because:
- Requires database design and configuration
- Adds maintenance burden for operations team
- Scope creep beyond email update
- No clear ownership for milestone data entry

#### Option 4: Use External API (e.g., Jira, Asana)
**Rejected** because:
- Not all projects use same project management tool
- Requires API integration and authentication
- Performance and reliability concerns
- Significant scope expansion

#### Option 5: Environment Variable Configuration
**Rejected** because:
- Milestones change frequently
- Not scalable (can't configure per contractor)
- Requires deployments to update milestones
- Doesn't solve the data source problem

### Why Hardcoded Mock Data is Best for Phase 1

1. **Unblocks Deployment**: Email update can be deployed immediately
2. **Clear Technical Debt**: TODO comment makes it obvious this is temporary
3. **Easy to Replace**: Replacing array assignment is straightforward
4. **Demonstrates Structure**: Shows contractors what to expect in future
5. **Minimal Risk**: Mock data is generic and not misleading
6. **Fast Implementation**: No additional infrastructure or integration needed

## Consequences

### Positive
1. **Fast Time to Market**: Phase 1 can be completed and deployed quickly
2. **Clear Separation**: Phase 1 (email update) and Phase 2 (milestone integration) are decoupled
3. **Low Risk**: Hardcoded data has no external dependencies or failure points
4. **Easy Replacement**: Future developer can easily find and replace mock data
5. **Template Structure Validated**: Email template can be finalized and tested

### Negative
1. **Static Content**: All contractors see same milestones (not personalized)
2. **Maintenance Debt**: Mock data becomes stale over time
3. **User Confusion Risk**: Contractors might think milestones are specific to their projects
4. **Technical Debt**: Creates known technical debt that must be addressed in Phase 2

### Mitigation Strategies

#### 1. Clear Technical Debt Marking
```go
// TODO: Replace with real milestone data source when available
// Phase 2: Implement one of the following approaches:
//   - Option A: Query Notion Project database for project-specific milestones
//   - Option B: Integrate with external project management API (Jira, Asana, etc.)
//   - Option C: Create dedicated Notion Milestones database with contractor relations
// See: docs/sessions/202601091229-update-task-order-email/planning/ADRs/ADR-003-milestone-data-approach.md
milestones := []string{
    // Generic mock data for Phase 1
    "Q1 Feature Release: Backend API completion by Jan 20",
    "Code Review Session: Architecture review on Jan 15",
    "Client Demo: Progress showcase scheduled for Jan 25",
}
```

#### 2. Generic Mock Data
Use intentionally generic milestones that:
- Don't reference specific projects or clients
- Use placeholder dates in near future
- Clearly illustrate the concept without misleading

#### 3. Phase 2 Planning
Document Phase 2 requirements:
- Identify real milestone data source
- Design data model and integration
- Implement contractor-specific milestone filtering
- Update handler to replace mock data

#### 4. Communication Strategy
- Internal documentation notes this is Phase 1 temporary implementation
- Operations team aware milestones are generic placeholders
- Future email improvements tracked in backlog

## Implementation Notes

### Data Model Extension
```go
// pkg/model/email.go
type TaskOrderConfirmationEmail struct {
    ContractorName string
    TeamEmail      string
    Month          string
    Clients        []TaskOrderClient
    InvoiceDueDay  string   // NEW: "10th" or "25th"
    Milestones     []string // NEW: Array of milestone descriptions
}
```

### Template Rendering
```html
<!-- pkg/templates/taskOrderConfirmation.tpl -->
<p>Here are some key client milestones coming up:</p>
<ul>
    {{range .Milestones}}
    <li>{{.}}</li>
    {{end}}
</ul>
```

### Edge Case: Empty Milestones
Template should handle empty array gracefully:
```html
{{if .Milestones}}
<p>Here are some key client milestones coming up:</p>
<ul>
    {{range .Milestones}}
    <li>{{.}}</li>
    {{end}}
</ul>
{{end}}
```

This ensures Phase 2 can set `Milestones: nil` for contractors with no applicable milestones.

## Phase 2 Considerations

### Real Data Source Requirements
When implementing real milestone data:

1. **Data Association**: Determine how milestones relate to contractors
   - Per project (contractor → project → milestones)
   - Per contractor (direct contractor → milestones relation)
   - Per time period (global milestones filtered by relevance)

2. **Data Freshness**: Define update frequency
   - Real-time query on each email send
   - Cached daily/weekly
   - Pre-computed during project planning

3. **Personalization**: Decide filtering logic
   - Only milestones for contractor's assigned projects
   - Team-wide milestones for all contractors
   - Mix of project-specific and team-wide

4. **Fallback Behavior**: Handle missing data
   - Show generic message if no milestones
   - Omit section entirely if no milestones
   - Use default organizational milestones

### Replacement Strategy
When real data source is ready:

1. Implement service method to fetch milestones
2. Update handler to call service method
3. Remove TODO comment and mock data array
4. Update tests to use real data or mocked service

Example:
```go
// Phase 2 implementation
milestones, err := notionService.GetContractorMilestones(ctx, contractorPageID, month)
if err != nil {
    l.Debug(fmt.Sprintf("failed to fetch milestones for %s: %v", contractorName, err))
    milestones = []string{} // Empty array, template handles gracefully
}
```

## Testing Strategy

### Phase 1 Tests
- Verify mock data renders correctly in email
- Test template with empty milestones array
- Validate HTML structure with 3-4 mock items

### Phase 2 Tests (Future)
- Test real data source integration
- Verify contractor-specific filtering
- Test empty/missing milestone scenarios
- Validate performance with large milestone lists

## Documentation Requirements
- Comment in code with TODO marker
- This ADR documenting decision and Phase 2 approach
- Requirements doc noting Phase 1 vs Phase 2 scope
- Planning summary tracking technical debt

## References
- Requirements: `docs/sessions/202601091229-update-task-order-email/requirements/overview.md` (Client Milestones section)
- Approved Plan: `/Users/quang/.claude/plans/glistening-roaming-fiddle.md` (Step 3: Update Handler Logic)
- Data Model: SPEC-001 (Data Model Extension)
- Template: SPEC-004 (Email Template Structure)

## Related Decisions
- SPEC-001: Data Model Extension (Milestones field definition)
- SPEC-004: Email Template Structure (Milestones rendering)
