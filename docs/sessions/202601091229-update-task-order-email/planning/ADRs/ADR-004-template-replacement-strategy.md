# ADR-004: Template Replacement Strategy

## Status
Accepted

## Context
The task order confirmation email requires significant content changes:
- Subject line: "Monthly Task Order - {Month}" → "Quick update for {Month} – Invoice reminder & client milestones"
- Body: Complete restructuring from work order confirmation to invoice reminder + milestones
- Tone: Formal acknowledgment request → Friendly encouragement
- Signature: "Team Dwarves, People Operations" → "Han Ngo, CTO & Managing Director"

### Current Template Structure
File: `pkg/templates/taskOrderConfirmation.tpl`
- MIME multipart/mixed format
- HTML content with quoted-printable encoding
- Template functions for date formatting and contractor name
- Shared signature.tpl template

### Extent of Changes
Analyzing the approved plan, the changes affect:
- Subject line (1 line)
- Email body content (ALL paragraphs replaced)
- Template function requirements (3 functions removed, 1 added)
- Signature details (name and title)

Only preservation requirements:
- MIME format structure
- From address: "Spawn @ Dwarves LLC" <spawn@d.foundation>
- HTML encoding and structure
- Use of signature.tpl template

## Decision
**We will perform a complete template file replacement rather than incremental edits.**

The old template content will be replaced entirely with the new structure in a single commit, with the old template preserved in git history for reference and potential rollback.

### Implementation Approach
1. Create complete new template content following approved design
2. Replace entire file content in single edit
3. Git commit contains full before/after diff
4. Old template remains in git history
5. Rollback possible via `git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl`

## Rationale

### Alternatives Considered

#### Option 1: Incremental Template Edits
Edit template section by section:
- Update subject line
- Replace greeting paragraph
- Replace body paragraphs
- Update closing

**Rejected** because:
- More complex to review (multiple small changes)
- Harder to test intermediate states
- Doesn't provide clear before/after comparison
- Increases risk of missed changes
- More time-consuming to implement

#### Option 2: Conditional Rendering with Feature Flag
Add template logic to render old or new content based on flag:
```html
{{if .UseNewFormat}}
    <!-- New content -->
{{else}}
    <!-- Old content -->
{{end}}
```

**Rejected** because:
- Significantly increases template complexity
- Doubles maintenance burden (two templates in one)
- Flag management adds operational overhead
- No business requirement for gradual rollout
- Template becomes harder to read and maintain
- Temporary code that must be removed later

#### Option 3: Versioned Templates
Create `taskOrderConfirmation_v2.tpl` alongside old template:

**Rejected** because:
- Route/handler logic must choose which template to use
- More complex file structure
- Old template remains in codebase after migration
- Cleanup debt (must remove old template later)
- No clear benefit over git history

#### Option 4: New Template File with Migration
Create new file, update handler, delete old file across multiple commits:

**Rejected** because:
- Unnecessarily complex commit history
- Harder to rollback (multiple commits)
- Risk of coordination issues between commits
- No benefit over direct replacement

### Why Complete Replacement is Best

1. **Clear Before/After**: Single commit shows full diff of old vs new template
2. **Easy Review**: Reviewers can see entire template change at once
3. **Simple Rollback**: Single file revert restores old template
4. **Clean Git History**: One commit for template replacement
5. **No Technical Debt**: No temporary code or cleanup required
6. **Easier Testing**: Test complete new template, not hybrid states
7. **Reduced Complexity**: Template remains simple and readable

## Consequences

### Positive
1. **Simple Implementation**: Single file edit, no complex logic
2. **Easy Code Review**: Reviewers see full template change
3. **Clean Rollback**: One-command revert if issues found
4. **Testable**: Can fully test new template in isolation
5. **No Maintenance Debt**: No feature flags or versioning to remove later
6. **Git History Preservation**: Old template preserved automatically
7. **Clear Deployment**: Single commit to deploy or revert

### Negative
1. **No Gradual Rollout**: All contractors get new email simultaneously
2. **Large Diff**: Full file replacement creates large PR diff
3. **Testing Scope**: Must fully test new template before deployment
4. **No A/B Testing**: Can't compare old vs new email effectiveness

### Mitigation Strategies

#### 1. Comprehensive Testing
Before deployment:
- Manual email tests to personal address
- Visual inspection of rendered HTML
- Test with multiple contractor scenarios
- Verify MIME format preservation
- Test email client rendering (Gmail, Outlook, etc.)

#### 2. Staged Rollout (Optional)
If business requires gradual rollout:
- Deploy to staging environment first
- Send test emails to operations team
- Use `test_email` parameter for controlled testing
- Monitor first batch before full deployment

#### 3. Quick Rollback Capability
Document rollback procedure:
```bash
# Rollback template only (keeps other changes)
git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl
git commit -m "chore: rollback task order template to previous version"
git push

# Full rollback (all changes)
git revert <commit-hash>
git push
```

#### 4. Clear Documentation
- Comment at top of new template noting replacement date
- Link to this ADR in commit message
- Update planning docs with deployment date
- Document rollback procedure in runbook

#### 5. Backup Old Template
Save old template content in documentation for easy reference:
```
docs/sessions/202601091229-update-task-order-email/planning/
├── old-template-backup.tpl  # Copy of old template for reference
└── new-template-preview.html  # Rendered sample of new email
```

## Implementation Notes

### Template File Structure
```
pkg/templates/
├── taskOrderConfirmation.tpl  # REPLACE entirely
└── signature.tpl              # Update template functions (signatureName, signatureTitle)
```

