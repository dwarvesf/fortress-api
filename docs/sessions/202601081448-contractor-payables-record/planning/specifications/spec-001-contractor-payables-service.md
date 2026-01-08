# Specification: ContractorPayablesService

## Overview
New Notion service for managing Contractor Payables database operations.

## File Location
`pkg/service/notion/contractor_payables.go`

## Data Structures

### CreatePayableInput
```go
type CreatePayableInput struct {
    ContractorPageID string   // Relation to Contractor (required)
    Total            float64  // Total amount in USD (required)
    Currency         string   // "USD" or "VND" (required)
    Period           string   // YYYY-MM-DD start of month (required)
    InvoiceDate      string   // YYYY-MM-DD (required)
    InvoiceID        string   // Invoice number e.g., CONTR-202512-A1B2 (required)
    PayoutItemIDs    []string // Relation to Payout Items (required)
    AttachmentURL    string   // Google Drive PDF URL (optional)
}
```

## Service Methods

### NewContractorPayablesService
```go
func NewContractorPayablesService(cfg *config.Config, logger logger.Logger) *ContractorPayablesService
```
- Validates Notion secret is configured
- Returns nil if not configured
- Logs service creation

### CreatePayable
```go
func (s *ContractorPayablesService) CreatePayable(ctx context.Context, input CreatePayableInput) (string, error)
```

**Notion Properties Mapping:**

| Input Field | Notion Property | Property Type | Notes |
|-------------|-----------------|---------------|-------|
| - | `Payable` | Title | Empty (Auto Name formula fills) |
| `Total` | `Total` | Number | `input.Total` |
| `Currency` | `Currency` | Select | `input.Currency` |
| `Period` | `Period` | Date | Parse `input.Period` |
| `InvoiceDate` | `Invoice Date` | Date | Parse `input.InvoiceDate` |
| `InvoiceID` | `Invoice ID` | Rich Text | `input.InvoiceID` |
| - | `Payment Status` | Status | "New" (hardcoded) |
| `ContractorPageID` | `Contractor` | Relation | Single relation |
| `PayoutItemIDs` | `Payout Items` | Relation | Multiple relations |
| `AttachmentURL` | `Attachments` | Files | External file URL |

**Returns:**
- `string`: Created page ID on success
- `error`: Error if creation fails

## Configuration

### Environment Variable
```
NOTION_CONTRACTOR_PAYABLES_DB_ID=2c264b29-b84c-8037-807c-000bf6d0792c
```

### Config Struct Addition
```go
// pkg/config/config.go - NotionDBs struct
ContractorPayables string
```

## Logging Requirements
- DEBUG: Service creation
- DEBUG: Input parameters on CreatePayable
- DEBUG: Notion API call details
- DEBUG: Success with page ID
- ERROR: Any failures with context
