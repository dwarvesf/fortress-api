# Concurrency Optimization: SyncTaskOrderLogs Endpoint

## Overview

Optimize the `/cronjobs/sync-task-order-logs` endpoint to process contractors concurrently instead of sequentially, reducing execution time from ~60s to ~6-15s for typical workloads (20 contractors, 60 projects).

## Implementation Approach

**Pattern**: Worker pool with configurable concurrency
**Model**: Follow existing `InitTaskOrderLogs` implementation (lines 729-800)
**Strategy**: Test-Driven Development (TDD)

## Architecture

```
Request
  |
  v
SyncTaskOrderLogs Handler
  |
  +-- Group timesheets by contractor
  |
  +-- Create job channel (buffered)
  |
  +-- Spawn N workers (configurable)
  |     |
  |     +-- Read jobs from channel
  |     +-- Call processContractorSync()
  |     +-- Send result to result channel
  |
  +-- Collect results from channel
  |
  +-- Aggregate counts and details
  |
  v
Response (unchanged format)
```

## Task Breakdown

### PHASE 1: Setup & Verification

#### Task 1.1: Verify Service Thread-Safety
**Status**: [ ] Not Started
**Files**:
- `pkg/service/notion/task_order_log.go`
- `pkg/service/openrouter/openrouter.go`

**Actions**:
- [ ] Review TaskOrderLogService for shared mutable state
- [ ] Review OpenRouter service HTTP client thread-safety
- [ ] Document findings and any concerns
- [ ] Identify if mutex protection needed

**Acceptance**:
- Services confirmed safe for concurrent use OR
- Mutex protection added where needed

---

### PHASE 2: Configuration Layer

#### Task 2.1: Add Configuration Field
**Status**: [ ] Not Started
**File**: `pkg/config/config.go`

**Changes**:
```go
type Config struct {
    // ... existing fields ...
    TaskOrderLogWorkerPoolSize int
}
```

**Implementation**:
- [ ] Add field to Config struct (after line 69)
- [ ] Add parsing in Generate() function with getIntWithDefault()
- [ ] Add validation: min 1, max 20, default 5
- [ ] Add environment variable: TASK_ORDER_LOG_WORKER_POOL_SIZE

**Validation Logic**:
```go
workerPoolSize := getIntWithDefault(v, "TASK_ORDER_LOG_WORKER_POOL_SIZE", 5)
if workerPoolSize < 1 {
    workerPoolSize = 1
}
if workerPoolSize > 20 {
    workerPoolSize = 20
}
```

**Acceptance**:
- Config field added
- Default value is 5
- Validation enforces bounds (1-20)
- Environment variable works

#### Task 2.2: Test Configuration
**Status**: [ ] Not Started
**File**: `pkg/config/config_test.go` (create if needed)

**Test Cases**:
- [ ] Test default value (no env var) = 5
- [ ] Test custom value (TASK_ORDER_LOG_WORKER_POOL_SIZE=10) = 10
- [ ] Test validation: 0 -> 1
- [ ] Test validation: 25 -> 20
- [ ] Test validation: -5 -> 1

**Acceptance**:
- All config tests pass
- Validation works correctly

---

### PHASE 3: Test-First Development

#### Task 3.1: Create Test File Structure
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log_test.go`

**Actions**:
- [ ] Create test file with package declaration
- [ ] Import required packages (testing, context, sync, mock)
- [ ] Set up test helper functions
- [ ] Create mock structures if needed

**Acceptance**:
- Test file compiles
- Test structure ready for test cases

#### Task 3.2: Write Correctness Tests (BEFORE Implementation)
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log_test.go`

**Test Cases** (write these BEFORE implementing concurrency):

**Test 1: Single Contractor Processing**
```go
func TestSyncTaskOrderLogs_SingleContractor(t *testing.T)
```
- Setup: 1 contractor, 3 projects
- Expected: Same result as sequential
- Verify: Counts, details, response format

