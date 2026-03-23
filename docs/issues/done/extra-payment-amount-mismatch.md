# Extra Payment Approval Email Amount Mismatch

**Date**: 2026-03-23
**Status**: Resolved
**Priority**: High
**Area**: Extra Payment Notifications (API + Discord + Email)

## Problem Statement

Extra payment notification emails and Discord preview embeds showed inconsistent USD amounts compared to the API preview. The same payout entry could appear as `$379.69` in one view and `$400.00` in another, depending on which code path produced the calculation.

### Key Issue

- **API preview** used the Wise live exchange rate to convert VND amounts to USD.
- **Discord preview and email sending** used a hardcoded fallback rate (`1 USD = 25,000 VND`) baked into `entry.AmountUSD` during Notion payout loading.
- The email body was built from the Discord flow, so the sent email carried the wrong fallback amount.

### Example

Extra payment entry:

| Field | Value |
|-------|-------|
| Amount | 10,000,000 VND |
| Wise live rate | ~1 USD = 26,337 VND |
| Hardcoded fallback rate | 1 USD = 25,000 VND |

| View | Amount shown |
|------|--------------|
| API preview | $379.69 |
| Discord preview | $400.00 |
| Email body (old) | $400.00 |

---

## Root Cause

### Two independent conversion paths

1. **API flow** (`pkg/handler/notify/handler.go`) called `h.service.Wise.Convert(...)` at runtime and used the live rate.

2. **Discord flow** (`pkg/handler/webhook/discord_interaction.go`) summed `entry.AmountUSD`, which was precomputed by `pkg/service/notion/contractor_payouts.go` using a hardcoded constant:

```go
// pkg/service/notion/contractor_payouts.go:1933
const DefaultVNDToUSDRate = 25000.0

// pkg/service/notion/contractor_payouts.go:2040-2041
amountUSD = amount / DefaultVNDToUSDRate
```

### The email used the wrong path

The Discord send path (`processExtraPaymentSend`) aggregated `entry.AmountUSD` into per-contractor totals and passed that into the email template. So the email always reflected the fallback rate, not the Wise rate.

---

## Additional Issue: Display Precision Mismatch

The Discord preview formatted contractor lines with `%.0f` (rounding to whole dollars) while the total used `%.2f` (two decimals). This caused a visible mismatch between the total and the contractor breakdown even when the underlying values were identical.

---

## Solution

### 1. Shared USD amount resolver

A new package `pkg/extrapayment/amount.go` introduces a single function:

```go
func ResolveAmountUSD(l logger.Logger, wiseSvc wise.IService, pageID string, amount float64, currency string) (float64, error)
```

This function:

- Keeps USD amounts unchanged.
- Calls `Wise.Convert(...)` for non-USD amounts.
- Rounds to two decimal places.
- Returns an error (not a fallback) if Wise is unavailable or conversion fails.
- Logs the entry ID, source currency, original amount, Wise rate, and resolved amount at `DEBUG` level.

### 2. Notify API flow updated

`pkg/handler/notify/handler.go` now calls `extrapayment.ResolveAmountUSD(...)` instead of its own inline conversion logic. The `math` package is no longer imported.

Changes:

| Function | Before | After |
|----------|--------|-------|
| `PreviewExtraPaymentNotification` | Inline `Wise.Convert` with fallback to `entry.AmountUSD` | `extrapayment.ResolveAmountUSD` with error return |
| `SendExtraPaymentNotification` | Inline conversion in each goroutine | Pre-resolve all amounts into `amountsByPageID` map before sending |
| `SendOneExtraPaymentNotification` | Inline conversion | `extrapayment.ResolveAmountUSD` with error return |

### 3. Discord flow updated

`pkg/handler/webhook/discord_interaction.go` now calls `extrapayment.ResolveAmountUSD(...)` in both the preview and send paths:

| Function | Before | After |
|----------|--------|-------|
| `processExtraPaymentPreview` | `entry.AmountUSD` | `extrapayment.ResolveAmountUSD` with error embed |
| `processExtraPaymentSend` | `entry.AmountUSD` | `extrapayment.ResolveAmountUSD` with error response |

### 4. Display precision fixed

Contractor lines in the Discord preview now use `%.2f` at `discord_interaction.go:1589`, matching the total amount format at `discord_interaction.go:1597`.

---

## Files Changed

| File | Change |
|------|--------|
| `pkg/extrapayment/amount.go` | New shared Wise-based USD resolver |
| `pkg/extrapayment/amount_test.go` | Unit tests for the resolver |
| `pkg/handler/notify/handler.go` | Uses shared resolver; removes inline conversion and `math` import |
| `pkg/handler/webhook/discord_interaction.go` | Uses shared resolver in preview and send paths; `%.2f` for contractor lines |

---

## Verification

```bash
go test ./pkg/extrapayment ./pkg/handler/notify ./pkg/handler/webhook
```

- `pkg/extrapayment` passes: tests cover USD passthrough, Wise conversion, nil Wise, and conversion failure.
- `pkg/handler/notify` compiles successfully.
- `pkg/handler/webhook` passes: existing tests continue to work with the updated code.

---

## What the user should see after the fix

- Discord preview and API preview show the same Wise-based USD amount.
- Contractor line items in Discord embeds use `%.2f`, matching the total.
- Email body amount matches the preview amount.
- If Wise conversion fails, the request fails with an error instead of silently using a stale fallback rate.