### Preservation Requirements
Must preserve from old template:
- MIME structure: `Mime-Version: 1.0`, `Content-Type: multipart/mixed`, boundaries
- From address: `"Spawn @ Dwarves LLC" <spawn@d.foundation>`
- To field: `{{.TeamEmail}}`
- HTML encoding: `charset="UTF-8"`, `Content-Transfer-Encoding: quoted-printable`
- Signature inclusion: `{{ template "signature.tpl" }}`
- Quoted-printable escaping: `=3D` for `=`, etc.

### New Template Requirements
Must implement from approved plan:
- Subject: "Quick update for {{formattedMonth}} – Invoice reminder & client milestones"
- Greeting: "Hi {{contractorLastName}},"
- Invoice section with {{invoiceDueDay}}
- Milestones section with range loop: `{{range .Milestones}}`
- Encouragement and support paragraphs
- Signature with updated name/title

### Template Function Changes
File: `pkg/service/googlemail/utils.go`

**Remove** (no longer used):
- `periodEndDay` - Calculated last day of month
- `monthName` - Extracted month name
- `year` - Extracted year

**Add**:
- `invoiceDueDay` - Returns `data.InvoiceDueDay` ("10th" or "25th")

**Keep** (still used):
- `formattedMonth` - Formats "2006-01" to "January 2006"
- `contractorLastName` - Extracts last name from full name

**Update** (new values):
- `signatureName` - "Team Dwarves" → "Han Ngo"
- `signatureTitle` - "People Operations" → "CTO & Managing Director"
- `signatureNameSuffix` - "" (unchanged)

## Testing Strategy

### Pre-Deployment Testing
1. **Unit Tests**: Template rendering with mock data
2. **Manual Testing**: Send to personal email using test_email parameter
3. **Visual Inspection**: Verify HTML rendering in multiple email clients
4. **MIME Validation**: Ensure email structure is valid
5. **Function Testing**: Verify all template functions work correctly

### Post-Deployment Monitoring
1. **Email Delivery Rate**: Monitor for any decrease
2. **Error Logs**: Check for template rendering errors
3. **User Feedback**: Monitor support channels for contractor questions
4. **Bounce Rate**: Verify no increase in bounced emails

### Rollback Testing
Before deployment, test rollback procedure:
1. Deploy new template to staging
2. Test rollback command
3. Verify old template works after rollback
4. Document timing (how fast can we rollback)

## Deployment Plan

### Deployment Steps
1. Merge PR with template replacement
2. Deploy to staging environment
3. Send test emails to operations team
4. Get approval from stakeholder (Han Ngo)
5. Deploy to production
6. Monitor first batch of emails
7. Verify delivery and rendering

### Deployment Timing
- Deploy outside of peak email sending hours
- Avoid month-end when task orders are sent
- Allow time for rollback if needed

### Success Criteria
- Email delivery rate >95%
- Zero template rendering errors
- Stakeholder approval of email content
- No increase in contractor support requests

## Rollback Scenarios

### Scenario 1: Template Rendering Error
**Trigger**: Template execution fails
**Action**: Immediate rollback
**Command**: `git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl`

### Scenario 2: Content Issues
**Trigger**: Stakeholder feedback on content
**Action**: Quick template edit or rollback
**Decision**: Based on severity (edit for minor, rollback for major)

### Scenario 3: Email Client Rendering Issues
**Trigger**: Email displays incorrectly in some clients
**Action**: Fix MIME/HTML encoding or rollback
**Decision**: If quick fix possible, hotfix; otherwise rollback

### Scenario 4: Contractor Confusion
**Trigger**: Multiple support requests about new format
**Action**: Communication + keep monitoring
**Decision**: Rollback only if confusion is widespread

## Documentation Requirements

### Code Comments
```go
// pkg/templates/taskOrderConfirmation.tpl
// Task Order Confirmation Email Template
// Updated: 2026-01-09 - Changed from work order confirmation to invoice reminder format
// See: docs/sessions/202601091229-update-task-order-email/planning/ADRs/ADR-004-template-replacement-strategy.md
```

### Commit Message
```
feat(email): update task order confirmation email template

Replace work order confirmation template with invoice reminder format:
- Update subject: "Invoice reminder & client milestones"
- Add dynamic invoice due date based on contractor Payday
- Add client milestones section (mock data for Phase 1)
- Update signature from "Team Dwarves" to "Han Ngo"
- Preserve MIME format and email structure

Complete template replacement for clear before/after comparison.
Rollback: git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl

See: docs/sessions/202601091229-update-task-order-email/planning/ADRs/ADR-004-template-replacement-strategy.md
```

### Runbook Entry
Document in operations runbook:
- What changed in email template
- How to rollback if needed
- Expected contractor questions and responses
- Monitoring metrics to watch

## References
- Requirements: `docs/sessions/202601091229-update-task-order-email/requirements/overview.md`
- Approved Plan: `/Users/quang/.claude/plans/glistening-roaming-fiddle.md` (Step 4: Update Email Template)
- Current Template: `pkg/templates/taskOrderConfirmation.tpl`
- Template Functions: `pkg/service/googlemail/utils.go`

## Related Decisions
- SPEC-004: Email Template Structure (detailed template specification)
- SPEC-005: Signature Update (signature template function changes)
- ADR-001: Payday Data Source Selection (invoice due date data)
- ADR-003: Milestone Data Approach (milestones rendering)
