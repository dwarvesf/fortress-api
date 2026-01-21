# Service Fee Classification Fix - Implementation Guide

**Session**: 202601211458-service-fee-classification
**Date**: 2026-01-21
**Status**: Ready for Implementation

---

## Quick Start

1. **Read Investigation**: Start with `/docs/issues/service-fee-invoice-split-classification.md`
2. **Review Summary**: Read `SUMMARY.md` for quick overview
3. **Understand Flow**: Review `FLOWCHART.md` for visual diagrams
4. **Check Tasks**: Read `tasks.md` for detailed task breakdown
5. **Copy Code**: Use `CODE_SNIPPETS.md` for ready-to-use code
6. **Track Progress**: Update `STATUS.md` as you work

---

## Document Index

### Primary Documents

| Document | Purpose | When to Use |
|----------|---------|-------------|
| **tasks.md** | Detailed task breakdown with acceptance criteria | Implementation planning and execution |
| **SUMMARY.md** | Quick reference with key changes | Fast lookup during coding |
| **CODE_SNIPPETS.md** | Copy-paste-ready code | Active implementation |
| **STATUS.md** | Progress tracking | Project management and status updates |
| **FLOWCHART.md** | Visual diagrams and flow charts | Understanding logic flow |

### Reference Documents

| Document | Location | Purpose |
|----------|----------|---------|
| **Investigation** | `/docs/issues/service-fee-invoice-split-classification.md` | Root cause analysis and solution design |
| **CLAUDE.md** | `/CLAUDE.md` | Project conventions and architecture |
| **README.md** | `/README.md` | General project setup and documentation |

---

## Implementation Workflow

### Phase 1: Pre-Implementation (15 minutes)

1. Review all documents in this directory
2. Confirm development environment is set up (`make dev`)
3. Identify test contractors with mixed payout types
4. Create feature branch:
   ```bash
   git checkout -b fix/service-fee-classification
   ```

### Phase 2: Code Changes (1-2 hours)

**Order of Implementation**:
1. Task 1: Update `determineSourceType()` in `contractor_payouts.go`
2. Task 4: Add debug logging (in parallel with Task 1)
3. Task 2: Update Fee section grouping in `contractor_invoice.go`
4. Task 3: Update Extra Payment section grouping
5. Task 9: Update code comments and documentation

**Use**: `CODE_SNIPPETS.md` for ready-to-use code

### Phase 3: Testing (2-4 hours)

1. Task 6: Write and run unit tests
2. Task 7: Write and run integration tests
3. Run full test suite: `make test`
4. Task 8: Manual end-to-end verification

**Update**: `STATUS.md` as each test phase completes

### Phase 4: Review and Deploy (1-2 hours)

1. Update `STATUS.md` with final status
2. Create PR with reference to investigation document
3. Request code review from @huynguyenh @lmquang
4. Address review comments
5. Deploy to staging
6. Deploy to production

---

## File Change Summary

### Files to Modify

| File | Lines | Changes | Risk |
|------|-------|---------|------|
| `pkg/service/notion/contractor_payouts.go` | 365-380 | Update classification logic | Medium |
| `pkg/controller/invoice/contractor_invoice.go` | 977-994 | Update Fee section | Low |
| `pkg/controller/invoice/contractor_invoice.go` | 996-1012 | Update Extra Payment | Low |

### Files to Create

