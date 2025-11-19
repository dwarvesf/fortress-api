# NocoDB Leave Integration - Setup & Test Guide

**Date:** 2025-11-19
**Feature:** On-Leave Request Migration from Basecamp to NocoDB

---

## Prerequisites

- NocoDB instance running and accessible
- PostgreSQL database (fortress_v2_prod) accessible
- Go 1.21+ installed
- Access to `.env` configuration file

---

## Part 1: Database Setup

### Step 1: Run Migration

```bash
# Navigate to project root
cd /Users/quang/workspace/dwarvesf/fortress-api-feat-nocodb-onleave-migration

# Run migration to add nocodb_id column
make migrate-up

# Verify migration applied
psql -h 127.0.0.1 -p 25432 -U postgres -d fortress_v2_prod -c "\d on_leave_requests"
# Should show: nocodb_id | integer |
```

**Expected Output:**
```
Column      | Type      | Nullable
------------+-----------+---------
id          | uuid      | not null
...
nocodb_id   | integer   |
```

---

## Part 2: NocoDB Setup

### Step 2: Sync Employees to NocoDB

```bash
# Set environment variables
export NOCO_BASE_ID=pin7oroe7to3o1l
export NOCO_TOKEN=your_nocodb_token_here
export NOCO_BASE_URL=https://app.nocodb.com

# Run employee sync script
./scripts/local/sync_employees_to_nocodb.sh
```

**Expected Output:**
```
✓ nc_employees table already exists with ID: mt4vxig5keqdzpc
✓ Found 26 full-time employees

Step 3: Syncing employees to NocoDB (upsert mode)...
  [1/26] ↻ Alice Smith (alice@d.foundation) - UPDATED
  [2/26] ↻ Bob Jones (bob@d.foundation) - UPDATED
  ...

Summary:
  Inserted: 0
  Updated:  26
  Failed:   0
  Total:    26

✓ Successfully synced all 26 employees to NocoDB

Store this ID in your .env:
NOCO_EMPLOYEES_TABLE_ID=mt4vxig5keqdzpc
```

### Step 3: Create Leave Requests Table

**Manual Step in NocoDB UI:**

1. Navigate to base `pin7oroe7to3o1l`
2. Verify `leave_requests` table exists (ID: `myvvv4swtdflfwq`)
3. Create link field `assignees`:
   - Click "+ Add Column"
   - Column Type: "Link to Another Record"
   - Target Table: `nc_employees`
   - Relationship: Many-to-Many
   - Display Field: `full_name` (with email)
   - Save

**Verify Schema:**
```
Field Name       | Type              | Required
-----------------|-------------------|----------
Id               | AutoNumber        | ✓
employee_email   | Email             | ✓
type             | SingleSelect      | ✓ (Off, Remote)
start_date       | Date              | ✓
end_date         | Date              | ✓
shift            | SingleSelect      | - (Morning, Afternoon, Full Day)
reason           | LongText          | -
status           | SingleSelect      | ✓ (Pending, Approved, Rejected)
approved_by      | Email             | -
approved_at      | DateTime          | -
assignees        | LinkToAnotherRec  | - (→ nc_employees)
CreatedAt        | DateTime          | ✓
UpdatedAt        | DateTime          | -
```

---

## Part 3: Configuration

### Step 4: Update Environment Variables

Add to `.env`:

```bash
# NocoDB Leave Integration
NOCO_LEAVE_TABLE_ID=myvvv4swtdflfwq
NOCO_LEAVE_WEBHOOK_SECRET=$(openssl rand -hex 32)

# Existing Discord config (reused)
# DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/...
# DISCORD_CHANNEL_ID=...
```

**Generate webhook secret:**
```bash
openssl rand -hex 32
# Example output: a7f3c9e1b2d4f8a6c3e5d7b9f1a3c5e7d9b2f4a6c8e0d2f4a6c8e0d2f4a6c8e0
```

### Step 5: Verify Configuration Loads

```bash
# Build the application
make build

# Check config (should not error)
./bin/fortress-api --help
```

---

## Part 4: Configure NocoDB Webhook

### Step 6: Create Leave Request Webhook (Unified Endpoint)

**In NocoDB UI:**

1. Go to `leave_requests` table → Settings → Webhooks
2. Click "+ New Webhook"
3. Configure:
   - **Name:** Leave Request Handler
   - **Event:** After Insert, After Update
   - **URL:** `https://your-fortress-api.com/webhooks/nocodb/leave`
   - **Method:** POST
   - **Headers:**
     ```
     X-NocoDB-Signature: <paste_webhook_secret_from_env>
     Content-Type: application/json
     ```
   - **Condition:** None (handler will route based on event type and status)
4. Click "Save"

