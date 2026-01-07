# Deployment Tracker

**Database ID:** `2b864b29-b84c-80f8-b16e-000b8e8ad2b4`

**Purpose:** Tracks contractor deployments to projects, including deployment types (Official, Part-time, Shadow), positions, dates, and associated team members (Delivery Leads, Account Managers).

**Created:** 2025-11-27
**Last Updated:** 2026-01-07

---

## Properties

### Core Properties

#### Name (Title)
- **ID:** `title`
- **Type:** Title
- **Description:** Primary identifier for the deployment record

#### Auto Name (Formula)
- **ID:** `i>n^`
- **Type:** Formula
- **Description:** Auto-generated name in format: `DPL :: [YYYYMM] :: [ProjectCode] :: [Discord]`
- **Format Example:** `DPL :: [202601] :: PROJ-CODE :: username#1234`

#### Type (Multi-Select)
- **ID:** `fJwa`
- **Type:** Multi-Select
- **Options:**
  - **Official** (Blue) - Full-time official deployment
  - **Part-time** (Purple) - Part-time engagement
  - **Shadow** (Yellow) - Shadow/training deployment
  - **Not started** (Default) - Deployment not yet started
- **Description:** Deployment type classification

#### Deployment Status (Status)
- **ID:** `phgs`
- **Type:** Status
- **Options:**
  - **Not started** (Gray/Default)
  - **Active** (Blue) - Currently active deployment
  - **Done** (Green) - Completed deployment
- **Status Groups:**
  - To-do: Not started
  - In progress: Active
  - Complete: Done

### Relations

#### Contractor (Relation)
- **ID:** `P~D>`
- **Type:** Single Relation
- **Links to:** Contractor database (`ed2b9224-97d9-4dff-97f9-82598b61f65d`)
- **Description:** The contractor assigned to this deployment

#### Project (Relation)
- **ID:** `V_:T`
- **Type:** Single Relation
- **Links to:** Project database (`2988f9de-9886-4c6f-a3ff-7f7ef74b3732`)
- **Description:** The project this deployment is for

#### Shadow For (Relation)
- **ID:** `G[V]`
- **Type:** Single Relation
- **Links to:** Contractor database (`ed2b9224-97d9-4dff-97f9-82598b61f65d`)
- **Description:** For Shadow deployments, the contractor being shadowed
- **Special Rule:** When Type includes "Shadow", this indicates which contractor is being shadowed

### Dates

#### Start Date (Date)
- **ID:** `;FZy`
- **Type:** Date
- **Description:** Deployment start date

#### End Date (Date)
- **ID:** `~]yI`
- **Type:** Date
- **Description:** Deployment end date (if applicable)

### Position

#### Position (Select)
- **ID:** `uOqq`
- **Type:** Select
- **Options:**
  - AI Engineer (Pink)
  - Quality Engineer (Purple)
  - Frontend Engineer (Brown)
  - Backend Engineer (Blue)
  - Fullstack Engineer (Green)
  - Technical Lead (Orange)
  - Web3 Developer (Default)
  - Product Designer (Gray)
  - Data Engineer (Red)
  - Business Analyst (Yellow)
  - Mobile Developer (Gray)
  - Executive Assistant (Pink)
  - Project Ops (Default)

### Team Members

#### Delivery Leads (Rollup)
- **ID:** `sJSw`
- **Type:** Rollup
- **Rolls up:** `Delivery Leads` from `Project` relation
- **Function:** Show Original

#### Final Delivery Lead (Formula)
- **ID:** `C_]<`
- **Type:** Formula
- **Description:** If Override DL is set, use it; otherwise use Delivery Leads from project

#### Override DL (Relation)
- **ID:** `zq^l`
- **Type:** Single Relation
- **Links to:** Contractor database (`ed2b9224-97d9-4dff-97f9-82598b61f65d`)
- **Description:** Override delivery lead for this specific deployment

#### Account Managers (Rollup)
- **ID:** `opdu`
- **Type:** Rollup
- **Rolls up:** `Account Managers` from `Project` relation
- **Function:** Show Original

#### Final AM (Formula)
- **ID:** `{LN=`
- **Type:** Formula
- **Description:** If Override AM is set, use it; otherwise use Account Managers from project

#### Override AM (Relation)
- **ID:** `yP[t`
- **Type:** Single Relation
- **Links to:** Contractor database (`ed2b9224-97d9-4dff-97f9-82598b61f65d`)
- **Description:** Override account manager for this specific deployment

### Sales & Upsell

#### Original Sales (Rollup)
- **ID:** `UNHk`
- **Type:** Rollup
- **Rolls up:** `Sales` from `Project` relation
- **Function:** Show Original

