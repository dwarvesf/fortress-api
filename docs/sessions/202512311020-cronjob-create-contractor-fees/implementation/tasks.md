# Implementation Tasks: Contractor Fees Cronjob

**Session**: 202512311020-cronjob-create-contractor-fees
**Specification**: `docs/sessions/202512311020-cronjob-create-contractor-fees/planning/specifications/contractor-fees-cronjob.md`

---

## Tasks

### Task 1: Extend TaskOrderLogService ✅ COMPLETED

- **File(s)**: `pkg/service/notion/task_order_log.go`
- **Description**: Add methods to query approved orders and update status
- **Changes**:
  1. Add `ApprovedOrderData` struct with fields: PageID, Name, Date, Contractor (from rollup), FinalHoursWorked, ProofOfWorks
  2. Add `QueryApprovedOrders(ctx) ([]*ApprovedOrderData, error)` method
     - Filter: Type=Order, Status=Approved
     - Extract Contractor from rollup property `q?kW`
  3. Add `UpdateOrderStatus(ctx, pageID, status) error` method
     - Update Status property to given value
- **Acceptance**: Can query approved orders and update their status

---

### Task 2: Extend ContractorRatesService ✅ COMPLETED

- **File(s)**: `pkg/service/notion/contractor_rates.go`
- **Description**: Add method to find active rate by contractor
- **Changes**:
  1. Add `FindActiveRateByContractor(ctx, contractorPageID, date) (*ContractorRateData, error)` method
     - Filter: Contractor relation contains contractorPageID
     - Filter: Status=Active
     - Application-level filter: StartDate <= date <= EndDate (or EndDate is nil)
  2. Return rate data including PageID, HourlyRate, FixedFee, Currency, BillingType
- **Acceptance**: Can find matching contractor rate for a given contractor and date

---

### Task 3: Extend ContractorFeesService ✅ COMPLETED

- **File(s)**: `pkg/service/notion/contractor_fees.go`
- **Description**: Add methods to check existence and create fees
- **Changes**:
  1. Add `CheckFeeExistsByTaskOrder(ctx, taskOrderPageID) (bool, string, error)` method
     - Query Contractor Fees by Task Order Log relation
     - Return (exists, existingFeeID, error)
  2. Add `CreateContractorFee(ctx, taskOrderPageID, contractorRatePageID) (string, error)` method
     - Create new page in Contractor Fees database
     - Set Task Order Log relation
     - Set Contractor Rate relation
     - Set Payment Status = "New"
     - Return created page ID
- **Acceptance**: Can check fee existence and create new fees

---

### Task 4: Create Cronjob Handler ✅ COMPLETED

- **File(s)**: `pkg/handler/notion/contractor_fees.go` (NEW)
- **Description**: Create handler for the cronjob endpoint
- **Changes**:
  1. Implement `CreateContractorFees(c *gin.Context)` method on handler struct
  2. Query approved orders, for each: find rate, check existence, create fee, update status
  3. Add Swagger documentation
  4. Add comprehensive DEBUG logging
- **Acceptance**: Handler processes approved orders and returns statistics

---

### Task 5: Update Handler Interface ✅ COMPLETED

- **File(s)**: `pkg/handler/notion/interface.go`
- **Description**: Add CreateContractorFees to interface
- **Changes**:
  1. Add `CreateContractorFees(c *gin.Context)` to IHandler interface
- **Acceptance**: Interface includes new method

---

### Task 6: Register Route ✅ COMPLETED

- **File(s)**: `pkg/routes/v1.go`
- **Description**: Register the cronjob endpoint
- **Changes**:
  1. Add route: `POST /cronjobs/create-contractor-fees`
  2. Apply `conditionalAuthMW` and `conditionalPermMW(model.PermissionCronjobExecute)` middleware
- **Acceptance**: Route is accessible and calls handler

---

### Additional: Update Services Struct ✅ COMPLETED

- **File(s)**:
  - `pkg/service/notion/notion_services.go`
  - `pkg/service/service.go`
- **Description**: Add ContractorRates and ContractorFees services to Services struct
- **Changes**:
  1. Add `ContractorRates *ContractorRatesService` and `ContractorFees *ContractorFeesService` to Services struct
  2. Initialize services in service.go

---

## Execution Order

```
Task 1 (TaskOrderLogService) ──┐
Task 2 (ContractorRatesService) ├─→ Task 4 (Handler) → Task 5 (Interface) → Task 6 (Route)
Task 3 (ContractorFeesService) ─┘
```

Tasks 1, 2, 3 can be done in parallel. Tasks 4, 5, 6 are sequential.

---

## Implementation Complete

**Build Status**: ✅ Passed (`go build ./...`)

**Endpoint**: `POST /cronjobs/create-contractor-fees`

**Files Modified**:
- `pkg/service/notion/task_order_log.go` - Added `QueryApprovedOrders`, `UpdateOrderStatus`, `getContractorName`
- `pkg/service/notion/contractor_rates.go` - Added `FindActiveRateByContractor`
- `pkg/service/notion/contractor_fees.go` - Added `CheckFeeExistsByTaskOrder`, `CreateContractorFee`
- `pkg/service/notion/notion_services.go` - Added ContractorRates and ContractorFees fields
- `pkg/service/service.go` - Initialize ContractorRates and ContractorFees services
- `pkg/handler/notion/contractor_fees.go` - NEW file with CreateContractorFees handler
- `pkg/handler/notion/interface.go` - Added CreateContractorFees to IHandler
- `pkg/routes/v1.go` - Registered `/create-contractor-fees` route
