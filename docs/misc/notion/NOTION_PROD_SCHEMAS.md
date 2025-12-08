# Notion Production Database Schemas

## Database Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           RELATIONSHIP GRAPH                                 │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────┐         ┌──────────────────┐         ┌──────────────────┐
│   Contractors    │◄────────│    Deployment    │────────►│     Project      │
│   (Profiles)     │         │     Tracker      │         │                  │
│                  │         │                  │         │                  │
│ 9d468753-ebb4-   │         │ 2b864b29-b84c-   │         │ 0ddadba5-bbf2-   │
│ 4977-a8dc-       │         │ 8079-9568-       │         │ 440c-a286-       │
│ 156428398a6b     │         │ dc17685f4f33     │         │ 9f607eca88db     │
└──────────────────┘         └────────┬─────────┘         └──────────────────┘
        ▲                             │                            ▲
        │                             │                            │
        │ Contractor                  │ Deployment Tracker         │ Project
        │ (via Deployment)            │                            │
        │                             ▼                            │
        │                    ┌──────────────────┐                  │
        └────────────────────│     Invoice      │──────────────────┘
                             │                  │
                             │ 2bf64b29-b84c-   │
                             │ 8087-9a52-       │
                             │ ed2f9d493096     │
                             │                  │
                             │ ┌──────────────┐ │
                             │ │ Parent item  │ │ (self-referential)
                             │ │      ↕       │ │
                             │ │  Line Item   │ │
                             │ └──────────────┘ │
                             └──────────────────┘
```

## Database IDs (Production)

| Database | ID | Title |
|----------|-----|-------|
| Invoice | `2bf64b29-b84c-8087-9a52-ed2f9d493096` | Invoices |
| Project | `0ddadba5-bbf2-440c-a286-9f607eca88db` | Projects |
| Deployment Tracker | `2b864b29-b84c-8079-9568-dc17685f4f33` | Deployment Tracker |
| Contractors | `9d468753-ebb4-4977-a8dc-156428398a6b` | Contractors |

---

## 1. Contractors Database

**ID:** `9d468753-ebb4-4977-a8dc-156428398a6b`
**Title:** Contractors
**Icon:** https://www.notion.so/icons/user-circle_gray.svg

### Properties

| Property | ID | Type | Details |
|----------|-----|------|---------|
| Full Name | `title` | title | Primary identifier |
| Team Email | `.%7Coa` | email | |
| Personal Email | `Q%40%40Z` | email | |
| Discord | `l\`p^` | rich_text | Discord username |
| First name | `P%40WN` | rich_text | |
| Last name | `ktsr` | rich_text | |
| Status | `Xfmt` | select | Active, Left, Probation, Offered, Intern, Apprentices |
| Position | `9Dl2` | multi_select | Managing Director, CEO, CTO, Head of Operations, Head of Business, Head of Engineering, DevOps Principal, Partner, Account Manager, Consultant, Technical Lead, Frontend Lead, Backend Lead, iOS Lead, Project Manager, Frontend, Bubble, Backend, Web, iOS, Android, Full-stack, QC Automation, QC Manual, Blockchain, Data Engineer, Data Analyst, Unity Game Client, Java Engineer, Design Ops, Product Designer, UX Designer, UI Designer, Business Ops, Community Ops, Technical Recruiter, Marketing Executive, Operation Executive, Community Executive, Apprentices, Partnership, Branding, React Native, Product Data, Software Architect, DevOps |
| Stack | `gCqt` | multi_select | Laravel, ReactJS, Wordpress, NodeJS, Blockchain, Android, Manual, Golang, Java, Automation, Designer, Unity, WebGL, Elixir, Art, Discord Bot, Product Design, Angular, Swift, PHP, Data, Devops, Modeling, React Native, PM, Operation Executive, Social Platform, Sales, Community, Selenium, Postman, Python, Typescript |
| Onboard Date | `IC%5E%3F` | date | |
| Offboard Date | `WKb%3C` | date | |
| Birthday | `z-hV` | date | |
| Mobile # | `x%2F%26D` | phone_number | |
| Referred By | `O%7BSI` | select | List of referrer names |
| Local employment | `J\`K]` | select | Vietnam |
| Blood Type | `%3BHDq` | select | A, A+, AB, B, O+, O, unknown |
| Horoscope | `Spl%5E` | select | Zodiac signs |
| MBTI | `%5CmKt` | select | Personality types (full names) |
| MBTI Abbr. | `\`sxw` | select | Personality types (abbreviations) |
| ID No. | `dYz%3E` | rich_text | |
| New CCCD No. | `ur%5Ev` | rich_text | |
| Permanent Address | `wZ%3CP` | rich_text | |
| Shelter Address | `9hd%3F` | rich_text | |
| Note | `\`%3FGu` | rich_text | |
| CV | `Z%7DBf` | files | |
| ID Image | `rM%5CG` | files | |
| New CCCD Image | `~g%5EA` | files | |
| Headshot | `%7COYL` | files | |
| Linkedin | `b%3B7o` | url | |
| Twitter | `RM%3Cm` | url | |
| Facebook | `vEay` | url | |
| GitHub | `776faea1-7940-4cac-95ff-e4a51269fcd4` | url | |

