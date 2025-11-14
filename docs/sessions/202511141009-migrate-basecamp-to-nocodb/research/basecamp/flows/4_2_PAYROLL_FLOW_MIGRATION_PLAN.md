# Payroll Flow Migration Plan

## 1. Inventory Current Behavior
- Review `pkg/handler/payroll/payroll_calculator.go` (`calculatePayrolls`, `getBonus`, `getAccountingExpense`, `getReimbursement`) and `pkg/handler/payroll/commit.go` (`markBonusAsDone`).
- List every Basecamp dependency: Ops/Expense/Accounting bucket IDs, todo/comment APIs, worker enqueue payloads, Basecamp mention helper, amount extraction helper.
- Capture data dependencies: `ProjectBonusExplain` reference fields, commission/bonus stores, reimbursement parsing rules, environment-based approver IDs.

## 2. Define Target NocoDB Data Sources
- Design tables/views for Ops expenses, team reimbursements, and accounting reimbursement feeds with normalized columns (reason, amount, currency, approver, assignees, approval status, related invoice/payroll).
- Ensure schema exposes approval metadata (approver ID + timestamp + note) to avoid parsing “approve” comments.
- Plan how reimbursements attach to employees (direct relation via employee ID rather than Basecamp assignee IDs).

## 3. Build Provider Abstraction
- Extend the shared provider layer with `PayrollIntegration` or reuse existing `TaskIntegration` to support: listing approved reimbursements, fetching accounting expenses, marking reimbursements resolved, and posting notifications.
- Implement NocoDB adapters that translate provider calls into REST queries (filters on month/batch/approval status) and updates (status = settled, add activity note).
- Keep Basecamp adapter for backward compatibility until cutover.

## 4. Update Payroll Calculation Logic
- Swap Basecamp `Todo`/`Comment` calls for provider methods that return structured reimbursement records (including approval info) rather than raw todos.
- Replace title parsing with direct field reads for reason/amount/currency; keep `getReimbursement` for backward compatibility but plan to phase out once NocoDB-only fields exist.
- Ensure provider returns necessary metadata (record ID, source bucket) so `ProjectBonusExplain` can still store references for later reconciliation.

## 5. Update Payroll Commit Actions
- Modify `markBonusAsDone()` to call provider abstraction for settlement: mark reimbursement records as paid, add activity notes mentioning the employee, and emit notifications (email/Slack) via provider-specific integrations.
- Update worker payload builder to be provider-agnostic (e.g., `TaskIntegration.EnqueueComment`), allowing comments/notifications in Basecamp or note creation in NocoDB.

## 6. Testing & Rollout
- Add unit tests mocking the provider to cover: reimbursement filtering, amount conversion, settlement actions, and worker enqueue logic.
- Create integration tests hitting a fake NocoDB API verifying pagination/filters for Ops, Team, and Accounting datasets.
- Plan dual-run period where payroll reads from both Basecamp and NocoDB to compare reimbursement totals; once parity achieved, flip provider flag and disable Basecamp references.
