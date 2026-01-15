# TASK-007: Integration Testing

## Priority
P3 - Validation

## Estimated Effort
20 minutes

## Description
Perform end-to-end integration testing to verify the complete flow from HTTP request to worker job enqueue.

## Dependencies
- TASK-001 through TASK-006 (All implementation tasks completed)

## Testing Scope

### Test Environment
- Local development environment
- Test Notion database (or mock)
- Worker in development mode

## Test Cases

### Test Case 1: Success Path - Valid Invoice

**Objective**: Verify successful invoice splits generation job enqueue

**Prerequisites:**
- Application running: `make dev`
- Valid invoice exists in Notion with Legacy Number "INV-2024-001"

**Steps:**
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}' \
  -v
```

**Expected Results:**
- HTTP Status: 200 OK
- Response Body:
  ```json
  {
    "data": {
      "legacy_number": "INV-2024-001",
      "invoice_page_id": "<notion-page-id>",
      "job_enqueued": true,
      "message": "Invoice splits generation job enqueued successfully"
    },
    "error": null
  }
  ```
- Log Output: Should show:
  - "handling generate invoice splits request"
  - "querying Notion for invoice by legacy number"
  - "found invoice in Notion: pageID=..."
  - "enqueuing invoice splits generation job"
  - "invoice splits generation job enqueued successfully"

**Verification:**
- Check worker logs for job processing
- Verify Notion invoice has splits generated

---

### Test Case 2: Error - Empty Legacy Number

**Objective**: Verify validation rejects empty legacy number

**Steps:**
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": ""}' \
  -v
```

**Expected Results:**
- HTTP Status: 400 Bad Request
- Response Body contains error message about invalid invoice number

---

### Test Case 3: Error - Missing Request Body

**Objective**: Verify validation rejects missing request body

**Steps:**
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -v
```

**Expected Results:**
- HTTP Status: 400 Bad Request
- Response Body contains JSON parsing error

---

### Test Case 4: Error - Invoice Not Found

**Objective**: Verify 404 response when invoice doesn't exist

**Prerequisites:**
- "INV-NONEXISTENT" does not exist in Notion

**Steps:**
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-NONEXISTENT"}' \
  -v
```

**Expected Results:**
- HTTP Status: 404 Not Found
- Response Body contains "not found" error message

---

### Test Case 5: Error - Invalid JSON

**Objective**: Verify handling of malformed JSON

**Steps:**
```bash
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{invalid json}' \
  -v
```

**Expected Results:**
- HTTP Status: 400 Bad Request
- Response Body contains JSON parsing error

---

### Test Case 6: Worker Job Processing

**Objective**: Verify worker actually processes the enqueued job

**Prerequisites:**
- Worker is running and processing jobs
- Valid invoice with Legacy Number "INV-2024-002"
- Invoice does not have splits generated yet

**Steps:**
1. Enqueue job:
   ```bash
   curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
     -H "Content-Type: application/json" \
     -d '{"legacy_number": "INV-2024-002"}'
   ```

2. Wait for worker to process (check logs)

3. Query Notion to verify splits were created

**Expected Results:**
- API returns 200 OK
- Worker logs show:
  - "processing invoice splits generation"
  - "querying line items with commissions"
  - (splits creation logs)
  - "invoice splits generated successfully"
- Notion invoice page:
  - "Splits Generated" checkbox is checked
  - Related invoice split records exist

---

### Test Case 7: Idempotency - Already Generated

**Objective**: Verify worker skips generation if already done

**Prerequisites:**
- Invoice "INV-2024-003" already has splits generated
- "Splits Generated" checkbox is checked

**Steps:**
1. Enqueue job:
   ```bash
   curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
     -H "Content-Type: application/json" \
     -d '{"legacy_number": "INV-2024-003"}'
   ```

2. Check worker logs

**Expected Results:**
- API returns 200 OK (job enqueued)
- Worker logs show:
  - "processing invoice splits generation"
  - "splits already generated, skipping"
- No duplicate splits created

---

### Test Case 8: Permission Check (Non-Local Environment)

**Objective**: Verify permission middleware works in production