### Relations

**None** - This is a base table referenced by other databases.

---

## 2. Project Database

**ID:** `0ddadba5-bbf2-440c-a286-9f607eca88db`
**Title:** Projects
**Icon:** https://super.so/icon/dark/briefcase.svg

### Properties

| Property | ID | Type | Details |
|----------|-----|------|---------|
| Project | `title` | title | Project name |
| ID | `W_nw` | unique_id | Auto-increment |
| Status | `u\`sn` | status | Groups: To-do (New), In progress (Paused, Active), Complete (Done, Closed, Failed) |
| Size | `A%3DQy` | select | Big, Medium +, Medium -, Small |
| Tags | `RsXJ` | multi_select | Partnership, Ventures, Enterprise, MVP, Pre-seed, Seed, Series A, Series B, CSR |
| Tech Stack | `aRYk` | multi_select | AI, Blockchain, Elixir, Golang, NodeJS, React, JavaScript, Ruby, React Native, Vue, TypeScript, Java, Web |
| Sales | `EpBL` | select | Inbound, Han, Nikki, Minh Le, Matt, Huy Tieu, Duc Nghiem, Khai Le, Son Le, N/A |
| Deal Closing (Account Manager) | `hKLo` | multi_select | Han, Minh Le, Nikki, Tom Nguyen, Thanh Pham, An Tran, Huy Tieu |
| PM/Delivery (Technical Lead) | `A%7Dea` | multi_select | Huy Nguyen, Thanh Pham, Tom Nguyen, Lap Nguyen, An Tran, Minh Luu, N/A, Huy Tieu, GiangT, Hung Vong, Tay Nguyen, Khai Le, Minh Le, Quang Le |
| Recipient Emails | `DHz%5E` | rich_text | For invoice delivery |
| Manual Codename | `XDFr` | rich_text | Override auto-generated codename |
| Change log | `zREA` | rich_text | |
| Closed Date | `d%7C%7BP` | date | |
| Codename | `_VO%3F` | formula | Auto-generates project codename |
| Redacted Code | `JUuv` | formula | Anonymized codename for external use |

### Formulas

**Codename** (`_VO?`): Generates a short code from project/client name
- Uses manual codename if provided
- Otherwise auto-generates from project name
- Handles special domains (.studio, .ai, .dev, .trade)
- Removes vowels for 8+ letter names

**Redacted Code** (`JUuv`): Generates anonymized codename
- Based on project name, client, tags, size
- Uses themed word lists (animals for Blockchain, gems for AI, plants for others)
- Format: `CODENAME-###`

### Relations

**None** - This is a base table referenced by other databases.

---

## 3. Deployment Tracker Database

**ID:** `2b864b29-b84c-8079-9568-dc17685f4f33`
**Title:** Deployment Tracker

### Properties

| Property | ID | Type | Details |
|----------|-----|------|---------|
| Name | `title` | title | Auto-populated via formula |
| ID | `Gucn` | unique_id | Auto-increment |
| Deployment Status | `phgs` | status | Groups: To-do (Not started), In progress (Active), Complete (Done) |
| Type | `fJwa` | multi_select | Official, Part-time, Shadow, Not started, Done |
| Position | `uOqq` | select | AI Engineer, Quality Engineer, Frontend Engineer, Backend Engineer, Fullstack Engineer, Technical Lead, Web3 Developer |
| Start Date | `%3BFZy` | date | |
| End Date | `~%5DyI` | date | |
| Charges | `G%5DDO` | rich_text | |
| Auto Name | `i%3En%5E` | formula | `{Project} :: {Discord}` |

### Relations