**Test 2: Multiple Contractors Processing**
```go
func TestSyncTaskOrderLogs_MultipleContractors(t *testing.T)
```
- Setup: 5 contractors, varying projects
- Expected: All contractors processed
- Verify: Total counts match sum of individual results

**Test 3: Error Isolation**
```go
func TestSyncTaskOrderLogs_ErrorIsolation(t *testing.T)
```
- Setup: 3 contractors, 1 fails
- Expected: Other 2 continue processing
- Verify: 2 successful results, 1 error logged

**Test 4: Empty Timesheets**
```go
func TestSyncTaskOrderLogs_EmptyTimesheets(t *testing.T)
```
- Setup: No timesheets
- Expected: Return zero counts
- Verify: No errors, proper response

**Test 5: Context Cancellation**
```go
func TestSyncTaskOrderLogs_ContextCancellation(t *testing.T)
```
- Setup: Cancel context during processing
- Expected: Graceful shutdown
- Verify: No panic, partial results returned

**Acceptance**:
- All test cases written and documented
- Tests fail initially (expected - no implementation yet)
- Test structure is clear and maintainable

---

### PHASE 4: Refactor for Concurrency

#### Task 4.1: Define Result Structures
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log.go`

**Add structures** (near existing contractorProcessResult at line 641):
```go
type contractorJob struct {
    contractorID string
    timesheets   []*notion.TimesheetEntry
}

type contractorSyncResult struct {
    contractorID     string
    ordersCreated    int
    lineItemsCreated int
    lineItemsUpdated int
    detail           map[string]any
    err              error
}
```

**Acceptance**:
- Structures defined
- Code compiles
- No changes to existing functionality yet

#### Task 4.2: Extract Processing Method
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log.go`

**Refactoring**:
Extract lines 118-282 into new method:
```go
func (h *handler) processContractorSync(
    ctx context.Context,
    contractorID string,
    contractorTimesheets []*notion.TimesheetEntry,
    month string,
) contractorSyncResult {
    // Move existing contractor processing logic here
    // Keep logic IDENTICAL to current implementation
    // Just wrapped in a method
}
```

**Critical**: Keep logic 100% identical, just extracted into method

**Changes in SyncTaskOrderLogs**:
- Replace lines 118-282 with call to processContractorSync()
- Keep sequential loop structure for now
- Aggregate results same way

**Acceptance**:
- Code compiles
- Existing behavior unchanged
- No test regressions
- Ready for concurrent refactor

#### Task 4.3: Verify No Regressions
**Status**: [ ] Not Started

**Actions**:
- [ ] Run: `make build`
- [ ] Run: `make test`
- [ ] Run: `make lint`
- [ ] Manual test if possible

**Acceptance**:
- All tests pass
- No compilation errors
- Linter happy
- Functionality unchanged

---

### PHASE 5: Implement Concurrency

#### Task 5.1: Implement Worker Pool
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log.go`

**Implementation Pattern** (follow InitTaskOrderLogs lines 729-800):

```go
// Get worker pool size from config
numWorkers := h.config.TaskOrderLogWorkerPoolSize
if numWorkers == 0 {
    numWorkers = 5 // fallback default
}
if numWorkers > len(contractorGroups) {
    numWorkers = len(contractorGroups) // don't spawn more workers than jobs
}

l.Info(fmt.Sprintf("processing %d contractors with %d workers",
    len(contractorGroups), numWorkers))

// Create channels
jobs := make(chan contractorJob, len(contractorGroups))
results := make(chan contractorSyncResult, len(contractorGroups))

// Spawn workers
var wg sync.WaitGroup
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go func(workerID int) {
        defer wg.Done()
        for job := range jobs {
            l.Debug(fmt.Sprintf("[Worker %d] processing contractor: %s",
                workerID, job.contractorID))

            result := h.processContractorSync(ctx, job.contractorID,
                job.timesheets, month)

            l.Debug(fmt.Sprintf("[Worker %d] finished contractor: %s",
                workerID, job.contractorID))

            results <- result
        }
    }(i)
}

