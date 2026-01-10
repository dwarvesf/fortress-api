# Test Cases Status - Contractor Invoice Email Listener

**Session**: 202601100842-contractor-invoice-email-listener
**Date**: 2026-01-10
**Status**: Test Specifications Complete

## Overview

Comprehensive unit test specifications have been created for the Contractor Invoice Email Listener feature. All test cases follow the existing testing patterns used in the fortress-api codebase (table-driven tests, testify assertions, mock-based testing).

## Test Coverage Summary

### Unit Test Specifications

| Component | Test File Spec | Test Cases | Coverage Focus |
|-----------|---------------|------------|----------------|
| Gmail Inbox Service | `unit/gmail-inbox-service-tests.md` | 40+ tests | API operations, error handling, helper functions |
| Invoice Email Processor | `unit/invoice-processor-tests.md` | 25+ tests | Business logic, batch processing, error paths |
| PDF Parser & Extractor | `unit/pdf-parser-tests.md` | 35+ tests | Regex patterns, PDF extraction, fallback logic |

**Total Test Cases**: 100+ unit tests specified

## Test Files to be Created

### Implementation Files

1. **pkg/service/googlemail/google_mail_inbox_test.go**
   - ListInboxMessages tests
   - GetMessage tests
   - GetAttachment tests
   - AddLabel tests
   - GetOrCreateLabel tests
   - Helper function tests (ParseEmailHeaders, FindPDFAttachment)

2. **pkg/service/invoiceemail/processor_test.go**
   - ProcessInvoiceEmails batch processing tests
   - ProcessSingleEmail end-to-end tests
   - findPayableByInvoiceID tests
   - Error scenario tests
   - Mock-based dependency testing

3. **pkg/service/invoiceemail/extractor_test.go**
   - ExtractInvoiceID integration tests
   - extractFromText regex tests
   - extractTextFromPDF tests
   - Subject vs PDF priority tests

### Test Data Files

**Location**: `pkg/service/invoiceemail/testdata/`

Required test files:
- `invoice_with_id.pdf` - Valid PDF containing "CONTR-202501-A1B2"
- `invoice_no_id.pdf` - Valid PDF without Invoice ID
- `malformed.pdf` - Corrupted/invalid PDF
- `empty.pdf` - Empty PDF file
- `multipage.pdf` - Multi-page PDF (Invoice ID on page 1)
- `image_only.pdf` - PDF with images but no text (optional)
- `encrypted.pdf` - Password-protected PDF (optional)

## Testing Strategy

### Test-Driven Development Approach

1. **Define Test Cases First**: All test specifications completed before implementation
2. **Implement to Pass Tests**: Feature implementation will make tests pass
3. **Refactor with Confidence**: Comprehensive tests enable safe refactoring

### Testing Layers

#### Unit Tests (Primary Focus)

**Scope**: Individual functions and methods in isolation

**Mocking Strategy**:
- Mock Gmail API service using testify/mock or gomock
- Mock Notion API service
- Test business logic independently of external dependencies

**Pattern**: Table-driven tests following existing codebase patterns

**Execution**: Fast, no external dependencies, runnable in CI

#### Integration Tests (Out of Scope)

**Note**: As per workflow instructions, integration test specifications are not included in this deliverable. Focus is on unit test specifications only.

**Future Considerations**:
- Integration tests with real Gmail API (test account)
- Integration tests with test PDF files
- End-to-end flow testing with Notion sandbox

## Test Coverage Goals

### Code Coverage Targets

| Component | Target | Priority |
|-----------|--------|----------|
| Gmail Inbox Service | 90%+ | High |
| Invoice Email Processor | 95%+ | Critical |
| PDF Parser & Extractor | 90%+ | High |

### Path Coverage

- **Happy Path**: All success scenarios tested
- **Error Paths**: All error conditions tested
- **Edge Cases**: Boundary conditions, invalid inputs, null handling
- **Business Logic**: Status validation, payable matching, label idempotency

### Critical Test Areas

1. **Invoice ID Extraction**
   - Regex pattern validation (20+ patterns tested)
   - Subject line parsing
   - PDF text extraction fallback
   - Error handling when ID not found

2. **Error Handling**
   - Gmail API failures
   - Notion API failures
   - PDF parsing errors
   - Partial failures (update succeeds, label fails)

3. **Business Logic**
   - Only process payables with status "New"
   - Skip already processed emails
   - Continue processing after individual failures
   - Idempotent label application

4. **Edge Cases**
   - Unicode characters in emails
   - Large PDFs
   - Malformed PDFs
   - Multiple Invoice IDs in text
   - Empty/missing data

## Testing Patterns Used

### 1. Table-Driven Tests

Used for testing multiple scenarios with similar structure:

```go
func TestMethod(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**Applied to**: Regex pattern matching, header parsing, validation logic

### 2. Mock-Based Testing

Used for isolating business logic from external dependencies:

```go
mockGmail := new(MockGmailService)
mockGmail.On("Method", args).Return(result, nil)

processor := NewProcessor(cfg, logger, mockGmail, mockNotion)
err := processor.ProcessEmail(ctx, messageID)

mockGmail.AssertExpectations(t)
```

**Applied to**: Invoice processor, Gmail/Notion service integration

### 3. Interface Verification Tests

Ensure methods exist with correct signatures:

```go
func TestMethodExists(t *testing.T) {
    var _ interface {
        Method(params) (return, error)
    } = (*Service)(nil)
}
```

**Applied to**: Gmail inbox service interface methods

### 4. File-Based Tests

Use test data files for realistic scenarios:

```go
pdfBytes, err := os.ReadFile("testdata/invoice.pdf")
require.NoError(t, err)

