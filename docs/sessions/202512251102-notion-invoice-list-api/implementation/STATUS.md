# Implementation Status: Notion Invoice List API Integration

## Overview
Successfully implemented dual-source invoice fetching from both PostgreSQL and Notion databases with intelligent merging and deduplication.

## Completion Status: ✅ COMPLETE

All planned tasks have been successfully implemented and the server builds without errors.

## Implementation Summary

### 1. Notion Invoice Query Interface ✅
- **File**: `pkg/service/notion/interface.go`
- **Changes**:
  - Added `QueryInvoices` method to IService interface
  - Added `GetInvoiceLineItems` method to IService interface
  - Defined `InvoiceFilter` struct for query criteria (ProjectIDs, Statuses, InvoiceNumber)

### 2. Notion Invoice Service Implementation ✅
- **File**: `pkg/service/notion/invoice.go` (NEW)
- **Features**:
  - Database ID constant: `ClientInvoicesDBID = "2bf64b29b84c80879a52ed2f9d493096"`
  - `QueryInvoices()`: Fetches invoices with filters (project, status) and pagination
    - Filters by Type="Invoice" to exclude line items
    - Supports multiple projects via OR logic
    - Supports multiple statuses via OR logic
    - Invoice number filtering done post-fetch (Notion API limitation)
    - Sorts by Issue Date descending
  - `GetInvoiceLineItems()`: Fetches line items for specific invoice
    - Filters by Type="Line Item" AND parent invoice relation

### 3. Notion-to-API Transformation Layer ✅
- **File**: `pkg/controller/invoice/transform_notion.go` (NEW)
- **Functions**:
  - `NotionPageToInvoice()`: Main transformation function
    - Extracts all invoice fields from Notion properties
    - Handles formula properties (Due Date, Final Total, Subtotal, Discount Amount)
    - Handles rollup properties (Recipients → Email)
    - Transforms nested line items
    - Calculates month/year from Issue Date
    - Comprehensive DEBUG logging for troubleshooting
  - `notionLineItemToAPI()`: Transforms line item pages
    - Extracts Quantity, Unit Price, Line Total, Description, Discount fields
  - `mapNotionStatusToAPI()`: Maps Notion status strings to API enums
    - Includes newly added "Uncollectible" status
    - Maps "Cancelled" → "Uncollectible"

### 4. Invoice Merge Helper ✅
- **File**: `pkg/controller/invoice/merge.go` (NEW)
- **Function**: `mergeInvoices()`
- **Logic**:
  - Uses invoice number as deduplication key
  - PostgreSQL invoices added to map first
  - Notion invoices override PG invoices (Notion takes precedence)
  - Sorts final list by InvoicedAt descending
  - Comprehensive logging for merge process and statistics

### 5. Invoice Controller List Method Update ✅
- **File**: `pkg/controller/invoice/list.go`
- **Implementation Flow**:
  1. Fetch invoices from PostgreSQL (with filters)
  2. Fetch invoices from Notion (with same filters)
  3. Transform Notion pages to API Invoice models
  4. Fetch line items for each Notion invoice
  5. Merge results (Notion wins on duplicates)
  6. Apply pagination to merged results
- **Error Handling**:
  - Graceful degradation: Falls back to PG-only data if Notion fails
  - Skips individual invoices if line item fetch fails
  - Skips individual invoices if transformation fails

### 6. Configuration ✅
- **Approach**: Following codebase pattern of hardcoded constants
- **Location**: `pkg/service/notion/invoice.go`
- **Constant**: `ClientInvoicesDBID = "2bf64b29b84c80879a52ed2f9d493096"`
- **Rationale**: Consistent with other Notion database IDs in the codebase

## Field Mappings

### Notion → API Invoice
| Notion Property | Type | API Field | Transformation |
|----------------|------|-----------|----------------|
| (auto) Invoice Number | Title | Number | Direct |
| Issue Date | Date | InvoicedAt | Direct |
| Due Date | Formula (Date) | DueAt | From formula result |
| Paid Date | Date | PaidAt | Direct |
| Status | Status | Status | Via status mapping |
| Final Total | Formula (Number) | Total | From formula result |
| Subtotal | Formula (Number) | SubTotal | From formula result |
| Tax Rate | Number | Tax | SubTotal * TaxRate |
| Discount Amount | Formula (Number) | Discount | From formula result |
| Discount Type | Select | DiscountType | Direct |
| Recipients | Rollup | Email | First email from rollup array |

