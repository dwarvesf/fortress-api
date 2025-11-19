# Fortress Invoice API Client Guide

## Overview

This document explains how to consume the Fortress Invoice endpoints exposed under `/api/v1/invoices`. It focuses on request/response shapes, permissions, and gotchas so you can build a client confidently.

## Authentication & Permissions

- All endpoints use Bearer JWT authentication (same token the rest of Fortress uses).
- Permission checks happen in the router (`pkg/routes/v1.go:233-239`). Make sure the token you use owns the required scopes:
  - `GET /api/v1/invoices` – `invoice.read`
  - `GET /api/v1/invoices/template` – `invoice.read`
  - `POST /api/v1/invoices/send` – `invoice.read` (same permission as listing because send is restricted to finance operators)
  - `PUT /api/v1/invoices/:id/status` – `invoice.edit`
  - `POST /api/v1/invoices/:id/calculate-commissions` – `projects.commission_rate.edit`

Send the token via `Authorization: Bearer <jwt>` on every call.

## Envelope & Common Types

Responses use `view.Response<T>` (`pkg/view/response.go`) which wraps:

```json
{
  "total": 42,          // provided when pagination is present
  "page": 0,
  "size": 20,
  "sort": "-invoicedAt",
  "data": { ... },
  "message": "ok",      // optional human message
  "error": "",          // string if an error occurred
  "errors": [            // validation errors (field + msg)
    {"field": "status", "msg": "invalid invoice status"}
  ]
}
```

Key data models from `pkg/view/invoice.go`:

| Type | Fields |
| --- | --- |
| `InvoiceStatus` (`pkg/model/invoice.go`) | `draft`, `sent`, `overdue`, `paid`, `error`, `scheduled` |
| `Invoice` | Number, status, invoiced/due/paid timestamps, totals (`subTotal`, `tax`, `discount`, `total`), CC emails, `lineItems[]`, month/year, `sentBy`, `threadID`, `scheduledDate`, `conversionRate`, `bankID`, `projectID`, file metadata |
| `InvoiceItem` | `quantity`, `unitCost`, `discount`, `cost`, `description`, `isExternal` |
| `BankAccount` | Account/bank identifiers plus nested `Currency` info |
| `CompanyInfo` | Registration details plus flexible `info` map (address/phone pairs) |
| `ClientInfo` | Client company/address plus contact list (name, emails, `isMainContact`) |
| `InvoiceData` | Combines `Invoice`, `projectName`, `bankAccount`, `companyInfo`, `client` |
| `ProjectInvoiceTemplate` | Project metadata + `invoiceNumber`, latest invoice snapshot, client/company/bank blocks |
| `EmployeeCommission` (`pkg/model/employee_commissions.go`) | `employeeID`, `invoiceID`, `project`, `amount` (VND), `conversionRate`, `formula`, `note`, `isPaid`, `paidAt` |

## Pagination & Sorting

`view.Pagination` accepts `page`, `size`, `sort`. `page` is zero-indexed. When `size` is omitted or invalid, the server defaults to `maxPageSize` (999) – most clients explicitly set a reasonable value. `sort` accepts comma-separated fields with optional `-` prefix for descending (e.g. `sort=-invoiced_at,status`).

## Endpoints

### 1. List Invoices

- **Method/Path:** `GET /api/v1/invoices`
- **Purpose:** Paginated list filtered by project(s) and status.
- **Query parameters:**

| Name | Type | Notes |
| --- | --- | --- |
| `projectID` | array of UUID (repeatable) | Filter by projects; invalid UUID → 400 (`ErrInvalidProjectID`). |
| `status` | array of `InvoiceStatus` | Optional; invalid values → 400 (`ErrInvalidInvoiceStatus`). |
| `page` | int (default 0) | Zero-indexed page. |
| `size` | int | Defaults to `maxPageSize` if ≤0. |
| `sort` | string | Converted into SQL `"field ASC/DESC"`. Use `-field` for descending. |

- **Success (200) payload:** pagination metadata + `data: InvoiceData[]`.