**How it works:**
- **Record Created** → Triggers validation
- **Record Updated (status: Pending → Approved)** → Creates DB record
- **Record Updated (status: Pending → Rejected)** → Sends rejection notification
- Other updates are ignored

---

## Part 5: Local Testing

### Step 9: Start Fortress API

```bash
# Start the API server
make dev

# Or run directly
go run cmd/server/main.go
```

**Expected Console Output:**
```
INFO  Starting Fortress API server
INFO  Environment: local
INFO  Port: 8080
INFO  Route registered: POST /webhooks/nocodb/leave
```

### Step 10: Expose Local Server (for Webhook Testing)

**Option A: Using ngrok**
```bash
ngrok http 8080
# Copy the HTTPS URL: https://abc123.ngrok.io
```

**Option B: Using CloudFlare Tunnel**
```bash
cloudflared tunnel --url http://localhost:8080
# Copy the HTTPS URL
```

Update NocoDB webhook URL to use the tunnel URL:
- `https://abc123.ngrok.io/webhooks/nocodb/leave`

---

## Part 6: End-to-End Test

### Test Case 1: Valid Leave Request Submission

**Action:**
1. In NocoDB, create new record in `leave_requests`:
   - `employee_email`: quang@d.foundation
   - `type`: Off
   - `start_date`: 2025-12-01
   - `end_date`: 2025-12-05
   - `shift`: Full Day
   - `reason`: Annual leave
   - `status`: Pending

**Expected Results:**
- ✅ Validation webhook triggered
- ✅ Fortress API logs: `leave request validated successfully: employee_id=... row_id=...`
- ✅ Discord notification sent: "📋 New Leave Request - Pending Approval"
- ✅ No errors in logs

**Verify:**
```bash
# Check Fortress API logs
tail -f logs/fortress-api.log | grep "leave request validated"

# Check NocoDB webhook logs
# Go to Webhooks → Leave Request Handler → View Logs
# Should show: 200 OK
```

---

### Test Case 2: Leave Approval

**Action:**
1. In NocoDB, update the record:
   - Change `status` from "Pending" → "Approved"
   - Set `approved_by`: nikki@d.foundation
   - `approved_at` auto-fills

**Expected Results:**
- ✅ Approval webhook triggered
- ✅ Fortress API logs: `leave request approved and persisted: id=... employee_id=... nocodb_id=...`
- ✅ Discord notification: "✅ Leave Request Approved"
- ✅ Database record created in `on_leave_requests` table

**Verify:**
```bash
# Check database
psql -h 127.0.0.1 -p 25432 -U postgres -d fortress_v2_prod -c \
  "SELECT id, title, creator_id, approver_id, nocodb_id, created_at
   FROM on_leave_requests
   WHERE nocodb_id IS NOT NULL
   ORDER BY created_at DESC LIMIT 1;"
```

**Expected Database Record:**
```
id         | uuid
title      | Quang Luong | Off | 2025-12-01 - 2025-12-05 | Full Day
creator_id | <employee_uuid>
approver_id| <approver_uuid>
nocodb_id  | 1 (matches NocoDB record ID)
```

---

### Test Case 3: Leave Rejection

**Action:**
1. Create another leave request (status: Pending)
2. Update `status` → "Rejected"

