# Implementation Tasks

## Overview

This document breaks down the implementation of the task order confirmation email update into actionable, sequenced tasks.

**Session**: 202601091229-update-task-order-email  
**Total Estimated Tasks**: 6 main tasks  
**Dependencies**: Follow sequence order

---

## Task 1: Add Service Layer Methods

**File**: `pkg/service/notion/task_order_log.go`  
**Priority**: HIGH  
**Dependencies**: None  
**Estimated Effort**: 1-2 hours

### Subtasks

#### 1.1: Add GetContractorPayday Method
- [ ] Add method signature with proper documentation
- [ ] Implement database config validation
- [ ] Build Notion query with filters (Contractor + Status="Active")
- [ ] Execute query and handle results
- [ ] Extract Payday field using Select property
- [ ] Parse Payday value ("01" → 1, "15" → 15)
- [ ] Implement graceful fallbacks (return 0 for missing/invalid)
- [ ] Add debug logging at key points
- [ ] Handle all error cases properly

**Location**: After `GetContractorPersonalEmail` (~line 1630)

**Reference**: SPEC-002, contractor_payables.go lines 580-630

#### 1.2: Add extractSelect Helper Method
- [ ] Add method to extract Select field from properties
- [ ] Handle missing property gracefully
- [ ] Handle null Select value
- [ ] Return empty string for invalid cases

**Location**: After `extractRichText` (~line 897)

### Acceptance Criteria
- [ ] Method compiles without errors
- [ ] Follows existing code patterns
- [ ] Debug logs present for all code paths
- [ ] Returns (0, nil) for fallback cases
- [ ] Returns (0, error) only for system errors

---

## Task 2: Update Data Model

**File**: `pkg/model/email.go`  
**Priority**: HIGH  
**Dependencies**: None (can run parallel with Task 1)  
**Estimated Effort**: 15 minutes

### Subtasks

#### 2.1: Add New Fields to TaskOrderConfirmationEmail
- [ ] Add `InvoiceDueDay string` field with comment
- [ ] Add `Milestones []string` field with comment
- [ ] Verify struct alignment
- [ ] Update any struct documentation

**Location**: Lines ~15-20

### Acceptance Criteria
- [ ] Fields added to struct
- [ ] Code compiles
- [ ] No breaking changes to existing code
- [ ] Comments explain field purpose

---

## Task 3: Update Handler Logic

**File**: `pkg/handler/notion/task_order_log.go`  
**Priority**: HIGH  
**Dependencies**: Task 1 (GetContractorPayday), Task 2 (Model)  
**Estimated Effort**: 1 hour

### Subtasks

#### 3.1: Fetch Payday and Calculate Due Date
- [ ] Call `GetContractorPayday()` after client info extraction
- [ ] Handle errors gracefully (log but continue)
- [ ] Implement due date calculation logic:
  - payday == 15 → "25th"
  - payday == 1 OR payday == 0 → "10th"
- [ ] Add debug logging for payday and due date
- [ ] Store result in `invoiceDueDay` variable

**Location**: After line 493 (after `detail["clients"]`)

#### 3.2: Create Mock Milestones Array
- [ ] Define milestones array with 2-3 mock entries
- [ ] Add TODO comment: "TODO: Replace with real data source"
- [ ] Make format consistent with template expectations

#### 3.3: Populate Email Data Model
- [ ] Add `InvoiceDueDay: invoiceDueDay` to struct initialization
- [ ] Add `Milestones: milestones` to struct initialization
- [ ] Verify all fields populated correctly

**Location**: Line ~500 (emailData creation)

### Acceptance Criteria
- [ ] Payday fetch works without blocking email
- [ ] Due date calculated correctly
- [ ] Mock milestones clearly marked
- [ ] All new fields populated
- [ ] Debug logs present

---

## Task 4: Update Template Functions and Signature

**Priority**: HIGH  
**Dependencies**: Task 2 (Model)  
**Estimated Effort**: 30 minutes

### Subtasks

#### 4.1: Update Template Functions in utils.go
**File**: `pkg/service/googlemail/utils.go`  
**Location**: Lines ~93-139

- [ ] Add `invoiceDueDay` function returning `data.InvoiceDueDay`
- [ ] Update `signatureName` to return "Han Ngo"
- [ ] Update `signatureTitle` to return "CTO & Managing Director"
- [ ] Remove unused functions: `periodEndDay`, `monthName`, `year`
- [ ] Verify template.FuncMap syntax correct

#### 4.2: Update Signature in task_order_log.go
**File**: `pkg/service/notion/task_order_log.go`  
**Location**: Lines ~1821-1830

- [ ] Update `signatureName` to return "Han Ngo"
- [ ] Update `signatureTitle` to return "CTO & Managing Director"
- [ ] Keep `signatureNameSuffix` as empty string
- [ ] Verify consistency with utils.go changes

### Acceptance Criteria
- [ ] Both files updated identically
- [ ] Template functions compile
- [ ] Signature values match requirements
- [ ] No syntax errors in FuncMap

---

## Task 5: Update Email Template

**File**: `pkg/templates/taskOrderConfirmation.tpl`  
**Priority**: MEDIUM  
**Dependencies**: Task 4 (Template functions)  
**Estimated Effort**: 30 minutes

### Subtasks