Example request / response:

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/v1/invoices?projectID=ec7a7288-8a55-4f79-a035-97dd1e3d4a8f&status=sent&size=20&sort=-invoiced_at"
```

```json
{
  "total": 4,
  "page": 0,
  "size": 20,
  "sort": "-invoiced_at",
  "data": [
    {
      "invoice": {
        "number": "202411-ACME-001",
        "status": "sent",
        "invoicedAt": "2024-11-01T00:00:00Z",
        "dueAt": "2024-11-30T00:00:00Z",
        "subTotal": 45000,
        "total": 45000,
        "lineItems": [ ... ],
        "projectID": "ec7a7288-8a55-4f79-a035-97dd1e3d4a8f",
        "bankID": "9f058542-7ab8-4c55-9da6-69272616babc"
      },
      "projectName": "ACME Revamp",
      "bankAccount": {
        "accountNumber": "123456789",
        "currency": {"name": "US Dollar", "symbol": "$"}
      },
      "companyInfo": {"name": "Dwarves Foundation"},
      "client": {
        "clientCompany": "ACME Corp.",
        "contacts": [{"name": "Jane Smith", "emails": ["jane@acme.com"], "isMainContact": true}]
      }
    }
  ],
  "message": ""
}
```

### 2. Get Invoice Template

- **Method/Path:** `GET /api/v1/invoices/template`
- **Purpose:** Fetch the latest invoice for a project plus the next invoice number suggestion and reference data.
- **Query parameters:** `projectID` (required UUID). Missing/invalid → 400 (`ErrInvalidProjectID`).
- **Success (200):**

```json
{
  "data": {
    "id": "ec7a7288-8a55-4f79-a035-97dd1e3d4a8f",
    "name": "ACME Revamp",
    "invoiceNumber": "202411-ACME-002",
    "lastInvoice": { ... Invoice ... },
    "client": { ... },
    "bankAccount": { ... },
    "companyInfo": { ... }
  }
}
```

Use this before creating/sending an invoice to pre-fill line items and confirm numbering.

### 3. Send Invoice

- **Method/Path:** `POST /api/v1/invoices/send`
- **Purpose:** Create and send (or draft) an invoice, trigger Discord logging, and persist records.
- **Body (`SendInvoiceRequest`, `pkg/handler/invoice/request/request.go`):**

| Field | Type | Notes |
| --- | --- | --- |
| `isDraft` | bool | When true, invoice status becomes `draft`, otherwise `sent` immediately. |
| `projectID` | UUID | Required. |
| `bankID` | UUID | Required. Must reference a project bank account. |
| `sentBy` | UUID | Optional override; defaults to the authenticated user. Useful for API keys. |
| `description`, `note` | string | Optional text shown in the PDF/email. |
| `cc` | string[] | Empty strings removed. Non-prod environments only accept `@dwarves...` emails (see `mailutils.IsDwarvesMail`). |
| `lineItems` | `InvoiceItem[]` | Provide quantity, unitCost, discount, cost, description, isExternal. Values are rounded to 2 decimals serverside. |
| `email` | string | Required, validated as email (and restricted to Dwarves domain when `cfg.Env != prod`). |
| `total`, `discount`, `tax`, `subtotal` | number ≥ 0 | Client-calculated totals; server persists as provided. |
| `invoiceDate`, `dueDate` | `YYYY-MM-DD` strings | Parsed via `time.Parse("2006-01-02", ...)`. |
| `invoiceMonth`, `invoiceYear` | int | Month 0-11, year ≥ 0. |

- **Response:** `200` with `{ "message": "ok" }`. Errors propagate via `errs.ConvertControllerErr` – expect 400 on validation or 404/500 on controller issues.

Example:

```bash
curl -X POST "$BASE/api/v1/invoices/send" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{
        "projectID": "ec7a7288-8a55-4f79-a035-97dd1e3d4a8f",
        "bankID": "9f058542-7ab8-4c55-9da6-69272616babc",
        "email": "billing@acme.com",
        "invoiceDate": "2024-11-01",
        "dueDate": "2024-11-30",
        "lineItems": [{"quantity": 160, "unitCost": 250, "cost": 40000, "description": "Engineering hours"}],
        "subTotal": 40000,
        "tax": 0,
        "discount": 0,
        "total": 40000
      }'
```

### 4. Update Invoice Status

- **Method/Path:** `PUT /api/v1/invoices/:id/status`
- **Body (`UpdateStatusRequest`):**

```json
{
  "status": "paid",               // optional; must be a valid InvoiceStatus when provided
  "sendThankYouEmail": true        // optional flag to trigger customer thank-you email
}
```

- **Validation:**
  - `id` path param must be a UUID, otherwise 400 (`ErrInvalidInvoiceID`).
  - `status` empty ⇒ keep current status.

- **Response:** 200 `{ "message": "ok" }`. Errors bubble up via controller:
  - 404 when invoice not found.
  - 400 for invalid transitions.

### 5. Calculate Commissions

- **Method/Path:** `POST /api/v1/invoices/:id/calculate-commissions`
- **Purpose:** Run commission distribution logic for an invoice. Optionally execute in dry-run mode.
- **Query params:** `dry_run=true|false` (defaults to `false`). Dry run returns calculations without writing to DB.
- **Response:** `[]model.EmployeeCommission` containing the calculated payouts. Example item:

```json
{
  "id": "0f5568b5-d52e-4d7e-81a2-5aeb41d9d15b",
  "employeeID": "f7e0c1c8-5a54-4e34-ae47-9872b72851b8",
  "invoiceID": "75343e2b-3a34-4a27-9a70-9fb86bd9447e",
  "project": "ACME Revamp",
  "amount": 125000000,       // stored in VND (integer)
  "conversionRate": 23500,
  "formula": "total*0.05",
  "isPaid": false,
  "note": "",
  "paidAt": null
}
```

- **Errors:** Missing `id` → 400, unknown invoice → 404, other issues → 500.

## Implementation Tips

1. **Consistent UUID handling:** Validate UUIDs client-side to avoid immediate 400 responses.
2. **Email restrictions in non-prod:** When working against staging/dev, only `@dwarvesvillage.com` (or other Dwarves domains accepted by `mailutils.IsDwarvesMail`) are allowed for `email` and `cc`. Inject environment awareness into your client to prevent confusion.
3. **Pagination defaults:** Because the backend defaults `size` to 999, always send a `size` to avoid unexpectedly large payloads.
4. **Idempotency:** Sending an invoice is not idempotent; if you need retry safety, add client-side guards (e.g., confirm `invoiceNumber` uniqueness via the template endpoint before posting).
5. **Sorting translation:** The server converts `sort` strings into SQL `ORDER BY` clauses. Stick to known invoice columns (e.g., `invoiced_at`, `due_at`, `status`) and prefix with `-` for descending order.

With these details you can build a robust client that lists invoices, drafts/sends new ones, updates status, and calculates commissions using the Fortress API.