**Prerequisites:**
- Application running in non-local environment
- User with valid JWT token
- User does NOT have `PermissionInvoiceEdit`

**Steps:**
```bash
curl -X POST https://staging-api.example.com/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token-without-permission>" \
  -d '{"legacy_number": "INV-2024-001"}' \
  -v
```

**Expected Results:**
- HTTP Status: 403 Forbidden
- Response indicates insufficient permissions

---

## Test Execution Checklist

### Pre-Testing Setup
- [ ] Application is running: `make dev`
- [ ] Database is seeded with test data
- [ ] Notion test database is accessible
- [ ] Worker is processing jobs
- [ ] Test invoice data exists in Notion

### Execution
- [ ] Test Case 1: Success Path - Valid Invoice
- [ ] Test Case 2: Error - Empty Legacy Number
- [ ] Test Case 3: Error - Missing Request Body
- [ ] Test Case 4: Error - Invoice Not Found
- [ ] Test Case 5: Error - Invalid JSON
- [ ] Test Case 6: Worker Job Processing
- [ ] Test Case 7: Idempotency - Already Generated
- [ ] Test Case 8: Permission Check (if applicable)

### Post-Testing Validation
- [ ] All test cases passed
- [ ] No unexpected errors in logs
- [ ] Worker jobs completed successfully
- [ ] Notion data is in expected state
- [ ] No memory leaks or goroutine leaks

## Automated Integration Test Script

Create test script: `scripts/test/integration_generate_splits.sh`

```bash
#!/bin/bash

set -e

BASE_URL="http://localhost:8080"
ENDPOINT="/api/v1/invoices/generate-splits"

echo "=== Integration Test: Generate Invoice Splits ==="

# Test 1: Valid request
echo "Test 1: Valid request"
RESPONSE=$(curl -s -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}')
echo "$RESPONSE" | jq .
if echo "$RESPONSE" | jq -e '.data.job_enqueued == true' > /dev/null; then
  echo "✓ Test 1 passed"
else
  echo "✗ Test 1 failed"
  exit 1
fi

# Test 2: Empty legacy number
echo "Test 2: Empty legacy number"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": ""}')
if [ "$HTTP_CODE" = "400" ]; then
  echo "✓ Test 2 passed"
else
  echo "✗ Test 2 failed (expected 400, got $HTTP_CODE)"
  exit 1
fi

# Test 3: Invoice not found
echo "Test 3: Invoice not found"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE_URL}${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-NONEXISTENT"}')
if [ "$HTTP_CODE" = "404" ]; then
  echo "✓ Test 3 passed"
else
  echo "✗ Test 3 failed (expected 404, got $HTTP_CODE)"
  exit 1
fi

echo "=== All tests passed ==="
```

Make executable:
```bash
chmod +x scripts/test/integration_generate_splits.sh
```

Run:
```bash
./scripts/test/integration_generate_splits.sh
```

## Acceptance Criteria

- [ ] All test cases pass successfully
- [ ] Success path (Test Case 1) works end-to-end
- [ ] All error cases return appropriate HTTP status codes
- [ ] Worker processes enqueued jobs correctly
- [ ] Idempotency works (no duplicate splits)
- [ ] Logs show expected messages at each step
- [ ] Notion data is updated correctly
- [ ] No errors or warnings in application logs
- [ ] Performance is acceptable (response time < 500ms)
- [ ] Integration test script runs successfully

## Verification Commands

```bash
# Start application
make dev

# Run integration test script
./scripts/test/integration_generate_splits.sh

# Manual test
curl -X POST http://localhost:8080/api/v1/invoices/generate-splits \
  -H "Content-Type: application/json" \
  -d '{"legacy_number": "INV-2024-001"}' | jq .

# Check application logs
tail -f logs/fortress-api.log | grep -i "generate.*splits"

# Check worker logs
tail -f logs/worker.log | grep -i "invoice.*splits"
```

## Reference Files
- Worker implementation: `pkg/worker/worker.go:104` (handleGenerateInvoiceSplits)
- Similar integration tests: Check existing test files in `scripts/test/` directory
