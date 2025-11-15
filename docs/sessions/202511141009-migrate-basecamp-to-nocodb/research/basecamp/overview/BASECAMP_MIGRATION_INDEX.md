# Basecamp to NocoDB Migration - Documentation Index

**Date:** 2025-11-14  
**Branch:** feat/migrate-bc-to-nocodb  
**Status:** Analysis Phase - Documentation Complete

---

## Quick Navigation

### For Quick Understanding
**Start here:** [`BASECAMP_WORKFLOWS_SUMMARY.md`](./BASECAMP_WORKFLOWS_SUMMARY.md) (302 lines, 10 min read)
- Quick reference for all 5 workflows
- Key endpoints, triggers, validations
- Hardcoded constants summary
- Environment-specific configuration
- Migration checklist

### For Deep Dive
**Detailed analysis:** [`BASECAMP_WORKFLOWS_ANALYSIS.md`](./BASECAMP_WORKFLOWS_ANALYSIS.md) (1,565 lines, 1-2 hour read)
- Complete end-to-end process for each flow
- All code files involved with line references
- Complete database schema documentation
- Business logic and validation rules
- Error handling strategies
- Detailed migration considerations per flow
- Implementation notes and timeline estimates

---

## Workflow Overview

| Workflow | Complexity | Trigger | Key Dependency | Migration Priority |
|----------|-----------|---------|-----------------|-------------------|
| **1. Invoice** | ⭐⭐⭐⭐ | Todo complete + "Paid" comment | Comment validation | 1st (foundation) |
| **2. Accounting** | ⭐⭐ | Todo created in Accounting | Title parsing | 2nd (simplest) |
| **3. Expense** | ⭐⭐⭐ | Todo create/complete/incomplete | Amount parsing, approver assignment | 3rd |
| **4. Payroll** | ⭐⭐⭐⭐ | Payroll calc + commit | Reimbursement approvals, todo completion | 4th (ties to Expense) |
| **5. On-Leave** | ⭐⭐⭐⭐⭐ | Todo complete + calendar sync | Calendar integration, date parsing | 5th (most complex) |

---

## Key Findings Summary

### 1. Invoice Flow
- **Current:** Basecamp todo → comment validation → invoice status update → accounting transaction
- **Key Pain:** Comment parsing regex, todo completion, multiple async tasks
- **Migration Impact:** Replace comment validation with NocoDB status field; keep currency conversion logic
- **Estimated Effort:** 3-4 weeks

### 2. Accounting Flow
- **Current:** Basecamp todo title → title parsing → accounting transaction
- **Key Pain:** Parent todo list API call for month/year extraction; thousand separator parsing
- **Migration Impact:** Use form fields for month/year; keep category auto-detection
- **Estimated Effort:** 1-2 weeks (simplest)

### 3. Expense Flow
- **Current:** Basecamp todo → validation comment → amount parsing → approver assignment → creation
- **Key Pain:** Complex amount parsing (k/tr/m units); title parsing; assignee lookup
- **Migration Impact:** Form-based entry eliminates parsing; cleaner approver assignment
- **Estimated Effort:** 2-3 weeks

### 4. On-Leave Request Flow
- **Current:** Basecamp todo title parsing → validation → approval → calendar creation (monthly chunks)
- **Key Pain:** Complex date range parsing; calendar integration; multiple approvers
- **Migration Impact:** Form entry simplifies workflow; calendar sync remains critical challenge
- **Estimated Effort:** 3-4 weeks (most complex)

### 5. Payroll Flow
- **Current:** Payroll calculation pulls Ops/Team/Accounting reimbursements from Basecamp todos/comments; commit phase completes todos or posts accounting comments when payroll pays out.
- **Key Pain:** Approval detection via unstructured “approve” comments; heavy reliance on bucket/list IDs; manual todo completion + comment queue.
- **Migration Impact:** Replace comment parsing with structured approval/status fields in NocoDB and abstract completion/notification steps via provider.
- **Estimated Effort:** 3-4 weeks (depends on Expense flow availability).

---

## Critical Migration Points

### Basecamp Service Layer
**Location:** `pkg/service/basecamp/`

Must migrate or replace:
- `Todo.Get()` → NocoDB query
- `Todo.Update()` → NocoDB update
- `Comment.Gets()` → NocoDB query (if storing comments)
- `Recording.Archive()` → NocoDB status change
- `Schedule.CreateScheduleEntry()` → NocoDB calendar or external sync
- `Subscription.Subscribe()` → N/A or NocoDB assignment

### Hardcoded Constants
**Location:** `pkg/service/basecamp/consts/consts.go`

Must remap all IDs:
```
16 Basecamp Project IDs (9403032, 12984857, 15258324, etc.)
9 Basecamp Todo List IDs
20+ Basecamp Person IDs
2 Schedule IDs per environment
```

### Title Parsing Logic
**Must be replaced with:**
- Invoice: Regex → direct number field read
- Accounting: Month/year parsing → form fields
- Expense: Amount parsing → number field + currency dropdown
- On-Leave: Complex date parsing → date picker + type dropdown

---

## Database Impact

### Tables to Migrate
1. `invoices` - 20+ fields, complex relationships
2. `accounting_transactions` - Simple, but core to all flows
3. `expenses` - Links to accounting transactions
4. `on_leave_requests` - JSONB array for assignee_ids

### Data Preservation Required
- Invoice status history + paid_at timestamps
- Accounting metadata (source, ID links)
- Expense basecamp_id mappings (for reconciliation)
- On-leave approver/creator relationships
- All currency conversion rates

### New Tables Possibly Needed
- NocoDB workflows/approvals table
- Calendar events table (if not using external calendar)
- Audit log/activity feed

---

## Implementation Roadmap

