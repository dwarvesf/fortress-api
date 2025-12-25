# Implementation Tasks: Notion Invoice List API Migration

**Session:** 202512251102-notion-invoice-list-api
**Created:** 2025-12-25
**Objective:** Migrate invoice List API from PostgreSQL to Notion as source of truth

## Overview

Update the `/api/v1/invoices` List endpoint to fetch invoice data from BOTH PostgreSQL and Notion Client Invoices database (ID: `2bf64b29b84c80879a52ed2f9d493096`), merging results with Notion data taking precedence. This is a transition phase before PostgreSQL is fully deprecated.

## Task Breakdown

### 1. Add Notion Invoice Query Interface ✅

**File:** `pkg/service/notion/interface.go`

- [x] Add `QueryInvoices` method signature
  - Parameters: `filter *InvoiceFilter, pagination model.Pagination`
  - Returns: `([]nt.Page, int64, error)`
- [x] Add `GetInvoiceLineItems` method signature
  - Parameters: `invoicePageID string`
  - Returns: `([]nt.Page, error)`
- [x] Define `InvoiceFilter` struct
  - Fields: `ProjectIDs []string`, `Statuses []string`, `InvoiceNumber string`
- [x] Add DEBUG logging for method calls

### 2. Implement Notion Invoice Service ✅

**File:** `pkg/service/notion/invoice.go` (new file)

- [x] Define `ClientInvoicesDBID` constant = `"2bf64b29b84c80879a52ed2f9d493096"`
- [x] Implement `QueryInvoices` method
  - [x] Build base filter for `Type = "Invoice"` (exclude Line Items)
  - [x] Add project filter if `filter.ProjectIDs` provided (use OR for multiple projects)
  - [x] Add status filter if `filter.Statuses` provided (use OR for multiple statuses)
  - [x] Add invoice number filter if `filter.InvoiceNumber` provided (post-fetch filtering - Notion API limitation)
  - [x] Combine all filters with AND logic
  - [x] Add sort by `Issue Date` descending
  - [x] Call `n.GetDatabase()` with constructed filter
  - [x] Add DEBUG logging: filter construction, query execution, result count
  - [x] Return results and total count
- [x] Implement `GetInvoiceLineItems` method
  - [x] Build filter for `Type = "Line Item"` AND `Parent item` relation contains invoicePageID
  - [x] Call `n.GetDatabase()` with filter and pageSize=100
  - [x] Add DEBUG logging: invoicePageID, line item count
  - [x] Return line item pages
- [x] Add comprehensive error handling with context

### 3. Create Notion-to-API Transformation Layer ✅

**File:** `pkg/controller/invoice/transform_notion.go` (new file)

- [x] Implement `NotionPageToInvoice` function
  - Parameters: `page nt.Page, lineItems []nt.Page`
  - Returns: `(*model.Invoice, error)`
  - [x] Extract `Number` from title property `(auto) Invoice Number`
  - [x] Extract `InvoicedAt` from `Issue Date` date property
  - [x] Extract `DueAt` from `Due Date` formula (date type)
  - [x] Extract `PaidAt` from `Paid Date` date property
  - [x] Extract `Status` and map via `mapNotionStatusToAPI()`
  - [x] Extract `Total` from `Final Total` formula (number)
  - [x] Extract `SubTotal` from `Subtotal` formula (number)
  - [x] Calculate `Tax` from `Tax Rate` number property (SubTotal * TaxRate)
  - [x] Extract `Discount` from `Discount Amount` formula (number)
  - [x] Extract `DiscountType` from `Discount Type` select property
  - [x] Extract `Email` from `Recipients` rollup (first email)
  - [x] Transform line items via `notionLineItemToAPI()`
  - [x] Marshal line items to JSON and assign to `invoice.LineItems`
  - [x] Set `Month` and `Year` from `InvoicedAt`
  - [x] Add DEBUG logging for each field extraction
  - [x] Handle nil checks for all optional fields
