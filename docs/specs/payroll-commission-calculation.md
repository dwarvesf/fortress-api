# Payroll Commission & Bonus Calculation

## Overview

When an invoice is paid, the system calculates commissions for eligible employees based on their roles in the project. These commissions accumulate as unpaid records in the `employee_commissions` table. During payroll calculation, all unpaid commissions are aggregated into the employee's payroll as the **commission amount**. Additionally, recurring **project bonuses** and **expense reimbursements** are included as the **project bonus amount**. Together, these form the "Bonus" line in accounting transactions.

## Source Code

| File | Purpose |
|---|---|
| `pkg/controller/invoice/commission.go` | Commission calculation from invoices |
| `pkg/handler/payroll/payroll_calculator.go` | Payroll aggregation (base salary + commissions + bonuses + expenses) |
| `pkg/handler/payroll/details.go` | Payroll detail preparation and formatting |
| `pkg/handler/payroll/commit.go` | Payroll commit and accounting transaction creation |
| `pkg/model/payroll.go` | Payroll model with CommissionExplain and ProjectBonusExplain |
| `pkg/model/employee_commissions.go` | EmployeeCommission model |
| `pkg/model/employee_bonus.go` | EmployeeBonus model (recurring project bonuses) |
| `pkg/store/employeebonus/employee_bonus.go` | EmployeeBonus store |
| `pkg/store/employeecommission/employee_commission.go` | EmployeeCommission store |

## Data Model

### `employee_commissions` table

Created when an invoice is marked as paid. Each record represents one commission entry for one employee from one invoice.

| Field | Description |
|---|---|
| `employee_id` | The employee receiving the commission |
| `invoice_id` | The source invoice |
| `project` | Project name |
| `amount` | Commission amount in VND |
| `conversion_rate` | Currency conversion rate used |
| `formula` | Human-readable calculation formula |
| `note` | Commission type label (e.g., "Lead - Name", "Hiring - Name") |
| `is_paid` | Whether this commission has been included in a committed payroll |

### `employee_bonuses` table

Recurring fixed bonuses assigned to employees. Active bonuses are included in every payroll cycle.

| Field | Description |
|---|---|
| `employee_id` | The employee receiving the bonus |
| `amount` | Fixed bonus amount in VND |
| `is_active` | Whether this bonus is currently active |
| `name` | Label (e.g., "2 Project Bonus") |

### `payrolls` table

| Field | Description |
|---|---|
| `commission_amount` | Sum of all unpaid `employee_commissions` for this employee |
| `commission_explain` | JSON array of `CommissionExplain` structs detailing each commission |
| `project_bonus_amount` | Sum of `employee_bonuses` + expense reimbursements |
| `project_bonus_explain` | JSON array of `ProjectBonusExplain` structs detailing each bonus/expense |

---

## Commission Types

Commissions are calculated in `pkg/controller/invoice/commission.go` when an invoice is paid. The system identifies eligible people from two sources:

1. **Project Heads** (`invoice.Project.Heads`) — leadership roles assigned to the project
2. **Project Members** (`ProjectMember` with `DeploymentType = Official`) — active team members

### Constants

```go
hiringCommissionRate       = 2    // 2%
saleReferralCommissionRate = 10   // 10%
inboundFundCommissionRate  = 0.04 // 4%
```

---

### 1. Lead Commission (Technical Lead)

**Source:** Project Heads with position `technical-lead`

**Who receives it:** The person assigned as Technical Lead on the project.

**Formula:**

```
CommissionRate% × InvoiceTotal(excluding bonus) × ConversionRate(to VND)
```

**Example:** Invoice `2025103-KAFI-008` total = 230,580,000 VND, Lead commission rate = 2%

```
2% × 230,580,000 × 1 (already VND) = 4,611,600 VND
```

**Note format:** `"Lead - {FullName}"`

**Code:** `calculateHeadCommission()` at `commission.go:353`

---

### 2. Account Manager Commission

**Source:** Project Heads with position `account-manager`

**Who receives it:** The person assigned as Account Manager on the project.

**Formula:** Same as Lead Commission.

```
CommissionRate% × InvoiceTotal(excluding bonus) × ConversionRate(to VND)
```

**Note format:** `"Account Manager"`

**Code:** `calculateHeadCommission()` at `commission.go:353`

---

### 3. Delivery Manager Commission

**Source:** Project Heads with position `delivery-manager`

**Who receives it:** The person assigned as Delivery Manager on the project.

