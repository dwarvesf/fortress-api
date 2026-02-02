# Contractor Invoice Generation Specification

## Document Information
- **Version**: 1.0
- **Date**: 2025-12-24
- **Status**: Draft
- **Author**: System

## Table of Contents
1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Database Schema](#database-schema)
4. [API Design](#api-design)
5. [Business Logic](#business-logic)
6. [Data Flow](#data-flow)
7. [Technical Implementation](#technical-implementation)
8. [Error Handling](#error-handling)
9. [Testing Strategy](#testing-strategy)
10. [Security Considerations](#security-considerations)

---

## 1. Overview

### 1.1 Purpose
Create an automated contractor invoice generation system that:
- Fetches contractor rate information from Notion Contractor Rates database
- Retrieves work logs from Notion Task Order Log database
- Generates professional PDF invoices based on billing type (Monthly Fixed or Hourly Rate)
- Uploads invoices to Google Drive
- Returns invoice metadata and file URL via REST API

### 1.2 Background
The Dwarves Foundation manages contractors through Notion databases. Currently, there is no automated way to generate invoices for contractors based on their work logs and billing agreements. This feature will integrate with existing Notion databases to automate invoice generation.

### 1.3 Scope
**In Scope:**
- REST API endpoint for invoice generation
- Integration with Notion Contractor Rates database (2c464b29b84c805bbcdedc052e613f4d)
- Integration with Notion Task Order Log database (2b964b29b84c801caccbdc8ca1e38a5f)
- PDF generation with conditional rendering based on billing type
- Google Drive upload for invoice storage
- Support for Monthly Fixed and Hourly Rate billing types
- VND and USD currency support

**Out of Scope:**
- Email delivery of invoices (future enhancement)
- Invoice payment tracking (use existing invoice system)
- Multi-currency conversion (use existing Wise integration)
- Invoice editing or regeneration (future enhancement)
- Support for Mixed, Retainer, Milestone-based billing types (future)

---

## 2. Requirements

### 2.1 Functional Requirements

#### FR-1: Invoice Generation Endpoint
**Priority**: P0 (Critical)

The system shall provide a REST API endpoint `POST /api/v1/invoices/contractor/generate` that:
- Accepts contractor identifier (Discord username) and month (YYYY-MM format)
- Validates input parameters
- Queries contractor rate information from Notion
- Queries work logs from Notion Task Order Log
- Generates PDF invoice
- Uploads PDF to Google Drive
- Returns invoice metadata and file URL

**Input Validation:**
- `contractorDiscord`: Required, non-empty string
- `month`: Required, must match regex `^\d{4}-\d{2}$` (YYYY-MM format)

**Success Response:**
```json
{
  "invoiceNumber": "INVC-adeki_-202512-001",
  "contractorName": "adeki_",
  "month": "2025-12",
  "billingType": "Monthly Fixed",
  "currency": "VND",
  "total": 48000000,
  "pdfFileUrl": "https://drive.google.com/file/d/...",
  "generatedAt": "2025-12-24T10:30:00Z"
}
```

#### FR-2: Monthly Fixed Billing Type Support
**Priority**: P0 (Critical)

For contractors with "Monthly Fixed" billing type, the system shall:
- Display line items with Project Name and Proof of Work only
- **NOT** show price, quantity, or amount per line item
- Display total as the "Monthly Fixed" amount from Contractor Rates
- Formula: `Monthly Fixed = Gross Fixed - Total Local`

**Example Invoice Structure:**
```
NO | PROJECT          | PROOF OF WORK
1  | Project Alpha    | Implemented user authentication module
2  | Project Beta     | Fixed payment gateway integration bugs
3  | Project Gamma    | Updated API documentation

                       Total: 48,000,000 VND
```

#### FR-3: Hourly Rate Billing Type Support
**Priority**: P0 (Critical)

For contractors with "Hourly Rate" billing type, the system shall:
- Display line items with Project Name, Proof of Work, Hours, Rate, and Amount
- Calculate amount per line: `Amount = Hours × Hourly Rate`
- Display total as sum of all line item amounts

**Example Invoice Structure:**
```
NO | PROJECT       | PROOF OF WORK           | HOURS | RATE    | TOTAL
1  | Project Alpha | User authentication     | 8     | $30     | $240
2  | Project Beta  | Payment integration     | 6     | $30     | $180
3  | Project Gamma | API documentation       | 4     | $30     | $120

                                               Total: $540
```

#### FR-4: Notion Data Retrieval
**Priority**: P0 (Critical)

The system shall query Notion databases with the following logic:

**Contractor Rates Query:**
- Database ID: `2c464b29b84c805bbcdedc052e613f4d`
- Filter by Discord username (rollup contains)
- Filter by date range: `Start Date <= month AND (End Date >= month OR End Date is empty)`
- Extract: Billing Type, Monthly Fixed, Hourly Rate, Currency, Contractor Page ID

**Task Order Log Query:**
- Database ID: `2b964b29b84c801caccbdc8ca1e38a5f`
- Find Order entry for contractor and month
- Query subitems (Type="Timesheet", Parent item = Order ID)
- Extract: Project Name, Line Item Hours, Proof of Works

#### FR-5: PDF Generation
**Priority**: P0 (Critical)

The system shall generate PDF invoices with:
- Professional layout matching contractor-invoice-template.html
- Dynamic table structure based on billing type
- Proper currency formatting (VND: no decimals, USD: 2 decimals)
- Invoice number format: `INVC-{discord}-{YYYYMM}-001`
- Invoice date: First day of the month
- Due date: Last day of the month

#### FR-6: Google Drive Upload
**Priority**: P1 (High)

The system shall upload generated PDFs to Google Drive with:
- Parent folder ID: Read from env `CONTRACTOR_INVOICE_DIR_ID` (default: `1_9Ai9erlvc39vDMoYCswroQxWzajKUX2`)
- Subfolder: `{contractor_full_name}/` (create if not exists)
- File naming: `{invoice_number}.pdf` (e.g., `INVC-202512-A7K9.pdf`)
- Full path: `{CONTRACTOR_INVOICE_DIR_ID}/{contractor_full_name}/INVC-202512-A7K9.pdf`
- Public access URL returned in response

### 2.2 Non-Functional Requirements

#### NFR-1: Performance
- Invoice generation should complete within 10 seconds
- Support pagination for querying Notion (100 items per page)
- Efficient property extraction from Notion responses

#### NFR-2: Reliability
- Handle Notion API rate limits gracefully
- Retry failed Google Drive uploads (up to 3 attempts)
- Comprehensive error logging at DEBUG level

#### NFR-3: Security
- Require authentication via JWT token
- Require `PermissionInvoicesCreate` permission
- Validate all input parameters
- Sanitize data before PDF rendering

#### NFR-4: Maintainability
- Follow existing codebase patterns (Handler → Controller → Service)
- Add comprehensive DEBUG logging at every step
- Use type-safe structs for all data transfers
- Document complex business logic with comments

---

## 3. Database Schema

### 3.1 Notion Contractor Rates Database

**Database ID**: `2c464b29b84c805bbcdedc052e613f4d`

**Properties:**

| Property Name | Type | Description | Example |
|---------------|------|-------------|---------|
| Name | Title | Auto-generated name | "Rate adeki_ :: 2025 Dec" |
| Contractor | Relation | Link to Contractor database | → Contractor page |
| Discord | Rollup | Contractor's Discord username | "adeki_" |
| Billing Type | Select | Type of billing | "Monthly Fixed", "Hourly Rate" |
| Gross Fixed | Number | Total monthly amount | 48000000 |
| Total Local | Number | Local payment portion | 0 |
| Monthly Fixed | Formula | Gross Fixed - Total Local | 48000000 |
| Hourly Rate | Number | Rate per hour | 30 |
| Currency | Select | Currency code | "VND", "USD" |
| Start Date | Date | Contract start date | 2025-12-15 |
| End Date | Date | Contract end date | null (ongoing) |
| Status | Status | Contract status | "Active", "Archived" |
| Payday | Select | Payment day of month | "01", "15" |

**Sample Data:**

```json
{
  "Name": "Rate adeki_ :: 2025 Dec",
  "Discord": "adeki_",
  "Billing Type": "Monthly Fixed",
  "Gross Fixed": 48000000,
  "Total Local": 0,
  "Monthly Fixed": 48000000,
  "Currency": "VND",
  "Start Date": "2025-12-15",
  "End Date": null,
  "Status": "Active"
}
```

### 3.2 Notion Task Order Log Database

**Database ID**: `2b964b29b84c801caccbdc8ca1e38a5f`

**Properties:**

| Property Name | Type | Description | Example |
|---------------|------|-------------|---------|
| Name | Title | Auto-generated name | "Order adeki_ :: 2025 Dec" |
| Type | Select | Entry type | "Order", "Timesheet" |
| Status | Select | Approval status | "Draft", "Approved", "Completed" |
| Date | Date | Work date | 2025-12-01 |
| Month | Formula | YYYY-MM from date | "2025-12" |
| Deployment | Relation | Link to deployment | → Deployment page |
| Contractor | Rollup | From Deployment | → Contractor page |
| Project | Rollup | From Deployment | → Project page |
| Line Item Hours | Number | Hours for timesheet | 8 |
| Proof of Works | Rich Text | Work description | "Implemented authentication" |
| Sub-item | Relation | Child timesheet entries | [→ Timesheet 1, → Timesheet 2] |
| Parent item | Relation | Parent order entry | → Order page |
| Subtotal Hours | Rollup | Sum of subitem hours | 24 |

**Data Structure:**

```
Order (Type="Order")
├── Timesheet 1 (Type="Timesheet", Parent item=Order)
│   ├── Project: "Project Alpha"
│   ├── Line Item Hours: 8
│   └── Proof of Works: "Implemented user authentication"
├── Timesheet 2 (Type="Timesheet", Parent item=Order)
│   ├── Project: "Project Beta"
│   ├── Line Item Hours: 6
│   └── Proof of Works: "Fixed payment bugs"
└── Timesheet 3 (Type="Timesheet", Parent item=Order)
    ├── Project: "Project Gamma"
    ├── Line Item Hours: 4
    └── Proof of Works: "Updated documentation"
```

---

## 4. API Design

### 4.1 Endpoint Specification

**Endpoint**: `POST /api/v1/invoices/contractor/generate`

**Authentication**: Required (JWT token)

**Authorization**: `PermissionInvoicesCreate`

**Request Body:**
```json
{
  "contractorDiscord": "adeki_",
  "month": "2025-12"
}
```

**Request Schema:**
```go
type GenerateContractorInvoiceRequest struct {
    ContractorDiscord string `json:"contractorDiscord" binding:"required"`
    Month             string `json:"month" binding:"required"` // YYYY-MM format
}
```

**Success Response (200 OK):**
```json
{
  "data": {
    "invoiceNumber": "INVC-202512-A7K9",
    "contractorName": "adeki_",
    "month": "2025-12",
    "billingType": "Monthly Fixed",
    "currency": "VND",
    "total": 48000000,
    "pdfFileUrl": "https://drive.google.com/file/d/ABC123/view",
    "generatedAt": "2025-12-24T10:30:00Z"
  },
  "error": null,
  "message": null,
  "pagination": null
}
```

**Error Response (400 Bad Request):**
```json
{
  "data": null,
  "error": "invalid month format, expected YYYY-MM",
  "message": "Validation failed",
  "pagination": null
}
```

**Error Response (404 Not Found):**
```json
{
  "data": null,
  "error": "contractor rates not found for the specified month",
  "message": "No active contractor rate found",
  "pagination": null
}
```

**Error Response (500 Internal Server Error):**
```json
{
  "data": null,
  "error": "failed to generate PDF",
  "message": "Internal server error",
  "pagination": null
}
```

### 4.2 Response Model

```go
type ContractorInvoiceResponse struct {
    InvoiceNumber  string  `json:"invoiceNumber"`  // INVC-{YYYYMM}-{random-4-chars}
    ContractorName string  `json:"contractorName"` // Discord username
    Month          string  `json:"month"`          // YYYY-MM
    BillingType    string  `json:"billingType"`    // "Monthly Fixed" or "Hourly Rate"
    Currency       string  `json:"currency"`       // "VND" or "USD"
    Total          float64 `json:"total"`          // Total invoice amount
    PDFFileURL     string  `json:"pdfFileUrl"`     // Google Drive public URL
    GeneratedAt    string  `json:"generatedAt"`    // RFC3339 timestamp
}
```

---

## 5. Business Logic

### 5.1 Invoice Number Generation

**Format**: `INVC-{YYYYMM}-{random-4-chars}`

**Algorithm**:
1. Extract year and month from request (remove hyphen) → YYYYMM
2. Generate 4 random alphanumeric characters (uppercase)
3. Combine into format

**Random Character Generation**:
- Use cryptographically secure random generator
- Character set: A-Z, 0-9 (36 characters)
- Length: 4 characters
- Total combinations: 36^4 = 1,679,616 unique IDs per month

**Examples**:
- Month: "2025-12" → `INVC-202512-A7K9`
- Month: "2025-11" → `INVC-202511-X3M2`
- Month: "2025-12" → `INVC-202512-P5Q1`

### 5.2 Date Calculation

**Invoice Date**: First day of the specified month
```go
invoiceDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
```

**Due Date**: Last day of the specified month
```go
dueDate := invoiceDate.AddDate(0, 1, -1) // Add 1 month, subtract 1 day
```

**Example**:
- Month: "2025-12"
- Invoice Date: 2025-12-01
- Due Date: 2025-12-31

### 5.3 Line Item Construction

#### 5.3.1 Monthly Fixed Billing

**Pseudocode**:
```
FOR EACH subitem IN taskOrderSubitems:
    lineItem = {
        ProjectName: subitem.ProjectName,
        Description: subitem.ProofOfWork,
        Hours: 0,        // Not displayed
        Rate: 0,         // Not displayed
        Amount: 0        // Not displayed
    }
    lineItems.append(lineItem)

total = contractorRate.MonthlyFixed
```

**Output**:
```json
{
  "lineItems": [
    {
      "projectName": "Project Alpha",
      "description": "Implemented user authentication module"
    },
    {
      "projectName": "Project Beta",
      "description": "Fixed payment gateway integration bugs"
    }
  ],
  "total": 48000000
}
```

#### 5.3.2 Hourly Rate Billing

**Pseudocode**:
```
total = 0
FOR EACH subitem IN taskOrderSubitems:
    amount = subitem.Hours × contractorRate.HourlyRate
    lineItem = {
        ProjectName: subitem.ProjectName,
        Description: subitem.ProofOfWork,
        Hours: subitem.Hours,
        Rate: contractorRate.HourlyRate,
        Amount: amount
    }
    lineItems.append(lineItem)
    total += amount
```

**Output**:
```json
{
  "lineItems": [
    {
      "projectName": "Project Alpha",
      "description": "Implemented authentication",
      "hours": 8,
      "rate": 30,
      "amount": 240
    },
    {
      "projectName": "Project Beta",
      "description": "Fixed payment bugs",
      "hours": 6,
      "rate": 30,
      "amount": 180
    }
  ],
  "total": 420
}
```

### 5.4 Currency Formatting

**VND (Vietnamese Dong)**:
- No decimal places
- Thousands separator: comma
- Format: `48,000,000`

**USD (US Dollar)**:
- 2 decimal places
- Thousands separator: comma
- Dollar sign prefix
- Format: `$1,234.56`

**Implementation**:
```go
import "github.com/Rhymond/go-money"

pound := money.New(1, currencyCode)
tmpValue := amount * math.Pow(10, float64(pound.Currency().Fraction))
result = pound.Multiply(int64(tmpValue)).Display()
```

---

## 6. Data Flow

### 6.1 High-Level Flow

```
Client Request
      ↓
[API Gateway]
      ↓
[Authentication & Authorization Middleware]
      ↓
[Invoice Handler]
      ↓
[Invoice Controller]
      ├─→ [Contractor Rates Service] → Notion API
      ├─→ [Task Order Log Service] → Notion API
      ├─→ [PDF Generator] → wkhtmltopdf
      └─→ [Google Drive Service] → Google Drive API
      ↓
[Response with PDF URL]
```

### 6.2 Detailed Sequence Diagram

```
Client                Handler              Controller           NotionService        GoogleDrive
  |                      |                      |                      |                   |
  |-- POST request ---→  |                      |                      |                   |
  |                      |-- Validate --------→ |                      |                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Query rates -----→ |                   |
  |                      |                      |                      |-- API call --→ Notion
  |                      |                      |                      |← rate data ---|
  |                      |                      |← ContractorRate ----|                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Query order -----→ |                   |
  |                      |                      |                      |-- API call --→ Notion
  |                      |                      |                      |← order data ---|
  |                      |                      |← OrderPageID -------|                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Query subitems --→ |                   |
  |                      |                      |                      |-- API call --→ Notion
  |                      |                      |                      |← subitems -----|
  |                      |                      |← Subitems ----------|                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Build line items   |                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Generate PDF       |                   |
  |                      |                      |(wkhtmltopdf)         |                   |
  |                      |                      |                      |                   |
  |                      |                      |-- Upload PDF -----------------------------→ |
  |                      |                      |                      |                   |-- Upload
  |                      |                      |                      |                   |← File URL
  |                      |                      |← PDF URL --------------------------------|
  |                      |                      |                      |                   |
  |                      |← InvoiceData -------|                      |                   |
  |                      |                      |                      |                   |
  |← 200 OK with data ---|                      |                      |                   |
```

### 6.3 Error Flow

```
Client                Handler              Controller           NotionService
  |                      |                      |                      |
  |-- POST request ---→  |                      |                      |
  |                      |-- Validate --------→ |                      |
  |                      |                      |                      |
  |                      |                      |-- Query rates -----→ |
  |                      |                      |                      |-- API call --→ Notion
  |                      |                      |                      |← 404 Not Found
  |                      |                      |← Error: Not Found --|
  |                      |                      |                      |
  |                      |← Error -------------|                      |
  |                      |                      |                      |
  |← 404 Not Found -----|                      |                      |
```

---

## 7. Technical Implementation

### 7.1 File Structure

```
pkg/
├── config/
│   └── config.go                              # Add ContractorRates DB ID
├── service/
│   ├── notion/
│   │   ├── contractor_rates.go                # NEW: Query contractor rates
│   │   └── task_order_log.go                  # MODIFY: Add QueryOrderSubitems()
│   └── googledrive/
│       └── google_drive.go                    # MODIFY: Add UploadContractorInvoicePDF()
├── controller/
│   └── invoice/
│       └── contractor_invoice.go              # NEW: Business logic
├── handler/
│   └── invoice/
│       ├── contractor_invoice.go              # NEW: HTTP handler
│       ├── request/
│       │   └── contractor_invoice.go          # NEW: Request model
│       └── errs/
│           └── errors.go                      # MODIFY: Add errors
├── view/
│   └── contractor_invoice.go                  # NEW: Response model
├── routes/
│   └── v1.go                                  # MODIFY: Add endpoint
└── templates/
    └── contractor-invoice-template.html       # MODIFY: Add conditionals
```

### 7.2 Configuration

**File**: `pkg/config/config.go`

**Changes**:
```go
type NotionDatabase struct {
    // ... existing fields ...
    TaskOrderLog    string
    Contractor      string
    ContractorRates string  // NEW
}
```

**Environment Variable**:
```bash
NOTION_CONTRACTOR_RATES_DB_ID=2c464b29b84c805bbcdedc052e613f4d
```

**Loading Logic** (follow existing pattern):
```go
Databases: NotionDatabase{
    // ... existing ...
    ContractorRates: viper.GetString("notion.contractor_rates_db_id"),
}
```

### 7.3 Contractor Rates Service

**File**: `pkg/service/notion/contractor_rates.go`

**Service Structure**:
```go
package notion

import (
    "context"
    "errors"
    "fmt"
    "time"

    nt "github.com/dstotijn/go-notion"
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
)

type ContractorRatesService struct {
    client *nt.Client
    cfg    *config.Config
    logger logger.Logger
}

func NewContractorRatesService(cfg *config.Config, logger logger.Logger) *ContractorRatesService {
    if cfg.Notion.Secret == "" {
        logger.Error(errors.New("notion secret not configured"), "notion secret is empty")
        return nil
    }

    logger.Debug("creating new ContractorRatesService")

    return &ContractorRatesService{
        client: nt.NewClient(cfg.Notion.Secret),
        cfg:    cfg,
        logger: logger,
    }
}

type ContractorRateData struct {
    ContractorPageID string
    Discord          string
    BillingType      string
    MonthlyFixed     float64
    HourlyRate       float64
    GrossFixed       float64
    Currency         string
    StartDate        time.Time
    EndDate          *time.Time
}
```

**Query Method**:
```go
func (s *ContractorRatesService) QueryRatesByDiscordAndMonth(
    ctx context.Context,
    discord string,
    month string,
) (*ContractorRateData, error) {
    dbID := s.cfg.Notion.Databases.ContractorRates
    if dbID == "" {
        return nil, errors.New("contractor rates database ID not configured")
    }

    s.logger.Debug("querying contractor rates",
        "discord", discord,
        "month", month,
        "database_id", dbID)

    // Parse month to get date range
    monthDate, err := time.Parse("2006-01", month)
    if err != nil {
        return nil, fmt.Errorf("invalid month format: %w", err)
    }

    // Build filters
    query := &nt.DatabaseQuery{
        Filter: &nt.DatabaseQueryFilter{
            And: []nt.DatabaseQueryFilter{
                // Filter by Discord username (rollup contains)
                {
                    Property: "Discord",
                    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                        Rollup: &nt.RollupDatabaseQueryFilter{
                            Any: &nt.DatabaseQueryPropertyFilter{
                                RichText: &nt.TextPropertyFilter{
                                    Contains: discord,
                                },
                            },
                        },
                    },
                },
                // Filter by Status = "Active"
                {
                    Property: "Status",
                    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                        Status: &nt.StatusDatabaseQueryFilter{
                            Equals: "Active",
                        },
                    },
                },
                // Filter by Start Date <= month
                {
                    Property: "Start Date",
                    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                        Date: &nt.DatePropertyFilter{
                            OnOrBefore: &monthDate,
                        },
                    },
                },
            },
        },
        PageSize: 100,
    }

    // Query database
    resp, err := s.client.QueryDatabase(ctx, dbID, query)
    if err != nil {
        s.logger.Error(err, "failed to query contractor rates database")
        return nil, fmt.Errorf("notion query failed: %w", err)
    }

    s.logger.Debug("query results", "count", len(resp.Results))

    // Process results
    for _, page := range resp.Results {
        props, ok := page.Properties.(nt.DatabasePageProperties)
        if !ok {
            continue
        }

        // Extract properties
        rateData := &ContractorRateData{
            ContractorPageID: page.ID,
        }

        // Extract Discord
        if discord := extractRollupString(props, "Discord"); discord != "" {
            rateData.Discord = discord
        }

        // Extract Billing Type
        if billingType := extractSelect(props, "Billing Type"); billingType != "" {
            rateData.BillingType = billingType
        }

        // Extract Monthly Fixed (formula)
        if monthlyFixed := extractFormulaNumber(props, "Monthly Fixed"); monthlyFixed != nil {
            rateData.MonthlyFixed = *monthlyFixed
        }

        // Extract Hourly Rate
        if hourlyRate := extractNumber(props, "Hourly Rate"); hourlyRate != nil {
            rateData.HourlyRate = *hourlyRate
        }

        // Extract Currency
        if currency := extractSelect(props, "Currency"); currency != "" {
            rateData.Currency = currency
        }

        // Extract dates
        if startDate := extractDate(props, "Start Date"); startDate != nil {
            rateData.StartDate = *startDate
        }
        rateData.EndDate = extractDate(props, "End Date")

        // Validate End Date (must be after month or empty)
        if rateData.EndDate != nil && rateData.EndDate.Before(monthDate) {
            s.logger.Debug("skipping rate: end date before month",
                "end_date", rateData.EndDate,
                "month", monthDate)
            continue
        }

        s.logger.Debug("found contractor rate",
            "billing_type", rateData.BillingType,
            "monthly_fixed", rateData.MonthlyFixed,
            "hourly_rate", rateData.HourlyRate,
            "currency", rateData.Currency)

        return rateData, nil
    }

    return nil, errors.New("contractor rates not found for the specified month")
}
```

**Helper Functions**:
```go
// extractRollupString extracts string from rollup property
func extractRollupString(props nt.DatabasePageProperties, propertyName string) string {
    prop, ok := props[propertyName]
    if !ok || prop.Rollup == nil {
        return ""
    }

    if len(prop.Rollup.Array) == 0 {
        return ""
    }

    firstItem := prop.Rollup.Array[0]
    if firstItem.RichText == nil || len(*firstItem.RichText) == 0 {
        return ""
    }

    return (*firstItem.RichText)[0].Text.Content
}

// extractFormulaNumber extracts number from formula property
func extractFormulaNumber(props nt.DatabasePageProperties, propertyName string) *float64 {
    prop, ok := props[propertyName]
    if !ok || prop.Formula == nil {
        return nil
    }

    if prop.Formula.Number == nil {
        return nil
    }

    return prop.Formula.Number
}

// extractNumber extracts number from number property
func extractNumber(props nt.DatabasePageProperties, propertyName string) *float64 {
    prop, ok := props[propertyName]
    if !ok || prop.Number == nil {
        return nil
    }

    return prop.Number
}

// extractSelect extracts select option name
func extractSelect(props nt.DatabasePageProperties, propertyName string) string {
    prop, ok := props[propertyName]
    if !ok || prop.Select == nil {
        return ""
    }

    return prop.Select.Name
}

// extractDate extracts date from date property
func extractDate(props nt.DatabasePageProperties, propertyName string) *time.Time {
    prop, ok := props[propertyName]
    if !ok || prop.Date == nil || prop.Date.Start == nil {
        return nil
    }

    date := *prop.Date.Start
    return &date
}
```

### 7.4 Task Order Log Service Extension

**File**: `pkg/service/notion/task_order_log.go`

**Add Structs**:
```go
type OrderSubitem struct {
    PageID      string
    ProjectName string
    ProjectID   string
    Hours       float64
    ProofOfWork string
}
```

**Add Method**:
```go
// QueryOrderSubitems queries timesheet line items (subitems) for a given order
func (s *TaskOrderLogService) QueryOrderSubitems(
    ctx context.Context,
    orderPageID string,
) ([]*OrderSubitem, error) {
    dbID := s.cfg.Notion.Databases.TaskOrderLog
    if dbID == "" {
        return nil, errors.New("task order log database ID not configured")
    }

    s.logger.Debug("querying order subitems", "order_page_id", orderPageID)

    // Build query to find subitems
    query := &nt.DatabaseQuery{
        Filter: &nt.DatabaseQueryFilter{
            And: []nt.DatabaseQueryFilter{
                // Filter by Type = "Timesheet"
                {
                    Property: "Type",
                    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                        Select: &nt.SelectDatabaseQueryFilter{
                            Equals: "Timesheet",
                        },
                    },
                },
                // Filter by Parent item contains orderPageID
                {
                    Property: "Parent item",
                    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                        Relation: &nt.RelationDatabaseQueryFilter{
                            Contains: orderPageID,
                        },
                    },
                },
            },
        },
        PageSize: 100,
    }

    var subitems []*OrderSubitem
    var cursor *string

    // Pagination loop
    for {
        if cursor != nil {
            query.StartCursor = *cursor
        }

        resp, err := s.client.QueryDatabase(ctx, dbID, query)
        if err != nil {
            s.logger.Error(err, "failed to query subitems")
            return nil, fmt.Errorf("notion query failed: %w", err)
        }

        s.logger.Debug("query page results", "count", len(resp.Results))

        // Process results
        for _, page := range resp.Results {
            props, ok := page.Properties.(nt.DatabasePageProperties)
            if !ok {
                continue
            }

            subitem := &OrderSubitem{
                PageID: page.ID,
            }

            // Extract Line Item Hours
            if hours := extractNumber(props, "Line Item Hours"); hours != nil {
                subitem.Hours = *hours
            }

            // Extract Proof of Works
            if pow := extractRichText(props, "Proof of Works"); pow != "" {
                subitem.ProofOfWork = pow
            }

            // Extract Project (rollup)
            if project := extractRollupRelation(props, "Project"); project != nil {
                subitem.ProjectID = project.ID
                // Fetch project name from relation
                if projectPage, err := s.client.FindPageByID(ctx, project.ID); err == nil {
                    if projectProps, ok := projectPage.Properties.(nt.DatabasePageProperties); ok {
                        if title := extractTitle(projectProps, "Name"); title != "" {
                            subitem.ProjectName = title
                        }
                    }
                }
            }

            subitems = append(subitems, subitem)
        }

        // Check if more pages exist
        if !resp.HasMore || resp.NextCursor == nil {
            break
        }
        cursor = resp.NextCursor
    }

    s.logger.Debug("found subitems", "count", len(subitems))

    return subitems, nil
}

// Helper functions
func extractRichText(props nt.DatabasePageProperties, propertyName string) string {
    prop, ok := props[propertyName]
    if !ok || len(prop.RichText) == 0 {
        return ""
    }

    return prop.RichText[0].Text.Content
}

func extractRollupRelation(props nt.DatabasePageProperties, propertyName string) *nt.Relation {
    prop, ok := props[propertyName]
    if !ok || prop.Rollup == nil || len(prop.Rollup.Array) == 0 {
        return nil
    }

    firstItem := prop.Rollup.Array[0]
    if firstItem.Relation == nil || len(*firstItem.Relation) == 0 {
        return nil
    }

    return &(*firstItem.Relation)[0]
}

func extractTitle(props nt.DatabasePageProperties, propertyName string) string {
    prop, ok := props[propertyName]
    if !ok || len(prop.Title) == 0 {
        return ""
    }

    return prop.Title[0].Text.Content
}
```

### 7.5 Controller Implementation

**File**: `pkg/controller/invoice/contractor_invoice.go`

```go
package invoice

import (
    "context"
    "fmt"
    "strings"
    "time"
    "bytes"
    "html/template"
    "math"
    "os/exec"

    "github.com/Rhymond/go-money"
    "github.com/dwarvesf/fortress-api/pkg/config"
    "github.com/dwarvesf/fortress-api/pkg/logger"
    "github.com/dwarvesf/fortress-api/pkg/service/notion"
    "github.com/dwarvesf/fortress-api/pkg/view"
)

// ContractorInvoiceData holds all data needed for contractor invoice PDF generation
type ContractorInvoiceData struct {
    InvoiceNumber   string
    ContractorName  string
    Month           string
    Date            time.Time
    DueDate         time.Time
    BillingType     string
    Currency        string
    LineItems       []view.ContractorInvoiceLineItem
    Total           float64
}

// GenerateContractorInvoice generates contractor invoice data from Notion
func (c *controller) GenerateContractorInvoice(
    ctx context.Context,
    discord string,
    month string,
) (*ContractorInvoiceData, error) {
    l := c.logger
    l.Debug("generating contractor invoice",
        "contractor_discord", discord,
        "month", month)

    // Step 1: Query Contractor Rates
    l.Debug("querying contractor rates from Notion")
    ratesService := notion.NewContractorRatesService(c.config, l)
    rateData, err := ratesService.QueryRatesByDiscordAndMonth(ctx, discord, month)
    if err != nil {
        l.Error(err, "failed to query contractor rates")
        return nil, fmt.Errorf("contractor rates query failed: %w", err)
    }

    l.Debug("found contractor rate",
        "billing_type", rateData.BillingType,
        "monthly_fixed", rateData.MonthlyFixed,
        "hourly_rate", rateData.HourlyRate,
        "currency", rateData.Currency)

    // Step 2: Query Task Order Log
    l.Debug("querying task order log")
    taskOrderService := notion.NewTaskOrderLogService(c.config, l)

    exists, orderPageID, err := taskOrderService.CheckOrderExistsByContractor(
        ctx,
        rateData.ContractorPageID,
        month,
    )
    if err != nil {
        l.Error(err, "failed to check task order log")
        return nil, fmt.Errorf("task order log query failed: %w", err)
    }

    if !exists {
        l.Error(nil, "no task order log found for contractor and month")
        return nil, fmt.Errorf("task order log not found")
    }

    l.Debug("found task order", "order_page_id", orderPageID)

    // Step 3: Query Order Subitems
    l.Debug("querying order subitems")
    subitems, err := taskOrderService.QueryOrderSubitems(ctx, orderPageID)
    if err != nil {
        l.Error(err, "failed to query subitems")
        return nil, fmt.Errorf("subitems query failed: %w", err)
    }

    l.Debug("found subitems", "count", len(subitems))

    // Step 4: Build Line Items based on Billing Type
    var lineItems []view.ContractorInvoiceLineItem
    var total float64

    if rateData.BillingType == "Monthly Fixed" {
        l.Debug("building line items for Monthly Fixed billing")

        // Monthly Fixed: Project + Proof of Work only, no amounts
        for _, subitem := range subitems {
            lineItems = append(lineItems, view.ContractorInvoiceLineItem{
                ProjectName: subitem.ProjectName,
                Description: subitem.ProofOfWork,
            })
        }
        total = rateData.MonthlyFixed

    } else if rateData.BillingType == "Hourly Rate" {
        l.Debug("building line items for Hourly Rate billing")

        // Hourly Rate: Project + PoW + Hours + Rate + Amount
        for _, subitem := range subitems {
            amount := subitem.Hours * rateData.HourlyRate
            lineItems = append(lineItems, view.ContractorInvoiceLineItem{
                ProjectName: subitem.ProjectName,
                Description: subitem.ProofOfWork,
                Hours:       subitem.Hours,
                Rate:        rateData.HourlyRate,
                Amount:      amount,
            })
            total += amount
        }

    } else {
        return nil, fmt.Errorf("unsupported billing type: %s", rateData.BillingType)
    }

    l.Debug("calculated total", "total", total, "line_items_count", len(lineItems))

    // Step 5: Generate Invoice Number
    randomChars := generateRandomAlphanumeric(4)
    invoiceNumber := fmt.Sprintf("INVC-%s-%s",
        strings.ReplaceAll(month, "-", ""),
        randomChars)

    l.Debug("generated invoice number", "invoice_number", invoiceNumber)

    // Step 6: Calculate Dates
    monthDate, _ := time.Parse("2006-01", month)
    invoiceDate := time.Date(monthDate.Year(), monthDate.Month(), 1, 0, 0, 0, 0, time.UTC)
    dueDate := invoiceDate.AddDate(0, 1, -1) // Last day of month

    // Step 7: Build Invoice Data
    invoiceData := &ContractorInvoiceData{
        InvoiceNumber:  invoiceNumber,
        ContractorName: discord,
        Month:          month,
        Date:           invoiceDate,
        DueDate:        dueDate,
        BillingType:    rateData.BillingType,
        Currency:       rateData.Currency,
        LineItems:      lineItems,
        Total:          total,
    }

    l.Info("contractor invoice generated successfully",
        "invoice_number", invoiceNumber,
        "total", total)

    return invoiceData, nil
}

// GenerateContractorInvoicePDF generates PDF from invoice data
func (c *controller) GenerateContractorInvoicePDF(
    data *ContractorInvoiceData,
) ([]byte, error) {
    l := c.logger
    l.Debug("generating contractor invoice PDF")

    // Setup currency formatter
    currencyCode := data.Currency
    pound := money.New(1, currencyCode)

    // Template functions
    funcMap := template.FuncMap{
        "formatMoney": func(amount float64) string {
            tmpValue := amount * math.Pow(10, float64(pound.Currency().Fraction))
            result := pound.Multiply(int64(tmpValue)).Display()
            return result
        },
        "formatDate": func(t time.Time) string {
            return t.Format("January 2, 2006")
        },
        "isMonthlyFixed": func() bool {
            return data.BillingType == "Monthly Fixed"
        },
        "isHourlyRate": func() bool {
            return data.BillingType == "Hourly Rate"
        },
        "add": func(a, b int) int {
            return a + b
        },
    }

    // Parse template
    templatePath := c.config.Invoice.TemplatePath
    if templatePath == "" {
        templatePath = "pkg/templates/contractor-invoice-template.html"
    }

    tmpl, err := template.New("contractor-invoice").Funcs(funcMap).ParseFiles(templatePath)
    if err != nil {
        l.Error(err, "failed to parse template")
        return nil, fmt.Errorf("template parse failed: %w", err)
    }

    // Render HTML
    var htmlBuffer bytes.Buffer
    templateData := struct {
        Invoice   *ContractorInvoiceData
        LineItems []view.ContractorInvoiceLineItem
    }{
        Invoice:   data,
        LineItems: data.LineItems,
    }

    if err := tmpl.Execute(&htmlBuffer, templateData); err != nil {
        l.Error(err, "failed to execute template")
        return nil, fmt.Errorf("template execution failed: %w", err)
    }

    l.Debug("template rendered", "html_size", htmlBuffer.Len())

    // Convert HTML to PDF using wkhtmltopdf
    cmd := exec.Command("wkhtmltopdf", "-", "-")
    cmd.Stdin = &htmlBuffer

    pdfBytes, err := cmd.Output()
    if err != nil {
        l.Error(err, "failed to generate PDF")
        return nil, fmt.Errorf("PDF generation failed: %w", err)
    }

    l.Debug("PDF generated", "pdf_size", len(pdfBytes))

    return pdfBytes, nil
}

// generateRandomAlphanumeric generates a random alphanumeric string of specified length
func generateRandomAlphanumeric(length int) string {
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)

    for i := range result {
        result[i] = charset[rand.Intn(len(charset))]
    }

    return string(result)
}
```

---

## 8. Error Handling

### 8.1 Error Definitions

**File**: `pkg/handler/invoice/errs/errors.go`

```go
package errs

import "errors"

var (
    // Contractor Rates Errors
    ErrContractorRatesNotFound = errors.New("contractor rates not found for the specified month")
    ErrContractorRatesInvalid  = errors.New("contractor rates data is invalid or incomplete")

    // Task Order Log Errors
    ErrTaskOrderLogNotFound = errors.New("task order log not found for the specified month")
    ErrNoSubitemsFound      = errors.New("no line items found in task order log")

    // Validation Errors
    ErrInvalidMonthFormat     = errors.New("invalid month format, expected YYYY-MM")
    ErrInvalidContractorInput = errors.New("contractor discord username is required")

    // Business Logic Errors
    ErrUnsupportedBillingType = errors.New("unsupported billing type")
    ErrInvoiceGenerationFailed = errors.New("failed to generate invoice")
    ErrPDFGenerationFailed    = errors.New("failed to generate PDF")

    // External Service Errors
    ErrNotionQueryFailed     = errors.New("notion query failed")
    ErrGoogleDriveUploadFailed = errors.New("google drive upload failed")
)
```

### 8.2 Error Handling Strategy

**Handler Level**:
- Validate all inputs before processing
- Return 400 Bad Request for validation errors
- Return 404 Not Found for missing resources
- Return 500 Internal Server Error for system errors
- Log all errors with context

**Controller Level**:
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Log errors with DEBUG information
- Return semantic errors from `errs` package

**Service Level**:
- Handle Notion API errors gracefully
- Retry transient failures (with exponential backoff)
- Return descriptive errors

**Example Error Response**:
```json
{
  "data": null,
  "error": "contractor rates not found for the specified month",
  "message": "No active contractor rate found for discord=adeki_ and month=2025-12",
  "pagination": null
}
```

---

## 9. Testing Strategy

### 9.1 Unit Tests

**Test Files**:
- `pkg/service/notion/contractor_rates_test.go`
- `pkg/service/notion/task_order_log_test.go`
- `pkg/controller/invoice/contractor_invoice_test.go`
- `pkg/handler/invoice/contractor_invoice_test.go`

**Test Cases**:

#### Contractor Rates Service
```go
func TestQueryRatesByDiscordAndMonth_MonthlyFixed(t *testing.T) {}
func TestQueryRatesByDiscordAndMonth_HourlyRate(t *testing.T) {}
func TestQueryRatesByDiscordAndMonth_NotFound(t *testing.T) {}
func TestQueryRatesByDiscordAndMonth_InvalidMonth(t *testing.T) {}
```

#### Task Order Log Service
```go
func TestQueryOrderSubitems_Success(t *testing.T) {}
func TestQueryOrderSubitems_Empty(t *testing.T) {}
func TestQueryOrderSubitems_Pagination(t *testing.T) {}
```

#### Controller
```go
func TestGenerateContractorInvoice_MonthlyFixed(t *testing.T) {
    // Verify line items have no amounts
    // Verify total equals monthly fixed
}

func TestGenerateContractorInvoice_HourlyRate(t *testing.T) {
    // Verify hours × rate calculation
    // Verify total is sum of amounts
}

func TestGenerateContractorInvoice_UnsupportedBillingType(t *testing.T) {}
```

#### Handler
```go
func TestGenerateContractorInvoice_Success(t *testing.T) {}
func TestGenerateContractorInvoice_InvalidMonth(t *testing.T) {}
func TestGenerateContractorInvoice_MissingContractor(t *testing.T) {}
```

### 9.2 Integration Tests

**Manual Testing with Real Data**:

1. **Test Monthly Fixed Billing**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/invoices/contractor/generate \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{
       "contractorDiscord": "adeki_",
       "month": "2025-12"
     }'
   ```

   Expected:
   - Invoice generated with project names and proof of work
   - No hours/rate/amount per line
   - Total = 48,000,000 VND

2. **Test Hourly Rate Billing**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/invoices/contractor/generate \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer $TOKEN" \
     -d '{
       "contractorDiscord": "nhidesign9",
       "month": "2025-11"
     }'
   ```

   Expected:
   - Invoice with hours, rate, amount per line
   - Total = sum of all amounts

3. **Test Error Cases**:
   ```bash
   # Invalid month format
   curl -X POST ... -d '{"contractorDiscord": "adeki_", "month": "2025/12"}'
   # Expected: 400 Bad Request

   # Non-existent contractor
   curl -X POST ... -d '{"contractorDiscord": "notexist", "month": "2025-12"}'
   # Expected: 404 Not Found
   ```

### 9.3 PDF Validation

- Verify PDF is generated without errors
- Check PDF contains all line items
- Verify currency formatting (VND vs USD)
- Confirm conditional table columns based on billing type

---

## 10. Security Considerations

### 10.1 Authentication & Authorization

- **Authentication**: JWT token required in Authorization header
- **Authorization**: User must have `PermissionInvoicesCreate` permission
- **Token Validation**: Handled by `conditionalAuthMW` middleware
- **Permission Check**: Handled by `conditionalPermMW` middleware

### 10.2 Input Validation

- **contractorDiscord**: Sanitize to prevent injection attacks
- **month**: Strict regex validation (`^\d{4}-\d{2}$`)
- **Notion API**: Use parameterized queries (SDK handles this)

### 10.3 Data Privacy

- **Sensitive Data**: Contractor rates are financial information
- **Access Control**: Only authorized users can generate invoices
- **Logging**: Do NOT log sensitive financial data in plain text
- **Google Drive**: Upload with restricted access (not public)

### 10.4 Rate Limiting

- **Notion API**: Respect rate limits (max 3 requests/second)
- **Endpoint**: Consider rate limiting per user (future)
- **Retry Logic**: Implement exponential backoff for retries

---

## Appendices

### Appendix A: Sample Notion Query Responses

**Contractor Rates Query Response**:
```json
{
  "object": "list",
  "results": [
    {
      "id": "2c464b29-b84c-8005-b990-d609c8ec7893",
      "properties": {
        "Discord": {
          "type": "rollup",
          "rollup": {
            "type": "array",
            "array": [{
              "type": "rich_text",
              "rich_text": [{"text": {"content": "adeki_"}}]
            }]
          }
        },
        "Billing Type": {
          "type": "select",
          "select": {"name": "Monthly Fixed"}
        },
        "Monthly Fixed": {
          "type": "formula",
          "formula": {"type": "number", "number": 48000000}
        },
        "Currency": {
          "type": "select",
          "select": {"name": "VND"}
        }
      }
    }
  ]
}
```

**Task Order Log Subitem Response**:
```json
{
  "object": "list",
  "results": [
    {
      "id": "2c464b29-b84c-8070-bf53-f6706bffbe71",
      "properties": {
        "Line Item Hours": {
          "type": "number",
          "number": 8
        },
        "Proof of Works": {
          "type": "rich_text",
          "rich_text": [{
            "text": {"content": "Implemented user authentication module"}
          }]
        },
        "Project": {
          "type": "rollup",
          "rollup": {
            "type": "array",
            "array": [{
              "type": "relation",
              "relation": [{"id": "project-page-id"}]
            }]
          }
        }
      }
    }
  ]
}
```

### Appendix B: Invoice Number Examples

| Month | Invoice Number | Notes |
|-------|----------------|-------|
| 2025-12 | INVC-202512-A7K9 | Random suffix: A7K9 |
| 2025-11 | INVC-202511-X3M2 | Random suffix: X3M2 |
| 2025-12 | INVC-202512-P5Q1 | Random suffix: P5Q1 |

**Format**: `INVC-{YYYYMM}-{random-4-chars}`
- Random characters: 4 uppercase alphanumeric (A-Z, 0-9)
- Total combinations per month: 1,679,616

### Appendix C: Currency Formatting Examples

**VND (Vietnamese Dong)**:
- Input: 48000000
- Output: "48,000,000"

**USD (US Dollar)**:
- Input: 1234.56
- Output: "$1,234.56"

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-24 | System | Initial specification document |

---

**End of Specification**