// Send jobs
for contractorID, timesheets := range contractorGroups {
    jobs <- contractorJob{
        contractorID: contractorID,
        timesheets:   timesheets,
    }
}
close(jobs)

// Close results when all workers done
go func() {
    wg.Wait()
    close(results)
}()

// Collect results
var (
    ordersCreated        = 0
    lineItemsCreated     = 0
    lineItemsUpdated     = 0
    contractorsProcessed = 0
    details              = []map[string]any{}
)

for result := range results {
    if result.err != nil {
        l.Error(result.err, fmt.Sprintf("error processing contractor: %s",
            result.contractorID))
        continue
    }

    if result.ordersCreated > 0 {
        ordersCreated += result.ordersCreated
    }
    lineItemsCreated += result.lineItemsCreated
    lineItemsUpdated += result.lineItemsUpdated
    contractorsProcessed++
    details = append(details, result.detail)
}
```

**Acceptance**:
- Worker pool implemented
- Jobs distributed to workers
- Results collected safely
- No data races

#### Task 5.2: Add Concurrency Logging
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log.go`

**Log Messages to Add**:
- Start: "processing X contractors with Y workers"
- Worker start: "[Worker N] processing contractor: ID"
- Worker end: "[Worker N] finished contractor: ID"
- Complete: "concurrent processing complete"

**Acceptance**:
- Clear visibility into concurrent execution
- Can trace which worker processed which contractor
- Logs helpful for debugging

