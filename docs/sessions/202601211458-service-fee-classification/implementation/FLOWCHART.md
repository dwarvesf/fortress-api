# Classification Flow Diagram

## Current (Incorrect) Behavior

```
┌─────────────────────────────────────────────────────────────────┐
│                    Payout Entry from Notion                      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │ Has TaskOrderID?     │
              └──────┬───────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
       YES                       NO
         │                       │
         ▼                       ▼
    ┌─────────┐      ┌──────────────────────┐
    │ Service │      │ Has InvoiceSplitID?  │
    │   Fee   │      └──────┬───────────────┘
    └─────────┘             │
         │          ┌───────┴───────┐
         │          │               │
         │        YES              NO
         │          │               │
         │          ▼               ▼
         │     ┌─────────┐   ┌──────────────────┐
         │     │Commission│   │Has RefundRequest?│
         │     └─────────┘   └──────┬───────────┘
         │          │               │
         │          │       ┌───────┴────────┐
         │          │       │                │
         │          │     YES               NO
         │          │       │                │
         │          │       ▼                ▼
         │          │  ┌────────┐    ┌────────────┐
         │          │  │ Refund │    │ExtraPayment│
         │          │  └────────┘    └────────────┘
         │          │       │                │
         ▼          ▼       ▼                ▼
    ┌──────────────────────────────────────────┐
    │         Invoice Section Grouping          │
    └──────────────┬───────────────────────────┘
                   │
        ┌──────────┼──────────┬─────────────┐
        │          │          │             │
        ▼          ▼          ▼             ▼
┌──────────┐ ┌─────────┐ ┌────────┐ ┌────────────┐
│Development│ │   Fee   │ │ Extra  │ │   Refund   │
│   Work    │ │(wrong!) │ │Payment │ │            │
└──────────┘ └─────────┘ └────────┘ └────────────┘
│           │           │           │
│ServiceFee │Commission │ExtraPaym. │  Refund
│ from      │ from      │           │
│TaskOrder  │InvoiceSplit            │
│           │  ⚠️ ISSUE: All        │
│           │  InvoiceSplit         │
│           │  → Commission         │
│           │  (ignores role type)  │
└───────────┴───────────┴───────────┴────────────┘

PROBLEM: All InvoiceSplit items become Commission,
even those with "Delivery Lead" or "Account Management" roles.
```

---

## Fixed (Correct) Behavior

```
┌─────────────────────────────────────────────────────────────────┐
│                    Payout Entry from Notion                      │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │ Has TaskOrderID?     │
              └──────┬───────────────┘
                     │
         ┌───────────┴───────────┐
         │                       │
       YES                       NO
         │                       │
         ▼                       ▼
    ┌─────────┐      ┌──────────────────────┐
    │ Service │      │ Has InvoiceSplitID?  │
    │   Fee   │      └──────┬───────────────┘
    └─────────┘             │
         │          ┌───────┴───────┐
         │          │               │
         │        YES              NO
         │          │               │
         │          ▼               ▼
         │  ┌───────────────────┐  │
         │  │Check Description  │  │
         │  │   for keywords    │  │
         │  └────────┬──────────┘  │
         │           │              │
         │  ┌────────┴────────┐    │
         │  │                 │    │
         │ "delivery lead"   Other │
         │      OR           words │
         │ "account mgmt"          │
         │  (case-insensitive)     │
         │  │                 │    │
         │  ▼                 ▼    ▼
         │ ┌─────────┐  ┌─────────────┐   ┌──────────────────┐
         │ │ Service │  │ Commission  │   │Has RefundRequest?│
         │ │   Fee   │  └─────────────┘   └──────┬───────────┘
         │ └─────────┘        │                   │
         │      │              │           ┌───────┴────────┐
         │      │              │           │                │
         │      │              │         YES               NO
         │      │              │           │                │
         │      │              │           ▼                ▼
         │      │              │      ┌────────┐    ┌────────────┐
         │      │              │      │ Refund │    │ExtraPayment│
         │      │              │      └────────┘    └────────────┘
         │      │              │           │                │
         ▼      ▼              ▼           ▼                ▼
    ┌────────────────────────────────────────────────────────┐
    │            Invoice Section Grouping                     │
    └──────────────┬─────────────────────────────────────────┘
                   │
        ┌──────────┼──────────┬─────────────┐
        │          │          │             │
        ▼          ▼          ▼             ▼
┌──────────┐ ┌─────────┐ ┌────────┐ ┌────────────┐
│Development│ │   Fee   │ │ Extra  │ │   Refund   │
│   Work    │ │         │ │Payment │ │            │
└──────────┘ └─────────┘ └────────┘ └────────────┘
│           │           │           │
│ServiceFee │ServiceFee │Commission │  Refund
│ from      │ from      │    +      │
│TaskOrder  │InvoiceSplit ExtraPaym.│
│           │(with      │           │
│           │keywords)  │           │
│           │           │           │
│ ✅ FIXED: Keywords in Description determine type
│           InvoiceSplit + keywords → ServiceFee → Fee section
│           InvoiceSplit - keywords → Commission → Extra Payment
└───────────┴───────────┴───────────┴────────────────────────────┘

SOLUTION: Check Description content for InvoiceSplit items
to determine correct classification (ServiceFee vs Commission).
```

