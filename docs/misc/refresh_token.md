# Refresh Token Usage

## ACCOUNTING_GOOGLE_REFRESH_TOKEN (10 usages)

pkg/service/googlemail/google_mail.go:

- Line 77: SendInvoiceMail() - Sending invoice emails
- Line 146: SendInvoiceThankYouMail() - Thank you emails
- Line 208: SendInvoiceOverdueMail() - Overdue reminders
- Line 348: SendPayrollPaidMail() - Payroll notifications

pkg/service/googledrive/google_drive.go:

- Line 37: UploadInvoicePDF() - Upload invoice PDFs
- Line 59: GetFile() - Get files from Drive
- Line 209: MoveFileToDestinationFolders() - Move files between folders

pkg/service/googlesheet/google_sheet.go:

- Line 62: ReadRows() - Read Google Sheets

pkg/service/googleadmin/google_admin.go:

- Line 44: Admin operations

## TEAM_GOOGLE_REFRESH_TOKEN (2 usages)

pkg/service/googlemail/google_mail.go:

- Line 435: SendInvitationMail() - Employee invitation emails
- Line 463: SendOffboardingMail() - Offboarding emails

## Configuration

pkg/config/config.go:

- Lines 75, 78: Field definitions in config struct
- Lines 252, 262: Loading from environment variables
