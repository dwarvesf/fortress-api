# Quick Reference Card

**One-page guide for implementing the Service Fee classification fix**

---

## The Problem (in 30 seconds)

Service Fee items from Invoice Split are appearing in the wrong section of contractor invoices because ALL InvoiceSplit items are classified as "Commission", even when they should be "Service Fee" (for Delivery Lead and Account Management roles).

---

## The Fix (3 code changes)

### 1. Update Classification Logic
**File**: `pkg/service/notion/contractor_payouts.go` (lines 365-380)

**Change**: Add Description check for InvoiceSplit items
```go
if entry.InvoiceSplitID != "" {
    if entry.Description != "" {
        desc := strings.ToLower(entry.Description)
        if strings.Contains(desc, "delivery lead") ||
           strings.Contains(desc, "account management") {
            return PayoutSourceTypeServiceFee  // ← NEW
        }
    }
    return PayoutSourceTypeCommission
}
```

### 2. Update Fee Section
**File**: `pkg/controller/invoice/contractor_invoice.go` (lines 977-994)

**Change**: Filter Service Fee from InvoiceSplit (not Commission)
```go
// OLD: if item.Type == string(notion.PayoutSourceTypeCommission)
// NEW:
if item.Type == string(notion.PayoutSourceTypeServiceFee) {
    if item.TaskOrderID == "" && item.ServiceRateID == "" {
        feeItems = append(feeItems, item)
    }
}
```

### 3. Update Extra Payment Section
**File**: `pkg/controller/invoice/contractor_invoice.go` (lines 996-1012)

**Change**: Include Commission items (moved from Fee section)
```go
// OLD: if item.Type == string(notion.PayoutSourceTypeExtraPayment)
// NEW:
if item.Type == string(notion.PayoutSourceTypeCommission) ||
   item.Type == string(notion.PayoutSourceTypeExtraPayment) {
    extraPaymentItems = append(extraPaymentItems, item)
}
```

---

## Keywords (Case-Insensitive)

- "delivery lead"
- "account management"

**CRITICAL**: Must match Notion Type formula

---

## Testing Commands

```bash
# Run unit tests
go test -v ./pkg/service/notion -run TestDetermineSourceType

# Run integration tests
go test -v ./pkg/controller/invoice -run TestGroupIntoSections

# Run all tests
make test

# Generate test invoice
curl -X POST http://localhost:8080/api/v1/invoices/contractor/{id}/generate \
  -H "Authorization: Bearer {token}" \
  -d '{"month": 1, "year": 2026}'

# Check debug logs
grep "\[CLASSIFICATION\]" logs/fortress-api.log
```

---

## Expected Results

| Before Fix | After Fix |
|------------|-----------|
| Fee section: Commission items ❌ | Fee section: Service Fee from InvoiceSplit ✅ |
| Extra Payment: ExtraPayment only | Extra Payment: Commission + ExtraPayment ✅ |

---

## Task Order

1. Change 1: Update `determineSourceType()` ← **Start here**
2. Change 2: Update Fee section grouping
3. Change 3: Update Extra Payment section grouping
4. Write unit tests (Task 6)
5. Write integration tests (Task 7)
6. Manual testing (Task 8)

**Estimated Time**: 6-8 hours total

---

## Files to Read

1. **SUMMARY.md** - Quick overview (5 min read)
2. **tasks.md** - Detailed breakdown (20 min read)
3. **CODE_SNIPPETS.md** - Copy-paste code (during implementation)
4. **STATUS.md** - Track progress (update as you go)

---

## Success Checklist

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual invoice generated successfully
- [ ] PDF has correct sections (Fee = Service Fee items only)
- [ ] Debug logs show correct classifications
- [ ] No linter errors
- [ ] Code reviewed and approved
- [ ] PR merged

---

## Common Pitfalls

- ⚠️ Forgetting to import `strings` package
- ⚠️ Using case-sensitive matching (use `strings.ToLower()`)
- ⚠️ Not filtering TaskOrder items from Fee section
- ⚠️ Forgetting to move "Bonus" → "Fee" replacement to Extra Payment section

---

## Rollback

```bash
git revert <commit-hash>
git push origin develop
```

No database changes to reverse.

---

## Help

- Investigation: `/docs/issues/service-fee-invoice-split-classification.md`
- Full guide: `README.md` (in this directory)
- Code reviewers: @huynguyenh @lmquang

---

**Last Updated**: 2026-01-21