**Formula:** Same as Lead Commission.

```
CommissionRate% × InvoiceTotal(excluding bonus) × ConversionRate(to VND)
```

**Note format:** `"Delivery Manager"`

**Code:** `calculateHeadCommission()` at `commission.go:353`

---

### 4. Sales Commission

**Source:** Project Heads with position `sale-person`

**Who receives it:** The person assigned as Sale Person on the project.

**Formula:** Same as Lead Commission.

```
CommissionRate% × InvoiceTotal(excluding bonus) × ConversionRate(to VND)
```

**Note format:** `"Sales - {FullName}"`

**Code:** `calculateHeadCommission()` at `commission.go:353`

**Special behavior:** If no sale person is assigned to the project, an **Inbound Fund Commission** (4%) is created instead (no specific employee, goes to the inbound fund).

---

### 5. Deal Closing Commission

**Source:** Project Heads with position `deal-closing`

**Who receives it:** The person assigned as Deal Closing on the project.

**Formula:** Same as Lead Commission.

```
CommissionRate% × InvoiceTotal(excluding bonus) × ConversionRate(to VND)
```

**Note format:** `"Deal Closing - {FullName}"`

**Code:** `calculateHeadCommission()` at `commission.go:353`

---

### 6. Hiring Commission (Supplier Referral)

**Source:** Project Members with `DeploymentType = Official` whose `Employee.Referrer` exists and has `WorkingStatus != Left`

**Who receives it:** The **referrer** of an active project member. This rewards the person who recruited a team member into the company. The commission goes to the referrer, not the member.

**Formula:**

```
2%(fixed) × MemberBillingRate × ConversionRate(to VND)
```

- `MemberBillingRate` = the member's `Rate` field from `project_members` (what the client pays for this member)
- Rate is **2%** (constant `hiringCommissionRate`)

**Example:** Invoice `2025103-KAFI-008`, member Le Anh Minh's billing rate = 99,000,000 VND

```
2% × 99,000,000 × 1 (already VND) = 1,980,000 VND
```

The referrer (Nguyễn Hoàng Huy) receives this for every invoice on the project where Le Anh Minh is deployed.

**Note format:** `"Hiring - {MemberFullName}"`

**Code:** `getPICs()` at `commission.go:310-318`, calculated via `calculateRefBonusCommission()` at `commission.go:412`

**Key detail:** This commission is generated per-invoice, per-member. If the referrer has multiple referred members on different projects, they accumulate separate hiring commissions from each invoice.

---

### 7. Upsell Commission

**Source:** Project Members with `DeploymentType = Official` and non-zero `UpsellCommissionRate`

**Who receives it:** The `UpsellPersonID` assigned to the project member.

**Formula:**

```
UpsellCommissionRate% × MemberBillingRate × ConversionRate(to VND)
```

**Note format:** `"Upsell"`

**Code:** `calculateRefBonusCommission()` at `commission.go:412`

---

### 8. Sale Referral Commission

**Source:** Project Heads with position `sale-person`, where the sale person has a `Referrer`

**Who receives it:** The **referrer of the sale person**. This creates a two-level referral chain: the sale person earns their sales commission, and whoever referred the sale person earns 10% of that commission.

**Formula:**

```
10%(fixed) × SalePersonCommission × ConversionRate(to VND)
```

Where `SalePersonCommission` = `SalePersonCommissionRate% × InvoiceTotal`

**Example:** Invoice `2025111-YOLO-012` total = $12,800.16, Sale person Le Anh Minh's commission rate on this project = 5%

```
Step 1: SalePersonCommission = 5% × $12,800.16 = $640.008
Step 2: SaleReferralCommission = 10% × $640.008 = $64.0008
Step 3: Convert to VND = $64.0008 × 26,375 = ~1,688,021 VND
```

The referrer (Nguyễn Hoàng Huy) earns this because they referred Le Anh Minh, who is the sale person on the Yolo Lab project.

**Note format:** `"Sale Referral - {SalePersonFullName}"`

**Code:** `getPICs()` at `commission.go:273-278`, calculated via `calculateSaleReferralCommission()` at `commission.go:437`

---

### 9. Upsell Sale Referral Commission

**Source:** Project Members with non-zero `UpsellCommissionRate`, where the upsell person has a `Referrer`

**Who receives it:** The **referrer of the upsell person**.

**Formula:**

```
10%(fixed) × UpsellPersonCommission × ConversionRate(to VND)
```

