# Payout Commit Command Requirements

## Source

Requirements extracted from `/docs/specs/payout-commit-command.md`

## Command Specification

**Command:** `?payout commit <month> <batch>`

| Parameter | Format | Example | Description |
|-----------|--------|---------|-------------|
| `month` | YYYY-MM | 2025-01 | The billing period |
| `batch` | 1 or 15 | 15 | Maps to contractor's PayDay from Service Rate |

## Functional Requirements

### FR1: Query Pending Payables
- Query Contractor Payables with Payment Status = "Pending" for the given period
- Filter by contractor's PayDay from Service Rate matching the batch parameter

### FR2: Preview Before Commit
- Show preview embed with:
  - Month and batch
  - Count of payables
  - Total amount
  - List of contractors with amounts
- Provide Confirm/Cancel buttons

### FR3: Cascade Status Updates
When committing, update status across multiple related tables:

1. **Contractor Payable**
   - Payment Status: Pending → Paid
   - Payment Date: Set to current date

2. **Contractor Payouts** (Payout Items)
   - Status: Pending → Paid

3. **Invoice Split** (via "02 Invoice Split" relation)
   - Status: Pending → Paid

4. **Refund Request** (via "01 Refund" relation)
   - Status: Approved → Paid

### FR4: Error Handling
- Invalid month format → error message
- Invalid batch (not 1 or 15) → error message
- No pending payables → info message with count=0
- Notion API error → log and return partial results
- Partial failure → return success + failed counts

### FR5: Permissions
- Command requires admin/ops role
- API endpoints require `PayrollsRead` and `PayrollsCreate` permissions

## Database References

| Database | ID | Key Properties |
|----------|-----|----------------|
| Contractor Payables | `2c264b29-b84c-8037-807c-000bf6d0792c` | Payment Status (Status), Period, Payment Date |
| Contractor Payouts | `2c564b29-b84c-8045-80ee-000bee2e3669` | Status (Status), 01 Refund, 02 Invoice Split |
| Invoice Split | `2c364b29-b84c-804f-9856-000b58702dea` | Status (Select) |
| Refund Request | `2cc64b29-b84c-8066-adf2-cc56171cedf4` | Status (Status) |
| Service Rate | `2c464b29-b84c-80cf-bef6-000b42bce15e` | PayDay (Select) |

## API Endpoints

### Preview Commit
```
GET /api/v1/contractor-payables/preview-commit?month=2025-01&batch=15
```

### Execute Commit
```
POST /api/v1/contractor-payables/commit
```

## Non-Functional Requirements

- Logging: DEBUG level logs throughout for tracing
- Atomicity: Best-effort updates with failure tracking
- Idempotency: Re-running should not cause issues (already Paid records skipped)