| Property | ID | Type | Target Database | Target ID |
|----------|-----|------|-----------------|-----------|
| **Contractor** | `P~D%3E` | relation | Contractors | `9d468753-ebb4-4977-a8dc-156428398a6b` |
| **Project** | `V_%3AT` | relation | Projects | `0ddadba5-bbf2-440c-a286-9f607eca88db` |
| **Upsell Person** | `a%3BqZ` | relation | Contractors | `9d468753-ebb4-4977-a8dc-156428398a6b` |

### Rollups

| Property | ID | Relation | Source Property | Function |
|----------|-----|----------|-----------------|----------|
| Discord | `lbu%3C` | Contractor | Discord (`l\`p^`) | show_original |
| Hiring Referral | `%5CMKl` | Contractor | Referred By (`O{SI`) | show_original |
| Original Sales | `UNHk` | Project | Sales (`EpBL`) | show_original |
| Account Managers | `opdu` | Project | Deal Closing (`hKLo`) | show_original |
| Delivery Leads | `sJSw` | Project | PM/Delivery (`A}ea`) | show_original |

### Formulas

**Final Sales Credit** (`\CtQK`): Returns Upsell Person if set, otherwise Original Sales
**Auto Name** (`i>n^`): `{Project} :: {Discord}` with blue styling

---

## 4. Invoice Database

**ID:** `2bf64b29-b84c-8087-9a52-ed2f9d493096`
**Title:** Invoices
**Icon:** https://www.notion.so/icons/receipt_gray.svg

### Data Model

Self-referential table containing both **Invoices** and **Line Items**:

```
Invoice (Type="Invoice")
├── Line Item 1 (Type="Line Item", Parent item → Invoice)
├── Line Item 2 (Type="Line Item", Parent item → Invoice)
└── Line Item N (Type="Line Item", Parent item → Invoice)
```

### Properties

#### Identity

| Property | ID | Type | Details |
|----------|-----|------|---------|
| Invoice Number | `title` | title | Format: `INV-YYYYMM-CODE-XXXX` |
| ID | `wX_b` | unique_id | Auto-increment |
| Type | `%3DBN%3A` | select | Invoice, Line Item |
| Auto Name | `IV%7BM` | formula | Display name with formatting |

#### Status & Dates

| Property | ID | Type | Options/Details |
|----------|-----|------|-----------------|
| Status | `nsb%40` | status | Groups: To-do (Draft), In progress (Sent, Overdue), Complete (Paid, Cancelled) |
| Issue Date | `U%3FNt` | date | |
| Due Date | `GsHF` | date | |
| Paid Date | `wXqV` | date | |

#### Financial - Line Item Level

| Property | ID | Type | Format |
|----------|-----|------|--------|
| Quantity | `ks~%3E` | number | number |
| Unit Price | `eQ%5EO` | number | number |
| Tax Rate | `%7Dzed` | number | percent |
| Currency | `vVG%3C` | select | USD, EUR, GBP, JPY, CNY, SGD, CAD, AUD |

#### Discounts

| Property | ID | Type | Options |
|----------|-----|------|---------|
| Discount Type | `%5DZC%3A` | select | None, Percentage, Fixed Amount, Bulk Discount, Seasonal, Loyalty, Early Payment |
| Discount Value | `UE%60Y` | number | number_with_commas |
| Discount Amount | `%3CTIG` | formula | Calculated based on type |
| Discount Display | `cQRk` | formula | "10%" or "$100" |

#### Commission Percentages

| Property | ID | Type |
|----------|-----|------|
| % Sales | `gZX~` | number (percent) |
| % Account Mgr | `iF%5D%5D` | number (percent) |
| % Delivery Lead | `r%40c%7B` | number (percent) |
| % Hiring Referral | `uyk%5B` | number (percent) |

#### Text Fields

| Property | ID | Type |
|----------|-----|------|
| Description | `_%3BtA` | rich_text |
| Role | `xbaN` | rich_text |
| Notes | `vL%3CE` | rich_text |
| Sent by | `%3Ccgl` | rich_text |

#### Payment

| Property | ID | Type | Options |
|----------|-----|------|---------|
| Payment Method | `kDRp` | select | Bank Transfer, Credit Card, Cash, Check, PayPal, Venmo, Crypto, Mobile Payment |

#### Other

| Property | ID | Type |
|----------|-----|------|
| Splits Generated | `%7DWM%7D` | checkbox |
| Generate Splits | `kS%3BX` | button |

### Relations

