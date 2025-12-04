# ADR-003: Task Provider Selection and Configuration

## Status
Proposed

## Context

The fortress-api application currently supports multiple task providers for expense management:

1. **Basecamp** (original): Native Basecamp API integration for todos
2. **NocoDB** (current): REST API integration with NocoDB database tables
3. **Notion** (new): Notion API integration with Notion databases

The system uses a provider abstraction pattern where the provider is selected at runtime based on configuration, allowing seamless switching between providers without code changes in the payroll calculator.

### Current Implementation

The application uses `TASK_PROVIDER` environment variable to select which provider to use:

```go
// From pkg/service/service.go
var payrollExpenseProvider basecamp.ExpenseProvider

if cfg.TaskProvider == "nocodb" {
    payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)
}
if payrollExpenseProvider == nil {
    // Fallback to Basecamp
    payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
}
```

### Integration Points

1. **Service Initialization**: Provider created during service setup
2. **Payroll Calculator**: Uses provider interface, agnostic to implementation
3. **Commit Handler**: Uses provider-specific logic for status updates
4. **Configuration**: Environment variables define provider and settings

### Requirements for Notion Provider

- Must implement existing `ExpenseProvider` interface (no changes)
- Must coexist with Basecamp and NocoDB providers (backward compatibility)
- Must support easy rollback to NocoDB if issues arise
- Must follow established configuration patterns

## Decision

We will add Notion as a third provider option using the value `"notion"` for the `TASK_PROVIDER` configuration variable.

### Configuration Pattern

```bash
# Environment variable
TASK_PROVIDER=notion  # Options: "basecamp", "nocodb", "notion"

# Notion-specific configuration
NOTION_SECRET=secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
NOTION_CONTRACTOR_DB_ID=2bfb69f8-f573-805a-8915-000bc44ce188  # Optional
```

### Service Initialization

```go
// In pkg/service/service.go
var payrollExpenseProvider basecamp.ExpenseProvider

// Initialize based on provider selection
switch cfg.TaskProvider {
case "nocodb":
    payrollExpenseProvider = nocodb.NewExpenseService(nocoSvc, cfg, store, repo, logger.L)

case "notion":
    // New: Notion provider initialization
    payrollExpenseProvider = notion.NewExpenseService(notionSvc, cfg, store, repo, logger.L)

default:
    // Fallback to Basecamp (empty string or unknown value)
    if basecampSvc != nil {
        payrollExpenseProvider = basecamp.NewExpenseAdapter(basecampSvc)
    }
}

// Validation: ensure provider is initialized
if payrollExpenseProvider == nil {
    logger.Fatal("No expense provider configured")
}
```

### Provider Hierarchy

```
TASK_PROVIDER environment variable
├── "nocodb"   → nocodb.ExpenseService
├── "notion"   → notion.ExpenseService
└── "" (empty) → basecamp.ExpenseAdapter (fallback)
```

### No Dual-Provider Support

Unlike NocoDB which uses separate providers for expenses and accounting todos, Notion will use a single provider for all expense-related data:

```go
// NocoDB pattern (two providers)
PayrollExpenseProvider:       nocodb.ExpenseService
PayrollAccountingTodoProvider: nocodb.AccountingTodoService

// Notion pattern (single provider)
PayrollExpenseProvider:       notion.ExpenseService
PayrollAccountingTodoProvider: nil  // Not used with Notion
```

**Rationale**: Notion database structure combines all expense data in one database, eliminating the need for separate accounting todo tracking.

## Alternatives Considered

### Option 1: Provider Per Feature

Configure separate providers for expenses and todos:

```bash
EXPENSE_PROVIDER=notion
TODO_PROVIDER=basecamp
```

**Rejected because**:
- Added complexity for minimal benefit
- Current providers don't mix-and-match well
- Increases configuration surface area
- Harder to reason about provider state
- No current use case for mixed providers

### Option 2: Multiple Active Providers

Allow multiple providers active simultaneously:

```bash
TASK_PROVIDERS=nocodb,notion  # Comma-separated
```

