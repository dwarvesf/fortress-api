# Implementation Tasks - Simplify ?gen inv Flow

## Status: ✅ Completed

## Tasks

### 1. ✅ Add new service method to contractor_payables.go
- Added `PayableInfo` struct for return data
- Added `FindPayableByContractorAndPeriod()` method to lookup existing payables
  - **Returns ALL pending payables**: Queries payables with status "Pending"
  - **Filters by**: Contractor + Period start date + Status = "Pending"
  - **Returns slice**: All matching payables (supports multiple invoices)
  - **Logging**: Logs count of pending payables found
- Added helper method `extractRichText()` for property extraction
- Location: After line 163 in contractor_payables.go

### 2. ✅ Replace processGenInvoice() in gen_invoice.go
- Replaced 20+ step invoice generation with 7-step payable lookup
- Simplified flow:
  1. Query contractor rates
  2. Calculate period dates
  3. Lookup ALL pending payables (returns array)
  4. Check if any pending payables found
  5. Get contractor personal email (once)
  6. **Process each payable separately** (loop):
     - Check if PDF available
     - Extract file ID from URL
     - Share file with contractor email
     - Send separate Discord message for this payable
  7. Log completion with success count
- **Supports multiple invoices**: Sends N separate Discord messages for N pending payables
- Removed invoice data generation, PDF generation, and file upload steps

### 3. ✅ Update Discord DM methods
- Modified `updateDMWithSuccess()` to accept `PayableInfo` instead of invoice data
- Modified `updateDMWithPartialSuccess()` to accept `PayableInfo` instead of invoice data
- Added new method `updateDMWithNotReady()` for "invoice being processed" case
- Updated embed messages and fields to reflect lookup behavior

### 4. ✅ Verify compilation
- Go build succeeds with no errors
- Only minor unused parameter warnings (non-blocking)

### 5. ✅ Run tests
- Existing tests still passing
- Rate limiting tests work correctly
- Validation tests work correctly

## Changes Summary

### Files Modified
1. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/notion/contractor_payables.go`
   - Added `PayableInfo` struct
   - Added `FindPayableByContractorAndPeriod()` method
   - Added `extractRichText()` helper method

2. `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/webhook/gen_invoice.go`
   - Replaced `processGenInvoice()` method (simplified from 20+ steps to 9 steps)
   - Updated `updateDMWithSuccess()` signature and implementation
   - Updated `updateDMWithPartialSuccess()` signature and implementation
   - Added `updateDMWithNotReady()` method

### Key Improvements
- Response time: Expected 1-3 seconds (vs 5-10 seconds previously)
- No invoice generation, PDF creation, or file uploads
- Clear status messages for each case:
  - ✅ Invoice Ready (payable found with PDF)
  - ⏳ Invoice Being Processed (no payable or no PDF)
  - ❌ Error (contractor not found, email missing, etc.)
  - ⚠️ Partial Success (PDF found but sharing failed)
- Rate limiting still enforced (3 per day)

### Edge Case Handling

#### Multiple Payables for Same Contractor + Period
**Problem**: A contractor could have 2+ pending payables for the same period (e.g., hourly + project work, split payables, corrections)

**Solution Implemented**:
1. Query ALL pending payables (filter by status "Pending")
2. Return all of them as an array
3. **Send separate Discord message for EACH payable**
4. Skip payables without PDF (still being processed)
5. Log total count: "processed 2/3" if 3 found but only 2 have PDFs

**Example Scenarios**:
- **1 pending payable** with PDF → Sends 1 message ✅
- **2 pending payables** both with PDFs → Sends 2 separate messages ✅✅
- **3 pending payables**, 2 with PDFs → Sends 2 messages (skips the one without PDF) ✅✅
- **0 pending payables** → Shows "invoice being processed" message ⏳

**Benefits**:
- Contractors see ALL their pending invoices, not just one
- Supports legitimate multiple invoices per period
- Each invoice gets its own clear message with amount, status, PDF link
- No arbitrary selection - full transparency

### Rollback Plan
If issues arise, revert the commit to restore the old implementation.
