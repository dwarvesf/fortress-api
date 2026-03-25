## Tasks

### Task 1: Create `discord_interaction_helpers.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_helpers.go` (new)
- **Description**: Move package-level and shared-utility functions out of `discord_interaction.go`:
  - `verifyDiscordSignature` (lines 1091–1109)
  - `respondToInteraction` (lines 615–624)
  - `formatCurrency` (lines 1965–1972)
  - `stripInvoiceIDFromReason` (lines 1976–1979)
  - Add correct imports: `crypto/ed25519`, `encoding/hex`, `net/http`, `regexp`, `discordgo`, `gin`, `github.com/leekchan/accounting`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 2: Create `discord_interaction_leave_nocodb.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_leave_nocodb.go` (new)
- **Description**: Move NocoDB leave handlers:
  - `handleLeaveApproveButton` (lines 337–366)
  - `handleLeaveRejectButton` (lines 369–385)
  - `updateLeaveMessageStatus` (lines 559–612)
  - Add correct imports: `fmt`, `net/http`, `time`, `discordgo`, `gin`, `logger`, `nocodb`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 3: Create `discord_interaction_leave_notion.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_leave_notion.go` (new)
- **Description**: Move Notion leave + calendar + contractor-lookup handlers:
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
  - Add correct imports: `context`, `errors`, `fmt`, `net/http`, `time`, `discordgo`, `nt "github.com/dstotijn/go-notion"`, `gin`, `logger`, `googleSvc`, `notionSvc`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 4: Create `discord_interaction_invoice.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_invoice.go` (new)
- **Description**: Move invoice-paid confirmation handlers:
  - `handleInvoicePaidConfirmButton` (lines 388–406)
  - `respondWithProcessing` (lines 409–448)
  - `processInvoicePaidConfirm` (lines 451–478)
  - `updateInvoiceMessageWithResult` (lines 481–524)
  - `updateInvoiceMessageWithError` (lines 527–556)
  - Add correct imports: `fmt`, `net/http`, `time`, `discordgo`, `gin`, `invoiceCtrl`, `logger`, `model`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 5: Create `discord_interaction_payout.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_payout.go` (new)
- **Description**: Move payout flow handlers:
  - `handlePayoutPreviewButton` (lines 1254–1354)
  - `handlePayoutCommitConfirmButton` (lines 1357–1384)
  - `processPayoutCommit` (lines 1387–1402)
  - `updatePayoutInteractionResponse` (lines 1405–1470)
  - `handlePayoutCommitCancelButton` (lines 1473–1493)
  - Add correct imports: `context`, `fmt`, `net/http`, `strings`, `time`, `discordgo`, `gin`, `ctrlcontractorpayables`, `logger`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 6: Create `discord_interaction_extra_payment.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction_extra_payment.go` (new)
- **Description**: Move extra-payment flow handlers:
  - `handleExtraPaymentPreviewButton` (lines 1496–1514)
  - `processExtraPaymentPreview` (lines 1517–1673)
  - `handleExtraPaymentConfirmButton` (lines 1676–1703)
  - `processExtraPaymentSend` (lines 1706–1859)
  - `updateExtraPaymentProgress` (lines 1862–1874)
  - `updateExtraPaymentInteractionResponse` (lines 1877–1938)
  - `handleExtraPaymentCancelButton` (lines 1941–1961)
  - Add correct imports: `context`, `errors`, `fmt`, `net/http`, `strings`, `time`, `discordgo`, `gin`, `logger`, `model`, `extrapayment`, `discordsvc`
- **Acceptance**: File compiles; `make lint` passes.

---

### Task 7: Trim `discord_interaction.go`
- **Status**: [x] Done
- **File(s)**: `pkg/handler/webhook/discord_interaction.go` (modify)
- **Description**: After Tasks 1–6 are complete, remove all functions that were moved. Keep only:
  - `HandleDiscordInteraction` (entry point)
  - `handleMessageComponentInteraction` (routing switch)
  - Remove the empty `init()` block (lines 1982–1984)
  - Trim imports to only what these two functions need
- **Acceptance**: File size reduced to ~150 lines; `go build ./...` succeeds; `make lint` passes; `make test` passes.