**Expected Results:**
- ✅ Rejection webhook triggered
- ✅ Fortress API logs: `leave request rejected: row_id=... employee_email=...`
- ✅ Discord notification: "❌ Leave Request Rejected"
- ✅ NO database record created (rejection doesn't persist)

**Verify:**
```bash
# Verify no DB record for rejected request
psql -h 127.0.0.1 -p 25432 -U postgres -d fortress_v2_prod -c \
  "SELECT COUNT(*) FROM on_leave_requests WHERE nocodb_id = <rejected_record_id>;"
# Should return: 0
```

---

## Part 7: Error Case Testing

### Test Case 4: Invalid Employee Email

**Action:**
1. Create leave request with `employee_email`: invalid@example.com

**Expected Results:**
- ✅ Validation webhook triggered
- ✅ Fortress API logs: `employee not found: email=invalid@example.com`
- ✅ Discord notification: "❌ Leave request validation failed - Employee not found"
- ✅ Response: `validation_failed:employee_not_found`

---

### Test Case 5: Invalid Date Range

**Action:**
1. Create leave request with:
   - `start_date`: 2025-12-10
   - `end_date`: 2025-12-05 (before start date)

**Expected Results:**
- ✅ Validation webhook triggered
- ✅ Fortress API logs: `end date before start date: start_date=2025-12-10 end_date=2025-12-05`
- ✅ Discord notification: "❌ Leave request validation failed - End date must be after start date"
- ✅ Response: `validation_failed:invalid_date_range`

---

### Test Case 6: Start Date in Past

**Action:**
1. Create leave request with `start_date`: 2025-01-01 (past date)

**Expected Results:**
- ✅ Validation webhook triggered
- ✅ Fortress API logs: `start date in past: start_date=2025-01-01`
- ✅ Discord notification: "❌ Leave request validation failed - Start date cannot be in the past"
- ✅ Response: `validation_failed:start_date_in_past`

---

### Test Case 7: Invalid Webhook Signature

**Action:**
1. Manually trigger webhook with incorrect signature:

```bash
curl -X POST https://your-api.com/webhooks/nocodb/leave \
  -H "Content-Type: application/json" \
  -H "X-NocoDB-Signature: invalid_signature" \
  -d '{
    "type": "record.created",
    "data": {
      "table_name": "leave_requests",
      "record": {
        "employee_email": "test@d.foundation"
      }
    }
  }'
```

**Expected Results:**
- ✅ HTTP 401 Unauthorized
- ✅ Response: `{"error": "invalid signature"}`
- ✅ Fortress API logs: `nocodb leave signature mismatch`

---

## Part 8: Monitoring & Debugging

### Check Webhook Logs

**In NocoDB:**
1. Go to `leave_requests` → Settings → Webhooks
2. Click on webhook name → View Logs
3. Review recent executions (status, response, timing)

**In Fortress API:**
```bash
# Tail logs for leave-related entries
tail -f logs/fortress-api.log | grep -i "leave"

# Or use structured logging
grep "method=ValidateNocodbLeave" logs/fortress-api.log | jq .
grep "method=ApproveNocodbLeave" logs/fortress-api.log | jq .
grep "method=RejectNocodbLeave" logs/fortress-api.log | jq .
```

### Common Debug Scenarios

**Webhook not triggering:**
```bash
# 1. Check webhook configuration in NocoDB
# 2. Verify URL is accessible
curl -I https://your-api.com/webhooks/nocodb/leave

# 3. Check firewall/ngrok is running
```

**Signature verification failing:**
```bash
# Verify secret matches between .env and NocoDB webhook header
echo $NOCO_LEAVE_WEBHOOK_SECRET

# Check NocoDB webhook header configuration
# X-NocoDB-Signature should match the secret
```

**Employee not found:**
```bash
# Verify employee email exists in DB
psql -h 127.0.0.1 -p 25432 -U postgres -d fortress_v2_prod -c \
  "SELECT id, full_name, team_email, personal_email
   FROM employees
   WHERE team_email = 'quang@d.foundation' OR personal_email = 'quang@d.foundation';"
```

---

## Part 9: Production Deployment Checklist

- [ ] Run migration on production database
- [ ] Update production `.env` with leave config
- [ ] Sync employees to NocoDB production base
- [ ] Create production webhook with production API URL (`/webhooks/nocodb/leave`)
- [ ] Test validation flow in production
- [ ] Test approval flow in production
- [ ] Verify Discord notifications working
- [ ] Monitor logs for 24 hours
- [ ] Document rollback procedure
- [ ] Disable Basecamp leave webhooks (if applicable)

---

## Troubleshooting

### Issue: Webhook returns 500 error

**Check:**
1. Database connection working
2. Employee table has required data
3. Logs show actual error message

```bash
# Check last error
grep "ERROR" logs/fortress-api.log | tail -20
```

### Issue: Discord notifications not sending

**Check:**
1. `h.service.Discord != nil`
2. Discord webhook URL configured
3. Discord channel ID valid

```bash
# Test Discord service directly
# Add debug log in sendLeaveDiscordNotification
```

### Issue: Database record not created on approval

**Check:**
1. Approval webhook triggered (check NocoDB logs)
2. Employee and approver found
3. No database constraint violations

```bash
# Check constraints
psql -h 127.0.0.1 -p 25432 -U postgres -d fortress_v2_prod -c \
  "\d on_leave_requests" | grep FOREIGN
```

---

## Success Criteria

✅ Validation webhook responds 200 OK
✅ Discord notification sent on validation
✅ Approval creates DB record with correct `nocodb_id`
✅ Rejection sends Discord notification
✅ Invalid data returns proper error responses
✅ Signature verification prevents unauthorized requests
✅ All logs show expected debug information

---

## Next Steps After Testing

1. Write unit tests for all three handlers
2. Add integration tests with mock NocoDB payloads
3. Performance test with multiple concurrent webhooks
4. Set up monitoring/alerting for webhook failures
5. Create runbook for production issues

---

## References

- **Implementation**: `pkg/handler/webhook/nocodb_leave.go`
- **Routes**: `pkg/routes/v1.go:74-79`
- **Schema**: `docs/sessions/.../plan/onleave/NOCODB_LEAVE_STRUCTURE.md`
- **Migration**: `migrations/schemas/20251119140000-add-nocodb-id-to-onleave-requests.sql`
