# Contractors Database Schema

## Overview

- **Database ID**: `9d468753-ebb4-4977-a8dc-156428398a6b`
- **Data Source ID**: `ed2b9224-97d9-4dff-97f9-82598b61f65d`
- **Title**: Contractors
- **Created**: 2021-06-24T05:42:00.000Z
- **Last Edited**: 2026-01-18T09:10:00.000Z
- **URL**: https://www.notion.so/9d468753ebb44977a8dc156428398a6b

## Purpose

The Contractors database stores contractor profiles including personal information, contact details, positions/roles, and employment status. It serves as the master record for all contractor-related operations.

## Properties

### Core Identity Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Full Name` | Title | `title` | Contractor's full name |
| `First name` | Rich Text | `P@WN` | First name |
| `Last name` | Rich Text | `ktsr` | Last name |
| `Discord` | Rich Text | `l\`p^` | Discord username/handle |
| `Person` | People | `R\[I` | Notion user reference |

### Contact Information

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Team Email` | Email | `.\|oa` | Work email address |
| `Personal Email` | Email | `Q@@Z` | Personal email address |
| `Mobile #` | Phone | `x/&D` | Mobile phone number |
| `Facebook` | URL | `vEay` | Facebook profile URL |
| `Twitter` | URL | `RM<m` | Twitter profile URL |
| `Linkedin` | URL | `b;7o` | LinkedIn profile URL |
| `GitHub` | URL | `776faea1-...` | GitHub profile URL |

### Position Property

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Position` | Multi-select | `9Dl2` | Contractor's positions/roles |

#### Position Options

##### Leadership Positions
| Position | Color |
|----------|-------|
| Managing Director | Brown |
| Partner | Purple |
| CEO | Blue |
| CTO | Red |
| Head of Operations | Blue |
| Head of Business | Blue |
| Head of Engineering | Blue |

##### Senior Technical Roles
| Position | Color |
|----------|-------|
| DevOps Principal | Blue |
| Software Architect | Purple |
| Technical Lead | Gray |
| Frontend Lead | Purple |
| Backend Lead | Purple |
| iOS Lead | Purple |

##### Engineering Positions
| Position | Color |
|----------|-------|
| Frontend | Red |
| Backend | Red |
| Web | Red |
| iOS | Red |
| Android | Red |
| Full-stack | Red |
| React Native | Default |
| Blockchain | Red |
| Java Engineer | Brown |
| Bubble | Red |

##### QA/Testing Positions
| Position | Color |
|----------|-------|
| QC Automation | Red |
| QC Manual | Red |

##### Data Positions
| Position | Color |
|----------|-------|
| Data Engineer | Red |
| Data Analyst | Red |
| Product Data | Orange |

##### Design Positions
| Position | Color |
|----------|-------|
| Design Ops | Yellow |
| Product Designer | Yellow |
| UX Designer | Yellow |
| UI Designer | Yellow |

##### Operations & Business Positions
| Position | Color |
|----------|-------|
| Account Manager | Purple |
| Consultant | Gray |
| Business Ops | Green |
| Community Ops | Green |
| Technical Recruiter | Green |
| Marketing Executive | Green |
| Operation Executive | Green |
| Community Executive | Green |

##### Other Positions
| Position | Color |
|----------|-------|
| Apprentices | Purple |
| Partnership | Gray |
| Branding | Gray |
| Unity Game Client | Red |
| DevOps | Pink |

### Employment Status

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Status` | Select | `Xfmt` | Employment status |

#### Status Options

| Status | Color | Description |
|--------|-------|-------------|
| Active | Green | Currently employed |
| Inactive | Gray | No longer employed |
| Intern | Blue | Internship |
| Apprentices | Blue | Apprenticeship program |

### Employment Dates

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Onboard Date` | Date | `IC^?` | Start date |
| `Offboard Date` | Date | `WKb<` | End date (if applicable) |
| `Birthday` | Date | `z-hV` | Date of birth |

### Payment Configuration

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Payday` | Select | `W:?j` | Preferred payment day |
| `Payment Method` | Relation | `S\xu` | Link to payment methods |

#### Payday Options

| Day | Color |
|-----|-------|
| 01 | Orange |
| 15 | Brown |

### Location & Identity

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Country` | Select | `_S?}` | Country of residence |
| `Local employment` | Select | `J\`K]` | Local employment jurisdiction |
| `Permanent Address` | Rich Text | `wZ<P` | Home address |
| `Shelter Address` | Rich Text | `9hd?` | Current address |
| `ID No.` | Rich Text | `dYz>` | ID card number |
| `New CCCD No.` | Rich Text | `ur^v` | New citizen ID number |

#### Country Options

| Country | Color |
|---------|-------|
| VN | Gray |
| US | Yellow |
| JP | Pink |
| SG | Default |
| Indo | Purple |