Where `UpsellPersonCommission` = `UpsellCommissionRate% × MemberBillingRate`

**Note format:** `"Sale Referral - {UpsellPersonName} Upsell {MemberName}"`

**Code:** `calculateUpsellSaleReferralCommission()` at `commission.go:462`

---

## Payroll Aggregation

During payroll calculation (`payroll_calculator.go:19`), the system aggregates all components for each employee:

### Step 1: Base Salary

```go
baseSalary, contract, _ = tryPartialCalculation(batchDate, dueDate, joinedDate, leftDate, personalAccountAmount, companyAccountAmount)
```

- `baseSalary` = personal account (TransferWise) portion
- `contract` = company account (BHXH / social insurance) portion
- Prorated if employee joined mid-cycle or left before cycle end

### Step 2: Recurring Bonuses (Project Bonus)

```go
bonusRecords, _ := h.store.Bonus.GetByUserID(h.repo.DB(), u.ID)
```

Fetches all active records from `employee_bonuses` table. Each record's `Amount` is summed into the project bonus.

### Step 3: Expense Reimbursements

```go
// Fetched from NocoDB or Basecamp
name, amount, _ := h.getReimbursement(expense.Title)
```

Approved expense submissions assigned to the employee. Parsed from expense title format: `"reason | amount | currency"`. Added to project bonus amount.

### Step 4: Invoice Commissions

```go
userCommissions, _ := h.store.EmployeeCommission.Get(h.repo.DB(), commissionStore.Query{
    EmployeeID: u.ID.String(),
    IsPaid:     false,
})
```

Fetches all **unpaid** `employee_commissions` for the employee. Each record's `Amount` is summed into the commission amount.

### Step 5: Total Calculation

For VND-based employees:

```go
total = baseSalary + contract + bonus + commission - salaryAdvance
```

For non-VND employees, amounts are converted using the Wise API before summing.

---

## Payroll Commit & Accounting Transactions

When payroll is committed (`commit.go:301`), three types of accounting transactions are created:

### Transaction 1: Payroll TW (Base Salary)

```
Name: "Payroll TW - {FullName} {BatchDate}"
Amount: baseSalaryAmount
ConversionAmount: total - contract - commission - projectBonus
```

### Transaction 2: Payroll BHXH (Social Insurance)

Only created if `ContractAmount != 0`.

```
Name: "Payroll BHXH - {FullName} {BatchDate}"
Amount: contractAmount
ConversionAmount: contractAmount
```

### Transaction 3: Bonus

Only created if `bonusConversionAmount != 0`.

```
bonusConversionAmount = commissionAmount + projectBonusAmount - reimbursementAmount
```

Where `reimbursementAmount = total - conversionAmount` (expenses excluded from wire transfer).

```
Name: "Bonus - {FullName} {BatchDate}"
Amount: bonusConversionAmount
Category: determined by commission notes (Lead/Sales/Account/Hiring)
```

**Key insight:** The "Bonus" accounting transaction **excludes expense reimbursements**. Reimbursements are tracked separately via the `conversionAmount` field on the payroll (which is `total - reimbursementAmount`). Reimbursements get their own accounting transactions elsewhere.

---

## Commission Lifecycle

```
Invoice Paid
    │
    ▼
calculateCommissionFromInvoice()
    │  Identifies PICs (project heads + members)
    │  Calculates commission per role
    │  Converts to VND via Wise API
    ▼
employee_commissions table (is_paid = false)
    │
    ▼
calculatePayrolls() → getBonus()
    │  Fetches unpaid commissions
    │  Fetches active employee_bonuses
    │  Fetches approved expense reimbursements
    │  Aggregates into payroll record
    ▼
payrolls table (commission_explain JSON, project_bonus_explain JSON)
    │
    ▼
commitPayrolls() → storePayrollTransaction()
    │  Creates accounting transactions (Payroll TW, BHXH, Bonus)
    │  Marks commissions as paid
    │  Marks expense tasks as done
    ▼
accounting_transactions table
```

---

## Referral Chain Summary

The system supports a two-level referral chain:

```
Employee A (referrer)
    └── referred Employee B (team member or sale person)
            └── deployed on Project X

When Project X invoice is paid:
  - If B is a team member: A gets 2% of B's billing rate (Hiring)
  - If B is a sale person: A gets 10% of B's sales commission (Sale Referral)
  - If B is an upsell person: A gets 10% of B's upsell commission (Upsell Sale Referral)
```
