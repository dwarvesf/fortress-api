# Expense Flow Unit Test Design

## 1. Handler Validation Tests
| Test ID | Scenario | Input | Expected Outcome |
|---------|----------|-------|------------------|
| U-HAND-001 | Valid todo create payload | Basecamp-format payload (reason|amount|currency) parsed via ExpenseIntegration mock returns valid struct | Handler enqueues success feedback, calls `CreateExpense`, returns nil |
| U-HAND-002 | Invalid title format (<3 parts) | Payload missing currency section | Handler posts failure feedback, no service calls, returns nil |
| U-HAND-003 | Unsupported currency | Currency `EUR` | Handler posts failure comment; ensures `CreateExpense` not invoked |
| U-HAND-004 | Zero/invalid amount | Amount string unparsable -> Extract returns 0 | Handler logs error, posts failure feedback |
| U-HAND-005 | Completed event by operations | `IsOperationComplete()` true, ensures `CreatorEmail` derived from `msg.Creator.Email` |

## 2. ExpenseIntegration Selection Tests
| Test ID | Scenario | Setup | Expected Outcome |
|---------|----------|-------|------------------|
| U-SEL-001 | TASK_PROVIDER=basecamp | Config + factory wiring | Handler receives Basecamp mock implementation |
| U-SEL-002 | TASK_PROVIDER=nocodb | Config uses Noco implementation | Handler receives Noco mock implementation |
| U-SEL-003 | Expense override flag (if implemented) | Override set to `basecamp` while global `nocodb` | Handler still uses Basecamp implementation |

## 3. Parser Tests
- U-PARSE-001: Ensure `ParseWebhookPayload` correctly maps NocoDB `row.created` to internal DTO (row_id, reason, amount, attachments).
- U-PARSE-002: Status transition detection for `row.updated` from `pending` → `completed` triggers completion path.
- U-PARSE-003: Uncomplete/deletion event translates to `UncompleteExpense` call.
- U-PARSE-004: Signature validation failure returns explicit error; handler returns 401.

## 4. Service Tests
- U-SVC-001: `CreateExpense` persists metadata with `task_provider=nocodb` and row ID.
- U-SVC-002: `CompleteExpense` ensures accounting transaction creation invoked once.
- U-SVC-003: Attachment upload helper fallback path when Noco returns 5xx; expect error propagation + metrics increment.

## 5. Config Tests
- U-CFG-001: Missing mandatory env (workspace/table/webhook secret) causes startup failure.
- U-CFG-002: Approver mapping parser rejects malformed entries.
