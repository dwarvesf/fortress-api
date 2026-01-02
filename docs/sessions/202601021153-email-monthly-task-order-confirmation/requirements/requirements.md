# Requirements: Email Monthly Task Order Confirmation

## Overview
Create a new cronjob endpoint that sends monthly task order confirmation emails to contractors with their active client assignments.

## Functional Requirements

### 1. API Endpoint
- **Endpoint**: `POST /api/v1/cronjobs/send-task-order-confirmation`
- **Query Parameters**:
  - `month` (optional): YYYY-MM format, defaults to current month if not specified
  - `discord` (optional): Discord username to filter specific contractor

### 2. Data Source
- Query active deployments from **Deployment Tracker** in Notion
- For each deployment, fetch:
  - Contractor information (Name, Team Email, Discord)
  - Project information
  - Client information (Name, Country/Headquarters)

### 3. Email Content
- Send email to contractor's "Team Email" field
- Email includes:
  - Contractor name
  - Month/Year period
  - List of active clients with headquarters location
  - Confirmation request

### 4. Email Template
```
Subject: Monthly Task Order – [Tháng/Năm]

Hi [Name],

This email outlines your planned assignments and work order for the upcoming month: [Tháng/Năm].

Period: 01 – [30/31] [Tháng] [Year]

Active clients & locations (all outside Vietnam):
- Client A – headquartered in [Country A]
- Client B – headquartered in [Country B]

Workflow & Tracking: All tasks and deliverables will be tracked in Notion/Jira as usual. Please ensure your capacity is updated to reflect these assignments.

Please reply "Confirmed – [Tháng/Năm]" to acknowledge this work order and confirm your availability.

Thanks, and looking forward to a productive month ahead!

[Your Name]
Dwarves LLC
```

## Response Format
```json
{
  "data": {
    "month": "2026-01",
    "emails_sent": 5,
    "emails_failed": 1,
    "details": [
      {
        "contractor": "John Doe",
        "discord": "johnd",
        "email": "john@example.com",
        "status": "sent",
        "clients": ["Client A (USA)", "Client B (Singapore)"]
      }
    ]
  }
}
```

## Technical Requirements

### Services Used
- `TaskOrderLogService` - Query deployments, get client info from Notion
- `GoogleMail` - Send emails via Gmail using accounting refresh token

### Notion Database Relations
- Deployment Tracker → Contractor (relation) → Name, Team Email, Discord
- Deployment Tracker → Project (relation) → Client (relation) → Name, Country

## Clarifications from User
1. **Email Action**: Send email via SendGrid (not just generate/save)
2. **Client Data**: Show Client name with Headquarters (Country)
3. **Data Source**: Query from Deployment Tracker (not Task Order Log)
4. **Month Parameter**: Optional, defaults to current month
5. **Discord Parameter**: Optional, for testing/filtering specific contractor
