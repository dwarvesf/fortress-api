# Payroll Calculation with Basecamp Expenses

## Overview

1. calculatePayrolls fetches approved Basecamp expense todos (lines 54-110)
2. For each employee, getBonus is called with expenses (line 144)
3. getBonus checks if expense is assigned to employee by matching u.BasecampID with expense assignee (lines 247-254)
4. If matched, parses amount from expense title via getReimbursement (line 256)
5. Adds expense amount to bonus and reimbursementAmount (lines 264-265)
6. Appends expense to bonusExplain with Basecamp todo/bucket IDs (lines 266-274)
7. Total payroll = base salary + bonus (includes expenses) + commission - salary advance (line 173)

## Key

- Expenses stored in expenses table are NOT directly counted
- Only approved Basecamp todos in expense lists are fetched and counted
- Matching uses employee.BasecampID vs expense assignee
- Amount parsed from Basecamp todo title, not from expense record