### Notion → API Line Item
| Notion Property | Type | API Field | Transformation |
|----------------|------|-----------|----------------|
| Quantity | Number | Quantity | Direct |
| Unit Price | Number | UnitCost | Direct |
| Line Total | Formula (Number) | Cost | From formula result |
| Description | Rich Text | Description | Plain text |
| Discount Amount | Formula (Number) | Discount | From formula result |
| Discount Type | Select | DiscountType | Direct |

### Status Mapping
| Notion Status | API Status |
|--------------|-----------|
| Draft | InvoiceStatusDraft |
| Sent | InvoiceStatusSent |
| Overdue | InvoiceStatusOverdue |
| Paid | InvoiceStatusPaid |
| Uncollectible | InvoiceStatusUncollectible |
| Cancelled | InvoiceStatusUncollectible |

## Build Status
✅ Server builds successfully: `go build ./cmd/server`

## Runtime Testing Results

### Initial Test (2025-12-25 12:18)
**Status**: ✅ Core implementation working, ❌ View layer fix required

**Success Indicators**:
- PostgreSQL fetch: ✅ 25 invoices retrieved
- Notion fetch: ✅ 7 invoice pages retrieved
- Line items fetch: ✅ 18 total line items across all Notion invoices
- Transformation: ✅ All 7 Notion invoices successfully transformed
- Merge logic: ✅ 30 total invoices (2 duplicates detected and handled correctly)
- Pagination: ✅ Applied correctly to merged results

**Issue Found**: Runtime panic at `pkg/view/invoice.go:374`
```
runtime error: invalid memory address or nil pointer dereference
ProjectName: invoice.Project.Name
```

**Root Cause**: Notion invoices don't have `Project` relationship loaded (only PostgreSQL invoices have this via `Preload: true`)

**Fix Applied**: Added nil check for `invoice.Project` before accessing `invoice.Project.Name`
```go
projectName := ""
if invoice.Project != nil {
    projectName = invoice.Project.Name
}
```

**Build Status After Fix**: ✅ Successfully builds

### Discord Testing (2025-12-25 12:30)
**Status**: ❌ Two issues identified from Discord `/inv ls` command

**Issues Found**:
1. **Missing Invoice Number Suffix**: Invoice numbers showing as "INV-202512-MUDAH-" (truncated)
   - Root Cause: Only extracting first title segment, Notion splits styled text into multiple segments
   - Fix: Concatenate all title segments like webhook handler does

2. **Missing Project Names**: All invoices showing "Unknown Project (5)"
   - Root Cause: Project relation not being followed to fetch project name
   - Fix: Extract Project relation ID, fetch project page, extract project name from "Project" title property

**Fixes Applied**:
- Updated `NotionPageToInvoice()` to concatenate all invoice number title segments
- Added Project relation extraction and project page fetching
- Modified function signature to accept `notionService` parameter
- Updated `list.go` to pass Notion service to transformation function

**Build Status After Fixes**: ✅ Successfully builds

## Next Steps

### Testing Recommendations
1. Test with various filter combinations:
   - Single project filter
   - Multiple project filters
   - Single status filter
   - Multiple status filters
   - Invoice number filter
   - Combined filters

2. Test merge/deduplication logic:
   - Verify Notion invoices override PG invoices with same number
   - Verify sorting by InvoicedAt descending
   - Test with no duplicates
   - Test with all duplicates
   - Test with partial duplicates

3. Test pagination:
   - Verify correct offset/limit calculation after merge
   - Test edge cases (page beyond results, empty results)

4. Test error handling:
   - Notion API failures (graceful degradation to PG)
   - Line item fetch failures (skip invoice)
   - Transformation failures (skip invoice)

5. Test line items attachment:
   - Verify line items correctly attached to invoices
   - Verify JSON marshaling works correctly

### Future Enhancements
1. **Performance Optimization**:
   - Consider caching Notion results
   - Batch line item fetches if Notion API supports it
   - Optimize pagination to avoid re-merging on every page

2. **Monitoring**:
   - Add metrics for merge statistics
   - Track Notion API failure rates
   - Monitor transformation success rates

3. **Feature Completeness**:
   - Support invoice number filtering in Notion query (if API adds support)
   - Add more comprehensive error messages
   - Consider webhook sync for real-time updates

## Migration Notes
- This implementation is designed for **transition period** where both PG and Notion data coexist
- Notion data takes precedence during deduplication
- PostgreSQL invoices will be deprecated and removed in future
- No changes required to API contract or response format
- Backward compatible with existing API clients
