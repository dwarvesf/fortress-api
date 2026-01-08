# ADR-001: ContractorPayablesService Architecture

## Status
Accepted

## Context
We need to create Contractor Payables records in Notion after successful invoice upload. The codebase has an established pattern for Notion services (e.g., `ContractorPayoutsService`, `ContractorRatesService`).

## Decision
Create a new `ContractorPayablesService` in `pkg/service/notion/` following the existing service pattern.

### Service Structure
```go
type ContractorPayablesService struct {
    client *nt.Client
    cfg    *config.Config
    logger logger.Logger
}
```

### Integration Points
1. **Config**: Add `ContractorPayables` to `NotionDBs` struct
2. **Services**: Add to `notion.Services` struct
3. **Initialization**: Initialize in `pkg/service/service.go`

## Consequences

### Positive
- Consistent with existing codebase patterns
- Reusable for future payables operations
- Testable in isolation

### Negative
- Additional service to maintain
- Requires config and initialization changes

## Alternatives Considered

### Alternative 1: Inline in Handler
- Rejected: Violates separation of concerns, not reusable

### Alternative 2: Add to ContractorPayoutsService
- Rejected: Different Notion database, different responsibility
