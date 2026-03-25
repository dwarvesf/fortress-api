# Specification: discord_interaction.go split

## Target directory
`pkg/handler/webhook/`

## File mapping

### 1. `discord_interaction.go` (keep, trimmed)
**Responsibility**: HTTP entry point, signature verification call, interaction type switch, custom-ID routing switch.

Functions to keep:
- `HandleDiscordInteraction` (lines 35–97)
- `handleMessageComponentInteraction` (lines 100–334)

Imports after trim: `context`, `crypto/ed25519`, `encoding/hex`, `encoding/json`, `errors`, `io`, `net/http`, `strconv`, `strings`, `discordgo`, `gin`, `logger`, `model`, `view`.

---

### 2. `discord_interaction_helpers.go` (new)
**Responsibility**: Pure utilities with no domain business logic.

Functions to move:
- `verifyDiscordSignature` (lines 1091–1109) — package-level func
- `respondToInteraction` (lines 615–624)
- `formatCurrency` (lines 1965–1972) — package-level func
- `stripInvoiceIDFromReason` (lines 1976–1979) — package-level func

Imports needed: `crypto/ed25519`, `encoding/hex`, `net/http`, `regexp`, `discordgo`, `gin`, `github.com/leekchan/accounting`.

---

### 3. `discord_interaction_leave_nocodb.go` (new)
**Responsibility**: NocoDB-backed leave approve/reject interactions.

Functions to move:
- `handleLeaveApproveButton` (lines 337–366)
- `handleLeaveRejectButton` (lines 369–385)
- `updateLeaveMessageStatus` (lines 559–612)

Imports needed: `fmt`, `net/http`, `time`, `discordgo`, `gin`, `logger`, `nocodb`, `view` (if used; check – `view` may not be needed here).

---

### 4. `discord_interaction_leave_notion.go` (new)
**Responsibility**: Notion-backed leave approve/reject, contractor lookup, Google Calendar event creation.

Functions to move:
- `handleNotionLeaveApproveButton` (lines 627–720)
- `handleNotionLeaveRejectButton` (lines 723–789)
- `getContractorPageIDByDiscordID` (lines 792–846)
- `respondWithNotionLeaveProcessingEmbed` (lines 849–898)
- `updateNotionLeaveMessageWithStatus` (lines 901–954)
- `updateNotionLeaveMessageWithError` (lines 957–998)
- `createCalendarEventForLeave` (lines 1001–1088)
- `getStakeholderEmailsFromDeployments` (lines 1112–1155)
- `getStakeholderIDsFromDeployments` (lines 1158–1221)
- `getContractorEmailFromNotion` (lines 1224–1251)

Imports needed: `context`, `errors`, `fmt`, `net/http`, `time`, `discordgo`, `nt "github.com/dstotijn/go-notion"`, `gin`, `logger`, `googleSvc`, `notionSvc`.

---

### 5. `discord_interaction_invoice.go` (new)
**Responsibility**: Invoice-paid confirmation button handling.

Functions to move:
- `handleInvoicePaidConfirmButton` (lines 388–406)
- `respondWithProcessing` (lines 409–448)
- `processInvoicePaidConfirm` (lines 451–478)
- `updateInvoiceMessageWithResult` (lines 481–524)
- `updateInvoiceMessageWithError` (lines 527–556)

Imports needed: `fmt`, `net/http`, `time`, `discordgo`, `gin`, `invoiceCtrl`, `logger`, `model`.

---

### 6. `discord_interaction_payout.go` (new)
**Responsibility**: Payout preview, confirm, and cancel button flows.

Functions to move:
- `handlePayoutPreviewButton` (lines 1254–1354)
- `handlePayoutCommitConfirmButton` (lines 1357–1384)
- `processPayoutCommit` (lines 1387–1402)
- `updatePayoutInteractionResponse` (lines 1405–1470)
- `handlePayoutCommitCancelButton` (lines 1473–1493)

Imports needed: `context`, `fmt`, `net/http`, `strings`, `time`, `discordgo`, `gin`, `ctrlcontractorpayables`, `logger`.

---

### 7. `discord_interaction_extra_payment.go` (new)
**Responsibility**: Extra payment preview, confirm, and cancel button flows.

Functions to move:
- `handleExtraPaymentPreviewButton` (lines 1496–1514)
- `processExtraPaymentPreview` (lines 1517–1673)
- `handleExtraPaymentConfirmButton` (lines 1676–1703)
- `processExtraPaymentSend` (lines 1706–1859)
- `updateExtraPaymentProgress` (lines 1862–1874)
- `updateExtraPaymentInteractionResponse` (lines 1877–1938)
- `handleExtraPaymentCancelButton` (lines 1941–1961)

Imports needed: `context`, `errors`, `fmt`, `net/http`, `strings`, `time`, `discordgo`, `gin`, `logger`, `model`, `extrapayment`, `discordsvc`.

## Rules for the refactor
1. **No logic changes** — pure mechanical move of functions between files.
2. **Same package** — all files use `package webhook`.
3. **Preserve all DEBUG/INFO/WARN/ERROR log calls** exactly as-is.
4. **Remove the empty `init()` at the bottom** of the original file (lines 1982–1984); it is a no-op.
5. After moving, each file must compile independently (correct imports).
6. Run `make lint` and `make test` to verify no regressions.