text, err := extractTextFromPDF(pdfBytes)
assert.NoError(t, err)
assert.Contains(t, text, "CONTR-202501-A1B2")
```

**Applied to**: PDF extraction tests

## Dependencies Required

### Testing Libraries

- **testify/assert**: Cleaner assertions (already in use)
- **testify/mock**: Mock generation (already in use)
- **testify/require**: Fatal assertions for setup

### Test Data Creation

- PDF generation tool for creating test PDFs
- Sample invoice templates with/without Invoice IDs

## Test Execution

### Running Tests

```bash
# Run all invoice email processor tests
go test ./pkg/service/invoiceemail -v

# Run all Gmail inbox service tests
go test ./pkg/service/googlemail -v -run TestListInboxMessages

# Run with coverage
go test ./pkg/service/invoiceemail -cover

# Run specific test
go test ./pkg/service/invoiceemail -run TestExtractInvoiceID_FromSubject_Success
```

### CI Integration

Tests should be added to existing CI pipeline:

```bash
make test  # Includes all new tests
```

## Quality Assurance Criteria

### Test Quality Standards

- [ ] All test cases have clear names describing what they test
- [ ] All test cases include setup, execution, and verification sections
- [ ] Error paths tested as thoroughly as success paths
- [ ] Edge cases identified and tested
- [ ] No flaky tests (deterministic, no timing dependencies)
- [ ] Tests are independent (can run in any order)
- [ ] Mocks properly cleaned up between tests

### Documentation Quality

- [ ] Each test case documented with purpose
- [ ] Input/setup clearly specified
- [ ] Expected output documented
- [ ] Edge cases explained
- [ ] Implementation examples provided

## Risk Areas and Mitigation

### Risk 1: PDF Parsing Complexity

**Risk**: pdfcpu library behavior may differ from expectations

**Mitigation**:
- Extensive test coverage with various PDF formats
- Clear error messages for debugging
- Fallback to manual processing documented

### Risk 2: Regex Pattern Validity

**Risk**: Invoice ID pattern may not match all real-world formats

**Mitigation**:
- 20+ regex test cases covering valid and invalid patterns
- Easy to update regex if requirements change
- Clear documentation of expected format

### Risk 3: Mock Limitations

**Risk**: Mocks may not represent real API behavior

**Mitigation**:
- Integration tests planned for future phase (out of scope now)
- Real API error scenarios documented in mocks
- Mock responses based on actual Gmail API documentation

### Risk 4: Partial Failure Handling

**Risk**: Label applied but Notion update fails (or vice versa)

**Mitigation**:
- Specific tests for partial failure scenarios
- Idempotent operations allow safe retries
- Clear logging for manual intervention

## Next Steps

### For Feature Implementer

1. **Review Test Specifications**: Understand all test cases before implementing
2. **Create Test Data**: Generate required PDF test files
3. **Implement Tests First**: Write tests following specifications
4. **Implement Features**: Make tests pass with minimal code
5. **Verify Coverage**: Ensure 90%+ code coverage achieved
6. **Run Full Suite**: Verify all tests pass

### For QA Agent

1. **Review Test Results**: Verify all tests pass
2. **Check Coverage**: Confirm coverage targets met
3. **Manual Testing**: Test edge cases not covered by unit tests
4. **Integration Testing**: Plan integration test scenarios (future phase)

## References

### Planning Documents

- **ADR**: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/ADRs/001-email-listener-architecture.md`
- **Gmail Service Spec**: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/specifications/gmail-inbox-service.md`
- **Processor Spec**: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/specifications/invoice-email-processor.md`
- **Configuration Spec**: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/specifications/configuration.md`

### Test Specifications

- **Gmail Inbox Tests**: `/docs/sessions/202601100842-contractor-invoice-email-listener/test-cases/unit/gmail-inbox-service-tests.md`
- **Processor Tests**: `/docs/sessions/202601100842-contractor-invoice-email-listener/test-cases/unit/invoice-processor-tests.md`
- **PDF Parser Tests**: `/docs/sessions/202601100842-contractor-invoice-email-listener/test-cases/unit/pdf-parser-tests.md`

### Existing Test Examples

- **Google Mail Tests**: `/pkg/service/googlemail/google_mail_sendas_test.go`
- **NocoDB Tests**: `/pkg/service/nocodb/accounting_todo_test.go`
- **Discord Tests**: `/pkg/service/discord/discord_test.go`

## Approval Checklist

- [x] All critical paths have test specifications
- [x] Error scenarios thoroughly covered
- [x] Edge cases identified and tested
- [x] Test patterns align with codebase standards
- [x] Test data requirements documented
- [x] Mock strategies defined
- [x] Coverage goals established
- [x] Implementation guidance provided

## Status Summary

**Test Specification Phase**: COMPLETE

All unit test specifications have been created following TDD principles and the project's established testing patterns. The specifications provide comprehensive coverage of:

- Gmail inbox operations (40+ tests)
- Invoice email processing business logic (25+ tests)
- PDF parsing and Invoice ID extraction (35+ tests)

**Total**: 100+ test cases specified, ready for implementation.

**Next Phase**: Feature Implementation (feature-implementer agent)
