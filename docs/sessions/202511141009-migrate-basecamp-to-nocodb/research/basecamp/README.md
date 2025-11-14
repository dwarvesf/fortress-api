# Session: Basecamp Service Exploration

**Date**: 2025-11-14
**Session ID**: 202511141009
**Type**: Codebase Exploration & Analysis

## Table of Contents

1. [Overview](#overview)
2. [Basecamp Service Analysis](#basecamp-service-analysis)
3. [Interface Architecture](#interface-architecture)
4. [Invoice Workflow Analysis](#invoice-workflow-analysis)
5. [Key Findings](#key-findings)
6. [References](#references)

---

## Overview

This session documents the exploration of the Basecamp service integration to understand its architecture, usage patterns, and potential for migration to alternative systems like NocoDB.

### Exploration Goals

- Understand current Basecamp integration architecture
- Map out all service interfaces and their implementations
- Document invoice creation workflow and Basecamp integration
- Identify usage patterns across the codebase
- Assess architectural strengths and technical debt

---

## Basecamp Service Analysis

### Architecture Overview

**Location**: `pkg/service/basecamp/`

The basecamp service is a comprehensive integration layer consisting of **28 Go files** organized into **12 sub-services**.

#### Core Components

1. **Main Service** (`basecamp.go:26-48`)
   - Aggregates all sub-services
   - Provides helper methods for comments and mentions
   - Handles expense-related business logic

2. **OAuth2 Client** (`client/client.go`)
   - Uses `golang.org/x/oauth2` for authentication
   - Manages token refresh automatically
   - Wraps http.Client with OAuth2 credentials

3. **Sub-Services** (All are interfaces)
   - Attachment, Campfire, Comment, MsgBoard
   - People, Project, Recording, Schedule
   - Subscription, Todo, Webhook
   - Each initialized with OAuth client

4. **Data Models** (`model/model.go`)
   - Complete type definitions for Basecamp API entities
   - Person, Todo, Project, Comment, Recording, Event types
   - Message types for worker queue integration

### Key Functionality

#### Expense Management (`integration.go:41-164`)

- Extracts expense amounts from formatted strings (supports k/tr/m suffixes)
- Creates expenses with currency conversion via Wise API
- Links expenses to accounting transactions
- Handles expense deletion on todo unchecking

**Flow**:
```
Basecamp Webhook → Extract Expense Data → Create Expense Record
  ↓
Wise Currency Conversion → Create Accounting Transaction → Update Expense
```

#### Integration Points

- **Database stores**: Employee, Currency, Expense, Accounting
- **Wise service**: Currency conversion
- **Worker queue**: Async comment posting
- **Discord**: Audit logging

### Usage Analysis

The basecamp service is used in **37 files** across the codebase:

#### Primary Consumers

1. **Webhook Handlers** (15 files)
   - `pkg/handler/webhook/basecamp_expense.go` - Expense validation/creation
   - `pkg/handler/webhook/basecamp_invoice.go` - Invoice marking as paid
   - `pkg/handler/webhook/basecamp_accounting.go` - Accounting transactions
   - `pkg/handler/webhook/onleave.go` - Leave request processing

2. **Controllers** (5 files)
   - `pkg/controller/invoice/send.go` - Posting invoices to Basecamp
   - `pkg/controller/invoice/commission.go` - Commission tracking
   - `pkg/controller/employee/update_employee_status.go` - Employee updates

3. **Background Workers**
   - `pkg/worker/worker.go` - Async processing of Basecamp comments/todos

4. **Other Handlers**
   - Payroll, Discord, Accounting integration points

### Architectural Strengths

✅ **Modular design** with clear separation of concerns (28 Go files)
✅ **Consistent patterns** across all sub-services
✅ **Type-safe models** with GORM tags and JSON serialization
✅ **Comprehensive logging** with structured fields
✅ **Proper error handling** with context propagation
✅ **Interface-based design** enabling testability and swappability

### Technical Debt & Issues

#### 🔴 Critical Priority

**1. Hardcoded Resource IDs** (`consts/consts.go:4-131`)

- **Issue**: 130+ hardcoded Basecamp resource IDs
- **Examples**:
  ```go
  WoodlandID = 9403032
  HanBasecampID = 21562923
  AccountingTodoID = 2329633561
  ```
- **Impact**: Any Basecamp reorganization requires code changes and redeployment
- **Recommendation**: Implement database-backed resource registry

#### 🟡 High Priority

**2. Duplicated Pagination Logic**

- **Issue**: Same pagination code in 7+ services
- **Locations**:
  - `todo/todo.go:171-195, 245-270`
  - `comment/comment.go:62-89`
  - `recording/recording.go:129-157`
- **Impact**: Bug fixes must be applied in multiple locations
- **Recommendation**: Extract to `pkg/utils/pagination.go`

**3. Synchronous API Bottlenecks** (`integration.go:100-108`)

- **Issue**: Blocking Wise API calls in critical path
- **Impact**: External API latency directly impacts user experience
- **Recommendation**: Async processing + cache exchange rates

**4. Missing Resilience Patterns**

- **Issue**: No rate limiting or circuit breaker for Basecamp API
- **Impact**: Vulnerable to API quota exhaustion and cascading failures
- **Recommendation**: Implement using `sony/gobreaker`

#### 🟢 Medium Priority

**5. No Caching Layer**

- **Issue**: Every request hits Basecamp API
- **Impact**: Higher latency and API usage
- **Recommendation**: Add Redis cache with TTL

**6. Scattered Environment Logic**

- **Issue**: Production vs non-prod checks throughout codebase
- **Recommendation**: Centralize environment-specific configuration

---

## Interface Architecture

### Investigation Summary

**Finding**: The root `basecamp.Service` is a **struct** (not an interface), but all 11 sub-services **are interfaces**.

### Service Structure

```go
// pkg/service/basecamp/basecamp.go

type Service struct {
    // Dependencies
    store  *store.Store
    repo   store.DBRepo
    config *config.Config
    logger logger.Logger

    Basecamp *model.Basecamp

    // All sub-services are INTERFACES ✅
    Client       client.Service
    Attachment   attachment.Service
    Campfire     campfire.Service
    Comment      comment.Service
    MsgBoard     messageboard.Service
    People       people.Service
    Project      project.Service
    Recording    recording.Service
    Schedule     schedule.Service
    Subscription subscription.Service
    Todo         todo.Service
    Webhook      webhook.Service

    Wise wise.IService
}
```

### Sub-Service Interfaces

All interfaces are defined in `service.go` files within each subdirectory:

#### Client Service (`client/service.go`)

```go
type Service interface {
    Get(url string) (*http.Response, error)
    Do(req *http.Request) (*http.Response, error)
    GetAccessToken(code, redirectURI string) (string, error)
}
```

#### Todo Service (`todo/service.go`)

```go
type Service interface {
    CreateList(projectID, todoSetID int, todoList model.TodoList) (*model.TodoList, error)
    CreateGroup(projectID, todoListID int, group model.TodoGroup) (*model.TodoGroup, error)
    Create(projectID, todoListID int, todo model.Todo) (*model.Todo, error)
    Get(url string) (*model.Todo, error)
    GetAllInList(todoListID, projectID int, query ...string) ([]model.Todo, error)
    GetGroups(todoListID, projectID int) ([]model.TodoGroup, error)
    GetLists(projectID, todoSetsID int) ([]model.TodoList, error)
    GetList(url string) (*model.TodoList, error)
    GetProjectsLatestIssue(projectNames []string) ([]*pkgmodel.ProjectIssue, error)
    CreateHiring(cv *pkgmodel.Candidate) error
    FirstOrCreateList(projectID, todoSetID int, todoListName string) (*model.TodoList, error)
    FirstOrCreateGroup(projectID, todoListID int, todoGroupName string) (*model.TodoGroup, error)
    FirstOrCreateTodo(projectID, todoListID int, todoName string) (*model.Todo, error)
    FirstOrCreateInvoiceTodo(projectID, todoListID int, invoice *pkgmodel.Invoice) (*model.Todo, error)
    Update(projectID int, todo model.Todo) (*model.Todo, error)
    Complete(projectID, todoID int) error
}
```

#### Comment Service (`comment/service.go`)

```go
type Service interface {
    Create(projectID int, recordingID int, comment *model.Comment) error
    Gets(projectID int, recordingID int) ([]model.Comment, error)
}
```

#### Project Service (`project/service.go`)

```go
type Service interface {
    GetAll() ([]model.Project, error)
    Get(id int64) (*model.Project, error)
}
```

#### People Service (`people/service.go`)

```go
type Service interface {
    GetByID(id int) (*model.Person, error)
    GetInfo(accessToken string) (model.UserInfo, error)
    Create(people model.PeopleEntry) error
    Remove(people model.PeopleEntry) error
    UpdateInProject(projectID int, people model.PeopleEntry) error
    GetAllOnProject(projectID int) ([]model.Person, error)
}
```

### Abstraction Pattern

**Current Implementation**:
- Each sub-service has:
  - **Service interface** (in `service.go`) - defines contract
  - **Concrete implementation struct** (in domain file)
  - **NewService() factory function** - returns interface type

**Example** (`comment/comment.go`):
```go
type CommentService struct {
    client client.Service  // Depends on interface
}

func NewService(client client.Service) Service {
    return &CommentService{client: client}
}

func (c *CommentService) Create(projectID int, recordingID int, comment *model.Comment) error {
    // Implementation using client.Service methods
}
```

### Dependency Injection Pattern

All sub-services:
- Accept `client.Service` interface in constructor
- Do NOT directly depend on HTTP client implementation
- Use only the interface for HTTP operations

**Root service factory** (`basecamp.go:50-77`):
```go
func New(store *store.Store, repo store.DBRepo, cfg *config.Config, bc *model.Basecamp, logger logger.Logger) *Service {
    c, err := client.NewClient(bc, cfg)
    if err != nil {
        logger.Error(err, "init basecamp service")
        return nil
    }

    return &Service{
        store:        store,
        repo:         repo,
        config:       cfg,
        logger:       logger,
        Basecamp:     bc,
        Client:       c,
        Attachment:   attachment.NewService(c),
        Campfire:     campfire.NewService(c, logger, cfg),
        Comment:      comment.NewService(c),
        MsgBoard:     messageboard.NewService(c),
        People:       people.NewService(c),
        Project:      project.NewService(c),
        Recording:    recording.NewService(c),
        Schedule:     schedule.NewService(c, logger),
        Subscription: subscription.NewService(c),
        Todo:         todo.NewService(c, cfg),
        Webhook:      webhook.NewService(c),
        Wise:         wise.New(cfg, logger),
    }
}
```

### Assessment for NocoDB Migration

**✅ EXCELLENT for adapter pattern**:

1. **Strong interface isolation** - Each sub-service depends only on abstract interfaces
2. **Factory pattern** - All services use `NewService()` returning interface types
3. **No hard dependencies** - No sub-service directly imports HTTP libraries
4. **Dependency inversion** - Services accept dependencies through constructor parameters

**What enables NocoDB migration**:
- Client interface can be swapped with NocoDB implementation
- Each sub-service can have NocoDB variant implementing same interface
- No changes needed to consumers (handlers, controllers)

---

## Invoice Workflow Analysis

### Invoice Creation: Manual vs Automated

**Finding**: Invoices are created **MANUALLY** via API endpoint. There is NO automatic monthly generation.

#### API Endpoint

**Route**: `POST /api/v1/invoices/send`
**Handler**: `pkg/handler/invoice/invoice.go:130` → `Send()`
**Permission**: `PermissionInvoiceRead`

#### Request Structure

**File**: `pkg/handler/invoice/request/request.go`

```go
type SendInvoiceRequest struct {
    IsDraft     bool
    ProjectID   view.UUID     // Required
    BankID      view.UUID     // Required
    Description string
    Note        string
    CC          []string
    LineItems   []InvoiceItem
    Email       string        // Required
    Total       float64       // Required
    Discount    float64
    Tax         float64
    SubTotal    float64
    InvoiceDate string        // Required
    DueDate     string        // Required
    Month       int           // 0-11 (0=January)
    Year        int
}
```

#### Invoice Creation Process

**File**: `pkg/controller/invoice/send.go`

1. **Validation** (lines 26-71)
   - Verify sender exists
   - Check bank account validity
   - Check project existence
   - Calculate next invoice number

2. **PDF Generation** (line 89)
   - Template: `/pkg/templates/invoice.html`
   - Tool: `go-wkhtmltopdf`

3. **Currency Conversion** (lines 94-98)
   - Converts to Vietnamese Dong using Wise API

4. **Save to Database** (lines 102-131)
   - Status: `InvoiceStatusDraft` or `InvoiceStatusSent`

5. **Asynchronous Tasks** (lines 136-175, when not draft)
   - Upload PDF to Google Cloud Storage
   - Send email via Google Mail API
   - Create/update Basecamp comment/todo
   - Post thread ID back to database

#### No Scheduled Generation

**Evidence**:
- ❌ No cron/scheduler library in `go.mod`
- ❌ No invoice generation in `/pkg/worker/worker.go`
- ❌ No invoice generation in cronjob routes
- ✅ Only manual API endpoint exists

**Existing Cronjobs** (for reference):
- Birthday notifications
- Leave request alerts
- Discord syncs
- Accounting todos

### How Invoices are Created in Basecamp

**File**: `pkg/controller/invoice/send.go:136-175`

When an invoice is sent (not draft), two Basecamp resources are created:

#### 1. Todo Item (Synchronous)

**Location**:
- **Project**: Accounting (15258324) or Playground (12984857)
- **List**: "Accounting | Month Year" (e.g., "Accounting | January 2025")
- **Group**: "In"

**Properties**:
- **Title**: `"{ProjectName} {Month}/{Year} - #{InvoiceNumber}"`
  - Example: `"Dwarves Foundation 1/2025 - #2025-FT-001"`
- **Description**: Invoice PDF as Basecamp attachment

**Method**: `service.Basecamp.Todo.FirstOrCreateInvoiceTodo()`
**File**: `pkg/service/basecamp/todo/todo.go:407-431`

**Logic**:
```go
func (t *TodoService) FirstOrCreateInvoiceTodo(projectID, todoListID int, invoice *pkgmodel.Invoice) (*model.Todo, error) {
    invoiceTodoName := fmt.Sprintf(`%v %v/%v - #%v`, invoice.Project.Name, invoice.Month, invoice.Year, invoice.Number)
    todos, err := t.GetAllInList(todoListID, projectID)

    // Check for exact match
    for i := range todos {
        if todos[i].Title == invoiceTodoName {
            return &todos[i], nil
        }
        // Check for same project/month/year - update description
        if re.MatchString(utils.RemoveAllSpace(todos[i].Title)) {
            todos[i].Content = invoiceTodoName
            todos[i].Description = fmt.Sprintf(`<div>%v%v</div>`, todos[i].Description, invoice.TodoAttachment)
            return t.Update(projectID, todos[i])
        }
    }

    // Create new if not found
    return t.Create(projectID, todoListID, model.Todo{
        Content: invoiceTodoName,
        Description: fmt.Sprintf(`<div>%v</div>`, invoice.TodoAttachment)
    })
}
```

#### 2. Comment on Todo (Asynchronous)

**Content**:
```
#Invoice {number} has been sent

Confirm Command: Paid @Giang #{number}
```

**Attachment**: Invoice PDF file

**Processing**:
- **Queue**: Worker queue message type `basecamp_comment`
- **Worker**: `pkg/worker/worker.go:60-68` → `handleCommentMessage()`
- **Method**: `service.Basecamp.Comment.Create()`

**Flow**:
```go
// pkg/controller/invoice/send.go:170-175
h.worker.Enqueue(bcModel.BasecampCommentMsg, bcModel.BasecampCommentMessage{
    ProjectID:   accountingProjectID,
    RecordingID: invoiceTodo.ID,
    Payload: &bcModel.Comment{
        Content: fmt.Sprintf("#Invoice %v has been sent\n\nConfirm Command: Paid @Giang #%v", iv.Number, iv.Number),
    },
})
```

#### Complete Flow Diagram

```
Invoice Send Request
  ↓
Generate PDF + Save to DB
  ↓
[Async Task 1] Upload to Google Drive
  ↓
[Async Task 2] Send Email + Create Basecamp Records
  ↓
  ├─ Create Basecamp Attachment (PDF)
  ├─ Get Accounting Project ID (prod/playground)
  ├─ Find/Create "Accounting | Month Year" TodoList
  ├─ Find/Create "In" TodoGroup
  ├─ Find or Create/Update Invoice Todo
  │    ↓
  │   Title: "{ProjectName} {Month}/{Year} - #{InvoiceNumber}"
  │   Description: <PDF Attachment>
  │
  └─ Enqueue Basecamp Comment Worker Job
       ↓
      [Worker] Create Comment
       ↓
      Content: "#Invoice {number} has been sent\n\nConfirm Command: Paid @Giang #{number}"
      Attachment: Invoice PDF
```

#### Related Webhook

**File**: `pkg/handler/webhook/basecamp_invoice.go`

The system has a webhook handler that:
- Listens for todo completion/changes in Basecamp
- Parses invoice number from todo title: `MM/YYYY - #InvoiceNumber`
- Updates invoice status to "Paid" when confirmation detected

---

## Key Findings

### 1. Migration Feasibility Assessment

**Question**: Can we create a NocoDB implementation using the same interfaces as Basecamp?

**Answer**: **YES - The architecture is excellently suited for this**

**Reasoning**:
- Root `basecamp.Service` is a struct, but all 11 sub-services are **interfaces**
- Each sub-service depends only on abstract `client.Service` interface
- Factory pattern with dependency injection throughout
- No hard dependencies on HTTP client implementation
- Consumers (handlers/controllers) depend on interfaces, not concrete types

**What Would Be Required**:
- Implement `client.Service` interface for NocoDB HTTP client
- Create NocoDB variant for each sub-service implementing the same interfaces
- Add service selection logic (feature flag or abstraction layer)
- Update webhook handlers to parse NocoDB webhook formats

### 2. Adapter Pattern Viability

The current architecture enables a clean adapter pattern:

```
pkg/service/basecamp/          pkg/service/nocodb/
├─ client/Service (interface)  ├─ client/Service (same interface)
├─ todo/Service (interface)    ├─ todo/Service (same interface)
├─ comment/Service (interface) ├─ comment/Service (same interface)
└─ ...                          └─ ...

Both implement identical contracts, enabling drop-in replacement
```

### 3. Invoice Workflow Understanding

**Creation**: Manual via API endpoint `POST /api/v1/invoices/send`
- No automatic monthly generation
- User provides all invoice details including month/year
- System auto-generates sequential invoice number

**Basecamp Integration** (when invoice sent):
1. **Todo created** (synchronous) in Accounting project
   - Title: `"{Project} {Month}/{Year} - #{Number}"`
   - Description: Invoice PDF attachment
2. **Comment created** (asynchronous via worker)
   - Content: Confirmation message with invoice number
   - Attachment: Invoice PDF

### 4. Usage Patterns

Basecamp service used in **37 files**:
- **Webhook handlers** (15 files) - Primary integration point
- **Controllers** (5 files) - Invoice, commission, employee updates
- **Workers** (1 file) - Async comment/todo processing
- **Other handlers** - Payroll, Discord, Accounting

### 5. Technical Debt Identified

**Critical**:
- 130+ hardcoded Basecamp resource IDs in `consts.go`

**High**:
- Duplicated pagination logic across 7+ services (~200 lines)
- Synchronous Wise API calls blocking invoice creation
- No rate limiting or circuit breaker for external APIs

**Medium**:
- No caching layer (every request hits Basecamp API)
- Environment-specific logic scattered throughout

### 6. Architectural Strengths

- ✅ Strong interface-based design
- ✅ Consistent service patterns across 28 files
- ✅ Proper dependency injection
- ✅ Comprehensive logging with structured fields
- ✅ Type-safe models with JSON/GORM tags
- ✅ Async processing via worker queues

---

## References

### Key Files

**Basecamp Service**:
- `pkg/service/basecamp/basecamp.go` - Main service
- `pkg/service/basecamp/client/client.go` - OAuth2 client
- `pkg/service/basecamp/integration.go` - Expense integration
- `pkg/service/basecamp/consts/consts.go` - Resource ID constants
- `pkg/service/basecamp/model/model.go` - Data models

**Sub-Services** (Interface + Implementation):
- `pkg/service/basecamp/todo/service.go` + `todo.go`
- `pkg/service/basecamp/comment/service.go` + `comment.go`
- `pkg/service/basecamp/project/service.go` + `project.go`
- `pkg/service/basecamp/people/service.go` + `people.go`
- Similar pattern for 7 other services

**Invoice Workflow**:
- `pkg/handler/invoice/invoice.go` - Invoice handler
- `pkg/controller/invoice/send.go` - Send controller
- `pkg/controller/invoice/update_status.go` - Status updates
- `pkg/handler/webhook/basecamp_invoice.go` - Invoice webhook
- `pkg/handler/webhook/basecamp_expense.go` - Expense webhook

**Worker**:
- `pkg/worker/worker.go` - Background task processing

### File Count Summary

- **Total Go files**: 28 in `pkg/service/basecamp/`
- **Interface files**: 11 `service.go` files defining contracts
- **Implementation files**: 11 domain-specific implementations
- **Model file**: 1 comprehensive model definition (309 lines)
- **Constants file**: 1 with 130+ resource IDs
- **Main service**: 1 factory file aggregating all sub-services

---

**Document Version**: 1.0
**Last Updated**: 2025-11-14
**Type**: Exploration & Analysis (No Implementation)