- [x] Implement `notionLineItemToAPI` helper function
  - Parameters: `page nt.Page`
  - Returns: `(model.InvoiceItem, error)`
  - [x] Extract `Quantity` from number property
  - [x] Extract `UnitCost` from `Unit Price` number property
  - [x] Extract `Cost` from `Line Total` formula (number)
  - [x] Extract `Description` from rich_text property
  - [x] Extract `Discount` from `Discount Amount` formula
  - [x] Extract `DiscountType` from select property
  - [x] Add DEBUG logging for line item transformation
  - [x] Handle nil checks
- [x] Implement `mapNotionStatusToAPI` helper function
  - Parameters: `notionStatus string`
  - Returns: `model.InvoiceStatus`
  - [x] Map "Draft" → `InvoiceStatusDraft`
  - [x] Map "Sent" → `InvoiceStatusSent`
  - [x] Map "Overdue" → `InvoiceStatusOverdue`
  - [x] Map "Paid" → `InvoiceStatusPaid`
  - [x] Map "Uncollectible" → `InvoiceStatusUncollectible`
  - [x] Map "Cancelled" → `InvoiceStatusUncollectible`
  - [x] Default to `InvoiceStatusDraft`
  - [x] Add DEBUG logging for status mapping

### 4. Create Invoice Merge Helper ✅

**File:** `pkg/controller/invoice/merge.go` (new file)

- [x] Implement `mergeInvoices` helper function
  - Parameters: `pgInvoices []*model.Invoice, notionInvoices []*model.Invoice`
  - Returns: `[]*model.Invoice`
  - [x] Create map for deduplication using invoice number as key
  - [x] Add all PG invoices to map
  - [x] Iterate through Notion invoices:
    - [x] If invoice number exists in map, replace with Notion version (Notion takes precedence)
    - [x] Add DEBUG log: "duplicate invoice found, preferring Notion: number=X"
    - [x] If invoice number not in map, add Notion invoice
  - [x] Convert map values to slice
  - [x] Sort by `InvoicedAt` descending
  - [x] Add DEBUG log: "merged results: PG=X, Notion=Y, deduplicated=Z, final=W"
  - [x] Return merged and sorted list
- [x] Add comprehensive nil checks and error handling

### 5. Update Invoice Controller List Method

**File:** `pkg/controller/invoice/list.go`

- [ ] Update `List` method implementation to fetch from BOTH sources
  - [ ] **Fetch from PostgreSQL:**
    - [ ] Add DEBUG log: "fetching invoices from PostgreSQL"
    - [ ] Keep existing PostgreSQL query logic (via store)
    - [ ] Add DEBUG log: "fetched X invoices from PostgreSQL"
    - [ ] Store PG results in `pgInvoices` variable
  - [ ] **Fetch from Notion:**
    - [ ] Add DEBUG log: "fetching invoices from Notion"
    - [ ] Build `notion.InvoiceFilter` from input parameters
      - Map `input.ProjectIDs` to `notionFilter.ProjectIDs`
      - Map `input.Statuses` to `notionFilter.Statuses`
      - Map `input.InvoiceNumber` to `notionFilter.InvoiceNumber`
    - [ ] Call `c.service.Notion.QueryInvoices(notionFilter, input.Pagination)`
    - [ ] Add DEBUG log: "fetched X invoices from Notion"
    - [ ] Iterate through `notionPages`:
      - [ ] Call `c.service.Notion.GetInvoiceLineItems(page.ID)` for each invoice
      - [ ] Add DEBUG log for line items fetch (success/failure, count)
      - [ ] Call `NotionPageToInvoice(page, lineItems)` to transform
      - [ ] Handle transformation errors gracefully (log and continue)
      - [ ] Append successfully transformed invoices to `notionInvoices` list
    - [ ] Add DEBUG log: "transformed X invoices successfully"
  - [ ] **Merge Results:**
    - [ ] Call `mergeInvoices(pgInvoices, notionInvoices)` helper function
    - [ ] Add DEBUG log: "merging PG (X) + Notion (Y) = Z total invoices"
    - [ ] Apply pagination to merged results (skip/limit based on input.Pagination)
    - [ ] Calculate total count from merged list
  - [ ] Add comprehensive error handling with context
  - [ ] Maintain existing return signature for backward compatibility

### 6. Configuration