| Property | ID | Type | Target Database | Target ID | Relation Type |
|----------|-----|------|-----------------|-----------|---------------|
| **Project** | `srY%3D` | relation | Projects | `0ddadba5-bbf2-440c-a286-9f607eca88db` | single_property |
| **Parent item** | `ES%5DD` | relation | Invoices (self) | `2bf64b29-b84c-8087-9a52-ed2f9d493096` | dual_property (synced: Line Item) |
| **Line Item** | `Z_%5Em` | relation | Invoices (self) | `2bf64b29-b84c-8087-9a52-ed2f9d493096` | dual_property (synced: Parent item) |
| **Deployment Tracker** | `OMj%3F` | relation | Deployment Tracker | `2b864b29-b84c-8079-9568-dc17685f4f33` | (referenced in formulas) |
| **Google Drive File** | `notion://...` | relation | Google Drive | `2bf64b29-b84c-80e2-8cc7-000bfe534203` | dual_property |

**Note:** Deployment Tracker relation (`OMj?`) is referenced in formulas but not visible in standard properties. It's accessed via formula expressions.

### Rollups

| Property | ID | Relation | Source Property | Function |
|----------|-----|----------|-----------------|----------|
| Total Amount | `%3Df%3Bt` | Line Item | Line Total | sum |
| Code | `%3CD%3ES` | Project | Codename | show_original |
| Client | `sYY%5D` | Project | Client | show_original |
| Recipients | `OLc%40` | Project | Recipient Emails | show_original |
| Redacted Codename | `psPe` | Project | Redacted Code | show_original |
| All Sales Amounts | `EARk` | Line Item | Sales Amount | sum |
| All AM Amounts | `Mu%5D%5D` | Line Item | Account Amount | sum |
| All DL Amounts | `~fv%5B` | Line Item | Delivery Lead Amount | sum |
| All Hiring Ref Amounts | `GuqF` | Line Item | Hiring Referral Amount | sum |

### Formulas

| Property | ID | Description |
|----------|-----|-------------|
| Line Total | `Cw%7Ck` | `Qty × Price × (1 + Tax) - Discount` (Line Item only) |
| Subtotal | `hDZr` | `Qty × Price` (Line Item only) |
| Final Total | `%3A%3CEa` | `Total Amount - Discount` with currency symbol |
| Sales Amount | `%5CLY%3F` | `Line Total × % Sales / count(sales)` |
| Account Amount | `zZE%3A` | `Line Total × % AM / count(AMs)` |
| Delivery Lead Amount | `%5B%7DgB` | `Line Total × % DL / count(leads)` |
| Hiring Referral Amount | `XM%3AA` | `Line Total × % Hiring Referral` |
| Total Commission Paid | `FvNv` | Sum of all commission amounts |
| Account Manager | `%3EOw%5E` | From Deployment Tracker |
| Delivery Lead | `%7CRjc` | From Deployment Tracker |
| Hiring Referral | `nYwg` | From Deployment Tracker |
| Auto Name | `IV%7BM` | Invoice: `INV-YYYYMM-CODE-XXXX`, Line Item: `{Project} :: {Discord} / {Date}` |
| Discount Amount | `%3CTIG` | Based on discount type (percentage vs fixed) |
| Discount Display | `cQRk` | "10%" or "$100" |

---

## Creation Order for Dev Environment

Due to relations, databases must be created in this order:

1. **Contractors** - No dependencies (base table)
2. **Project** - No dependencies (base table)
3. **Deployment Tracker** - Depends on: Contractors, Project
4. **Invoice** - Depends on: Project, Deployment Tracker (+ self-reference)

---

## API Limitations

The following property types **CANNOT** be created via Notion API:

| Type | Workaround |
|------|------------|
| `status` | Use `select` with similar options |
| `formula` | Must be added manually in Notion UI |
| `rollup` | Must be added manually in Notion UI |
| `unique_id` | Auto-created by Notion, cannot be set |
| `button` | Must be added manually in Notion UI |

### Creatable via API

- title, rich_text, number, select, multi_select, date, checkbox, url, email, phone_number, files, relation

---

## Property ID Reference

Property IDs are URL-encoded. Common patterns:
- `%3D` = `=`
- `%3A` = `:`
- `%3C` = `<`
- `%3E` = `>`
- `%40` = `@`
- `%5B` = `[`
- `%5C` = `\`
- `%5D` = `]`
- `%5E` = `^`
- `%60` = `` ` ``
- `%7B` = `{`
- `%7C` = `|`
- `%7D` = `}`
- `%7E` = `~`
