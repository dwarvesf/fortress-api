# Planning Phase Status

## Phase Complete ✅

Date: 2026-01-09  
Status: **COMPLETE - Ready for Phase 2 (Test Case Design)**

## Summary

Planning phase successfully completed for task order confirmation email update. All architecture decisions documented, detailed specifications created, and implementation approach defined.

## Deliverables Created

### Architecture Decision Records (4/4)

1. **ADR-001: Payday Data Source Selection**
   - Decision: Use Service Rate database (not Task Order Log)
   - Rationale: Payday field exists there with proper Contractor relation
   - Pattern: Follow contractor_payables implementation

2. **ADR-002: Default Fallback Strategy**
   - Decision: Default to "10th" for missing/invalid Payday
   - Rationale: Email must always send successfully
   - Impact: Requires monitoring but ensures zero email failures

3. **ADR-003: Milestone Data Approach**
   - Decision: Hardcode mock data with TODO marker (Phase 1)
   - Rationale: Real data source not yet defined
   - Future: Easy to replace when source available

4. **ADR-004: Template Replacement Strategy**
   - Decision: Complete template replacement (not incremental)
   - Rationale: Significant content changes, easier rollback
   - Benefit: Clear diff and simpler code review

### Specifications (5/5)

1. **SPEC-001: Data Model Extension**
   - Add `InvoiceDueDay string` field
   - Add `Milestones []string` field
   - Backward compatible (additive only)

2. **SPEC-002: Payday Fetching Service**
   - `GetContractorPayday(ctx, contractorPageID)` method
   - Query Service Rate with Contractor + Status filters
   - Return (0, nil) for graceful fallbacks
   - Debug logging for monitoring

3. **SPEC-003: Handler Logic Update**
   - Fetch Payday and calculate due date
   - Build mock milestones array
   - Populate new email model fields
   - Graceful error handling

4. **SPEC-004: Email Template Structure**
   - Complete template replacement
   - New content: invoice reminder + milestones
   - MIME format preserved
   - Template functions: formattedMonth, contractorLastName, invoiceDueDay

5. **SPEC-005: Signature Update**
   - Change signatureName: "Team Dwarves" → "Han Ngo"
   - Change signatureTitle: "People Operations" → "CTO & Managing Director"
   - Update in 2 locations (utils.go + task_order_log.go)

## Key Decisions

### Data Flow
```
Handler
  ↓ Call GetContractorPayday()
Service (Notion)
  ↓ Query Service Rate database
Data Model
  ↓ Populate InvoiceDueDay + Milestones
Template Rendering
  ↓ Compose email with new content
Gmail Service
  ↓ Send email
```

### Error Handling Strategy
- **Payday Missing**: Default to "10th", log debug message
- **API Errors**: Log error, use default, continue
- **Email Always Sends**: No blocking on data issues

### Implementation Order
1. Add GetContractorPayday service method
2. Update TaskOrderConfirmationEmail model
3. Update handler to fetch and calculate
4. Update template functions (signature)
5. Update email template (last for easy rollback)

## Dependencies Identified

### External Services
- Notion API (Service Rate database queries)
- Gmail API (existing, no changes)

### Configuration Requirements
- `config.Notion.Databases.ServiceRate` must be configured
- No new config needed

### Database Schema
- No changes needed (all fields exist)
- Service Rate: Contractor (Relation), Status (Status), Payday (Select)

## Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Service Rate data missing | Emails show wrong due date | Medium | Default to 10th, monitor coverage |
| Payday API timeout | Email send delays | Low | Add timeout, use fallback |
| Template rendering errors | Email not sent | Low | Unit tests, staged rollout |
| Signature inconsistency | Mixed branding | Low | Update both locations, verify |

## Technical Constraints

### Performance
- Target: <100ms additional latency per contractor
- Payday fetch adds one API call per contractor
- Acceptable for monthly cronjob

### Compatibility
- Maintain existing endpoint `/cronjobs/send-task-order-confirmation`
- Preserve query parameters (month, discord, test_email)
- No breaking changes to API contract

### Rollback
- Template-only rollback: <5 minutes
- Full rollback: <15 minutes (git revert)
- Independent signature rollback possible

## Assumptions

1. Service Rate database is configured and accessible
2. >90% of contractors have Payday configured
3. Gmail API quota sufficient for email volume
4. Notion API stability meets SLA
5. No concurrent updates to template system

## Open Questions

✅ All questions resolved during planning:
- Payday source confirmed (Service Rate, not Task Order Log)
- Default fallback agreed (10th for Payday 1 or missing)
- Milestone approach finalized (mock data with TODO)
- Template strategy decided (complete replacement)

## Next Phase

### Phase 2: Test Case Design

**Status**: READY TO PROCEED

**Scope**:
- Unit test specifications for service methods
- Integration test scenarios for handler
- Template rendering test cases
- End-to-end email flow testing

**Skip Criteria**: None - tests required for data fetching and email content changes

**Expected Duration**: 1-2 hours

### Phase 3: Task Breakdown

After test case design, break down into implementation tasks:
- Task 1: Add service methods
- Task 2: Update model
- Task 3: Update handler
- Task 4: Update templates
- Task 5: Add tests
- Task 6: Integration testing

## Files Modified (Planned)

1. `pkg/model/email.go` - Add fields
2. `pkg/service/notion/task_order_log.go` - Add methods, update signature
3. `pkg/handler/notion/task_order_log.go` - Update logic
4. `pkg/templates/taskOrderConfirmation.tpl` - Replace content
5. `pkg/service/googlemail/utils.go` - Update functions

**Estimated LOC**: ~150 lines added, ~30 lines modified

## Success Criteria

Planning phase complete when:
- ✅ All ADRs documented
- ✅ All specifications written
- ✅ Dependencies identified
- ✅ Risks assessed
- ✅ Implementation approach clear
- ✅ Ready for test case design

**Status**: ALL CRITERIA MET ✅

---

**Prepared by**: Claude (Project Manager Agent)  
**Reviewed by**: Development Team  
**Approved for**: Test Case Design Phase
