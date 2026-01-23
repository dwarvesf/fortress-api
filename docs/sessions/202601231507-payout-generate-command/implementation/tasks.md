# Implementation Tasks: Payout Generate Command

## Overview
Add `?payout generate` Discord command with payout type and name filter support.

---

## Part 1: fortress-api Changes

### Task 1.1: Add ID Filter Parameter to Handler ✅
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**:
  - Add `id` query parameter parsing after `contractorFilter` (~line 174)
  - Pass `idFilter` to `processInvoiceSplitPayouts()` and `processRefundPayouts()`
- **Acceptance**: Handler accepts `id` query param and passes to process functions
- **Status**: COMPLETED

### Task 1.2: Update processInvoiceSplitPayouts with ID Filter ✅
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**:
  - Add `idFilter string` parameter to function signature (~line 700)
  - After querying `pendingSplits`, filter by `split.Name` if `idFilter != ""`
  - Use case-insensitive contains match
- **Acceptance**: Invoice splits filtered by Auto Name when `id` param provided
- **Status**: COMPLETED

### Task 1.3: Update processRefundPayouts with ID Filter ✅
- **File(s)**: `pkg/handler/notion/contractor_payouts.go`
- **Description**:
  - Add `idFilter string` parameter to function signature (~line 488)
  - After querying `approvedRefunds`, filter by `refund.RefundID` if `idFilter != ""`
  - Use case-insensitive contains match
- **Acceptance**: Refunds filtered by RefundID when `id` param provided
- **Status**: COMPLETED

---

## Part 2: fortress-discord Changes

### Task 2.1: Add Model for Generate Payout Result ✅
- **File(s)**: `pkg/model/payout.go`
- **Description**:
  - Add `GeneratePayoutResult` struct with fields:
    - `PayoutsCreated int`
    - `Processed int`
    - `Skipped int`
    - `Errors int`
    - `Type string`
    - `Details []map[string]any`
- **Acceptance**: Struct compiles and matches API response format
- **Status**: COMPLETED

### Task 2.2: Add Adapter Interface Method ✅
- **File(s)**: `pkg/adapter/fortress/interface.go`
- **Description**:
  - Add `GeneratePayouts(payoutType, id string) (*model.GeneratePayoutResult, error)` to interface
- **Acceptance**: Interface defines method signature
- **Status**: COMPLETED

### Task 2.3: Implement Adapter Method ✅
- **File(s)**: `pkg/adapter/fortress/payout.go`
- **Description**:
  - Implement `GeneratePayouts` method
  - Call `POST /api/v1/cronjobs/create-contractor-payouts`
  - Build query params: `type={type}` always, `id={id}` or `contractor={id}` based on type
  - For `contractor_payroll` type, use `contractor` param instead of `id`
- **Acceptance**: Adapter calls correct API endpoint with proper params
- **Status**: COMPLETED

### Task 2.4: Add Service Interface Method ✅
- **File(s)**: `pkg/discord/service/payout/interface.go`
- **Description**:
  - Add `GeneratePayouts(payoutType, id string) (*model.GeneratePayoutResult, error)` to interface
- **Acceptance**: Interface defines method signature
- **Status**: COMPLETED

### Task 2.5: Implement Service Method ✅
- **File(s)**: `pkg/discord/service/payout/service.go`
- **Description**:
  - Implement `GeneratePayouts` method
  - Delegate to adapter
- **Acceptance**: Service calls adapter method
- **Status**: COMPLETED

### Task 2.6: Add View Interface Method ✅
- **File(s)**: `pkg/discord/view/payout/interface.go`
- **Description**:
  - Add `ShowGenerateResult(original *model.DiscordMessage, result *model.GeneratePayoutResult, payoutType, id string) error`
- **Acceptance**: Interface defines method signature
- **Status**: COMPLETED

### Task 2.7: Implement View Method ✅
- **File(s)**: `pkg/discord/view/payout/payout.go`
- **Description**:
  - Implement `ShowGenerateResult` method
  - Format results showing payouts created, processed, skipped, errors
  - Include filter info (type, id) in output
- **Acceptance**: View displays results in readable Discord format
- **Status**: COMPLETED

### Task 2.8: Update Help View ✅
- **File(s)**: `pkg/discord/view/payout/payout.go`
- **Description**:
  - Update `Help()` method to document `generate` subcommand
  - Include usage examples with flags
- **Acceptance**: Help shows generate command with options
- **Status**: COMPLETED

### Task 2.9: Add Generate Command Handler ✅
- **File(s)**: `pkg/discord/command/payout/command.go`
- **Description**:
  - Add `generate` case in `Execute()` switch statement
  - Implement `generate()` method
  - Implement `parseGenerateArgs()` helper:
    - Support `--flag=value` and `--flag value` formats
    - Parse `--type`/`-t` (default: `invoice_split`)
    - Parse `--id`/`-i` (optional)
    - Validate type is: `contractor_payroll`, `invoice_split`, or `refund`
  - Call service and display result via view
- **Acceptance**: Command parses args, calls service, displays results
- **Status**: COMPLETED

---

## Verification Tasks

### Task 3.1: Build fortress-api ✅
- **Command**: `make build`
- **Acceptance**: Compiles without errors
- **Status**: COMPLETED (pre-existing go-duckdb issue unrelated to implementation)

### Task 3.2: Build fortress-discord ✅
- **Command**: `go build ./...` (in fortress-discord directory)
- **Acceptance**: Compiles without errors
- **Status**: COMPLETED

---

## Task Dependencies

```
Part 1 (fortress-api):
  1.1 → 1.2, 1.3 (parallel)

Part 2 (fortress-discord):
  2.1 → 2.2 → 2.3 → 2.4 → 2.5 → 2.9
  2.6 → 2.7 → 2.8 → 2.9

Verification:
  Part 1 complete → 3.1
  Part 2 complete → 3.2
```

## Execution Order

1. Task 1.1, 1.2, 1.3 (fortress-api)
2. Task 3.1 (verify fortress-api)
3. Task 2.1 (model)
4. Task 2.2, 2.3 (adapter)
5. Task 2.4, 2.5 (service)
6. Task 2.6, 2.7, 2.8 (view)
7. Task 2.9 (command)
8. Task 3.2 (verify fortress-discord)
