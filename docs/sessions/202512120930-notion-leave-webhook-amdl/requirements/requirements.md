# Notion Leave Webhook - AM/DL Integration Requirements

## Summary

Enhance the Notion leave request webhook to automatically fetch Account Managers (AM) and Delivery Leads (DL) from the Deployment Tracker and notify them via Discord.

## Functional Requirements

### FR1: Replace Assignees with AM/DL

- Remove current `Assignees` multi_select functionality
- Fetch AM/DL dynamically from Deployment Tracker based on employee's active deployments
- Use these AM/DL for Discord notifications instead of manually set assignees

### FR2: Discord Mention Lookup

- Get `Team Email` from Contractors DB in Notion
- Lookup employee in fortress DB by email
- Get Discord ID from fortress DB
- Format as `<@discord_id>` for mentions

### FR3: Auto-fill Approved/Rejected By

- When leave is approved/rejected via Discord button
- Fill `Approved/Rejected By` relation field on Notion Leave Request page
- Lookup contractor page by Discord username of the approver
- Set relation to that contractor page ID

### FR4: No Auto-update for Submitted Discord

- `Submitted Discord` field remains manually filled
- No automation needed for this field

## Data Flow

```
Leave Request Created (Notion)
    ↓
Get Team Email from Leave Request
    ↓
Lookup Contractor in Contractors DB by Team Email
    ↓
Query Deployment Tracker for active deployments
    ↓
Extract AM/DL from each deployment (Override AM/DL or rollup)
    ↓
For each AM/DL: Get Team Email → Lookup in fortress DB → Get Discord ID
    ↓
Send Discord notification with AM/DL mentions
    ↓
On Approve/Reject button click → Update Notion with approver's Discord username
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `NOTION_LEAVE_REQUEST_DB_ID` | Leave Requests database ID |
| `NOTION_DEPLOYMENT_TRACKER_DB_ID` | Deployment Tracker database ID |
| `NOTION_CONTRACTOR_DB_ID` | Contractors database ID |

## Notion Properties

### Leave Request
- `Team Email` (email) - Employee's email
- `Contractor` (relation) - Link to Contractors DB
- `Approved/Rejected By` (relation) - Link to Contractors DB (approver/rejecter)

### Deployment Tracker
- `Contractor` (relation) - Link to Contractors DB
- `Deployment Status` (status) - Filter by "Active"
- `Override AM` (relation) - Override Account Manager
- `Override DL` (relation) - Override Delivery Lead
- `Account Managers` (rollup) - From Project
- `Delivery Leads` (rollup) - From Project

### Contractors
- `Team Email` (email) - Employee's team email