### Personal Information

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Blood Type` | Select | `;\HDq` | Blood type |
| `Horoscope` | Select | `Spl^` | Zodiac sign |
| `MBTI` | Select | `\mKt` | Personality type (full name) |
| `MBTI Abbr.` | Select | `\`sxw` | MBTI abbreviation |

### Skills & Technical

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Stack` | Multi-select | `gCqt` | Technical stack/skills |

#### Stack Options

| Stack | Color |
|-------|-------|
| Laravel | Orange |
| ReactJS | Yellow |
| Wordpress | Pink |
| NodeJS | Blue |
| Blockchain | Default |
| Android | Green |
| Golang | Gray |
| Java | Red |
| Python | Red |
| Typescript | Gray |
| React Native | Pink |
| Swift | Orange |
| PHP | Gray |
| Angular | Green |
| Unity | Gray |
| Elixir | Blue |
| Data | Red |
| Devops | Pink |
| Designer | Default |
| ... | ... |

### Files & Documents

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `CV` | Files | `Z}Bf` | Curriculum vitae |
| `Headshot` | Files | `\|OYL` | Profile photo |
| `ID Image` | Files | `rM\G` | ID card image |
| `New CCCD Image` | Files | `~g^A` | New citizen ID image |

### Relations

| Property | Type | ID | Related Database | Description |
|----------|------|-----|------------------|-------------|
| `Candidate Record` | Relation | `g_CZ` | Candidates (`2b764b29-b84c-802c-...`) | Link to hiring record |
| `Payment Method` | Relation | `S\xu` | Payment Methods (`2c864b29-b84c-8036-...`) | Payment configuration |

### Rollup Properties

| Property | Type | ID | Source | Description |
|----------|------|-----|--------|-------------|
| `Referred By` | Rollup | `iEgD` | `Candidate Record` → `Referred by` | Referral source |
| `Onboard` | Rollup | `uwfA` | `Candidate Record` → `Onboarding date` | From candidate record |

### Other Properties

| Property | Type | ID | Description |
|----------|------|-----|-------------|
| `Note` | Rich Text | `\`?Gu` | General notes |
| `Rate Generated?` | Checkbox | `mNHZ` | Whether rate has been generated |
| `Gen Rate` | Button | `XmzJ` | Button to generate rate |

## API Integration

### Query Contractor by Person Page ID

```go
// Fetch contractor page directly by ID
page, err := client.FindPageByID(ctx, contractorPageID)
props := page.Properties.(nt.DatabasePageProperties)

// Extract Position multi-select
if positionProp, ok := props["Position"]; ok {
    for _, opt := range positionProp.MultiSelect {
        fmt.Println(opt.Name) // e.g., "Frontend", "Backend", "Product Designer"
    }
}
```

### Query Contractor by Discord

```javascript
{
  "filter": {
    "property": "Discord",
    "rich_text": {
      "equals": "username"
    }
  }
}
```

### Query Active Contractors

```javascript
{
  "filter": {
    "property": "Status",
    "select": {
      "equals": "Active"
    }
  }
}
```

## Usage in Fortress API

### Position-Based Service Fee Description

The `Position` field is used to determine the service fee description type:
- Contains "design" (case-insensitive) → "Design Consulting Services Rendered"
- Contains "Operation Executive" → "Operational Consulting Services Rendered"
- Default → "Software Development Services Rendered"

### Related Services

- `ContractorPayoutsService.GetContractorPositions()` - Fetches positions for invoice description generation
- `TimesheetService.FindContractorByPersonID()` - Looks up contractor by Notion person ID

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                   Contractors Table                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ Contractor Profile                                   │    │
│  │ - Full Name, Discord, Email                         │    │
│  │ - Position (Multi-select) ← Used for invoice desc   │    │
│  │ - Status (Active/Inactive)                          │    │
│  │ - Payday (01/15)                                    │    │
│  └─────────────────────────────────────────────────────┘    │
│          │                                                   │
│          │ referenced by                                     │
│          ▼                                                   │
│  ┌─────────────────┐    ┌─────────────────┐                 │
│  │ Contractor      │    │ Contractor      │                 │
│  │ Rates           │    │ Payouts         │                 │
│  │ (billing info)  │    │ (Person field)  │                 │
│  └─────────────────┘    └─────────────────┘                 │
│          │                      │                            │
│          └──────────┬───────────┘                            │
│                     ▼                                        │
│          ┌─────────────────┐                                 │
│          │ Contractor      │                                 │
│          │ Payables        │                                 │
│          │ (invoice gen)   │                                 │
│          └─────────────────┘                                 │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Notes

- `Position` is a **multi-select** field - contractors can have multiple positions
- The `Person` field links to Notion user accounts
- `Discord` username is the primary identifier for contractor operations
- `Payday` determines when the contractor prefers to receive payment (1st or 15th)
- Status should be "Active" for current contractors
