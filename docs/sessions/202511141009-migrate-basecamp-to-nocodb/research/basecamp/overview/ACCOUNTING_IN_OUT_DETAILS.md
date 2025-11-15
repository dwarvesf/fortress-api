# Accounting In/Out Todo Management - Detailed Analysis

**File**: `pkg/handler/accounting/accounting.go:42-249`
**Endpoint**: `POST /api/v1/cronjobs/sync-monthly-accounting-todo`
**Type**: Monthly Cronjob
**Permission**: `PermissionCronjobExecute`

---

## Overview

This cronjob runs monthly to create a structured Basecamp todo list for the upcoming month's accounting tasks. It automatically generates todos for:
- **Income tracking** ("In" group) - Project invoices
- **Expense tracking** ("Out" group) - Service payments and salaries

---

## Flow Diagram

```
Cronjob Triggered
  ↓
Get Next Month (e.g., January 2025)
  ↓
Select Basecamp Project (Accounting/Playground)
  ↓
Create Monthly TodoList: "Accounting | January 2025"
  ↓
Create Two Groups:
  ├─ "In" Group (Revenue)
  └─ "Out" Group (Expenses)
  ↓
Populate "Out" Group:
  ├─ Fetch service templates from operational_service table
  ├─ Create todo for each service
  ├─ Special: Office Rental creates 3 todos (Rental, CBRE, Tiền điện)
  └─ Add 2 salary todos (15th, 1st)
  ↓
Populate "In" Group:
  ├─ Fetch all active Time & Material projects
  └─ Create invoice todo for each project
  ↓
Return 200 OK
```

---

## Code Breakdown

### 1. Environment Configuration (lines 48-56)

```go
month, year := timeutil.GetMonthAndYearOfNextMonth()  // Next month

accountingTodo := consts.PlaygroundID      // Default: dev
todoSetID := consts.PlaygroundTodoID

if h.config.Env == "prod" {
    accountingTodo = consts.AccountingID    // 15258324
    todoSetID = consts.AccountingTodoID     // 2329633561
}
```

**Constants Used:**
- **Production**:
  - Project: `AccountingID` = 15258324
  - Todo Set: `AccountingTodoID` = 2329633561
- **Development**:
  - Project: `PlaygroundID` = 12984857
  - Todo Set: `PlaygroundTodoID` = 1941398075

### 2. Monthly TodoList Creation (lines 58-76)

```go
todoList := bcModel.TodoList{
    Name: fmt.Sprintf("Accounting | %s %v", time.Month(month).String(), year)
}
// Example: "Accounting | January 2025"

createTodo, err := h.service.Basecamp.Todo.CreateList(
    accountingTodo,  // Project ID
    todoSetID,       // Todo Set ID
    todoList
)
```

### 3. Group Creation (lines 78-92)

```go
// "In" Group for Revenue/Income
todoGroupInFoundation := bcModel.TodoGroup{Name: "In"}
inGroup, err := h.service.Basecamp.Todo.CreateGroup(
    accountingTodo,
    createTodo.ID,
    todoGroupInFoundation
)

// "Out" Group for Expenses/Payments
todoGroupOut := bcModel.TodoGroup{Name: "Out"}
outGroup, err := h.service.Basecamp.Todo.CreateGroup(
    accountingTodo,
    createTodo.ID,
    todoGroupOut
)
```

---

## "Out" Group Population

### Source Data (lines 64-69)

```go
// Fetch service templates from database
outTodoTemplates, err := h.store.OperationalService.FindOperationByMonth(
    h.repo.DB(),
    time.Month(month)
)
```

**Database Table**: `operational_services`
**Schema**:
```sql
id, name, amount, currency_id, month, created_at, updated_at
```

### Service Todo Creation (lines 120-167)

**Function**: `createTodoInOutGroup()`

**Logic**:
1. Iterate through each service template
2. Format todo content: `"{Name} | {Amount} | {Currency}"`
3. Set due date to last day of month
4. Assign to Quang (22659105)

**Special Case - Office Rental** (lines 128-147):
Creates **3 separate todos**:
```go
if strings.Contains(v.Name, "Office Rental") {
    // 1. Office Rental
    // 2. CBRE
    // 3. Tiền điện
}
```

**Example Todos**:
```
Office Rental | 1.500.000 | VND          (due: 31/1/2025, assigned: Quang)
CBRE | 800.000 | VND                     (due: 31/1/2025, assigned: Quang)
Tiền điện | 300.000 | VND                (due: 31/1/2025, assigned: Quang)
IT Services | 2.000.000 | VND            (due: 31/1/2025, assigned: Quang)
```

### Salary Todo Creation (lines 169-193)

**Function**: `createSalaryTodo()`

**Two Todos Created**:

1. **Salary 15th** (line 171-179):
```go
salary15 := bcModel.Todo{
    Content:     "salary 15th",
    DueOn:       fmt.Sprintf("%v-%v-%v", 12, year, month),  // 12th of month
    AssigneeIDs: []int{
        consts.QuangBasecampID,  // 22659105
        consts.HanBasecampID      // 21562923
    },
}
```

2. **Salary 1st** (line 182-191):
```go
salary1 := bcModel.Todo{
    Content:     "salary 1st",
    DueOn:       fmt.Sprintf("%v-%v-%v", 27, year, month),  // 27th of month
    AssigneeIDs: []int{
        consts.QuangBasecampID,  // 22659105
        consts.HanBasecampID      // 21562923
    },
}
```

---

## "In" Group Population

### Source Data (lines 200-208)

