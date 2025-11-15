# NocoDB Table Provisioning Notes

## Meta API basics
- Table creation goes through the **Meta** endpoints: `POST /api/v1/db/meta/projects/{baseId}/tables` and `POST /api/v1/db/meta/tables/{tableId}/columns`.
- Use the base ID you see in the address bar (e.g. `pk5crwekcume3yb`). In scripts you can reference this as `NOCO_BASE_ID` or `NOCO_PROJECT_ID`.
- The `uidt` values must match NocoDB’s enum (`AutoNumber`, `SingleLineText`, `SingleSelect`, `JSON`, etc.). There is no `Int` type even though the UI shows integers.
- When creating tables via API, **set the primary key explicitly** (`"pk": true`) and use `AutoNumber` if you need auto increment. NocoDB will not infer it automatically.

## Environment variables
Scripts expect:
```
NOCO_BASE_URL=https://app.nocodb.com      # no trailing /api/v2
NOCO_TOKEN=...                            # Personal access token
NOCO_BASE_ID=pk5crwekcume3yb              # or NOCO_PROJECT_ID
```

## Accounting tables script (`scripts/local/create_nocodb_accounting_tables.sh`)
- Creates two tables and all columns:
  - `accounting_todos` (title matches table name to stay snake_case like the invoice tables) with fields `board_label`, `task_group (pre-seeded with choices), `template_id`, `title`, `description`, `due_on`, `assignee_ids`, `status`, `metadata` (+ AutoNumber primary key `task_id`).
  - `accounting_transactions` with fields `todo_row_id`, `txn_group`, `bucket`, `amount`, `currency`, `actor`, `status`, `occurred_at`, `metadata` (+ AutoNumber primary key `transaction_id`).
- Reuses `.env` credentials; if `NOCO_PROJECT_ID` is absent it falls back to `NOCO_BASE_ID`.
- Run: `NOCO_BASE_URL=... NOCO_TOKEN=... NOCO_BASE_ID=... ./scripts/local/create_nocodb_accounting_tables.sh`
- Script prints the generated table IDs so you can copy them back into `.env` (`NOCO_ACCOUNTING_TODOS_TABLE_ID`, `NOCO_ACCOUNTING_TRANSACTIONS_TABLE_ID`).

### Renaming legacy tables
If you already created tables before the script was updated (e.g. `AccountingTodos` with camel-case title), rename them in NocoDB so automation names stay consistent:
1. In the NocoDB UI, click the table name ▸ **Rename** ▸ enter the snake_case version.
2. Or via API:
   ```bash
   curl -sS -X PATCH \
     -H "xc-token: $NOCO_TOKEN" \
     -H "Content-Type: application/json" \
     "$NOCO_BASE_URL/api/v1/db/meta/tables/<tableId>" \
     -d '{"title":"accounting_todos"}'
   ```
   Repeat for `accounting_transactions`.

## Lessons for other flows
1. **Use AutoNumber for PKs** – NocoDB rejects `"uidt":"Int"`; pick from the allowed set.
2. **Stick to base host in `NOCO_BASE_URL`** – e.g. `https://app.nocodb.com`. The script appends the Meta path automatically; do not include `/api/v2`.
3. **Share schema via arrays** – define column payloads as JSON strings in bash arrays. For more complex flows, it may be easier to switch to a small Node/TS script using the official SDK.
4. **Surface IDs immediately** – print the returned IDs so other services can update `.env` without hunting through API responses.
5. **Webhooks use data table IDs** – the Meta API response also gives you `table_alias` (e.g., `tbl_xxx`) which is what the data APIs/webhooks use. Capture it if needed.

With these conventions we can add new scripts for Expense, Payroll, etc. by copying the structure and updating the column list.
