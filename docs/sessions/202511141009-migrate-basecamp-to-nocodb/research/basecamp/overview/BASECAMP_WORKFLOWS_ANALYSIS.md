# Fortress API: Basecamp Integration Workflows Analysis
## Migration to NocoDB - Comprehensive Flow Documentation

**Analysis Date:** 2025-11-14  
**Target Codebase:** fortress-api (feat/migrate-bc-to-nocodb)  
**Scope:** 4 critical workflows dependent on Basecamp integration

---

## TABLE OF CONTENTS
1. [Invoice Flow](#invoice-flow)
2. [Accounting Flow](#accounting-flow)
3. [Expense Flow](#expense-flow)
4. [On-Leave Request Flow](#on-leave-request-flow)
5. [Cross-Flow Integration Points](#cross-flow-integration-points)
6. [Migration Considerations](#migration-considerations)

---

# 1. INVOICE FLOW

## 1.1 End-to-End Process

### Trigger Points
- **API Endpoint:** `PUT /webhooks/basecamp/operation/invoice` (Basecamp webhook)
- **Trigger Event:** Basecamp todo marked as complete with specific format
- **Webhook Source:** Basecamp event: `todo_completed` or manual API call

### Process Flow

```
1. User/System marks todo as complete in Basecamp
   ↓
2. Basecamp sends webhook to: PUT /webhooks/basecamp/operation/invoice
   ↓
3. Handler: pkg/handler/webhook.MarkInvoiceAsPaidViaBasecamp()
   ├─ Decodes BasecampWebhookMessage
   ├─ Validates message sender (not from AutoBotID)
   └─ Parses todo title for invoice number
   ↓
4. Handler: pkg/handler/webhook.markInvoiceAsPaid()
   ├─ Calls: webhook.GetInvoiceViaBasecampTitle(msg)
   │  ├─ Extracts invoice number from todo title format: MM/YYYY - #XXXX-XX-XXXX
   │  ├─ Queries: store.Invoice.One() → finds invoice by number
   │  ├─ Validates invoice status (must be 'sent' or 'overdue')
   │  └─ Fetches comments from Basecamp
   │      └─ Validates "Paid" comment from specific user (HanBasecampID or similar)
   │
   └─ Calls: controller.Invoice.MarkInvoiceAsPaidByBasecampWebhookMessage()
      ├─ Updates invoice status to 'paid'
      ├─ Records PaidAt timestamp
      ├─ Processes invoice async tasks (via goroutines):
      │  ├─ processPaidInvoiceData()
      │  │  ├─ Updates database: invoice.status = 'paid', invoice.paid_at = now
      │  │  ├─ Stores commission (if applicable)
      │  │  ├─ Creates AccountingTransaction for income
      │  │  │  └─ Converts currency using Wise API
      │  │  ├─ Enqueues Basecamp comment (success/failed)
      │  │  └─ Handles transaction rollback on failure
      │  │
      │  ├─ sendThankYouEmail()
      │  │  └─ Sends email to project contact (via GoogleMail service)
      │  │
      │  └─ movePaidInvoiceGDrive()
      │     └─ Moves invoice PDF from 'Sent' → 'Paid' folder
      │
      └─ Returns 200 OK (even if async tasks fail)
   ↓
5. Worker processes enqueued Basecamp comment message
```

## 1.2 Basecamp Integration Points

### Basecamp Service Calls

**Service:** `pkg/service/basecamp/integration.go` + subservices

| Operation | Basecamp API Call | Purpose |
|-----------|-------------------|---------|
| Get Todo | `service.Basecamp.Todo.Get(msg.Recording.URL)` | Fetch full todo details |
| Get Comments | `service.Basecamp.Comment.Gets(bucketID, todoID)` | Validate "Paid" confirmation comment |
| Complete Todo | `service.Basecamp.Todo.Complete(bucketID, todoID)` | Mark invoice todo as done in Basecamp |
| Post Comment | `service.Basecamp.BuildCommentMessage()` + worker | Send success/failure feedback to Basecamp |
| Archive Todo (error) | `service.Basecamp.Recording.Archive(bucketID, todoID)` | Archive todo if invoice marked as error |

### Basecamp Resources Created/Updated

**In Basecamp (Project: Woodland or Playground)**
- **Todo Title Format:** `MM/YYYY - #INVOICE_NUMBER` (e.g., "01/2025 - #2025-DWF-001")
- **Bucket (Project) ID:** 
  - Prod: `9403032` (Woodland)
  - Dev: `12984857` (Fortress | Playground)
- **Todo List ID:** 
  - Prod: `1346305133` (Woodland)
  - Dev: `1941398075` (Playground)
- **Comment Validation:** Requires comment starting with "Paid" from approved user
- **User IDs (hardcoded):**
  - AutoBotID: `25727627` (system bot, ignored)
  - HanBasecampID: `21562923` (approval authority in prod)
  - NamNguyenBasecampID: `21581534` (approval authority in dev)

### Data Sent to Basecamp

**Comments Posted:**
- Success: "Invoice marked as paid" (with green indicator if using message types)
- Error: Error message describing validation failure
- Message Type Tags: `success` / `failed`

## 1.3 Database Models

### Primary Model: Invoice
**Table:** `invoices`
**Fields:**
```
id (UUID)
number (TEXT) - UNIQUE, format: YYYY-XXX-###
status (ENUM) - draft|sent|overdue|paid|error|scheduled
project_id (UUID) - FK to projects
bank_id (UUID) - FK to bank_accounts
sent_by (UUID) - FK to employees
paid_at (TIMESTAMP) - set when marked as paid
invoiced_at (DATE) - invoice date
due_at (TIMESTAMP) - due date
month (INT) - month of invoice
year (INT) - year of invoice
total (NUMERIC)
conversion_amount (INT8)
conversion_rate (FLOAT4)
email (VARCHAR) - recipient email
cc (JSON) - CC list
thread_id (VARCHAR) - Gmail thread ID
invoice_file_url (TEXT)
line_items (JSON) - invoice line details
metadata (JSON)
created_at (TIMESTAMP)
updated_at (TIMESTAMP)
deleted_at (TIMESTAMP) - soft delete
```

### Related Models
- **BankAccount:** Contains currency info
- **Currency:** Used for conversion calculations
- **Project:** Linked to invoice, provides organization context
- **Employee:** Sender of invoice

### Relationships
```
Invoice
├─ Project (belongs to)
│  ├─ Organization
│  └─ BankAccount
│     └─ Currency
├─ Bank (belongs to)
│  └─ Currency
└─ Sender (Employee - belongs to)
```

## 1.4 Code Files Involved

### Handler Layer
- **File:** `pkg/handler/webhook/basecamp.go`
- **Methods:**
  - `MarkInvoiceAsPaidViaBasecamp()` - HTTP handler
  - `markInvoiceAsPaid()` - Business logic delegator

- **File:** `pkg/handler/webhook/basecamp_invoice.go`
- **Methods:**
  - `GetInvoiceViaBasecampTitle()` - Parse title and fetch invoice

### Controller Layer
- **File:** `pkg/controller/invoice/update_status.go`
- **Methods:**
  - `MarkInvoiceAsPaidByBasecampWebhookMessage()` - Entry point
  - `processPaidInvoice()` - Orchestrates async tasks
  - `processPaidInvoiceData()` - Database updates + accounting
  - `sendThankYouEmail()` - Email notification
  - `movePaidInvoiceGDrive()` - Google Drive file management

### Service Layer
- **File:** `pkg/service/basecamp/integration.go`
- Methods for Basecamp API interactions:
  - `service.Basecamp.Todo.Get()`
  - `service.Basecamp.Todo.Complete()`
  - `service.Basecamp.Comment.Gets()`
  - `service.Basecamp.Recording.Archive()`
  - `service.Basecamp.BuildCommentMessage()`
  - `service.Basecamp.Wise.Convert()` - Currency conversion

- **Subservices:**
  - `pkg/service/basecamp/todo/service.go`
  - `pkg/service/basecamp/comment/service.go`
  - `pkg/service/basecamp/recording/service.go`

### Store/Repository Layer
- **File:** `pkg/store/invoice/invoice.go`
- **Methods:**
  - `One(db, query)` - Get invoice by ID or number
  - `UpdateSelectedFieldsByID()` - Update status and paid_at
  - `Update()` - Full update

- **File:** `pkg/store/accounting/accounting.go`
- **Methods:**
  - `CreateTransaction()` - Record income transaction

### Models
- `pkg/model/invoice.go` - Invoice struct with validation
- `pkg/model/accounting.go` - AccountingTransaction struct

### Routing
- **File:** `pkg/routes/v1.go` (lines 71-90)
```go
operationGroup := basecampGroup.Group("/operation")
{
    operationGroup.PUT("/invoice", h.Webhook.MarkInvoiceAsPaidViaBasecamp)
}
```

## 1.5 Business Logic

### Validation Rules
1. **Todo Title Format:** Must match regex `.*([1-9]|0[1-9]|1[0-2])/(20[0-9]{2}) - #(20[0-9]+-[A-Z0-9]+-[0-9]+)`
   - Pattern: `MM/YYYY - #INVOICE_NUMBER`
   - Example: `01/2025 - #2025-DWF-001`

2. **Invoice State Validation:**
   - Invoice must exist in database
   - Invoice status must be `sent` OR `overdue` (not draft, error, paid, scheduled)
   - Prevents double-payment

3. **Comment Validation:**
   - Requires comment starting with "Paid" in todo comments
   - In production: must be from `HanBasecampID` (21562923)
   - In dev: can be from `NamNguyenBasecampID` (21581534)
   - Prevents accidental marking without explicit confirmation

4. **Creator Validation:**
   - Ignores webhooks from AutoBotID (prevents loops)

### State Transitions
```
DRAFT
  ↓ (after PDF generation & sending)
SENT ──→ OVERDUE (time-based, via cron)
  ↓
PAID ← (via Basecamp webhook with validation)
  ↓ (optional error handling)
ERROR ← (manual override or PDF generation failure)
```

### Calculations/Transformations

1. **Currency Conversion:**
   - Invoice total (in project's currency) → VND conversion
   - Uses Wise API: `service.Wise.Convert(amount, fromCurrency, toCurrency)`
   - Stores: `conversion_amount` (VND) + `conversion_rate`

2. **Commission Calculation:**
   - Call to `storeCommission()` if project has commission structure
   - Creates accounting transaction(s) for commission payouts

3. **Accounting Transaction Creation:**
   - Type: `AccountingIncome`
   - Category: `In`
   - Links to invoice via metadata: `{source: "invoice", id: invoice_uuid}`
   - Converts to VND for internal accounting

### Error Handling

| Error Scenario | Response | Basecamp Action |
|----------------|----------|-----------------|
| Invalid title format | 200 OK (fail silently) | Fail comment posted |
| Invoice not found | 200 OK (fail silently) | Fail comment posted |
| Wrong invoice status | 200 OK (fail silently) | Fail comment posted |
| Missing "Paid" comment | 200 OK (fail silently) | Fail comment posted |
| Wrong approval user | 200 OK (fail silently) | Fail comment posted |
| DB update fails | 200 OK (async failure) | Fail comment posted |
| Email send fails | Logged, continues | N/A |
| GDrive move fails | Logged, continues | N/A |
| Currency conversion fails | Transaction rolls back | Fail comment posted |

**Note:** Handler always returns 200 OK for webhook reliability; failures are logged and commented in Basecamp.

## 1.6 Migration Considerations

### Basecamp-Specific Dependencies
1. **Hardcoded Basecamp IDs:**
   - Bucket IDs: `9403032` (prod), `12984857` (dev)
   - Todo list IDs: `1346305133`, `1941398075`
   - User IDs: `25727627`, `21562923`, `21581534`
   - Must be mapped to NocoDB record IDs

2. **Basecamp API Integration Points:**
   - `Comment.Gets()` - validates approval comment
   - `Todo.Get()` - fetches full todo
   - `Todo.Complete()` - marks as done in Basecamp
   - Need NocoDB webhook/API equivalents

3. **Title Parsing Logic:**
   - Current: regex on Basecamp todo title
   - Future: could read from NocoDB record directly (no parsing needed)
   - Would allow more flexible title format

### Data That Would Change
| Current (Basecamp) | Future (NocoDB) | Migration Strategy |
|-------------------|-----------------|-------------------|
| Todo title parsing | Direct field read | Map from title field or separate invoice_number field |
| Approval comment validation | NocoDB record status/field | Use NocoDB "approved_by" field + timestamp |
| Mark todo complete | Update NocoDB record status | Direct status update in NocoDB |
| Post comment feedback | N/A or NocoDB activity feed | Optional: log to NocoDB comments table |
| Currency conversion | No change needed | Keep existing Wise API calls |

### Hardcoded Constants to Migrate
```go
// consts/consts.go
WoodlandID = 9403032           // Prod Basecamp bucket
PlaygroundID = 12984857        // Dev Basecamp bucket
WoodlandTodoID = 1346305133    // Prod todo list
PlaygroundTodoID = 1941398075  // Dev todo list
HanBasecampID = 21562923       // Prod approver
NamNguyenBasecampID = 21581534 // Dev approver
```

### Webhook URL Changes
```
Current:  PUT /webhooks/basecamp/operation/invoice
Future:   PUT /webhooks/nocodb/operation/invoice
          (or /webhooks/n8n/operation/invoice if using n8n)
```

### Testing Implications
- Invoice creation/update workflows must be tested without Basecamp API
- Mock NocoDB responses for unit tests
- Integration tests should use test NocoDB instance
- Approval workflow validation becomes simpler (no comment parsing)

---

# 2. ACCOUNTING FLOW

## 2.1 End-to-End Process

### Trigger Points
- **API Endpoint:** `POST /webhooks/basecamp/operation/accounting-transaction`
- **Trigger Event:** Basecamp todo created in Accounting project
- **Webhook Source:** Basecamp event: `todo_created`

### Process Flow

```
1. User creates new todo in Basecamp Accounting project
   ├─ Todo location: Accounting | MONTH YEAR (e.g., "Accounting | January 2025")
   └─ Todo title format: "Description | Amount | Currency"
        e.g., "Office Rental | 1.000.000 | VND"
   ↓
2. Basecamp sends webhook to: POST /webhooks/basecamp/operation/accounting-transaction
   ↓
3. Handler: pkg/handler/webhook.StoreAccountingTransaction()
   ├─ Decodes BasecampWebhookMessage
   ├─ Validates bucket/project ID
   │  └─ Prod: AccountingID (15258324)
   │  └─ Dev: PlaygroundID (12984857)
   └─ Calls webhook handler
   ↓
4. Webhook Handler: pkg/handler/webhook.StoreAccountingTransactionFromBasecamp()
   ├─ Calls getManagementTodoInfo(msg)
   │  ├─ Fetches todo list from Basecamp.Todo.GetList()
   │  ├─ Parses parent category from title: "Accounting | MONTH YEAR"
   │  ├─ Validates category matches expected pattern
   │  ├─ Extracts month and year
   │  └─ Returns managementTodoInfo{month, year}
   │
   ├─ Parses todo title with regex: [S|s]alary\s*(1st|15th)|(.*)\|\s*([0-9\.]+)\s*\|\s*([a-zA-Z]{3})
   │  └─ Extracts: description, amount, currency
   │
   └─ Calls storeAccountingTransaction(operationInfo, data, basecampTodoID)
      ├─ Parses amount (removes dots for thousands)
      ├─ Fetches currency from database
      ├─ Converts to VND using Wise API
      ├─ Categorizes expense (office rental, office services, etc.)
      ├─ Creates AccountingTransaction record
      │  ├─ name: "description"
      │  ├─ amount: parsed amount
      │  ├─ currency: extracted currency
      │  ├─ conversion_amount: VND equivalent
      │  ├─ date: now()
      │  ├─ type: "OP" (Operation)
      │  ├─ category: auto-categorized
      │  └─ metadata: {source: "basecamp_accounting", id: todoID}
      │
      └─ Stores in database via store.Accounting.CreateTransaction()
   ↓
5. Returns 500 on error, 200 on success
```

## 2.2 Basecamp Integration Points

### Basecamp Service Calls

| Operation | Basecamp API Call | Purpose |
|-----------|-------------------|---------|
| Get Todo List | `service.Basecamp.Todo.GetList(url)` | Fetch parent category to extract month/year |
| Parse Title | Regex on `msg.Recording.Title` | Extract description, amount, currency |

### Basecamp Resources Used

**In Basecamp (Project: Accounting or Playground)**
- **Bucket (Project) ID:**
  - Prod: `15258324` (Accounting)
  - Dev: `12984857` (Fortress | Playground)
- **Todo List ID:**
  - Prod: `2329633561` (Accounting)
  - Dev: Not specified (uses playground)
- **Parent Category Title Format:** `Accounting | MONTH_NAME YEAR`
  - Example: `Accounting | January 2025`
  - Must match regex: `Accounting \| (.+) ([0-9]{4})`

### Data Sent to Basecamp
- No data is sent to Basecamp
- Basecamp is the data source only
- One-way integration (read-only)

## 2.3 Database Models

### Primary Model: AccountingTransaction
**Table:** `accounting_transactions`
**Fields:**
```
id (UUID)
name (TEXT) - description of transaction
date (TIMESTAMP) - transaction date
amount (FLOAT8) - original amount
currency (TEXT) - currency name (VND, USD, etc.)
currency_id (UUID) - FK to currencies
conversion_amount (INT8) - amount in VND
conversion_rate (FLOAT8) - exchange rate used
organization (TEXT) - optional organization name
category (TEXT) - FK to accounting_categories (name)
type (TEXT) - accounting type (SE, OP, OV, CA, In)
metadata (JSON) - {source: "basecamp_accounting", id: basecamp_todo_id}
created_at (TIMESTAMP)
deleted_at (TIMESTAMP) - soft delete
```

### Related Models
- **Currency:** Contains name and conversion info
- **AccountingCategory:** Lookup table with type and name

### Relationships
```
AccountingTransaction
├─ Currency (belongs to)
└─ AccountingCategory (belongs to via type)
```

### Enum Values
**Accounting Types:**
- `SE` - Salary Expense
- `OP` - Operation (e.g., office rental, utilities)
- `OV` - Overhead (expense reimbursements)
- `CA` - Capital (fixed assets)
- `In` - Income

**Categories (from Basecamp title parsing):**
- `Office Supply` - office supplies
- `Office Services` - general office services
- `Office Space` - office rental/CBRE
- `Tools`, `Assets`, etc.

## 2.4 Code Files Involved

### Handler Layer
- **File:** `pkg/handler/webhook/basecamp_accounting.go`
- **Methods:**
  - `StoreAccountingTransaction()` - HTTP handler wrapper
  - `StoreAccountingTransactionFromBasecamp()` - Main logic
  - `getManagementTodoInfo()` - Parse parent category
  - `storeAccountingTransaction()` - Create DB record
  - `StoreOperationAccountingTransaction()` - DB wrapper
  - `checkCategory()` - Auto-categorize

### Service Layer
- **File:** `pkg/service/basecamp/todo/service.go`
- **Methods:**
  - `service.Basecamp.Todo.GetList(url)` - Fetch parent todo list

### Store/Repository Layer
- **File:** `pkg/store/accounting/accounting.go`
- **Methods:**
  - `CreateTransaction(db, transaction)` - Insert into DB
  - `GetAccountingTransactions()` - List all
  - `DeleteTransaction()` - Soft delete

### Models
- `pkg/model/accounting.go` - AccountingTransaction and Category structs

### Routing
- **File:** `pkg/routes/v1.go` (lines 79-82)
```go
operationGroup := basecampGroup.Group("/operation")
{
    operationGroup.POST("/accounting-transaction", h.Webhook.StoreAccountingTransaction)
}
```

## 2.5 Business Logic

### Validation Rules

1. **Todo List Parent Validation:**
   - Parent must be in Accounting project (ID 15258324 in prod)
   - Parent title must match: `Accounting | MONTH YEAR`
   - Example: `Accounting | January 2025`

2. **Title Format Validation:**
   - Must contain pipe separators: `description | amount | currency`
   - Alternative: `Salary 1st` or `Salary 15th` (for salary entries)
   - Regex: `[S|s]alary\s*(1st|15th)|(.*)\|\s*([0-9\.]+)\s*\|\s*([a-zA-Z]{3})`

3. **Amount Parsing:**
   - Accepts: `1000`, `1.000`, `1.000.000` (with thousands separator)
   - Dots removed before parsing: `1.000.000` → `1000000`
   - Must be numeric

4. **Currency Validation:**
   - Must exist in currencies table
   - Common: VND, USD, EUR, SGD, etc.

5. **Bucket/Project Validation:**
   - In prod: must be Accounting project (15258324)
   - In dev: uses Playground (12984857)
   - Ignores webhooks from wrong project

### Categorization Logic

```go
if strings.Contains(content, "office rental") || 
   strings.Contains(content, "cbre") {
    return AccountingOfficeSpace
} else {
    return AccountingOfficeServices  // default
}
```

### Calculations/Transformations

1. **Amount Parsing:**
   - Input: "1.000.000"
   - Remove dots: "1000000"
   - Parse as integer

2. **Currency Conversion:**
   - Uses Wise API: `service.Wise.Convert(amount, currency, "VND")`
   - Stores: `conversion_amount` (VND) + `conversion_rate`
   - If conversion fails: returns nil (no transaction created)

3. **Date Extraction:**
   - Extracted from parent todo list title month/year
   - Sets transaction date to parsed month

### Error Handling

| Error Scenario | Response | Action |
|----------------|----------|--------|
| Wrong project | 200 OK (silent) | Transaction not created |
| Invalid title format | 400 Bad Request | Rejected with error |
| Currency not found | 400 Bad Request | Rejected with error |
| Conversion fails | 200 OK | Transaction not created |
| DB insert fails | 500 Internal Error | Rolled back |

## 2.6 Migration Considerations

### Basecamp-Specific Dependencies
1. **Hardcoded Project IDs:**
   - Prod: `15258324` (Accounting)
   - Dev: `12984857` (Playground)
   - Must map to NocoDB table/workspace IDs

2. **Basecamp API Calls:**
   - `Todo.GetList()` - fetches parent category
   - No other Basecamp interactions
   - Could be simplified in NocoDB (direct field reads)

3. **Title Parsing Logic:**
   - Parent category extraction from title
   - Could be simplified: NocoDB can have separate "month" and "year" fields
   - Amount parsing could use number fields

### Data That Would Change
| Current (Basecamp) | Future (NocoDB) | Migration Strategy |
|-------------------|-----------------|-------------------|
| Title parsing for month/year | Separate month/year fields | Use NocoDB columns directly |
| Amount string parsing | Number field | Direct number input |
| Category auto-detection | Dropdown/select field | Pre-categorized in NocoDB |
| Regex validation | Data validation rules | Configure in NocoDB |

### Hardcoded Constants
```go
// consts/consts.go
AccountingID = 15258324        // Prod Accounting project
PlaygroundID = 12984857        // Dev project
AccountingTodoID = 2329633561  // Prod todo list
```

### Webhook Logic Changes
1. **Simpler validation:** No todo list fetching needed if month/year are fields
2. **Direct field mapping:** Amount, currency, description from NocoDB fields
3. **Optional:** Keep category auto-detection or let users select

---

# 3. EXPENSE FLOW

## 3.1 End-to-End Process

### Trigger Points
- **Validation Endpoint:** `POST /webhooks/basecamp/expense/validate`
- **Creation Endpoint:** `POST /webhooks/basecamp/expense`
- **Deletion Endpoint:** `DELETE /webhooks/basecamp/expense`
- **Trigger Events:** 
  - `todo_created` (Woodland/Playground expense todo)
  - `todo_completed` (expense marked as done)
  - `todo_uncompleted` (expense marked as undone)

### Process Flow

#### 3.1.1 VALIDATION Flow

```
1. User creates expense todo in Basecamp Expense list
   ├─ Bucket: Woodland (9403032) or Playground (12984857)
   ├─ TodoList: ExpenseTodoID (2353511928) or similar
   └─ Title format: "Reason | Amount | Currency" (e.g., "Office supplies | 500.000 | VND")
   ↓
2. Basecamp sends webhook: POST /webhooks/basecamp/expense/validate
   ├─ Event: todo_created
   └─ Action: Validate format and assign to approver
   ↓
3. Handler: pkg/handler/webhook.ValidateBasecampExpense()
   ↓
4. Webhook Handler: pkg/handler/webhook.basecampExpenseValidate()
   ├─ Validates bucket name matches:
   │  └─ Prod: "Woodland"
   │  └─ Dev: "Fortress | Playground"
   │
   ├─ If todo_created:
   │  ├─ Fetch todo from Basecamp
   │  ├─ Set AssigneeIDs to [HanBasecampID] (prod) or [NamNguyenBasecampID] (dev)
   │  └─ Update todo in Basecamp
   │
   ├─ Calls extractExpenseData(msg)
   │  ├─ Parses title: split by "|" → [reason, amount, currency]
   │  ├─ Validates >= 3 parts
   │  ├─ Extracts reason: trim + add date
   │  ├─ Extracts amount: custom parser for k/tr/m units
   │  ├─ Validates currency: VND or USD only
   │  ├─ Fetches image URL from todo attachments
   │  └─ Returns BasecampExpenseData or error
   │
   └─ If validation fails:
      ├─ Posts format error comment to todo
      └─ Returns 200 OK
   └─ If validation succeeds:
      ├─ Posts "Format looks good" comment
      └─ Returns 200 OK
   ↓
5. Basecamp comment posted with feedback
```

#### 3.1.2 CREATION Flow

```
1. User marks expense todo as complete in Basecamp
   └─ Basecamp sends webhook: POST /webhooks/basecamp/expense
      └─ Event: todo_completed
   ↓
2. Handler: pkg/handler/webhook.CreateBasecampExpense()
   ↓
3. Webhook Handler: pkg/handler/webhook.createBasecampExpense()
   ├─ Calls extractExpenseData(msg) again
   ├─ Unmarshals metadata from webhook payload
   │
   ├─ Calls service.Basecamp.CreateBasecampExpense(*obj)
   │  ├─ In: pkg/service/basecamp/integration.go
   │  ├─ Fetches employee by Basecamp creator ID
   │  ├─ Fetches currency by name
   │  │
   │  ├─ Creates Expense record:
   │  │  ├─ employee_id: creator employee
   │  │  ├─ amount: parsed amount
   │  │  ├─ reason: description
   │  │  ├─ currency_id: found currency
   │  │  ├─ invoice_image_url: extracted image
   │  │  ├─ basecamp_id: todo ID
   │  │  ├─ metadata: JSON with basecamp info
   │  │  └─ issued_date: now()
   │  │
   │  ├─ Converts currency to VND via Wise API
   │  │  ├─ Handles conversion failure gracefully
   │  │  └─ Logs critical error if fails
   │  │
   │  ├─ Creates AccountingTransaction:
   │  │  ├─ name: "Expense - " + description
   │  │  ├─ amount: original amount
   │  │  ├─ type: "OV" (Overhead/Variance)
   │  │  ├─ category: "Office Supply"
   │  │  ├─ currency_id: found currency
   │  │  ├─ conversion_amount: VND
   │  │  ├─ conversion_rate: rate from Wise
   │  │  ├─ date: now()
   │  │  └─ metadata: links to expense
   │  │
   │  └─ Links expense to accounting transaction
   │
   ├─ Enqueues comment message (success or failure)
   └─ Returns 200 OK
   ↓
4. Worker processes comment posting
```

#### 3.1.3 DELETION Flow

```
1. User marks expense todo as incomplete (unchecks)
   └─ Basecamp sends webhook: DELETE /webhooks/basecamp/expense
      └─ Event: todo_uncompleted
   ↓
2. Handler: pkg/handler/webhook.UncheckBasecampExpense()
   ↓
3. Webhook Handler: pkg/handler/webhook.UncheckBasecampExpenseHandler()
   ├─ Calls extractExpenseData(msg)
   ├─ Unmarshals metadata
   ├─ Calls service.Basecamp.UncheckBasecampExpenseHandler(*obj)
   │
   └─ In service:
      ├─ Finds expense by basecamp_id
      ├─ Deletes expense record (soft delete)
      └─ Related accounting transaction remains (orphaned)
   │
   ├─ Enqueues comment message (success or failure)
   └─ Returns 200 OK
```

## 3.2 Basecamp Integration Points

### Basecamp Service Calls

| Operation | Basecamp API Call | Purpose |
|-----------|-------------------|---------|
| Get Todo | `service.Basecamp.Todo.Get(msg.Recording.URL)` | Fetch full details for assignment |
| Update Todo | `service.Basecamp.Todo.Update(projectID, todo)` | Assign to approver |
| Get Todo List | `service.Basecamp.Todo.GetList(msg.Recording.Parent.URL)` | Get parent category |
| Get Image URL | `service.Basecamp.Recording.TryToGetInvoiceImageURL()` | Extract receipt/invoice image |
| Post Comment | Worker → `service.Basecamp.BuildCommentMessage()` | Feedback comments |

### Basecamp Resources Used

**In Basecamp (Project: Woodland or Playground)**
- **Bucket (Project) ID:**
  - Prod: `9403032` (Woodland)
  - Dev: `12984857` (Fortress | Playground)
- **Todo List ID:**
  - Prod: `2353511928` (Expense)
  - Dev: `2436015405` (PlaygroundExpenseTodoID)
- **Assignee IDs:**
  - Prod approver: `21562923` (HanBasecampID)
  - Dev approver: `21581534` (NamNguyenBasecampID)

### Data Sent to Basecamp

**Comments Posted:**
- Validation success: "Your format looks good 👍"
- Validation failure: Error message with format instructions
- Creation success: "✓ Expense created successfully"
- Creation failure: "✗ Expense creation failed"
- Deletion success: "✓ Expense deleted successfully"
- Deletion failure: "✗ Deletion failed"

**Message Types:** `success` / `failed` / `` (default)

## 3.3 Database Models

### Primary Model: Expense
**Table:** `expenses`
**Fields:**
```
id (UUID)
employee_id (UUID) - FK to employees (who submitted)
amount (INT8) - original amount (stored as integer)
currency_id (UUID) - FK to currencies
currency (TEXT) - currency name (VND, USD, etc.)
reason (TEXT) - description of expense
issued_date (TIMESTAMPTZ) - when submitted
invoice_image_url (TEXT) - receipt/invoice image URL
metadata (JSON) - {basecamp_id, creator_id, etc.}
basecamp_id (INT) - Basecamp todo ID
accounting_transaction_id (UUID) - FK to accounting_transactions
created_at (TIMESTAMPTZ)
updated_at (TIMESTAMPTZ)
deleted_at (TIMESTAMPTZ) - soft delete
```

### Related Models
- **Employee:** Who submitted the expense
- **Currency:** Currency of expense
- **AccountingTransaction:** Linked accounting entry

### Relationships
```
Expense
├─ Employee (belongs to, created by)
├─ Currency (belongs to)
└─ AccountingTransaction (belongs to, one-to-one optional)
```

## 3.4 Code Files Involved

### Handler Layer
- **File:** `pkg/handler/webhook/basecamp_expense.go`
- **Methods:**
  - `ValidateBasecampExpense()` - HTTP handler for validation
  - `CreateBasecampExpense()` - HTTP handler for creation
  - `UncheckBasecampExpense()` - HTTP handler for deletion
  - `basecampExpenseValidate()` - Validation logic + todo assignment
  - `createBasecampExpense()` - Creation logic
  - `UncheckBasecampExpenseHandler()` - Deletion logic
  - `extractExpenseData()` - Parse title and fetch details

### Service Layer
- **File:** `pkg/service/basecamp/integration.go`
- **Methods:**
  - `CreateBasecampExpense(data)` - Main creation logic
  - `UncheckBasecampExpenseHandler(data)` - Deletion logic
  - `ExtractBasecampExpenseAmount()` - Parse amount with k/tr/m units

- **Subservices:**
  - `pkg/service/basecamp/todo/service.go`
  - `pkg/service/basecamp/recording/service.go`

### Store/Repository Layer
- **File:** `pkg/store/expense/expense.go`
- **Methods:**
  - `Create()` - Insert expense record
  - `Update()` - Update expense
  - `Delete()` - Soft delete expense
  - `GetByQuery()` - Find by basecamp_id

- **File:** `pkg/store/accounting/accounting.go`
- **Methods:**
  - `CreateTransaction()` - Create accounting entry

### Models
- `pkg/model/expense.go` - Expense struct
- `pkg/model/accounting.go` - AccountingTransaction struct

### Routing
- **File:** `pkg/routes/v1.go` (lines 73-78)
```go
expenseGroup := basecampGroup.Group("/expense")
{
    expenseGroup.POST("/validate", h.Webhook.ValidateBasecampExpense)
    expenseGroup.POST("", h.Webhook.CreateBasecampExpense)
    expenseGroup.DELETE("", h.Webhook.UncheckBasecampExpense)
}
```

## 3.5 Business Logic

### Amount Parsing Logic

**Supported Formats:**
- Plain: `500`, `1000`
- Thousands with `k`: `500k`, `500k500`
- Millions with `m` or `tr`: `5m`, `5m500k`, `5tr`, `5tr500k`

**Examples:**
- `500` → 500
- `500.000` → 500000 (dots removed)
- `500k` → 500,000
- `500k500` → 500,500
- `5m` → 5,000,000
- `5m500k` → 5,500,000
- `5tr` (trillion) → 5,000,000 (converted to m)

**Regex:** `(\\d+(k|tr|m)\\d+|\\d+(k|tr|m)|\\d+)`

### Validation Rules

1. **Title Format:**
   - Must have 3+ pipe-separated parts: `reason | amount | currency`
   - Reason: any text (trimmed)
   - Amount: any numeric format (k/tr/m supported)
   - Currency: VND or USD only (case-insensitive)

2. **Bucket Validation:**
   - Prod: must be "Woodland"
   - Dev: must be "Fortress | Playground"

3. **Amount Validation:**
   - Must parse to non-zero value
   - Returns 0 if parsing fails → validation error

4. **Image URL:**
   - Attempts to extract from todo attachments
   - Logs error but continues if not found

5. **Employee Lookup:**
   - Must find employee by basecamp_id (creator)
   - If not found → creation fails with error

6. **Currency Lookup:**
   - Must exist in currencies table
   - If not found → creation fails with error

### Categorization

All expenses automatically categorized as:
- **Category:** `Office Supply`
- **Type:** `OV` (Overhead/Variance)

No dynamic categorization (hardcoded).

### Error Handling

| Error Scenario | Validation | Creation |
|----------------|-----------|----------|
| Invalid title format | ✗ Fail + comment | N/A |
| Invalid amount | ✗ Fail + comment | N/A |
| Invalid currency | ✗ Fail + comment | N/A |
| Employee not found | ✓ Pass | ✗ Fail + comment |
| Currency not found | ✓ Pass | ✗ Fail + comment |
| Image URL missing | ✓ Pass | ✓ Pass (logged) |
| Conversion fails | ✓ Pass | ✓ Pass (logged as critical) |
| DB insert fails | N/A | ✗ Fail + comment |

**Pattern:** Failures always post comment to Basecamp for user feedback.

## 3.6 Migration Considerations

### Basecamp-Specific Dependencies
1. **Hardcoded Project/Todo IDs:**
   - Prod: Woodland (9403032), ExpenseTodoID (2353511928)
   - Dev: Playground (12984857), PlaygroundExpenseTodoID (2436015405)
   - Approver IDs: HanBasecampID, NamNguyenBasecampID

2. **Basecamp APIs:**
   - `Todo.Get()` - fetch for assignment
   - `Todo.Update()` - assign to approver
   - `Todo.GetList()` - get parent category
   - `Recording.TryToGetInvoiceImageURL()` - attachment handling
   - Comments system - feedback mechanism

3. **Amount Parsing:**
   - Custom logic for k/tr/m abbreviations
   - Specific to Basecamp title format
   - Could be simplified with number field in NocoDB

### Data That Would Change
| Current (Basecamp) | Future (NocoDB) | Migration Strategy |
|-------------------|-----------------|-------------------|
| Title parsing for amount | Number field | Direct number input |
| Title parsing for currency | Currency select/dropdown | Pre-selected in form |
| Reason in title | Reason text field | Direct text input |
| Image URL from attachments | URL/file field | Direct upload or URL paste |
| Approval comment feedback | Activity feed or status field | NocoDB update notifications |
| Todo assignment | Assignment field | NocoDB @mention or assign field |

### Simplifications Possible
1. **Form-based entry:** Replace Basecamp todo title parsing with NocoDB form
2. **Approval workflow:** Replace todo assignment with NocoDB workflow/status
3. **Feedback:** Replace Basecamp comments with NocoDB activity feed
4. **Image handling:** Replace attachment fetching with direct URL field

---

# 4. ON-LEAVE REQUEST FLOW

## 4.1 End-to-End Process

### Trigger Points
- **Validation Endpoint:** `POST /webhooks/basecamp/onleave/validate`
- **Approval Endpoint:** `POST /webhooks/basecamp/onleave`
- **Trigger Events:**
  - `todo_created` (user submits leave request)
  - `todo_completed` (approver marks as complete/approved)

### Process Flow

#### 4.1.1 VALIDATION Flow

```
1. User creates on-leave todo in Basecamp
   ├─ Location: Basecamp project (Woodland or Playground)
   ├─ Parent Group: OnLeave (OnleaveID: 6935836756 prod, 2243342506 dev)
   └─ Title format: "Name | Type | Date [| Shift]"
        Examples:
        - "John Doe | Off | 15/01/2025"
        - "Jane Smith | Remote | 10/01/2025 - 15/01/2025 | Morning"
   ↓
2. Basecamp sends webhook: POST /webhooks/basecamp/onleave/validate
   └─ Event: todo_created
   ↓
3. Handler: pkg/handler/webhook.ValidateOnLeaveRequest()
   ↓
4. Webhook Handler: pkg/handler/webhook.handleOnLeaveValidation()
   ├─ Validates environment: Prod only checks "Woodland" bucket
   │
   ├─ Calls validateOnLeaveData(msg)
   │  ├─ Calls parseOnLeaveDataFromMessage(nil, msg)
   │  │  ├─ Splits title by "|"
   │  │  ├─ Extracts: name, type, date range, shift
   │  │  └─ Returns OnLeaveData struct
   │  │
   │  ├─ Validates parent group ID
   │  │  ├─ Prod: OnleaveID (6935836756)
   │  │  └─ Dev: OnleavePlaygroundID (2243342506)
   │  │
   │  ├─ Validates title format (min 3 pipe parts)
   │  ├─ Validates off type: "off" or "remote" only
   │  ├─ Validates date range:
   │  │  ├─ Start date: not in past (or today is OK)
   │  │  ├─ End date: >= start date
   │  └─ Returns error if any validation fails
   │
   └─ Posts comment based on result:
      ├─ Success: "Your format looks good to go 👍"
      └─ Failure: Error message with format requirements
   ↓
5. Basecamp comment posted
```

#### 4.1.2 APPROVAL Flow

```
1. Approver marks on-leave todo as complete in Basecamp
   └─ Basecamp sends webhook: POST /webhooks/basecamp/onleave
      └─ Event: todo_completed
   ↓
2. Handler: pkg/handler/webhook.ApproveOnLeaveRequest()
   ↓
3. Webhook Handler: pkg/handler/webhook.handleApproveOnLeaveRequest()
   ├─ Fetches full todo details from Basecamp
   ├─ Calls parseOnLeaveDataFromMessage(todo, msg)
   │  ├─ Parses title and extracts data
   │  ├─ Reads todo assignees and description
   │  └─ Returns OnLeaveData with full details
   │
   ├─ Creates OnLeaveRequest record:
   │  ├─ type: parsed type (Off/Remote)
   │  ├─ start_date: parsed start date
   │  ├─ end_date: parsed end date
   │  ├─ shift: optional shift (Morning/Afternoon/etc)
   │  ├─ title: full todo title
   │  ├─ description: todo description (HTML cleaned)
   │  ├─ creator_id: UUID of request creator
   │  ├─ approver_id: UUID of approver (message creator)
   │  └─ assignee_ids: JSONB array of assignee UUIDs
   │
   ├─ Creates schedule entries in Basecamp (chunked by month)
   │  ├─ Title format: "⚠️ ORIGINAL_TITLE"
   │  ├─ Schedule: Woodland (prod) or Playground (dev)
   │  ├─ Marks as all-day event
   │  └─ Sets subscribers: assignees + ops team
   │      ├─ Prod ops: HuyNguyenBasecampID, GiangThanBasecampID
   │      └─ Dev ops: NamNguyenBasecampID
   │
   └─ Returns 200 OK (or error logged)
   ↓
4. Schedule created in Basecamp, OnLeaveRequest stored in DB
```

## 4.2 Basecamp Integration Points

### Basecamp Service Calls

| Operation | Basecamp API Call | Purpose |
|-----------|-------------------|---------|
| Get Todo | `service.Basecamp.Todo.Get(url)` | Fetch full details for approval |
| Create Schedule | `service.Basecamp.Schedule.CreateScheduleEntry()` | Add to team calendar |
| Subscribe to Event | `service.Basecamp.Subscription.Subscribe()` | Set event subscribers |

### Basecamp Resources Used

**In Basecamp (Project: Woodland or Playground)**
- **Bucket (Project) ID:**
  - Prod: `9403032` (Woodland)
  - Dev: `12984857` (Fortress | Playground)
- **Parent Group (OnLeave):**
  - Prod: `6935836756`
  - Dev: `2243342506`
- **Schedule ID:**
  - Prod: `1346305137` (Woodland schedule)
  - Dev: `1941398077` (Playground schedule)
- **Ops Team Members (subscribers):**
  - Prod: `22658825` (HuyNguyenBasecampID), `26160802` (GiangThanBasecampID)
  - Dev: `21581534` (NamNguyenBasecampID)

### Data Sent to Basecamp

**Schedule Entry Created:**
```
Title: "⚠️ [ORIGINAL_TODO_TITLE]"
AllDay: true
StartsAt: start_date (ISO 8601)
EndsAt: end_date (ISO 8601)
Description: todo description
```

**Subscribers:**
- Request creator
- Request assignees
- Ops team members

## 4.3 Database Models

### Primary Model: OnLeaveRequest
**Table:** `on_leave_requests`
**Fields:**
```
id (UUID)
type (TEXT) - "Off" or "Remote"
start_date (DATE)
end_date (DATE)
shift (TEXT) - optional (Morning, Afternoon, Full Day, etc.)
title (TEXT) - full todo title
description (TEXT) - todo description (HTML cleaned)
creator_id (UUID) - FK to employees (who requested)
approver_id (UUID) - FK to employees (who approved)
assignee_ids (JSONB) - array of employee UUIDs
created_at (TIMESTAMP)
updated_at (TIMESTAMP)
deleted_at (TIMESTAMP) - soft delete
```

### Related Models
- **Employee:** Creator, approver, and assignees

### Relationships
```
OnLeaveRequest
├─ Creator (Employee, belongs to)
├─ Approver (Employee, belongs to)
└─ Assignees (Employee array via JSONB)
```

## 4.4 Code Files Involved

### Handler Layer
- **File:** `pkg/handler/webhook/onleave.go`
- **Methods:**
  - `ValidateOnLeaveRequest()` - HTTP handler for validation
  - `ApproveOnLeaveRequest()` - HTTP handler for approval
  - `handleOnLeaveValidation()` - Validation logic
  - `validateOnLeaveData()` - Data validation
  - `handleApproveOnLeaveRequest()` - Approval logic
  - `parseOnLeaveDataFromMessage()` - Parse title and todo

### Service Layer
- **File:** `pkg/service/basecamp/schedule/service.go`
- **Methods:**
  - `service.Basecamp.Schedule.CreateScheduleEntry()` - Create schedule
  - `service.Basecamp.Subscription.Subscribe()` - Add subscribers

- **File:** `pkg/service/basecamp/todo/service.go`
- **Methods:**
  - `service.Basecamp.Todo.Get()` - Fetch full todo

### Store/Repository Layer
- **File:** `pkg/store/onleaverequest/onleave_request.go`
- **Methods:**
  - `Create()` - Insert on-leave request
  - `All()` - List on-leave requests

### Models
- `pkg/model/onleave_request.go` - OnLeaveRequest struct

### Routing
- **File:** `pkg/routes/v1.go` (lines 84-88)
```go
onLeaveGroup := basecampGroup.Group("/onleave")
{
    onLeaveGroup.POST("/validate", h.Webhook.ValidateOnLeaveRequest)
    onLeaveGroup.POST("", h.Webhook.ApproveOnLeaveRequest)
}
```

## 4.5 Business Logic

### Title Parsing Logic

**Format:** `Name | Type | DateRange [| Shift]`

**Sections (pipe-separated):**
1. **Name:** Employee name
2. **Type:** "Off" or "Remote"
3. **DateRange:** Single date or range
   - Single: `15/01/2025` → same day
   - Range: `15/01/2025 - 20/01/2025`
4. **Shift:** (Optional) Morning, Afternoon, Full Day, etc.

**Examples:**
- `John Doe | Off | 15/01/2025`
- `Jane Smith | Remote | 10/01/2025 - 15/01/2025`
- `Bob Johnson | Off | 20/01/2025 | Morning`

**Date Format:** `DD/MM/YYYY`

### Validation Rules

1. **Title Format:**
   - Must have >= 3 pipe-separated parts
   - Each part trimmed

2. **Off Type Validation:**
   - Case-insensitive check
   - Allowed: "off", "remote" only
   - Rejects: "sick", "vacation", "work-from-home", etc.

3. **Parent Group Validation:**
   - Must be under OnLeave group
   - Prod: 6935836756
   - Dev: 2243342506
   - Checks both direct parent and via GetList() API call

4. **Date Range Validation:**
   - Start date: not in past (today is OK via `IsSameDay()`)
   - End date: >= start date
   - Single date: end_date = start_date
   - Supports range with "-" separator

5. **Environment Check:**
   - Prod: only processes "Woodland" bucket
   - Dev: processes all buckets

### Employee Lookup

1. **Creator Lookup:**
   - Uses `msg.Recording.Creator.ID` (Basecamp ID)
   - Finds in employees table by `basecamp_id`

2. **Approver Lookup:**
   - Uses `msg.Creator.ID` (Basecamp ID of todo completer)
   - Finds in employees table by `basecamp_id`

3. **Assignee Lookup:**
   - From todo.Assignees list (Basecamp user objects)
   - Maps Basecamp IDs to employee UUIDs

### Schedule Creation Logic

```go
// For each month in date range:
dateChunks := timeutil.ChunkDateRange(startDate, endDate)
for _, dateChunk := range dateChunks {
    createScheduleEntry(
        title: "⚠️ " + originalTitle,
        startsAt: dateChunk[0],
        endsAt: dateChunk[1],
        allDay: true,
        description: todoDescription,
        projectID: woodlandID,
        scheduleID: woodlandScheduleID,
    )
    
    // Set subscribers
    subscribe(
        subscribers: assigneeIDs + completionSubscriberIDs + opsTeamIDs
    )
}
```

**Purpose:** Breaks multi-month leave requests into monthly calendar entries (Basecamp API limitation).

### Error Handling

| Error Scenario | Validation | Approval |
|----------------|-----------|----------|
| Invalid format | ✗ Error comment | N/A |
| Wrong parent group | ✗ Error comment | N/A |
| Invalid off type | ✗ Error comment | N/A |
| Date in past | ✗ Error comment | N/A |
| Invalid date range | ✗ Error comment | N/A |
| Creator not found | ✓ Pass | ✗ Error logged, continues |
| Approver not found | ✓ Pass | ✗ Error logged, continues |
| Assignee not found | ✓ Pass | ✗ Error logged, continues |
| Schedule creation fails | N/A | ✗ Error logged, continues |
| DB insert fails | N/A | ✗ Error logged |

**Pattern:** Returns 200 OK for webhooks even on failure; errors logged and sometimes commented.

## 4.6 Migration Considerations

### Basecamp-Specific Dependencies
1. **Hardcoded IDs:**
   - Prod: OnleaveID (6935836756), WoodlandID (9403032), WoodlandScheduleID (1346305137)
   - Dev: OnleavePlaygroundID (2243342506), PlaygroundID (12984857), PlaygroundScheduleID (1941398077)
   - Approver/Ops IDs: various HuyNguyenBasecampID, etc.

2. **Basecamp APIs:**
   - `Todo.Get()` - fetch full details
   - `Schedule.CreateScheduleEntry()` - create calendar events
   - `Subscription.Subscribe()` - set attendees

3. **Title Parsing:**
   - Complex regex for date ranges
   - Hard to read from todo title; could be form fields

### Data That Would Change
| Current (Basecamp) | Future (NocoDB) | Migration Strategy |
|-------------------|-----------------|-------------------|
| Title parsing for name | Employee select field | Dropdown of employees |
| Title parsing for type | Type select (Off/Remote) | Pre-selected in form |
| Title parsing for dates | Date range picker | Native date fields |
| Title parsing for shift | Shift select field | Optional dropdown |
| Todo description | Description field | Text area in NocoDB |
| Todo completion = approval | Status field (Approved) | Workflow status change |
| Schedule creation | Calendar integration | NocoDB calendar view or external sync |

### Simplifications Possible
1. **Form-based entry:** Replace title parsing with dedicated form fields
2. **Employee sync:** NocoDB can preload employee dropdown
3. **Approval workflow:** Replace todo completion with status field
4. **Calendar sync:** Use NocoDB calendar view or webhook to external calendar
5. **Better UX:** Separate fields are clearer than pipe-delimited format

---

# CROSS-FLOW INTEGRATION POINTS

## Shared Infrastructure

### 1. Basecamp Service Package
- **Location:** `pkg/service/basecamp/`
- **Core Components:**
  - `basecamp.go` - Main service struct
  - `integration.go` - Expense & accounting logic
  - `consts/consts.go` - All hardcoded IDs
  - Subservices: `todo/`, `comment/`, `recording/`, `schedule/`, `subscription/`

### 2. Basecamp Webhook Message Structure
```go
type BasecampWebhookMessage struct {
    Kind       string                    // "todo_created", "todo_completed", etc.
    Recording  BasecampWebhookRecording  // The actual object
    Creator    BasecampWebhookPerson     // Who triggered the event
}

type BasecampWebhookRecording struct {
    ID         int
    Title      string
    Type       string
    URL        string
    Creator    BasecampWebhookPerson
    Bucket     BasecampBucket
    Parent     BasecampParent
    UpdatedAt  time.Time
}
```

### 3. Worker/Queue System
- **Type:** Background job queue
- **Usage:** Asynchronous comment posting to Basecamp
- **Enqueued Messages:**
  - `BasecampCommentMsg` - post comment to todo
  - Contains: bucket ID, recording ID, message text, message type

### 4. Currency Conversion
- **Service:** `service.Wise`
- **Used by:** Invoice, Expense, Accounting flows
- **Function:** Convert any currency → VND
- **Stores:** conversion_amount, conversion_rate

### 5. Employee/Basecamp ID Mapping
- **Store:** `pkg/store/employee/`
- **Methods:**
  - `OneByBasecampID(bcID int)` - Get employee by Basecamp ID
  - `GetByBasecampIDs([]int)` - Get multiple employees
- **Used by:** All flows for user identification

### 6. Accounting Transaction Creation
- **Common pattern across flows:**
  - Invoice paid → Income transaction
  - Expense created → Overhead transaction
  - Accounting todo → Operation transaction
- **Metadata pattern:** `{source: "source_type", id: source_id}`

## Environment-Specific Constants

```
PRODUCTION (env == "prod"):
├─ Woodland: 9403032 (main project)
├─ Accounting: 15258324
├─ OnLeave: 6935836756
├─ Approvers: HanBasecampID, HuyNguyenBasecampID, etc.
└─ Schedules: 1346305137, etc.

DEVELOPMENT/LOCAL (env != "prod"):
├─ Playground: 12984857 (dev project)
├─ OnLeave: 2243342506
├─ Approvers: NamNguyenBasecampID, etc.
└─ Schedules: 1941398077, etc.
```

## Webhook Routing Pattern
```
/webhooks/basecamp/
├─ /expense
│  ├─ POST /validate
│  ├─ POST (create)
│  └─ DELETE (uncheck)
├─ /operation
│  ├─ POST /accounting-transaction
│  └─ PUT /invoice
└─ /onleave
   ├─ POST /validate
   └─ POST (approve)
```

---

# MIGRATION CONSIDERATIONS

## Summary Table: Basecamp → NocoDB Migration Impact

| Aspect | Invoice | Accounting | Expense | On-Leave |
|--------|---------|-----------|---------|----------|
| **Complexity** | High | Medium | High | Very High |
| **API Calls** | 4 types | 1 type | 5 types | 3 types |
| **Hardcoded IDs** | 4 | 2 | 4 | 6 |
| **Regex Parsing** | 1 pattern | 1 pattern | 1 pattern | 2 patterns |
| **Form Entry** | Basecamp todo | Basecamp todo | Basecamp todo | Basecamp todo |
| **Approval Method** | Comment validation | N/A | Auto-pass | Todo completion |
| **Calendar Sync** | None | None | None | Basecamp schedule |

## General Migration Strategy

### Phase 1: Infrastructure
1. Set up NocoDB tables/views for each entity
2. Create NocoDB webhooks endpoints
3. Map Basecamp IDs to NocoDB record IDs
4. Configure n8n workflows (if using automation)

### Phase 2: Per-Flow Migration
1. **Invoice:** Replace Basecamp API calls with NocoDB queries
   - Title parsing → direct field reads
   - Comment validation → status field check
   - Todo completion → status update

2. **Accounting:** Simplify significantly
   - Replace regex parsing with form fields
   - Remove todo list API call
   - Direct amount/currency field reads

3. **Expense:** Form-based entry
   - Replace title parsing with form submission
   - Validation via form constraints
   - Approval via workflow status

4. **On-Leave:** Biggest simplification
   - Form-based entry (name, type, dates, shift)
   - Approval via workflow status
   - Calendar integration via n8n/workflow

### Phase 3: Testing & Cutover
1. Run parallel with Basecamp (dual writes)
2. Test each flow end-to-end
3. Migrate historical data
4. Disable Basecamp webhooks
5. Monitor for issues

## Critical Data Preservation

### Must Migrate:
1. **Invoice records:** All invoice data + status history
2. **Accounting transactions:** Full transaction history with metadata
3. **Expense records:** All expenses with basecamp_id mapping
4. **On-leave requests:** Request history and approvals
5. **Currency conversions:** Historical rates preserved in metadata

### Must Update:
1. Basecamp ID references (if keeping dual system)
2. Links between related records (e.g., expense → accounting transaction)
3. Metadata JSON fields (source and ID)

### Can Drop:
1. Temporary webhook parsing strings
2. Basecamp service layer (no longer needed)
3. Basecamp subservices (todo, comment, recording, etc.)

## Implementation Notes

### 1. NocoDB Table Setup
Create tables for:
- Invoices (with status, approval tracking)
- Accounting Transactions (with metadata)
- Expenses (with approval tracking)
- On-Leave Requests (with approval, calendar entries)

### 2. Workflow Setup
Replace Basecamp logic with:
- **n8n workflows** for complex logic (especially on-leave calendar)
- **NocoDB automations** for simple triggers
- **Webhooks** for event notifications

### 3. Approval Workflow
```
Basecamp (current):
  todo_created → validation comment
  todo_completed → approval
  
NocoDB (future):
  record_created → validation via formula/automation
  status_updated → approval workflow
  approval_done → calendar/notification webhook
```

### 4. Calendar Integration
- **Invoice:** None needed
- **Accounting:** None needed
- **Expense:** None needed
- **On-Leave:** CRITICAL
  - Current: Basecamp schedule entries
  - Future: NocoDB calendar view or external calendar (Google Calendar via n8n)

### 5. Notification System
- **Basecamp comments:** Were internal communication
- **Future:** Use NocoDB activity feed + Discord/email webhooks
- **n8n:** Can post to Discord, send emails, update external calendars

## Testing Checklist

- [ ] Invoice workflow: Create, mark as paid, verify accounting
- [ ] Accounting workflow: Create transaction, verify conversion
- [ ] Expense workflow: Create, approve, verify accounting link
- [ ] On-Leave workflow: Create, approve, verify calendar
- [ ] Currency conversion: Test multiple currencies
- [ ] Error handling: Test all validation paths
- [ ] Data integrity: Verify metadata links preserved
- [ ] Performance: Test with bulk operations
- [ ] Audit trail: Verify all changes logged

---

## CONCLUSION

This migration represents a move from **Basecamp-driven workflows** to **NocoDB-driven workflows**. 

### Key Takeaways:

1. **Invoice Flow** is tightly coupled to Basecamp for approval validation
2. **Accounting Flow** is simple read-only from Basecamp  
3. **Expense Flow** uses Basecamp as entry point with complex parsing
4. **On-Leave Flow** is most complex, creates calendar events in Basecamp

**Simplification Opportunity:** Form-based entry (instead of Basecamp title parsing) will significantly improve UX and reduce code complexity.

**Critical Success Factor:** Approval workflow redesign is essential. NocoDB status fields + automation rules can replace Basecamp todo completion.

**Estimated Effort:**
- Invoice: 3-4 weeks (API replacement + testing)
- Accounting: 1-2 weeks (simplification)
- Expense: 2-3 weeks (form redesign)
- On-Leave: 3-4 weeks (calendar integration critical)

**Total:** 10-13 weeks for full migration with parallel testing.