#### 5.1: Replace Template Content
- [ ] Update From header (keep "Spawn @ Dwarves LLC")
- [ ] Update Subject line
- [ ] Replace body with new content structure:
  - Greeting
  - Opening message
  - Invoice reminder with due date
  - Client milestones section
  - Encouragement
  - Support offer
  - Closing
- [ ] Keep signature template include
- [ ] Preserve MIME format and boundaries
- [ ] Verify HTML structure valid

### Acceptance Criteria
- [ ] Subject matches requirement
- [ ] From header unchanged
- [ ] All template variables present
- [ ] Milestones section uses {{range}}
- [ ] Signature included via template
- [ ] MIME format preserved

---

## Task 6: Add Unit Tests

**Priority**: HIGH  
**Dependencies**: Tasks 1-5 complete  
**Estimated Effort**: 2-3 hours

### Subtasks

#### 6.1: Test GetContractorPayday Method
**File**: `pkg/service/notion/task_order_log_test.go`

- [ ] Test Payday="01" returns (1, nil)
- [ ] Test Payday="15" returns (15, nil)
- [ ] Test no Service Rate found returns (0, nil)
- [ ] Test empty Payday field returns (0, nil)
- [ ] Test invalid Payday returns (0, nil)
- [ ] Test config error returns (0, error)
- [ ] Test API error returns (0, error)
- [ ] Use table-driven test pattern

**Reference**: `test-cases/unit/get-contractor-payday-tests.md`

#### 6.2: Test Invoice Due Date Calculation
**File**: `pkg/handler/notion/task_order_log_test.go`

- [ ] Test payday=1 → "10th"
- [ ] Test payday=15 → "25th"
- [ ] Test payday=0 → "10th"
- [ ] Test invalid payday → "10th"
- [ ] Test error from service → "10th"

**Reference**: `test-cases/unit/invoice-due-date-calculation-tests.md`

#### 6.3: Test Template Rendering (Optional)
**File**: `pkg/service/googlemail/utils_test.go`

- [ ] Test template renders with all fields
- [ ] Test template with empty milestones
- [ ] Test signature functions return correct values
- [ ] Verify HTML structure

### Acceptance Criteria
- [ ] All unit tests pass
- [ ] Test coverage >90% for new code
- [ ] Table-driven tests used where appropriate
- [ ] Mocks properly configured

---

## Verification & Testing Tasks

### Manual Testing Checklist

- [ ] **Local Email Test**
  ```bash
  curl -X POST "http://localhost:8080/api/v1/cronjobs/send-task-order-confirmation?month=2026-01&test_email=your@email.com"
  ```

- [ ] **Verify Email Content**
  - [ ] Subject: "Quick update for January 2026 – Invoice reminder & client milestones"
  - [ ] From: "Spawn @ Dwarves LLC"
  - [ ] Due date shows "10th" or "25th"
  - [ ] Milestones render as bullet list
  - [ ] Signature: "Han Ngo, CTO & Managing Director"
  - [ ] HTML renders correctly in Gmail

- [ ] **Test Contractor-Specific**
  ```bash
  curl -X POST "http://localhost:8080/api/v1/cronjobs/send-task-order-confirmation?month=2026-01&discord=contractor_username&test_email=your@email.com"
  ```

- [ ] **Test Payday Variations**
  - [ ] Contractor with Payday=1 shows "10th"
  - [ ] Contractor with Payday=15 shows "25th"
  - [ ] Contractor without Payday shows "10th" (default)

### Code Quality Checks

- [ ] Run linter: `make lint`
- [ ] Run tests: `make test`
- [ ] Check test coverage
- [ ] Verify no compiler warnings
- [ ] Review debug log output

### Pre-Deployment Checklist

- [ ] All unit tests passing
- [ ] Manual email tests successful
- [ ] Code reviewed by team
- [ ] Documentation updated
- [ ] Rollback plan confirmed
- [ ] Monitoring/alerts configured

---

## Implementation Sequence

### Phase 1: Core Functionality (Tasks 1-3)
1. Task 1: Service methods (can run parallel with Task 2)
2. Task 2: Data model (can run parallel with Task 1)
3. Task 3: Handler logic (depends on 1 & 2)

### Phase 2: Presentation (Task 4-5)
4. Task 4: Template functions & signature
5. Task 5: Email template

### Phase 3: Validation (Task 6)
6. Task 6: Unit tests

### Phase 4: Verification
7. Manual testing
8. Code quality checks
9. Pre-deployment validation

---

## Rollback Procedures

### If Issues Found in Testing

#### Template Only
```bash
git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl
```

#### Full Rollback
```bash
git revert <commit-hash>
```

### Emergency Disable
Comment out route registration temporarily if critical issue found.

---

## Success Criteria

Implementation complete when:
- [ ] All 6 main tasks completed
- [ ] All unit tests passing
- [ ] Manual email tests successful
- [ ] Code reviewed and approved
- [ ] No linting errors
- [ ] Documentation updated
- [ ] Ready for deployment

---

**Total Estimated Time**: 6-8 hours  
**Recommended Approach**: Complete in sequence, test incrementally  
**Critical Path**: Task 1 → Task 3 → Task 6

---

**Created**: 2026-01-09  
**Session**: 202601091229-update-task-order-email  
**Ready for**: Implementation via `proceed` command