---

## Classification Decision Tree

```
┌─────────────────────────────────────┐
│        Payout Entry                 │
│  (from Notion database)             │
└─────────────┬───────────────────────┘
              │
              ▼
      ╔═══════════════════╗
      ║ Priority Check 1: ║
      ║  TaskOrderID?     ║
      ╚═══════╦═══════════╝
              ║
        ┌─────╨─────┐
        │           │
      YES          NO
        │           │
        │           ▼
        │   ╔═══════════════════╗
        │   ║ Priority Check 2: ║
        │   ║  InvoiceSplitID?  ║
        │   ╚═══════╦═══════════╝
        │           ║
        │     ┌─────╨─────┐
        │     │           │
        │   YES          NO
        │     │           │
        │     │           ▼
        │     │   ╔═══════════════════╗
        │     │   ║ Priority Check 3: ║
        │     │   ║ RefundRequestID?  ║
        │     │   ╚═══════╦═══════════╝
        │     │           ║
        │     │     ┌─────╨─────┐
        │     │     │           │
        │     │   YES          NO
        │     │     │           │
        │     │     │           │
        │     ▼     │           │
        │  ╔══════════════════════╗   │
        │  ║ NEW LOGIC:           ║   │
        │  ║ Check Description    ║   │
        │  ║ (case-insensitive)   ║   │
        │  ╚═══════╦══════════════╝   │
        │          ║                  │
        │    ┌─────╨─────┐            │
        │    │           │            │
        │ Contains    Does NOT        │
        │ keywords   contain           │
        │  │         keywords         │
        │  │           │              │
        ▼  ▼           ▼              ▼
     ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
     │ Service  │  │Commission│  │  Refund  │  │  Extra   │
     │   Fee    │  │          │  │          │  │ Payment  │
     └──────────┘  └──────────┘  └──────────┘  └──────────┘

Keywords:
• "delivery lead" (case-insensitive)
• "account management" (case-insensitive)

Match logic: strings.Contains(strings.ToLower(description), keyword)
```

---

## Section Grouping Logic

### Before Fix (Incorrect)

```
All Line Items
      │
      ├─► Type = "ServiceFee" + TaskOrderID ───► Development Work
      │
      ├─► Type = "Commission" ───────────────► Fee ❌ WRONG
      │
      ├─► Type = "ExtraPayment" ─────────────► Extra Payment
      │
      └─► Type = "Refund" ───────────────────► Expense Reimbursement

Problem: All InvoiceSplit items (Commission) go to Fee section,
         even if they should be in Extra Payment.
```

### After Fix (Correct)

```
All Line Items
      │
      ├─► Type = "ServiceFee" + TaskOrderID ──────────────────────────────┐
      │                                                                    │
      ├─► Type = "ServiceFee" + NO TaskOrderID + NO ServiceRateID ────────┤
      │   (= from InvoiceSplit with keywords)                             │
      │                                                                    ▼
      │                                                          ┌──────────────┐
      │                                                          │Development   │
      │                                                          │Work Section  │
      │                                                          └──────────────┘
      │
      ├─► Type = "ServiceFee" + TaskOrderID = "" ──────────────────────────────┐
      │   + ServiceRateID = ""                                                  │
      │   (= from InvoiceSplit with keywords)                                   │
      │                                                                          ▼
      │                                                                 ┌──────────────┐
      │                                                                 │Fee Section   │
      │                                                                 │✅ FIXED      │
      │                                                                 └──────────────┘
      │
      ├─► Type = "Commission" ───────────────────────────────────────────────────┐
      │   (from InvoiceSplit without keywords)                                   │
      │                                                                           │
      ├─► Type = "ExtraPayment" ─────────────────────────────────────────────────┤
      │                                                                           │
      │                                                                           ▼
      │                                                                  ┌──────────────┐
      │                                                                  │Extra Payment │
      │                                                                  │Section       │
      │                                                                  │✅ CORRECT    │
      │                                                                  └──────────────┘
      │
      └─► Type = "Refund" ──────────────────────────────────────────────────────┐
                                                                                  │
                                                                                  ▼
                                                                         ┌──────────────┐
                                                                         │Expense       │
                                                                         │Reimbursement │
                                                                         └──────────────┘
```

---

## Example Data Flow

### Example 1: Delivery Lead (Service Fee from InvoiceSplit)

```
Notion Entry:
┌────────────────────────────────────────────┐
│ PageID: abc-123                            │
│ Description: "[FEE :: MUDAH] Delivery Lead"│
│ 00 Task Order: (empty)                     │
│ 02 Invoice Split: invoice-split-789        │
│ 01 Refund: (empty)                         │
└────────────────────────────────────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ determineSourceType()  │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ InvoiceSplitID != ""  │
        │                       │
        │ Description contains  │
        │   "delivery lead"     │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
            ┌───────────────┐
            │ ServiceFee    │
            └───────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ Section Grouping       │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ Type = "ServiceFee"   │
        │ TaskOrderID = ""      │
        │ ServiceRateID = ""    │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
            ┌───────────────┐
            │ Fee Section   │
            │   ✅ CORRECT  │
            └───────────────┘
```

