# Specification: Contractor Rates Service Updates

## Overview
Add new method to query contractor rates by contractor page ID and month.

## File
`pkg/service/notion/contractor_rates.go`

## New Method

### QueryRatesByContractorPageID

```go
// QueryRatesByContractorPageID queries contractor rates by contractor page ID and month
func (s *ContractorRatesService) QueryRatesByContractorPageID(ctx context.Context, contractorPageID, month string) (*ContractorRateData, error)
```

**Parameters:**
- `contractorPageID`: Contractor page ID from Task Order Log
- `month`: Month in YYYY-MM format

**Behavior:**
1. Query Contractor Rates database with filter:
   - Contractor relation contains contractorPageID
   - Month formula equals month parameter
2. Extract rate data: BillingType, MonthlyFixed, HourlyRate, Currency
3. Return ContractorRateData or error if not found

**Filter Structure:**
```go
Filter: &nt.DatabaseQueryFilter{
    And: []nt.DatabaseQueryFilter{
        {
            Property: "Contractor",
            Relation: &nt.RelationDatabaseQueryFilter{
                Contains: contractorPageID,
            },
        },
        {
            Property: "Month",
            Formula: &nt.FormulaDatabaseQueryFilter{
                String: &nt.TextPropertyFilter{
                    Equals: month,
                },
            },
        },
    },
},
```

**Reference:** Similar to `QueryRatesByDiscordAndMonth` but uses Contractor relation instead of Discord rollup.