**Rejected because**:
- Duplicate expense tracking (same expense in both systems)
- Complex reconciliation logic needed
- Risk of double-payment
- No clear conflict resolution strategy
- Unnecessary for migration use case

### Option 3: Provider Hierarchy with Fallback Chain

Attempt providers in order until one succeeds:

```go
providers := []ExpenseProvider{notionProvider, nocodbProvider, basecampProvider}
for _, provider := range providers {
    todos, err := provider.GetAllInList(listID, projectID)
    if err == nil {
        return todos, nil
    }
}
```

**Rejected because**:
- Hides configuration errors (silent fallback masks issues)
- Unpredictable behavior (which provider actually used?)
- Performance overhead (multiple API calls on failure)
- Difficult to debug in production
- Fail-fast is better than silent fallback for misconfiguration

### Option 4: Provider Auto-Detection

Automatically detect available provider based on configuration:

```go
if cfg.NotionExpenseDBID != "" {
    provider = notion.NewExpenseService(...)
} else if cfg.NocoDBTableID != "" {
    provider = nocodb.NewExpenseService(...)
} else {
    provider = basecamp.NewExpenseAdapter(...)
}
```

**Rejected because**:
- Implicit behavior is harder to understand
- Risk of unintended provider selection
- Cannot have multiple providers configured for staged rollout
- Explicit configuration is clearer and more maintainable

## Consequences

### Positive

1. **Clean Separation**: Each provider is independent, no cross-provider dependencies
2. **Easy Switching**: Change single environment variable to switch providers
3. **Backward Compatible**: Existing Basecamp and NocoDB configs unchanged
4. **Explicit Configuration**: Provider selection is clear and intentional
5. **Fail-Fast**: Missing configuration causes startup failure, not runtime errors
6. **Simple Rollback**: Revert `TASK_PROVIDER` to "nocodb" if Notion has issues

### Negative

1. **No Gradual Migration**: Cannot run both NocoDB and Notion simultaneously
   - **Mitigation**: Use feature flags or separate environments for testing

2. **Single Point of Configuration**: Entire system depends on one env var
   - **Mitigation**: Validate configuration on startup, fail early if misconfigured

3. **Provider-Specific Code**: Commit handler needs type assertion for status updates
   - **Impact**: Acceptable trade-off, already exists for NocoDB

### Migration Strategy

#### Phase 1: Development Environment

```bash
# Test with Notion provider
TASK_PROVIDER=notion
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
```

#### Phase 2: Staging Environment

```bash
# Validate against production-like data
TASK_PROVIDER=notion
NOTION_EXPENSE_DB_ID=<staging-database-id>
```

#### Phase 3: Production Deployment

```bash
# Switch production to Notion
TASK_PROVIDER=notion  # Changed from "nocodb"
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
```

#### Rollback Plan

```bash
# Revert to NocoDB if issues occur
TASK_PROVIDER=nocodb  # Changed back from "notion"
# Keep Notion config for future retry
NOTION_EXPENSE_DB_ID=2bfb69f8-f573-81cb-a2da-f06d28896390
```

## Implementation Details

### Configuration Structure

```go
// In pkg/config/config.go
type Config struct {
    // Provider selection
    TaskProvider string  // "basecamp", "nocodb", or "notion"

    // Provider-specific configs
    ExpenseIntegration ExpenseIntegration
}

type ExpenseIntegration struct {
    Noco   ExpenseNocoIntegration
    Notion ExpenseNotionIntegration
}

type ExpenseNotionIntegration struct {
    ExpenseDBID    string  // NOTION_EXPENSE_DB_ID
    ContractorDBID string  // NOTION_CONTRACTOR_DB_ID (optional)
}
```

### Service Structure

```go
// In pkg/service/service.go
type Service struct {
    // Dual providers (used by NocoDB only)
    PayrollExpenseProvider       basecamp.ExpenseProvider
    PayrollAccountingTodoProvider basecamp.ExpenseProvider  // nil for Notion

    // Other services...
}
```

### Commit Handler Updates

