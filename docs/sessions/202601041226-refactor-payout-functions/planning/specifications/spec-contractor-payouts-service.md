# Specification: contractor_payouts.go Changes

## File
`pkg/service/notion/contractor_payouts.go`

## Changes Required

### 1. PayoutEntry Struct (line 22-34)

**Current:**
```go
type PayoutEntry struct {
    PageID           string
    Name             string
    PersonPageID     string
    SourceType       PayoutSourceType
    Direction        PayoutDirection    // REMOVE
    Amount           float64
    Currency         string
    Status           string
    ContractorFeesID string             // RENAME to TaskOrderID
    InvoiceSplitID   string
    RefundRequestID  string
}
```

**New:**
```go
type PayoutEntry struct {
    PageID          string
    Name            string
    PersonPageID    string
    SourceType      PayoutSourceType
    Amount          float64
    Currency        string
    Status          string
    TaskOrderID     string  // Was ContractorFeesID, maps to "00 Task Order"
    InvoiceSplitID  string  // Maps to "02 Invoice Split"
    RefundRequestID string  // Maps to "01 Refund"
}
```

---

### 2. QueryPendingPayoutsByContractor (line 52-166)

**Remove Direction filter (line 81-88):**
```go
// DELETE THIS BLOCK:
{
    Property: "Direction",
    DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
        Select: &nt.SelectDatabaseQueryFilter{
            Equals: string(PayoutDirectionOutgoing),
        },
    },
},
```

**Update property extraction (around line 133-144):**
```go
// OLD:
Direction:        PayoutDirection(s.extractSelect(props, "Direction")),
ContractorFeesID: s.extractFirstRelationID(props, "Billing"),
InvoiceSplitID:   s.extractFirstRelationID(props, "Invoice Split"),
RefundRequestID:  s.extractFirstRelationID(props, "Refund"),

// NEW:
TaskOrderID:     s.extractFirstRelationID(props, "00 Task Order"),
InvoiceSplitID:  s.extractFirstRelationID(props, "02 Invoice Split"),
RefundRequestID: s.extractFirstRelationID(props, "01 Refund"),
```

---

### 3. determineSourceType (line 168-184)

**Update return value:**
```go
// OLD:
if entry.ContractorFeesID != "" {
    return PayoutSourceTypeContractorPayroll
}

// NEW:
if entry.TaskOrderID != "" {
    return PayoutSourceTypeServiceFee
}
```

---

### 4. CheckPayoutExistsByContractorFee (line 234-269)

**Update property name:**
```go
// OLD:
Property: "Billing",

// NEW:
Property: "00 Task Order",
```

---

### 5. CreatePayoutInput Struct (line 271-281)

**Current:**
```go
type CreatePayoutInput struct {
    Name              string
    ContractorPageID  string
    ContractorFeeID   string
    Amount            float64
    Currency          string
    Month             string  // REMOVE
    Date              string
    Type              string  // REMOVE
}
```

**New:**
```go
type CreatePayoutInput struct {
    Name             string
    ContractorPageID string
    TaskOrderID      string  // Was ContractorFeeID
    Amount           float64
    Currency         string
    Date             string
    Description      string  // NEW optional
}
```

---

### 6. CreatePayout (line 285-386)

**Remove these property writes:**
- `"Month"` (line 306-310) - formula
- `"Type"` (line 318-322) - formula
- `"Direction"` (line 324-329) - removed

**Update relation name:**
```go
// OLD:
"Billing": nt.DatabasePageProperty{...}

// NEW:
"00 Task Order": nt.DatabasePageProperty{...}
```

**Add Description if provided:**
```go
if input.Description != "" {
    props["Description"] = nt.DatabasePageProperty{
        RichText: []nt.RichText{
            {Text: &nt.Text{Content: input.Description}},
        },
    }
}
```

---

### 7. CheckPayoutExistsByRefundRequest (line 399-436)

**Update property name:**
```go
// OLD:
Property: "Refund",

// NEW:
Property: "01 Refund",
```

---

### 8. CreateRefundPayoutInput Struct (line 388-397)

**Remove Month field** (now calculated from Date by formula)

---

### 9. CreateRefundPayout (line 440-541)

**Remove these property writes:**
- `"Month"` - formula
- `"Type"` - formula
- `"Direction"` - removed

**Update relation name:**
```go
// OLD:
"Refund": nt.DatabasePageProperty{...}

// NEW:
"01 Refund": nt.DatabasePageProperty{...}
```

---

### 10. CheckPayoutExistsByInvoiceSplit (line 543-580)

**Update property name:**
```go
// OLD:
Property: "Invoice Split",

// NEW:
Property: "02 Invoice Split",
```

---

### 11. CreateCommissionPayout (line 592-689)

**Remove these property writes:**
- `"Type"` - formula
- `"Direction"` - removed

**Update relation name:**
```go
// OLD:
"Invoice Split": nt.DatabasePageProperty{...}

// NEW:
"02 Invoice Split": nt.DatabasePageProperty{...}
```

---

### 12. CreateBonusPayout (line 701-798)

**Remove these property writes:**
- `"Type"` - formula
- `"Direction"` - removed

**Update relation name:**
```go
// OLD:
"Invoice Split": nt.DatabasePageProperty{...}

// NEW:
"02 Invoice Split": nt.DatabasePageProperty{...}
```