```go
// Only Time & Material projects (NOT Fixed-Cost)
activeProjects, _, err := h.store.Project.All(h.repo.DB(), project.GetListProjectInput{
    Statuses: []string{model.ProjectStatusActive.String()},
    Types:    []string{model.ProjectTypeTimeMaterial.String()},
}, model.Pagination{})
```

**Filter Criteria**:
- Status: `active`
- Type: `time-material` (excludes `fixed-cost`)

### Invoice Todo Creation (lines 215-224)

**Function**: `createTodoInInGroup()`

**Logic**:
```go
for _, p := range activeProjects {
    assigneeIDs := []int{consts.GiangThanBasecampID}  // 26160802

    todo := buildInvoiceTodo(p.Name, month, year, assigneeIDs)
    h.service.Basecamp.Todo.Create(projectID, inGroupID, todo)
}
```

### Todo Content Builder (lines 227-248)

**Function**: `buildInvoiceTodo()`

**Content Format** (line 247):
```go
content := fmt.Sprintf("%s %v/%v", name, month, year)
// Example: "Voconic 1/2025"
```

**Due Date Logic** (lines 236-245):
```go
func getProjectInvoiceDueOn(name string, month, year int) string {
    var day int

    if strings.ToLower(name) == "voconic" {
        day = 23  // Special case: Voconic due on 23rd
    } else {
        day = timeutil.LastDayOfMonth(month, year).Day()  // Last day
    }

    return fmt.Sprintf("%v-%v-%v", day, month, year)
}
```

**Example Todos**:
```
Voconic 1/2025                    (due: 23/1/2025, assigned: Giang Than)
Dwarves Foundation 1/2025         (due: 31/1/2025, assigned: Giang Than)
ConsultingWare 1/2025             (due: 31/1/2025, assigned: Giang Than)
```

---

## Complete Example Output

**For January 2025 in Production:**

```
Basecamp Project: Accounting (15258324)
├─ TodoList: "Accounting | January 2025"
   ├─ Group: "In"
   │  ├─ Voconic 1/2025
   │  │  ├─ Due: 23/1/2025
   │  │  └─ Assigned: Giang Than (26160802)
   │  ├─ Dwarves Foundation 1/2025
   │  │  ├─ Due: 31/1/2025
   │  │  └─ Assigned: Giang Than
   │  └─ ConsultingWare 1/2025
   │     ├─ Due: 31/1/2025
   │     └─ Assigned: Giang Than
   │
   └─ Group: "Out"
      ├─ Office Rental | 1.500.000 | VND
      │  ├─ Due: 31/1/2025
      │  ├─ Assigned: Quang (22659105)
      │  └─ Description: "Hado Office Rental 1/2025"
      ├─ CBRE | 800.000 | VND
      │  ├─ Due: 31/1/2025
      │  ├─ Assigned: Quang
      │  └─ Description: "I3.18.08 thanh toan CBRE 1/2025"
      ├─ Tiền điện | 300.000 | VND
      │  ├─ Due: 31/1/2025
      │  ├─ Assigned: Quang
      │  └─ Description: "I3.18.08 thanh toan Tiền điện 1/2025"
      ├─ IT Services | 2.000.000 | VND
      │  ├─ Due: 31/1/2025
      │  └─ Assigned: Quang
      ├─ salary 15th
      │  ├─ Due: 12/1/2025
      │  └─ Assigned: Quang + Han (22659105, 21562923)
      └─ salary 1st
         ├─ Due: 27/1/2025
         └─ Assigned: Quang + Han
```

---

## Hardcoded Constants

### People IDs (from pkg/service/basecamp/consts/consts.go)

```go
QuangBasecampID       = 22659105   // Out todos + salary
HanBasecampID         = 21562923   // Salary todos
GiangThanBasecampID   = 26160802   // In todos
```

### Project IDs

```go
// Production
AccountingID          = 15258324   // Accounting project
AccountingTodoID      = 2329633561 // Accounting todo set

// Development
PlaygroundID          = 12984857   // Playground project
PlaygroundTodoID      = 1941398075 // Playground todo set
```

### Group Names (hardcoded strings)

```go
"In"   // Revenue/income todos
"Out"  // Expense/payment todos
```

---

## Migration to NocoDB Considerations

### Current Basecamp Approach
- Monthly structure: List → Groups → Todos
- Hardcoded person assignments
- Service templates from database
- Project list from database
- Special case handling (Voconic, Office Rental)

### NocoDB Alternative Approach

**Option 1: Replicate Structure**
```
NocoDB Table: "Accounting Monthly"
├─ Fields:
│  ├─ Month/Year
│  ├─ Type (In/Out)
│  ├─ Project/Service Name
│  ├─ Amount
│  ├─ Currency
│  ├─ Due Date
│  └─ Assignee (link to Users)
├─ Views:
│  ├─ "In - January 2025"
│  └─ "Out - January 2025"
```

**Option 2: Simplified Form Entry**
- Replace cronjob with manual form submission
- Pre-fill from database (services, projects)
- Direct NocoDB record creation
- Workflow automation via n8n or NocoDB automation

### Data Sources to Preserve
1. **operational_services table** - Service templates
2. **projects table** - Active T&M projects
3. **Assignment logic**:
   - Out → Finance team
   - In → Accounting team
   - Salary → Multiple assignees

### Benefits of Migration
- ✅ No hardcoded Basecamp IDs
- ✅ Flexible assignment rules
- ✅ Better reporting/analytics
- ✅ Form-based entry option
- ✅ Configurable due date logic

---

**Last Updated**: 2025-11-14