#### Task 5.3: Handle Context Cancellation
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log.go`

**Implementation**:
Add context check in processContractorSync before expensive operations:
```go
func (h *handler) processContractorSync(...) contractorSyncResult {
    // Check context at start
    if ctx.Err() != nil {
        return contractorSyncResult{
            contractorID: contractorID,
            err: ctx.Err(),
        }
    }

    // ... existing logic ...

    // Check context before OpenRouter call
    if ctx.Err() != nil {
        return contractorSyncResult{
            contractorID: contractorID,
            err: ctx.Err(),
        }
    }
}
```

**Acceptance**:
- Context cancellation handled gracefully
- No goroutine leaks
- Partial results returned if cancelled mid-processing

---

### PHASE 6: Testing & Validation

#### Task 6.1: Run Unit Tests
**Status**: [ ] Not Started

**Commands**:
```bash
go test ./pkg/handler/notion -v
go test ./pkg/handler/notion -run TestSyncTaskOrderLogs -v
```

**Acceptance**:
- All tests pass
- Tests from Phase 3 now pass with concurrent implementation

#### Task 6.2: Run Race Detection
**Status**: [ ] Not Started

**Commands**:
```bash
go test ./pkg/handler/notion -race
go test ./pkg/handler/notion -race -run TestSyncTaskOrderLogs
```

**Acceptance**:
- No race conditions detected
- All tests pass with -race flag

#### Task 6.3: Write Concurrency-Specific Tests
**Status**: [ ] Not Started
**File**: `pkg/handler/notion/task_order_log_test.go`

**Additional Test Cases**:

**Test 6: Different Worker Pool Sizes**
```go
func TestSyncTaskOrderLogs_WorkerPoolSizes(t *testing.T)
```
- Test with 1, 5, 10, 20 workers
- Verify same results regardless of pool size
- Verify no more workers spawned than contractors

**Test 7: Concurrent Result Aggregation**
```go
func TestSyncTaskOrderLogs_ConcurrentAggregation(t *testing.T)
```
- Setup: Many contractors to force concurrent execution
- Verify: Total counts are correct
- Verify: No results lost during aggregation

**Test 8: Performance Improvement**
```go
func TestSyncTaskOrderLogs_PerformanceImprovement(t *testing.T)
```
- Measure: Sequential vs concurrent execution time
- Verify: Concurrent is faster (with artificial delays)

**Acceptance**:
- All concurrency tests pass
- Performance improvement demonstrated

#### Task 6.4: Integration Testing
**Status**: [ ] Not Started

**Test Scenarios**:
- [ ] Test with realistic data (20 contractors, 60 projects)
- [ ] Measure execution time improvement
- [ ] Verify response format unchanged
- [ ] Test error scenarios (API failures, timeouts)
- [ ] Test with actual Notion/OpenRouter services (if safe)

**Performance Baseline**:
- Sequential: ~60s (20 contractors, 60 projects)
- Target: 6-15s (70-90% improvement)

**Acceptance**:
- Performance target met
- Response format unchanged
- All error scenarios handled gracefully

---

### PHASE 7: Documentation & Handoff

#### Task 7.1: Update Implementation Status
**Status**: [ ] Not Started
**File**: `docs/sessions/202601161133-update-timesheet-to-project-updates/implementation/STATUS.md`

**Updates**:
- Add section: "Concurrency Optimization"
- Document performance improvements
- Note configuration changes
- List all modified files

**Acceptance**:
- STATUS.md updated
- Changes documented clearly

#### Task 7.2: Update API Documentation
**Status**: [ ] Not Started

**Actions**:
- [ ] Review Swagger annotations if changes needed
- [ ] Update any performance notes in documentation
- [ ] Note configuration option (TASK_ORDER_LOG_WORKER_POOL_SIZE)

**Acceptance**:
- Documentation accurate
- Configuration documented

---

## Success Criteria

### Functional Requirements
- [ ] Same response format as sequential version
- [ ] Isolated error handling (one failure doesn't stop others)
- [ ] All existing tests pass
- [ ] New tests pass

### Performance Requirements
- [ ] Execution time reduced from ~60s to ~6-15s (typical workload)
- [ ] Configurable worker pool size (1-20)
- [ ] No performance degradation for small workloads

### Quality Requirements
- [ ] No race conditions (`-race` flag passes)
- [ ] No goroutine leaks
- [ ] Proper context handling
- [ ] Clear logging with concurrency markers

### Configuration Requirements
- [ ] TASK_ORDER_LOG_WORKER_POOL_SIZE environment variable
- [ ] Default: 5 workers
- [ ] Min: 1, Max: 20
- [ ] Validation enforced

## Files Modified

1. `pkg/config/config.go` - Configuration
2. `pkg/handler/notion/task_order_log.go` - Concurrent implementation
3. `pkg/handler/notion/task_order_log_test.go` - Tests (new file)
4. `pkg/config/config_test.go` - Config tests (if needed)
5. `docs/sessions/202601161133-update-timesheet-to-project-updates/implementation/STATUS.md` - Documentation

## Risk Mitigation

### Identified Risks
1. **Service Thread-Safety**: Notion/OpenRouter services might not be thread-safe
   - Mitigation: Phase 1 verification, add mutex if needed

2. **API Rate Limiting**: Concurrent requests might trigger rate limits
   - Mitigation: Configurable pool size (can reduce to 1)

3. **Memory Overhead**: Holding results during processing
   - Mitigation: Controlled by worker pool size

4. **Context Cancellation**: Goroutines might leak if not handled
   - Mitigation: Proper context checking in workers

### Rollback Plan
If issues arise:
1. Set TASK_ORDER_LOG_WORKER_POOL_SIZE=1 (sequential processing)
2. Revert code changes if critical bugs found
3. Feature flag could be added if gradual rollout needed

## Testing Strategy

### Test Pyramid
```
              Integration Tests (Performance, E2E)
                    /               \
              Unit Tests              Concurrency Tests
           (Correctness)              (Race Detection)
                  |                           |
            Config Tests                 Thread Safety
```

### Test Execution Order
1. Config tests (verify configuration)
2. Unit tests (verify correctness)
3. Concurrency tests (verify thread safety)
4. Race detection (verify no race conditions)
5. Integration tests (verify performance)

## Notes

- **TDD Approach**: Tests written BEFORE implementation
- **Pattern Reuse**: Follow existing InitTaskOrderLogs implementation
- **No Breaking Changes**: Response format unchanged
- **Backward Compatible**: Can run sequentially with pool size 1
- **Observable**: Clear logging for debugging and monitoring
