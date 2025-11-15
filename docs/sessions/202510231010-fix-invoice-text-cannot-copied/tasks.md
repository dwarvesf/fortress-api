# PDF Text Selection Fix - Implementation Tasks

## Problem
Generated invoice PDFs have non-copyable/non-selectable text because wkhtmltopdf is rendering text as images instead of text layers.

## Root Cause (UPDATED AFTER INVESTIGATION)
**PRIMARY ROOT CAUSE**: The Dockerfile installs the **unpatched Qt version** of wkhtmltopdf from Debian's apt repository. This version lacks the proper Qt webkit patches needed to embed text layers in PDFs, causing text rasterization (conversion to images).

**NOT the root cause** (but previously suspected):
- ~~WOFF font format~~ - Fonts are fine, patched wkhtmltopdf handles them correctly
- ~~High DPI setting (600)~~ - DPI is fine for quality, not the cause of text rasterization

## Implementation Tasks

### ✅ Task 1: Create tasks.md with implementation plan
**Status**: COMPLETED
**Description**: Document all tasks and track progress

---

### ✅ Task 2: Investigate Dockerfile wkhtmltopdf installation
**Status**: COMPLETED
**Finding**: Dockerfile installs unpatched wkhtmltopdf from `apt-get`, which is the root cause

---

### ✅ Task 3: Update Dockerfile to install patched wkhtmltopdf
**Status**: COMPLETED
**File**: `Dockerfile`
**Changes**:
- Removed `wkhtmltopdf` from apt-get install list (line 37)
- Added `wget` to apt-get dependencies
- Added new RUN command to download and install patched wkhtmltopdf 0.12.6.1-3 from GitHub
- Added comments explaining why patched version is required

**Implementation**:
```dockerfile
# Install wkhtmltopdf with patched Qt (required for selectable text in PDFs)
# DO NOT use apt-get install wkhtmltopdf - it installs unpatched version that rasterizes text
RUN wget -q https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltox_0.12.6.1-3.bookworm_amd64.deb && \
  dpkg -i wkhtmltox_0.12.6.1-3.bookworm_amd64.deb && \
  rm wkhtmltox_0.12.6.1-3.bookworm_amd64.deb
```

**Why this fixes the issue**: The patched Qt version properly embeds text layers in PDFs instead of rasterizing them as images

---

### ✅ Task 4: Create CLI command for testing PDF generation
**Status**: COMPLETED ✅
**Files Created**:
- `cmd/test-invoice-pdf/main.go` - CLI tool for testing PDF generation (fixed compilation errors)
- Added `GenerateInvoicePDFForTest` method to `pkg/controller/invoice` interface

**Build & Usage**:
```bash
# Build the CLI tool
go build -o test-invoice-pdf ./cmd/test-invoice-pdf

# Run with an existing invoice ID (requires .env file for DB connection)
./test-invoice-pdf --invoice-id=<uuid> --output=test-invoice.pdf
```

**Example**:
```bash
./test-invoice-pdf --invoice-id=123e4567-e89b-12d3-a456-426614174000 --output=my-test-invoice.pdf
```

**Requirements**:
- `.env` file with database configuration
- Existing invoice ID in the database
- Local wkhtmltopdf installation OR run inside Docker container

---

### ⏳ Task 5: Test the PDF generation with selectable text
**Status**: PENDING (Ready for testing after Docker rebuild)
**Prerequisites**:
1. Rebuild Docker image with updated Dockerfile
2. Have an existing invoice ID from the database

**Testing Steps**:
```bash
# 1. Rebuild Docker image
docker build -t fortress-api .

# 2. Run the test CLI command (inside container or with local build)
go install ./cmd/test-invoice-pdf
test-invoice-pdf --invoice-id=<your-invoice-id> --output=test-invoice.pdf

# 3. Open the generated PDF and verify:
- Text is selectable (try to highlight text)
- Text is copyable (Cmd+C / Ctrl+C)
- Text is searchable (Cmd+F / Ctrl+F)
- Visual appearance matches original invoices
```

---

## Expected Outcome
- ✅ Text in PDFs is fully selectable and copyable
- ✅ Text is searchable
- ✅ Visual appearance remains identical
- ✅ High DPI (600) maintained for quality
- ✅ WOFF fonts work correctly with patched wkhtmltopdf

## Files Modified
- ✅ `Dockerfile` - Install patched wkhtmltopdf from GitHub releases
- ✅ `pkg/controller/invoice/new.go` - Added test interface method
- ✅ `pkg/controller/invoice/send.go` - Added public test wrapper method
- ✅ `cmd/test-invoice-pdf/main.go` - New CLI tool for testing
- ✅ `tasks.md` - This file, documenting all changes

## Testing Checklist
- [ ] Generate invoice PDF via CLI
- [ ] Verify text is selectable
- [ ] Verify text is copyable
- [ ] Verify text is searchable (Cmd+F / Ctrl+F)
- [ ] Compare visual appearance with old PDF
- [ ] Test on different platforms (Mac/Windows/Linux)
- [ ] Check PDF file size reduction

## Rollback Plan
If issues arise:
1. Revert `invoice.html` to use WOFF fonts
2. Revert `send.go` DPI to 600
3. Keep original font files as backup
