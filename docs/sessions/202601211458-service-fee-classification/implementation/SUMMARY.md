# Implementation Summary: Service Fee Classification Fix

**Quick Reference Guide**

---

## The Problem

Service Fee items from Invoice Split (with "Delivery Lead" or "Account Management" roles) are incorrectly appearing in the "Fee" section of contractor invoices because they are misclassified as Commission.

---

## The Solution

Update classification logic to check Description content for InvoiceSplit items and return the correct source type:
- Keywords found ("delivery lead", "account management") → Service Fee → Fee section
- No keywords → Commission → Extra Payment section

---

## Files to Modify

1. `/pkg/service/notion/contractor_payouts.go` (lines 365-380)
   - Update `determineSourceType()` function
   - Add Description content check for InvoiceSplit items

2. `/pkg/controller/invoice/contractor_invoice.go` (lines 977-1012)
   - Update Fee section to filter Service Fee items (from InvoiceSplit only)
   - Update Extra Payment section to include Commission items

---

## Key Changes

### Change 1: Classification Logic
```go
// Before: All InvoiceSplit → Commission
if entry.InvoiceSplitID != "" {
    return PayoutSourceTypeCommission
}

// After: Check Description for keywords
if entry.InvoiceSplitID != "" {
    desc := strings.ToLower(entry.Description)
    if strings.Contains(desc, "delivery lead") ||
       strings.Contains(desc, "account management") {
        return PayoutSourceTypeServiceFee
    }
    return PayoutSourceTypeCommission
}
```

### Change 2: Fee Section
```go
// Before: Filter Commission items
if item.Type == string(notion.PayoutSourceTypeCommission) {
    feeItems = append(feeItems, item)
}

// After: Filter Service Fee from InvoiceSplit
if item.Type == string(notion.PayoutSourceTypeServiceFee) {
    if item.TaskOrderID == "" && item.ServiceRateID == "" {
        feeItems = append(feeItems, item)
    }
}
```

### Change 3: Extra Payment Section
```go
// Before: Only ExtraPayment type
if item.Type == string(notion.PayoutSourceTypeExtraPayment) {
    extraPaymentItems = append(extraPaymentItems, item)
}

// After: Commission + ExtraPayment
if item.Type == string(notion.PayoutSourceTypeCommission) ||
   item.Type == string(notion.PayoutSourceTypeExtraPayment) {
    extraPaymentItems = append(extraPaymentItems, item)
}
```

---

## Expected Behavior After Fix

| Payout Source | Description Contains | Classification | Invoice Section |
|---------------|---------------------|----------------|-----------------|
| TaskOrder | Any | Service Fee | Development Work |
| InvoiceSplit | "Delivery Lead" | Service Fee | **Fee** |
| InvoiceSplit | "Account Management" | Service Fee | **Fee** |
| InvoiceSplit | No keywords | Commission | **Extra Payment** |
| No relation | Any | ExtraPayment | Extra Payment |
| RefundRequest | Any | Refund | Expense Reimbursement |

---

## Testing Checklist

### Unit Tests (Task 6)
- [ ] TaskOrder → ServiceFee
- [ ] InvoiceSplit + "Delivery Lead" → ServiceFee
- [ ] InvoiceSplit + "Account Management" → ServiceFee
- [ ] InvoiceSplit + no keywords → Commission
- [ ] Case-insensitive matching works
- [ ] Empty Description handled safely

### Integration Tests (Task 7)
- [ ] Fee section contains only Service Fee from InvoiceSplit
- [ ] Extra Payment contains Commission + ExtraPayment
- [ ] Service Fee from TaskOrder in Development Work
- [ ] "Bonus" → "Fee" replacement works

### Manual Testing (Task 8)
- [ ] Generate test invoice
- [ ] Verify PDF sections correct
- [ ] Check debug logs
- [ ] Cross-reference with Notion data

---

## Quick Commands

```bash
# Run unit tests
go test -v ./pkg/service/notion -run TestDetermineSourceType

# Run integration tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Run full test suite
make test

# Generate test invoice
curl -X POST http://localhost:8080/api/v1/invoices/contractor/{id}/generate \
  -H "Authorization: Bearer {token}" \
  -d '{"month": 1, "year": 2026}'

# Check debug logs
grep "\[CLASSIFICATION\]" logs/fortress-api.log
grep "created Fee section" logs/fortress-api.log
grep "created Extra Payment section" logs/fortress-api.log
```

---

## Task Sequence

**Critical Path** (must be done in order):
1. Task 1: Update `determineSourceType()` ← START HERE
2. Task 2: Update Fee section grouping
3. Task 3: Update Extra Payment section grouping
4. Task 6: Write unit tests
5. Task 7: Write integration tests
6. Task 8: Manual end-to-end verification

**Can be done in parallel**:
- Task 4: Debug logging (with Task 1)
- Task 9: Documentation (anytime during implementation)
- Task 10: ADR (anytime, recommended)

**Blocked**:
- Task 5: GroupFeeByProject (needs stakeholder decision)

---

## Estimated Timeline

- **Code Changes**: 1-2 hours (Tasks 1-4)
- **Unit Tests**: 1-2 hours (Task 6)
- **Integration Tests**: 1-2 hours (Task 7)
- **Manual Testing**: 1 hour (Task 8)
- **Documentation**: 30 minutes (Task 9)
- **Total**: 6-8 hours

---

## Rollback Plan

```bash
# Quick revert if issues found
git revert <commit-hash>
git push origin develop

# No database changes to reverse
# Historical PDFs unaffected
```

---

## Success Criteria

After deployment, verify:
- ✅ Fee section has ONLY Service Fee items (with keywords)
- ✅ Extra Payment has Commission + ExtraPayment items
- ✅ No Commission items in Fee section
- ✅ Debug logs show correct classifications
- ✅ No increase in errors or performance issues

---

## Open Questions / Decisions Needed

**Task 5: GroupFeeByProject**
- **Question**: What to do with the GroupFeeByProject functionality?
- **Options**:
  - A) Remove it (recommended)
  - B) Adapt for Service Fee items
  - C) Keep but disable
- **Impact**: Not blocking critical path, can deploy without resolving
- **Decision Required From**: Product Owner / Tech Lead

---

## Contact

- **Code Reviewers**: @huynguyenh @lmquang (per CODEOWNERS)
- **Questions**: See investigation doc or reach out to tech lead

---

## Related Documents

- **Detailed Tasks**: `tasks.md` (in same directory)
- **Status Tracking**: `STATUS.md` (in same directory)
- **Investigation**: `/docs/issues/service-fee-invoice-split-classification.md`

---

**Version**: 1.0
**Date**: 2026-01-21
