# Accounting Provider Cutover Checklist

1. **Staging validation**
   - Update `.env.staging` with `TASK_PROVIDER=nocodb` and NocoDB table IDs/secret.
   - Run `make migrate-up` to ensure `accounting_task_refs` exists.
   - Trigger `POST /api/v1/cronjobs/sync-monthly-accounting-todo` and verify records appear in both NocoDB (`AccountingTodos`) and `accounting_task_refs`.
   - Hit `/webhooks/nocodb/accounting` with a sample payload to confirm accounting transactions are created.

2. **Production rollout**
   - Deploy config change with `TASK_PROVIDER=nocodb`, NocoDB credentials, and webhook secret.
   - Trigger cronjob once (manual) and confirm NocoDB board + task refs updated for the next month.
   - Switch Basecamp automations off for accounting and enable the NocoDB webhook (Authorization header using `NOCO_WEBHOOK_SECRET`).

3. **Monitoring & rollback**
   - Watch application logs for `StoreNocoAccountingTransaction` and ensure accounting transaction counts match Basecamp history for the same day.
   - If issues arise, set `TASK_PROVIDER=basecamp`, re-enable Basecamp webhook, and rerun the cronjob; NocoDB data remains for analysis.
   - After one full cycle without issues, remove Basecamp accounting constants/IDs from environment configs.