**File:** Check if Notion database ID needs to be configurable

- [ ] Verify if `ClientInvoicesDBID` should be added to config
- [ ] If yes, add to `pkg/config/config.go` under Notion section
- [ ] Update environment variable documentation if needed
- [ ] Add DEBUG logging for config loading

### 7. Runtime Bug Fixes ✅

**File:** `pkg/view/invoice.go`

- [x] Fix nil pointer dereference at line 374
  - Added nil check for `invoice.Project` before accessing `invoice.Project.Name`
  - Notion invoices don't have Project relationship loaded
  - Set `projectName = ""` as default for Notion invoices
- [x] Verify build after fix
- [x] Update STATUS.md with runtime testing results

### 8. Testing & Validation

**Manual Testing Tasks:**

- [x] **Initial runtime test completed (2025-12-25 12:18)**
  - ✅ PostgreSQL fetch: 25 invoices
  - ✅ Notion fetch: 7 invoices
  - ✅ Line items: 18 total across Notion invoices
  - ✅ Transformation: All 7 Notion invoices successful
  - ✅ Merge: 30 total (2 duplicates handled correctly)
  - ✅ Pagination: Applied correctly
  - ✅ View layer fix: Nil Project handling
- [ ] Test List API with no filters
  - Verify invoices returned from Notion
  - Verify pagination works
  - Verify response structure matches existing API contract
- [ ] Test List API with project filter
  - Single project ID
  - Multiple project IDs
- [ ] Test List API with status filter
  - Single status
  - Multiple statuses
  - Each status type (draft, sent, overdue, paid, uncollectible)
- [ ] Test List API with invoice number filter
  - Exact match
  - Partial match (contains)
- [ ] Test List API with combined filters
  - Project + Status
  - Project + Status + Invoice Number
- [ ] **Test merge/deduplication logic:**
  - [ ] Verify invoices only in PG are returned
  - [ ] Verify invoices only in Notion are returned
  - [ ] Verify duplicate invoices (same number in both sources) use Notion version
  - [ ] Verify merged results are sorted correctly
  - [ ] Verify pagination works on merged results
  - [ ] Check DEBUG logs show deduplication happening
- [ ] Verify line items are correctly attached to invoices
- [ ] Verify all field mappings are correct
  - Number, dates, status, totals, discounts
  - Line item details
- [ ] Test error scenarios
  - Invalid Notion page structure
  - Missing required properties
  - Network failures
- [ ] Check DEBUG logs are properly emitted at each step
- [ ] Verify performance with large result sets

### 9. Documentation ✅

- [x] Add comments to new code explaining Notion property mappings
- [x] Document status mapping in code comments
- [x] Update API documentation if response structure changed (no changes needed)
- [x] Add runtime testing results in `docs/sessions/202512251102-notion-invoice-list-api/implementation/STATUS.md`

## Dependencies

- Existing Notion service (`pkg/service/notion/notion.go`)
- Notion client library (`github.com/dstotijn/go-notion`)
- Invoice model (`pkg/model/invoice.go`)
- Invoice controller interface

## Success Criteria

- [ ] `/api/v1/invoices` List endpoint returns merged data from both PostgreSQL and Notion
- [ ] Deduplication works correctly (Notion data takes precedence over PG)
- [ ] All filters (project, status, invoice number) work correctly on merged results
- [ ] Response structure matches existing API contract
- [ ] Line items are properly nested in invoice response
- [ ] All status mappings work correctly (including new "Uncollectible")
- [ ] DEBUG logs provide clear trace of data flow from both sources
- [ ] No breaking changes to API consumers
- [ ] Pagination works correctly on merged results

## Notes

- **Migration Phase:** Fetching from BOTH PostgreSQL and Notion during transition period
- PostgreSQL invoice data will be deprecated - Notion is becoming source of truth
- Deduplication: If invoice exists in both sources, Notion version is used (more recent)
- Maintain backward compatibility with existing API consumers
- Focus on read-only operations (List endpoint only)
- Other invoice endpoints (GetTemplate, Send, UpdateStatus, CalculateCommissions) not affected by this migration
- Status "Uncollectible" added to Notion, maps to existing API status