| File | Purpose |
|------|---------|
| `pkg/service/notion/contractor_payouts_test.go` | Unit tests (if doesn't exist) |
| Tests in `pkg/controller/invoice/contractor_invoice_test.go` | Integration tests |

---

## Expected Outcomes

### Before Fix

```
Invoice Sections:
├── Development Work: Service Fee from TaskOrder
├── Fee: ❌ Commission items (WRONG - should be Service Fee)
├── Extra Payment: ExtraPayment items only
└── Expense Reimbursement: Refund items
```

### After Fix

```
Invoice Sections:
├── Development Work: Service Fee from TaskOrder
├── Fee: ✅ Service Fee from InvoiceSplit (with keywords)
├── Extra Payment: ✅ Commission + ExtraPayment items
└── Expense Reimbursement: Refund items
```

---

## Key Concepts

### Classification Rules

| Source | Description Contains | Type | Invoice Section |
|--------|---------------------|------|-----------------|
| TaskOrder | Any | Service Fee | Development Work |
| InvoiceSplit | "delivery lead" | Service Fee | Fee |
| InvoiceSplit | "account management" | Service Fee | Fee |
| InvoiceSplit | Other | Commission | Extra Payment |
| No relation | Any | ExtraPayment | Extra Payment |
| RefundRequest | Any | Refund | Expense Reimbursement |

### Keywords (Case-Insensitive)

- "delivery lead"
- "account management"

**IMPORTANT**: These must match Notion Type formula logic

---

## Testing Strategy

### Unit Tests (Task 6)
- Test all classification paths
- Test case-insensitive matching
- Test edge cases (empty Description, partial keywords)

### Integration Tests (Task 7)
- Test section grouping with mixed payout types
- Test Service Fee filtering (TaskOrder vs InvoiceSplit)
- Test "Bonus" → "Fee" replacement

### Manual Tests (Task 8)
- Generate actual invoice PDF
- Verify sections visually
- Check debug logs
- Cross-reference with Notion data

---

## Success Metrics

Monitor after deployment:

- ✅ Fee section contains ONLY Service Fee items from InvoiceSplit
- ✅ No Commission items in Fee section
- ✅ Commission items appear in Extra Payment
- ✅ No increase in invoice generation errors
- ✅ Debug logs show correct classifications
- ✅ Performance unchanged (invoice generation time)

---

## Common Issues and Solutions

### Issue 1: Import Error for `strings` Package
**Solution**: Add to imports in `contractor_payouts.go`:
```go
import (
    "strings"
    // ... other imports
)
```

### Issue 2: Test Compilation Errors
**Solution**: Ensure test logger helper exists:
```go
func newTestLogger(t *testing.T) Logger {
    // Use your project's logger
}
```

### Issue 3: Debug Logs Not Appearing
**Solution**: Check log level configuration. Debug logs may be disabled in production.

### Issue 4: Items in Wrong Section
**Solution**: Check debug logs to trace classification:
```bash
grep "\[CLASSIFICATION\]" logs/fortress-api.log
```

---

## Rollback Plan

If issues are found after deployment:

1. **Immediate Rollback** (< 1 hour):
   ```bash
   git revert <commit-hash>
   git push origin develop
   ```

2. **No Database Changes**: No migrations to reverse

3. **Historical Data**: Previously generated PDFs are unaffected

4. **Investigation**: Use debug logs to identify edge cases not covered in testing

---

## Open Questions

### Task 5: GroupFeeByProject

**Status**: Blocked - Decision Required

**Question**: What should we do with the `GroupFeeByProject` functionality?

**Options**:
- A) Remove it (recommended - simplifies code)
- B) Adapt for Service Fee items (more complex)
- C) Keep but disable (interim solution)

**Impact**: Not on critical path, can deploy without resolving

**Decision Required From**: Product Owner / Tech Lead

---

## Contact Information

### Code Reviewers (per CODEOWNERS)
- @huynguyenh
- @lmquang

### Stakeholders
- Product Owner: (TBD)
- Tech Lead: (TBD)

### Questions?
- Review investigation document: `/docs/issues/service-fee-invoice-split-classification.md`
- Check this implementation guide
- Reach out to tech lead

---

## Related Links

### Documentation
- Investigation: `/docs/issues/service-fee-invoice-split-classification.md`
- Project Guidelines: `/CLAUDE.md`
- Main README: `/README.md`

### Code Files
- Classification Logic: `/pkg/service/notion/contractor_payouts.go`
- Section Grouping: `/pkg/controller/invoice/contractor_invoice.go`

### Testing
- Unit Tests: `/pkg/service/notion/contractor_payouts_test.go`
- Integration Tests: `/pkg/controller/invoice/contractor_invoice_test.go`

---

## Revision History

| Date | Version | Changes | Author |
|------|---------|---------|--------|
| 2026-01-21 | 1.0 | Initial implementation guide created | Claude Code |

---

## Next Steps

1. Review all documents in this directory
2. Start with Task 1 in `tasks.md`
3. Use `CODE_SNIPPETS.md` for implementation
4. Update `STATUS.md` as you progress
5. Complete all tests before requesting review

---

**Ready to start? Begin with `SUMMARY.md` for a quick overview, then dive into `tasks.md` for detailed implementation steps.**