#### Upsell Person (Relation)
- **ID:** `a;qZ`
- **Type:** Single Relation
- **Links to:** Contractor database (`ed2b9224-97d9-4dff-97f9-82598b61f65d`)
- **Description:** Person who upsold this deployment

#### Final Sales Credit (Formula)
- **ID:** `\tQK`
- **Type:** Formula
- **Description:** If Upsell Person is set, use it; otherwise use Original Sales from project

### Hiring & Referrals

#### Hiring Referral (Formula)
- **ID:** `\MKl`
- **Type:** Formula
- **Description:** Complex formula that retrieves hiring referrer from contractor's candidate record

#### Hiring Referral Status (Formula)
- **ID:** `dJ{b`
- **Type:** Formula
- **Description:** Status of the hiring referral from contractor's candidate record

### Contractor Info (Rollups)

#### Person (Rollup)
- **ID:** `UD}F`
- **Type:** Rollup
- **Rolls up:** `Person` from `Contractor` relation
- **Function:** Show Original

#### Discord (Rollup)
- **ID:** `lbu<`
- **Type:** Rollup
- **Rolls up:** `Discord` from `Contractor` relation
- **Function:** Show Original

### Project Info (Rollups)

#### Code (Rollup)
- **ID:** `_{<}`
- **Type:** Rollup
- **Rolls up:** `Codename` from `Project` relation
- **Function:** Show Original

### Other

#### ID (Unique ID)
- **ID:** `Gucn`
- **Type:** Unique ID
- **Description:** Auto-generated unique identifier

#### Charges (Rich Text)
- **ID:** `G]DO`
- **Type:** Rich Text
- **Description:** Additional charges or notes

---

## Special Rules for Email Sending

### Shadow Deployment Client Handling

When sending task order confirmation emails, deployments with `Type = "Shadow"` have special client handling:

**Rule:** For Shadow deployments, the client should be considered as **"Dwarves LLC"** with headquarters in **"USA"**, regardless of the actual project's client.

**Reason:** Shadow deployments are internal training engagements where contractors work alongside official team members. Since these are not billable client projects, they are treated as Dwarves LLC internal assignments.

**Implementation Location:** `pkg/handler/notion/task_order_log.go` in `SendTaskOrderConfirmation` handler

**Current Implementation for Vietnam Clients:**
```go
// If client is in Vietnam, use "Dwarves LLC" (USA) instead
if strings.TrimSpace(clientInfo.Country) == "Vietnam" {
    clientInfo.Name = "Dwarves LLC"
    clientInfo.Country = "USA"
}
```

**TODO:** Add similar logic for Shadow deployments:
```go
// For Shadow deployments, use "Dwarves LLC" (USA)
// Check deployment type from DeploymentData
```

---

## Usage in Codebase

### Query Functions

**Location:** `pkg/service/notion/task_order_log.go`

#### `QueryActiveDeploymentsByMonth`
Queries active deployments filtered by:
- `Deployment Status = "Active"`
- Optional: Discord username filter

Returns:
```go
type DeploymentData struct {
    PageID           string // Deployment page ID
    ContractorPageID string // From Contractor relation
    ProjectPageID    string // From Project relation
    Status           string // Deployment status
}
```

#### `GetDeploymentByContractor`
Gets deployment ID for a specific contractor

#### `GetDeploymentByContractorAndProject`
Gets deployment ID for a specific contractor-project pair

### Related Handlers

- **`SendTaskOrderConfirmation`** (`pkg/handler/notion/task_order_log.go:311`): Bulk sends monthly task order confirmation emails based on active deployments

---

## Database Relations

### Outgoing Relations
- **Contractor** → Contractor database
- **Project** → Project database
- **Shadow For** → Contractor database
- **Upsell Person** → Contractor database
- **Override DL** → Contractor database
- **Override AM** → Contractor database

### Incoming Relations
- Used by Task Order Log generation
- Referenced in hiring tracker
- Linked from contractor invoice generation

---

## Notes

1. **Unique ID System:** Each deployment has an auto-generated unique ID for tracking
2. **Override System:** DL and AM can be overridden at deployment level, overriding project defaults
3. **Shadow Deployments:** When Type includes "Shadow", the `Shadow For` relation indicates the mentor/contractor being shadowed
4. **Sales Attribution:** Supports both original sales and upsell tracking with Final Sales Credit formula
5. **Hiring Tracking:** Complex formula system links back to candidate records to track referrals
6. **Auto Naming:** Auto Name formula generates standardized naming format for easy identification