```go
// In pkg/handler/payroll/commit.go
func (h *handler) markExpenseSubmissionsAsCompleted(payrolls []*model.Payroll) {
    if h.service.PayrollExpenseProvider == nil {
        return
    }

    // Type assertion based on provider type
    switch provider := h.service.PayrollExpenseProvider.(type) {
    case *nocodb.ExpenseService:
        // NocoDB status update logic
        for _, expenseID := range extractExpenseIDs(payrolls) {
            provider.MarkExpenseAsCompleted(expenseID)
        }

    case *notion.ExpenseService:
        // Notion status update logic
        for _, pageID := range extractNotionPageIDs(payrolls) {
            provider.MarkExpenseAsCompleted(pageID)
        }

    default:
        // Basecamp or unknown provider (no action needed)
        h.logger.Debug("Provider does not support status updates",
            "provider_type", fmt.Sprintf("%T", provider))
    }
}
```

## Validation

### Configuration Validation

Add startup checks to validate provider configuration:

```go
func validateExpenseProviderConfig(cfg *config.Config) error {
    switch cfg.TaskProvider {
    case "nocodb":
        if cfg.ExpenseIntegration.Noco.TableID == "" {
            return fmt.Errorf("NOCO_EXPENSE_TABLE_ID required for nocodb provider")
        }

    case "notion":
        if cfg.ExpenseIntegration.Notion.ExpenseDBID == "" {
            return fmt.Errorf("NOTION_EXPENSE_DB_ID required for notion provider")
        }
        // Validate UUID format
        if !isValidNotionDatabaseID(cfg.ExpenseIntegration.Notion.ExpenseDBID) {
            return fmt.Errorf("NOTION_EXPENSE_DB_ID has invalid format")
        }

    case "", "basecamp":
        // Basecamp config validated elsewhere

    default:
        return fmt.Errorf("unknown TASK_PROVIDER: %s (expected: basecamp, nocodb, or notion)", cfg.TaskProvider)
    }

    return nil
}
```

### Testing Strategy

1. **Unit Tests**: Mock each provider, verify interface compliance
2. **Integration Tests**: Test with each provider configuration
3. **Config Tests**: Validate all provider combinations
4. **Fallback Tests**: Verify Basecamp fallback when provider not configured

### Acceptance Criteria

- [ ] `TASK_PROVIDER=notion` initializes Notion expense service
- [ ] `TASK_PROVIDER=nocodb` continues to work unchanged
- [ ] `TASK_PROVIDER=basecamp` or empty defaults to Basecamp
- [ ] Invalid provider value fails fast on startup
- [ ] Missing required config (e.g., `NOTION_EXPENSE_DB_ID`) fails fast
- [ ] Provider selection visible in startup logs
- [ ] Can switch providers via config change without code changes
- [ ] Rollback to previous provider works immediately

## References

- **Service Initialization**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/service.go` (Lines 243-279)
- **Provider Interface**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/basecamp/basecamp.go` (Lines 50-55)
- **NocoDB Provider**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/service/nocodb/expense.go`
- **Commit Handler**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/handler/payroll/commit.go` (Lines 772-790)
- **Configuration**: `/Users/quang/workspace/dwarvesf/fortress-api/pkg/config/config.go`

## Monitoring and Observability

### Startup Logs

```
INFO: Initializing expense provider: provider=notion
INFO: Notion expense service configured: database_id=2bfb69f8-f573-81cb-a2da-f06d28896390
INFO: Expense provider ready: type=*notion.ExpenseService
```

### Runtime Metrics

Track provider usage for observability:

```go
// Metrics to track
- expense_provider_type{provider="notion"}
- expense_fetch_count{provider="notion"}
- expense_fetch_duration_seconds{provider="notion"}
- expense_transformation_errors{provider="notion"}
- expense_status_update_count{provider="notion"}
```

### Health Checks

Add provider-specific health check:

```go
func (s *Service) HealthCheck() error {
    if s.PayrollExpenseProvider == nil {
        return fmt.Errorf("expense provider not initialized")
    }

    // Provider-specific health check
    if notionSvc, ok := s.PayrollExpenseProvider.(*notion.ExpenseService); ok {
        return notionSvc.Ping()  // Test Notion API connectivity
    }

    return nil
}
```
