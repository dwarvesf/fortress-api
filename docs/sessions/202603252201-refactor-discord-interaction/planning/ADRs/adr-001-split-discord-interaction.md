# ADR-001: Split `discord_interaction.go` into domain-focused files

## Status
Proposed

## Context
`pkg/handler/webhook/discord_interaction.go` is a 1985-line monolith containing handler logic for every Discord button interaction the application supports. It covers 5 clearly distinct domains:

| Domain | Custom ID prefixes | Lines (approx.) |
|--------|-------------------|-----------------|
| Leave (NocoDB) | `leave_approve_`, `leave_reject_` | ~50 |
| Leave (Notion) | `notion_leave_approve_`, `notion_leave_reject_` | ~450 |
| Invoice | `invoice_paid_confirm_` | ~175 |
| Payout | `payout_preview:`, `payout_commit_confirm:`, `payout_commit_cancel:` | ~350 |
| Extra Payment | `ep_p:`, `ep_c:`, `ep_x:` | ~500 |
| Shared utilities | `verifyDiscordSignature`, `respondToInteraction`, `formatCurrency`, etc. | ~80 |
| Entry point / router | `HandleDiscordInteraction`, `handleMessageComponentInteraction` | ~100 |

The file is hard to navigate, violates the single-responsibility principle, and will keep growing as more button types are added.

## Decision
Split the file into **7 focused files**, all in the same `webhook` package, following the existing pattern used by e.g. `notion_leave.go`, `nocodb_leave.go`.

```
discord_interaction.go                    ← entry point + signature verify + router
discord_interaction_helpers.go            ← shared helpers (respondToInteraction, formatCurrency, stripInvoiceIDFromReason, verifyDiscordSignature)
discord_interaction_leave_nocodb.go       ← leave_approve_ / leave_reject_ handlers
discord_interaction_leave_notion.go       ← notion_leave_* handlers + calendar + contractor lookup
discord_interaction_invoice.go            ← invoice_paid_confirm_ handlers
discord_interaction_payout.go             ← payout_preview: / payout_commit_* handlers
discord_interaction_extra_payment.go      ← ep_p: / ep_c: / ep_x: handlers
```

## Consequences
- **Positive**: Each file is independently readable and testable; clear domain ownership.
- **Positive**: Easier to locate and modify a specific button flow.
- **Positive**: Consistent with the existing `notion_leave.go` / `nocodb_leave.go` naming pattern.
- **Neutral**: All files remain in the same `webhook` package, so no import changes outside the package are needed.
- **Risk (low)**: Moving functions between files is pure mechanical work; no logic changes, so regression risk is minimal.
