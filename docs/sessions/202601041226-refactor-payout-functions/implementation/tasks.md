# Implementation Tasks: Refactor Payout Functions

## Overview
Refactor payout-related Go code to align with updated Notion Contractor Payouts schema.

---

## Task 1: Update payout_types.go

**File:** `pkg/service/notion/payout_types.go`

### 1.1 Rename constant
- [ ] Change `PayoutSourceTypeContractorPayroll` to `PayoutSourceTypeServiceFee`
- [ ] Update value from `"Contractor Payroll"` to `"Service Fee"`

---

## Task 2: Update PayoutEntry struct

**File:** `pkg/service/notion/contractor_payouts.go`

### 2.1 Modify PayoutEntry struct (line 22-34)
- [ ] Remove `Direction` field
- [ ] Rename `ContractorFeesID` to `TaskOrderID`
- [ ] Update field comments

---

## Task 3: Update QueryPendingPayoutsByContractor

**File:** `pkg/service/notion/contractor_payouts.go`

### 3.1 Remove Direction filter (line 81-88)
- [ ] Delete the Direction filter block from the And query

### 3.2 Update property extraction (line 133-144)
- [ ] Remove `Direction` extraction
- [ ] Change `"Billing"` → `"00 Task Order"`
- [ ] Change `"Invoice Split"` → `"02 Invoice Split"`
- [ ] Change `"Refund"` → `"01 Refund"`

---

## Task 4: Update determineSourceType

**File:** `pkg/service/notion/contractor_payouts.go`

### 4.1 Update field reference and return value (line 168-184)
- [ ] Change `entry.ContractorFeesID` to `entry.TaskOrderID`
- [ ] Change `PayoutSourceTypeContractorPayroll` to `PayoutSourceTypeServiceFee`

---

## Task 5: Update CheckPayoutExistsByContractorFee

**File:** `pkg/service/notion/contractor_payouts.go`

### 5.1 Update property name (line 245)
- [ ] Change `"Billing"` → `"00 Task Order"`
- [ ] Consider renaming function to `CheckPayoutExistsByTaskOrder`

---

## Task 6: Update CreatePayoutInput and CreatePayout

**File:** `pkg/service/notion/contractor_payouts.go`

### 6.1 Update CreatePayoutInput struct (line 271-281)
- [ ] Remove `Month` field
- [ ] Remove `Type` field
- [ ] Rename `ContractorFeeID` to `TaskOrderID`
- [ ] Add `Description` field

### 6.2 Update CreatePayout function (line 285-386)
- [ ] Remove `"Month"` property write
- [ ] Remove `"Type"` property write
- [ ] Remove `"Direction"` property write
- [ ] Change `"Billing"` → `"00 Task Order"`
- [ ] Update field reference from `ContractorFeeID` to `TaskOrderID`
- [ ] Add `Description` property write (if provided)
- [ ] Update debug logs

---

## Task 7: Update Refund Payout Functions

**File:** `pkg/service/notion/contractor_payouts.go`

### 7.1 Update CheckPayoutExistsByRefundRequest (line 409-410)
- [ ] Change `"Refund"` → `"01 Refund"`

### 7.2 Update CreateRefundPayoutInput struct (line 388-397)
- [ ] Remove `Month` field

### 7.3 Update CreateRefundPayout function (line 440-541)
- [ ] Remove `"Month"` property write
- [ ] Remove `"Type"` property write
- [ ] Remove `"Direction"` property write
- [ ] Change `"Refund"` → `"01 Refund"`
- [ ] Update debug logs

---

## Task 8: Update Invoice Split Payout Functions

**File:** `pkg/service/notion/contractor_payouts.go`

### 8.1 Update CheckPayoutExistsByInvoiceSplit (line 553-554)
- [ ] Change `"Invoice Split"` → `"02 Invoice Split"`

### 8.2 Update CreateCommissionPayout (line 592-689)
- [ ] Remove `"Type"` property write
- [ ] Remove `"Direction"` property write
- [ ] Change `"Invoice Split"` → `"02 Invoice Split"`

### 8.3 Update CreateBonusPayout (line 701-798)
- [ ] Remove `"Type"` property write
- [ ] Remove `"Direction"` property write
- [ ] Change `"Invoice Split"` → `"02 Invoice Split"`

---

## Task 9: Update Handler

**File:** `pkg/handler/notion/contractor_payouts.go`

### 9.1 Update processContractorPayrollPayouts (around line 191-200)
- [ ] Update `CreatePayoutInput` usage:
  - Rename `ContractorFeeID` to `TaskOrderID`
  - Remove `Month` field
  - Remove `Type` field

### 9.2 Update processRefundPayouts (around line 363-371)
- [ ] Update `CreateRefundPayoutInput` usage:
  - Remove `Month` field

---

## Task 10: Build and Verify

### 10.1 Build verification
- [ ] Run `make build` to ensure no compile errors
- [ ] Run `make lint` to check code style

### 10.2 Manual testing
- [ ] Test with `scripts/test-payout-services/main.go` if available
- [ ] Verify payout creation works with new schema

---

## Execution Order

1. Task 1 (constants)
2. Task 2 (struct)
3. Tasks 3-8 (service methods - can be done together)
4. Task 9 (handler)
5. Task 10 (verification)

## Estimated Scope
- ~15 discrete code changes
- 3 files modified
- Low risk (property name changes, struct field renames)