### Example 2: Performance Bonus (Commission from InvoiceSplit)

```
Notion Entry:
┌────────────────────────────────────────────┐
│ PageID: xyz-456                            │
│ Description: "[BONUS] Q4 Performance"      │
│ 00 Task Order: (empty)                     │
│ 02 Invoice Split: invoice-split-999        │
│ 01 Refund: (empty)                         │
└────────────────────────────────────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ determineSourceType()  │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ InvoiceSplitID != ""  │
        │                       │
        │ Description DOES NOT  │
        │   contain keywords    │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
            ┌───────────────┐
            │ Commission    │
            └───────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ Section Grouping       │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ Type = "Commission"   │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Extra Payment Section │
        │   ✅ CORRECT          │
        └───────────────────────┘
```

### Example 3: Development Work (Service Fee from TaskOrder)

```
Notion Entry:
┌────────────────────────────────────────────┐
│ PageID: def-789                            │
│ Description: "Development work"            │
│ 00 Task Order: task-order-123              │
│ 02 Invoice Split: (empty)                  │
│ 01 Refund: (empty)                         │
└────────────────────────────────────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ determineSourceType()  │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ TaskOrderID != ""     │
        │ (Highest priority)    │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
            ┌───────────────┐
            │ ServiceFee    │
            └───────────────┘
                    │
                    ▼
       ┌────────────────────────┐
       │ Section Grouping       │
       └────────────────────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
        │ Type = "ServiceFee"   │
        │ TaskOrderID != ""     │
        │ ServiceRateID != ""   │
        │                       │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Development Work      │
        │ Section               │
        │   ✅ UNCHANGED        │
        └───────────────────────┘
```

---

## Code Change Impact Map

```
┌──────────────────────────────────────────────────────────────┐
│                  CODE CHANGE LOCATIONS                        │
└──────────────────────────────────────────────────────────────┘

Change 1: Classification Logic
┌─────────────────────────────────────────────────────────────┐
│ File: pkg/service/notion/contractor_payouts.go              │
│ Function: determineSourceType()                             │
│ Lines: 365-380                                              │
│                                                             │
│ IMPACT:                                                     │
│ • Affects all invoice generation                           │
│ • Affects payroll calculation                              │
│ • Changes payout classification at source                  │
│                                                             │
│ RISK: Medium                                               │
│ • Wrong logic = all invoices affected                      │
│ • Good test coverage mitigates risk                        │
└─────────────────────────────────────────────────────────────┘

Change 2: Fee Section Grouping
┌─────────────────────────────────────────────────────────────┐
│ File: pkg/controller/invoice/contractor_invoice.go          │
│ Section: Fee section grouping                               │
│ Lines: 977-994                                              │
│                                                             │
│ IMPACT:                                                     │
│ • Only affects invoice PDF generation                      │
│ • Changes what appears in Fee section                      │
│                                                             │
│ RISK: Low                                                  │
│ • Isolated to PDF display logic                            │
│ • No data persistence changes                              │
└─────────────────────────────────────────────────────────────┘

Change 3: Extra Payment Section Grouping
┌─────────────────────────────────────────────────────────────┐
│ File: pkg/controller/invoice/contractor_invoice.go          │
│ Section: Extra Payment section grouping                     │
│ Lines: 996-1012                                             │
│                                                             │
│ IMPACT:                                                     │
│ • Only affects invoice PDF generation                      │
│ • Changes what appears in Extra Payment section            │
│                                                             │
│ RISK: Low                                                  │
│ • Isolated to PDF display logic                            │
│ • Complements Change 2                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## Testing Flow

```
┌─────────────────────────────────────────────────────────────┐
│                      TESTING STRATEGY                        │
└─────────────────────────────────────────────────────────────┘

Unit Tests (Task 6)
┌──────────────────────────────────────────┐
│ Test: determineSourceType()              │
│ Focus: Classification logic              │
│ Coverage: All branches                   │
└──────────────┬───────────────────────────┘
               │
               ▼
Integration Tests (Task 7)
┌──────────────────────────────────────────┐
│ Test: groupIntoSections()                │
│ Focus: Section grouping                  │
│ Coverage: All payout type combinations   │
└──────────────┬───────────────────────────┘
               │
               ▼
Manual E2E Tests (Task 8)
┌──────────────────────────────────────────┐
│ Test: Full invoice generation            │
│ Focus: Real data, PDF output             │
│ Coverage: Edge cases, visual inspection  │
└──────────────────────────────────────────┘

Pass all tests? ───YES───► Deploy to Staging
       │
       NO
       │
       ▼
   Fix issues
       │
       └──► Re-test
```

---

**Last Updated**: 2026-01-21