### Phase 1: Infrastructure (Week 1-2)
- [ ] Set up NocoDB tables (structure from migrations/)
- [ ] Create NocoDB webhook endpoints
- [ ] Map Basecamp IDs to NocoDB records
- [ ] Set up n8n workflows (if using automation)
- [ ] Database migration scripts

### Phase 2: Accounting Flow (Week 3-4) ⭐ Start Here
- [ ] Simplest flow
- [ ] No calendar integration
- [ ] Good learning experience
- [ ] Unblocks other flows

### Phase 3: Invoice Flow (Week 5-8)
- [ ] Replace Basecamp API calls
- [ ] Form-based entry (optional)
- [ ] Status field validation
- [ ] Keep currency conversion
- [ ] Test with real invoices

### Phase 4: Expense Flow (Week 9-11)
- [ ] Form-based entry
- [ ] Approval workflow
- [ ] Keep amount parsing (or simplify to form field)
- [ ] Link to accounting transactions

### Phase 5: Payroll Flow (Week 12-14)
- [ ] Provider abstraction for reimbursement queries + settlement
- [ ] Structured approval data replacing comment parsing
- [ ] Worker/comment replacement for notifications
- [ ] Dual-read against Basecamp + NocoDB for reimbursement totals
- [ ] Update payroll commit logic to close/notify via provider

### Phase 6: On-Leave Request Flow (Week 15-18)
- [ ] Form-based entry
- [ ] Date picker for ranges
- [ ] Calendar integration (critical path)
- [ ] Approval workflow
- [ ] Ops team notifications

### Phase 7: Testing & Cutover (Week 19-20)
- [ ] Parallel system testing
- [ ] Data reconciliation
- [ ] Disable Basecamp webhooks
- [ ] Monitor for issues
- [ ] Rollback plan ready

---

## Success Criteria

### Per-Flow
- ✓ All webhook endpoints working with NocoDB
- ✓ No data loss (accounting transactions linked)
- ✓ Approval workflows functioning
- ✓ Notifications/comments working
- ✓ Historical data accessible

### Overall
- ✓ Zero invoice payment misses
- ✓ All accounting transactions recorded
- ✓ All expenses tracked to accounting
- ✓ Payroll reimbursements settled with provider-agnostic approvals
- ✓ All on-leave requests in calendar
- ✓ Team productivity unaffected (actually improved via forms)

---

## Key Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Comment validation loss | Invoice approval fails | Use NocoDB status field + audit trail |
| Title parsing bugs | Data entry failures | Form-based entry eliminates parsing |
| Calendar sync missing | On-leave visibility lost | n8n workflow to Google Calendar |
| Approver assignment broken | Expenses pending forever | Workflow automation with rules |
| Data loss during migration | Accounting corruption | Parallel run 2+ weeks; validate totals |
| Hardcoded ID references broken | Multiple endpoints fail | Comprehensive constant remapping + tests |

---

## Testing Strategy

### Unit Tests
- [ ] Workflow trigger validation
- [ ] Form field validation
- [ ] Approval logic
- [ ] Currency conversion (keep existing)

### Integration Tests
- [ ] End-to-end flow with mock NocoDB
- [ ] Webhook payload handling
- [ ] Database transaction linking
- [ ] Async operation completion

### Parallel Testing
- [ ] Run Basecamp + NocoDB simultaneously
- [ ] Compare results (totals, counts, fields)
- [ ] Reconciliation audit
- [ ] 2-4 week minimum parallel period

### User Acceptance Testing
- [ ] Finance team (invoices, accounting)
- [ ] Operations team (expenses)
- [ ] HR team (on-leave)
- [ ] Form UX validation

---

## Documentation Structure

```
BASECAMP_MIGRATION_INDEX.md (this file)
├─ Quick navigation & overview
├─ Key findings summary
├─ Implementation roadmap
└─ Success criteria & risks

BASECAMP_WORKFLOWS_SUMMARY.md (302 lines, quick reference)
├─ Per-flow quick facts
├─ Key files & endpoints
├─ Hardcoded constants
├─ Validation rules
└─ Migration checklist

BASECAMP_WORKFLOWS_ANALYSIS.md (1,565 lines, detailed)
├─ Complete process flows
├─ Code file references
├─ Database schema docs
├─ Business logic rules
├─ Migration considerations
├─ Implementation notes
└─ Conclusion & timeline
```

---

## Next Steps

### For Immediate Action
1. **Read:** Start with BASECAMP_WORKFLOWS_SUMMARY.md (10 min)
2. **Understand:** Review each flow's trigger → endpoint mapping
3. **Plan:** Identify which flow to start with (recommend Accounting first)
4. **Setup:** Prepare NocoDB structure based on migrations/

### For Migration Planning
1. **Read:** BASECAMP_WORKFLOWS_ANALYSIS.md for your chosen flow
2. **Map:** Create NocoDB table structure
3. **Code:** Update webhook handlers
4. **Test:** Parallel testing with real data
5. **Deploy:** Cutover when ready

### For Architecture Decisions
- Decide: Form-based entry vs. direct Basecamp migration
- Decide: Calendar integration solution (n8n → Google Calendar, native NocoDB, etc.)
- Decide: Approval workflow implementation (NocoDB automations, n8n, custom logic)
- Decide: Activity feed/notification strategy

---

## Contact & Questions

For questions about specific flows or implementation details, refer to:
- **Code locations:** See file reference sections in each document
- **Current implementation:** Available in `feat/migrate-bc-to-nocodb` branch
- **Database schema:** See migrations/ directory for authoritative table definitions

---

**Total Documentation:** 1,867 lines covering all aspects of the Basecamp integration for NocoDB migration planning.

Last updated: 2025-11-14
