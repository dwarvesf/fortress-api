# Technical Specification: Send Task Order Confirmation Email

## Overview

This specification defines the implementation for a cronjob endpoint that sends monthly task order confirmation emails to contractors with their active client assignments.

**Endpoint**: `POST /cronjobs/send-task-order-confirmation`

**Purpose**: Query active deployments from Notion Deployment Tracker and send confirmation emails via Gmail (using accounting refresh token) to contractors listing their client assignments for the month.

## API Contract

### HTTP Method

`POST`

### URL Path

`/api/v1/cronjobs/send-task-order-confirmation`

### Authentication

- **Required**: Bearer token authentication
- **Permission**: `model.PermissionCronjobExecute`
- **Bypass**: Auth bypassed in local environment (per existing pattern)

### Request

**Headers**:
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**: None (no request body required)

**Query Parameters**:

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `month` | string | No | Current month | Target month in YYYY-MM format (e.g., "2026-01") |
| `discord` | string | No | None | Discord username to filter specific contractor |

**Examples**:
```
POST /api/v1/cronjobs/send-task-order-confirmation
POST /api/v1/cronjobs/send-task-order-confirmation?month=2026-01
POST /api/v1/cronjobs/send-task-order-confirmation?discord=johnd
POST /api/v1/cronjobs/send-task-order-confirmation?month=2026-02&discord=janedoe
```

### Response

**Success Response** (HTTP 200):

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
        "clients": [
          "Acme Corp (USA)",
          "Tech Solutions (Singapore)"
        ]
      },
      {
        "contractor": "Jane Smith",
        "discord": "janes",
        "email": "jane@example.com",
        "status": "failed",
        "error": "Gmail API error: invalid email address",
        "clients": [
          "Global Industries (UK)"
        ]
      }
    ]
  },
  "error": null,
  "message": "ok"
}
```

**Field Descriptions**:

| Field | Type | Description |
|-------|------|-------------|
| `month` | string | Target month in YYYY-MM format |
| `emails_sent` | number | Count of successfully sent emails |
| `emails_failed` | number | Count of failed email attempts |
| `details` | array | Per-contractor processing results |
| `details[].contractor` | string | Contractor full name |
| `details[].discord` | string | Contractor Discord username |
| `details[].email` | string | Team email address used |
| `details[].status` | string | "sent" or "failed" |
| `details[].clients` | array | List of client names with headquarters |
| `details[].error` | string | Error message if status is "failed" |

**Error Response** (HTTP 400):

```json
{
  "data": null,
  "error": "invalid month format, expected YYYY-MM (e.g., 2026-01)",
  "message": ""
}
```

**Error Response** (HTTP 500):

```json
{
  "data": null,
  "error": "task order log service not configured",
  "message": ""
}
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ STEP 1: Parse and Validate Parameters                          │
│ - Parse month (default to current if not provided)             │
│ - Validate month format (YYYY-MM)                              │
│ - Parse discord filter (optional)                              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 2: Query Active Deployments                               │
│ Service: TaskOrderLogService.QueryActiveDeploymentsByMonth()   │
│ Filter: Status=Active AND (Discord=? if provided)              │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 3: Group Deployments by Contractor                        │
│ Group by: Contractor Page ID                                   │
│ Result: Map[contractorID] → []DeploymentData                   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4: For Each Contractor                                    │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4a: Extract Contractor Info                               │
│ - Contractor Name (from contractor page)                       │
│ - Team Email (from contractor page)                            │
│ - Discord username (from contractor page)                      │
│ Skip if: Team Email is empty                                   │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4b: Extract Client Information                            │
│ For each deployment:                                           │
│   - Get Project page ID from deployment                        │
│   - Fetch Client from project relation                         │
│   - Extract Client Name and Country                            │
│ Result: []string of "Client Name (Country)"                    │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4c: Generate Email Content                                │
│ - Build HTML email from template                               │
│ - Insert contractor name, month, client list                   │
│ - Calculate period end date (last day of month)                │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 4d: Send Email via Gmail                                  │
│ Service: GoogleMail.SendTaskOrderConfirmation()                │
│ To: Contractor Team Email                                      │
│ Subject: Monthly Task Order – [Tháng/Năm]                      │
│ Note: Log error but continue if send fails                     │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         v
┌─────────────────────────────────────────────────────────────────┐
│ STEP 5: Aggregate Results and Return Response                  │
│ - Count emails_sent, emails_failed                             │
│ - Build detailed per-contractor results                        │
└─────────────────────────────────────────────────────────────────┘
```

## Service Layer Specification

### TaskOrderLogService Extensions

**File**: `pkg/service/notion/task_order_log.go`

#### New Data Structure

```go
// DeploymentData represents an active deployment from Deployment Tracker
type DeploymentData struct {
    PageID           string // Deployment page ID
    ContractorPageID string // From Contractor relation
    ProjectPageID    string // From Project relation
    Status           string // Deployment status
}

// ContractorWithClients represents a contractor with their client assignments
type ContractorWithClients struct {
    ContractorPageID string
    Name             string
    TeamEmail        string
    Discord          string
    Clients          []ClientInfo
}

// ClientInfo represents client information
type ClientInfo struct {
    Name    string
    Country string
}
```

#### Method: QueryActiveDeploymentsByMonth

```go
// QueryActiveDeploymentsByMonth queries active deployments from Deployment Tracker
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - month: Target month in YYYY-MM format (used for logging/filtering if needed)
//   - contractorDiscord: Optional Discord username filter (empty string = all contractors)
//
// Returns:
//   - []*DeploymentData: Slice of active deployments
//   - error: Error if query fails
//
// Filters:
//   - Deployment Status = "Active"
//   - Contractor Discord (rollup) = contractorDiscord (if provided)
//
// Extracts:
//   - Contractor page ID from Contractor relation
//   - Project page ID from Project relation
//   - Deployment status
//
// Handles:
//   - Pagination (page size 100)
//   - Empty relations
//   - Missing properties
func (s *TaskOrderLogService) QueryActiveDeploymentsByMonth(ctx context.Context, month string, contractorDiscord string) ([]*DeploymentData, error)
```

**Implementation Notes**:
- Query Deployment Tracker database (ID from `cfg.Notion.Databases.DeploymentTracker`)
- Filter by Deployment Status = "Active"
- If contractorDiscord provided, add rollup filter on Discord property
- Extract Contractor and Project relations from each deployment
- Return slice of DeploymentData

**Notion Query** (no discord filter):
```go
query := &nt.DatabaseQuery{
    Filter: &nt.DatabaseQueryFilter{
        Property: "Deployment Status",
        DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
            Status: &nt.StatusDatabaseQueryFilter{
                Equals: "Active",
            },
        },
    },
    PageSize: 100,
}
```

**Notion Query** (with discord filter):
```go
query := &nt.DatabaseQuery{
    Filter: &nt.DatabaseQueryFilter{
        And: []nt.DatabaseQueryFilter{
            {
                Property: "Deployment Status",
                DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                    Status: &nt.StatusDatabaseQueryFilter{
                        Equals: "Active",
                    },
                },
            },
            {
                Property: "Discord",
                DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
                    Rollup: &nt.RollupDatabaseQueryFilter{
                        Any: &nt.DatabaseQueryPropertyFilter{
                            RichText: &nt.TextPropertyFilter{
                                Contains: contractorDiscord,
                            },
                        },
                    },
                },
            },
        },
    },
    PageSize: 100,
}
```

#### Method: getClientInfo

```go
// getClientInfo fetches client information from a project page
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//   - projectPageID: Project page ID
//
// Returns:
//   - *ClientInfo: Client name and country, nil if not found
//   - error: Error if fetch fails
//
// Extracts:
//   - Client relation from Project page
//   - Client Name from Client page (Title property)
//   - Client Country from Client page (Country property)
//
// Handles:
//   - Missing Client relation
//   - Missing Client properties
func (s *TaskOrderLogService) getClientInfo(ctx context.Context, projectPageID string) (*ClientInfo, error)
```

**Implementation Notes**:
- Fetch Project page by ID
- Extract Client relation (first relation in "Client" property)
- Fetch Client page by ID
- Extract "Name" (title) and "Country" (rich text or select) properties
- Return ClientInfo struct

**Pseudo-code**:
```go
// Fetch project page
projectPage, err := s.client.FindPageByID(ctx, projectPageID)
if err != nil {
    return nil, err
}

// Extract client relation
props := projectPage.Properties.(nt.DatabasePageProperties)
clientPageID := s.extractFirstRelationID(props, "Client")
if clientPageID == "" {
    return nil, errors.New("no client relation found")
}

// Fetch client page
clientPage, err := s.client.FindPageByID(ctx, clientPageID)
if err != nil {
    return nil, err
}

// Extract client info
clientProps := clientPage.Properties.(nt.DatabasePageProperties)
name := s.extractTitle(clientProps, "Name")
country := s.extractRichText(clientProps, "Country") // or extractSelect if Country is select

return &ClientInfo{
    Name:    name,
    Country: country,
}, nil
```

## Handler Implementation

**File**: `pkg/handler/notion/task_order_confirmation.go` (NEW FILE)

### Handler Signature

```go
// SendTaskOrderConfirmation godoc
// @Summary Send monthly task order confirmation emails
// @Description Sends task order confirmation emails to contractors with active client assignments
// @Tags Cronjobs
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param month query string false "Target month in YYYY-MM format (default: current month)"
// @Param discord query string false "Discord username to filter specific contractor"
// @Success 200 {object} view.Response
// @Failure 400 {object} view.Response
// @Failure 500 {object} view.Response
// @Router /cronjobs/send-task-order-confirmation [post]
func (h *handler) SendTaskOrderConfirmation(c *gin.Context)
```

### Handler Logic

```go
func (h *handler) SendTaskOrderConfirmation(c *gin.Context) {
    l := h.logger.Fields(logger.Fields{
        "handler": "Notion",
        "method":  "SendTaskOrderConfirmation",
    })
    ctx := c.Request.Context()

    // Step 1: Parse and validate parameters
    month := c.Query("month")
    if month == "" {
        now := time.Now()
        month = now.Format("2006-01")
    }

    // Validate month format (YYYY-MM)
    if !isValidMonthFormat(month) {
        l.Error(fmt.Errorf("invalid month format"), fmt.Sprintf("month=%s", month))
        c.JSON(http.StatusBadRequest, view.CreateResponse[any](nil, nil,
            fmt.Errorf("invalid month format, expected YYYY-MM (e.g., 2026-01)"), nil, ""))
        return
    }

    contractorDiscord := strings.TrimSpace(c.Query("discord"))
    l.Info(fmt.Sprintf("sending task order confirmations: month=%s discord=%s", month, contractorDiscord))

    // Step 2: Get services
    taskOrderLogService := h.service.Notion.TaskOrderLog
    if taskOrderLogService == nil {
        err := fmt.Errorf("task order log service not configured")
        l.Error(err, "service is nil")
        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    googleMailService := h.service.GoogleMail
    if googleMailService == nil {
        err := fmt.Errorf("google mail service not configured")
        l.Error(err, "service is nil")
        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    // Step 3: Query active deployments
    l.Debug(fmt.Sprintf("querying active deployments for month: %s", month))
    deployments, err := taskOrderLogService.QueryActiveDeploymentsByMonth(ctx, month, contractorDiscord)
    if err != nil {
        l.Error(err, "failed to query deployments")
        c.JSON(http.StatusInternalServerError, view.CreateResponse[any](nil, nil, err, nil, ""))
        return
    }

    l.Debug(fmt.Sprintf("found %d active deployments", len(deployments)))

    if len(deployments) == 0 {
        l.Info("no active deployments found")
        c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
            "month":         month,
            "emails_sent":   0,
            "emails_failed": 0,
            "details":       []any{},
        }, nil, nil, nil, "ok"))
        return
    }

    // Step 4: Group deployments by contractor
    contractorGroups := groupDeploymentsByContractor(deployments)
    l.Debug(fmt.Sprintf("grouped into %d contractors", len(contractorGroups)))

    // Step 5: Process each contractor
    var (
        emailsSent   = 0
        emailsFailed = 0
        details      = []map[string]any{}
    )

    for contractorID, contractorDeployments := range contractorGroups {
        detail := map[string]any{
            "contractor": "",
            "discord":    "",
            "email":      "",
            "status":     "",
            "clients":    []string{},
        }

        // Step 5a: Get contractor info
        name, discord := taskOrderLogService.getContractorInfo(ctx, contractorID)
        if name == "" {
            l.Warn(fmt.Sprintf("skipping contractor %s: no name found", contractorID))
            continue
        }

        // Get team email
        teamEmail := getContractorTeamEmail(ctx, taskOrderLogService, contractorID)
        if teamEmail == "" {
            l.Warn(fmt.Sprintf("skipping contractor %s: no team email", name))
            detail["contractor"] = name
            detail["discord"] = discord
            detail["status"] = "failed"
            detail["error"] = "no team email found"
            emailsFailed++
            details = append(details, detail)
            continue
        }

        detail["contractor"] = name
        detail["discord"] = discord
        detail["email"] = teamEmail

        // Step 5b: Extract client info from deployments
        var clients []string
        for _, deployment := range contractorDeployments {
            clientInfo, err := taskOrderLogService.getClientInfo(ctx, deployment.ProjectPageID)
            if err != nil {
                l.Error(err, fmt.Sprintf("failed to get client for project %s", deployment.ProjectPageID))
                continue
            }
            if clientInfo != nil {
                clients = append(clients, fmt.Sprintf("%s (%s)", clientInfo.Name, clientInfo.Country))
            }
        }

        if len(clients) == 0 {
            l.Warn(fmt.Sprintf("skipping contractor %s: no clients found", name))
            detail["status"] = "failed"
            detail["error"] = "no clients found"
            emailsFailed++
            details = append(details, detail)
            continue
        }

        detail["clients"] = clients

        // Step 5c: Deduplicate clients
        uniqueClients := deduplicateClients(clients)
        detail["clients"] = uniqueClients

        // Step 5d: Send email via Gmail (using accounting refresh token)
        taskOrderEmail := &model.TaskOrderConfirmationEmail{
            ContractorName: name,
            TeamEmail:      teamEmail,
            Month:          month,
            Clients:        uniqueClients,
        }

        err = googleMailService.SendTaskOrderConfirmationMail(taskOrderEmail)
        if err != nil {
            l.Error(err, fmt.Sprintf("failed to send email to %s (%s)", name, teamEmail))
            detail["status"] = "failed"
            detail["error"] = err.Error()
            emailsFailed++
        } else {
            l.Info(fmt.Sprintf("sent email to %s (%s)", name, teamEmail))
            detail["status"] = "sent"
            emailsSent++
        }

        details = append(details, detail)
    }

    // Step 6: Return response
    l.Info(fmt.Sprintf("email sending complete: sent=%d failed=%d", emailsSent, emailsFailed))

    c.JSON(http.StatusOK, view.CreateResponse[any](map[string]any{
        "month":         month,
        "emails_sent":   emailsSent,
        "emails_failed": emailsFailed,
        "details":       details,
    }, nil, nil, nil, "ok"))
}
```

## Email Template

### Helper Functions

```go
// formatMonthVietnamese converts YYYY-MM to Vietnamese format (Tháng MM/YYYY)
func formatMonthVietnamese(month string) string {
    parts := strings.Split(month, "-")
    if len(parts) != 2 {
        return month
    }
    return fmt.Sprintf("Tháng %s/%s", parts[1], parts[0])
}

// calculatePeriodEndDate calculates the last day of the month
func calculatePeriodEndDate(month string) string {
    // Parse month string
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return "31"
    }
    // Get last day of month
    lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC)
    return fmt.Sprintf("%02d", lastDay.Day())
}

// generateTaskOrderEmail generates HTML email content
func generateTaskOrderEmail(contractorName, month string, clients []string) string {
    monthYear := formatMonthVietnamese(month)
    endDay := calculatePeriodEndDate(month)

    // Parse month for period display
    t, _ := time.Parse("2006-01", month)
    monthName := t.Format("January")
    year := t.Format("2006")

    clientList := ""
    for _, client := range clients {
        clientList += fmt.Sprintf("- %s<br>", client)
    }

    return fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; color: #333;">
    <p>Hi %s,</p>

    <p>This email outlines your planned assignments and work order for the upcoming month: <b>%s</b>.</p>

    <p><b>Period:</b> 01 – %s %s %s</p>

    <p><b>Active clients & locations (all outside Vietnam):</b></p>
    <p>%s</p>

    <p><b>Workflow & Tracking:</b> All tasks and deliverables will be tracked in Notion/Jira as usual. Please ensure your capacity is updated to reflect these assignments.</p>

    <p>Please reply <b>"Confirmed – %s"</b> to acknowledge this work order and confirm your availability.</p>

    <p>Thanks, and looking forward to a productive month ahead!</p>

    <p>
        Dwarves LLC<br>
        <a href="https://d.foundation">d.foundation</a>
    </p>
</body>
</html>
    `, contractorName, monthYear, endDay, monthName, year, clientList, monthYear)
}

// getContractorTeamEmail fetches team email from contractor page
func getContractorTeamEmail(ctx context.Context, service *notion.TaskOrderLogService, contractorPageID string) string {
    page, err := service.client.FindPageByID(ctx, contractorPageID)
    if err != nil {
        return ""
    }

    props, ok := page.Properties.(nt.DatabasePageProperties)
    if !ok {
        return ""
    }

    // Extract Team Email property (email type)
    if prop, ok := props["Team Email"]; ok && prop.Email != "" {
        return prop.Email
    }

    return ""
}
```

### Email Example

**Subject**: Monthly Task Order – Tháng 01/2026

**Body**:
```
Hi John Doe,

This email outlines your planned assignments and work order for the upcoming month: Tháng 01/2026.

Period: 01 – 31 January 2026

Active clients & locations (all outside Vietnam):
- Acme Corp (USA)
- Tech Solutions (Singapore)
- Global Industries (UK)

Workflow & Tracking: All tasks and deliverables will be tracked in Notion/Jira as usual. Please ensure your capacity is updated to reflect these assignments.

Please reply "Confirmed – Tháng 01/2026" to acknowledge this work order and confirm your availability.

Thanks, and looking forward to a productive month ahead!

Dwarves LLC
d.foundation
```

## Helper Functions

```go
// groupDeploymentsByContractor groups deployments by contractor page ID
func groupDeploymentsByContractor(deployments []*notion.DeploymentData) map[string][]*notion.DeploymentData {
    groups := make(map[string][]*notion.DeploymentData)
    for _, deployment := range deployments {
        if deployment.ContractorPageID == "" {
            continue
        }
        groups[deployment.ContractorPageID] = append(groups[deployment.ContractorPageID], deployment)
    }
    return groups
}

// isValidMonthFormat validates month format as YYYY-MM
func isValidMonthFormat(month string) bool {
    if len(month) != 7 {
        return false
    }
    if month[4] != '-' {
        return false
    }
    // Parse to validate
    _, err := time.Parse("2006-01", month)
    return err == nil
}
```

## Error Scenarios

| Scenario | HTTP Status | Response | Handler Behavior | User Impact |
|----------|-------------|----------|------------------|-------------|
| No active deployments | 200 | Success with zero counts | Return empty results | None (expected) |
| Invalid month format | 400 | Error response | Return error immediately | Request fails, user must fix |
| Notion API unavailable | 500 | Error response | Return error immediately | Cronjob fails, retry needed |
| Missing contractor email | 200 | Success with failed in details | Skip contractor, continue | Email not sent, logged |
| Missing client data | 200 | Success (partial data) | Continue with available clients | Email sent with partial info |
| Gmail API error | 200 | Success with failed in details | Skip email, continue | Email not sent, logged |
| Service not configured | 500 | Error response | Return error immediately | Cronjob fails, config issue |

## Configuration

### Required Notion Database IDs

These should already be configured in `pkg/config/config.go`:

```go
type NotionDatabase struct {
    DeploymentTracker string // Deployment Tracker database ID
    Contractor        string // Contractor database ID
    Project           string // Project database ID
    // ... other databases
}
```

### Required Google Mail Configuration

```go
type Config struct {
    Google Google
    // ... other config
}

type Google struct {
    AccountingGoogleRefreshToken string // Refresh token for Gmail OAuth
    AccountingEmailID            string // Email user ID (e.g., "me" or email address)
}
```

**Validation**: Services should validate configuration in initialization.

**Email Template**: A new template file `taskOrderConfirmation.tpl` will be created in `pkg/templates/`.

## Logging Strategy

### Log Levels

**DEBUG**:
- Number of deployments found
- Grouping by contractor
- Client extraction per deployment

**INFO**:
- Cronjob started with parameters
- Email sent successfully
- Processing complete with statistics

**WARNING**:
- Contractor skipped (missing email)
- No clients found for contractor

**ERROR**:
- Failed to query deployments
- Failed to get client info
- Failed to send email

### Example Log Output

```
[INFO] sending task order confirmations: month=2026-01 discord=
[DEBUG] querying active deployments for month: 2026-01
[DEBUG] found 15 active deployments
[DEBUG] grouped into 5 contractors
[INFO] sent email to John Doe (john@example.com)
[WARN] skipping contractor Jane Smith: no team email
[ERROR] failed to send email to Bob Wilson (bob@example.com): Gmail API error
[INFO] email sending complete: sent=3 failed=2
```

## Routes Configuration

**File**: `pkg/routes/v1.go`

Add to cronjob group:

```go
cronjob := r.Group("/cronjobs")
{
    // ... existing routes
    cronjob.POST("/send-task-order-confirmation", conditionalAuthMW, conditionalPermMW(model.PermissionCronjobExecute), h.Notion.SendTaskOrderConfirmation)
}
```

**Line**: After line 64 (after create-contractor-payouts)

## Interface Updates

**File**: `pkg/handler/notion/interface.go`

Add method to IHandler:

```go
type IHandler interface {
    // ... existing methods
    SyncTaskOrderLogs(c *gin.Context)
    CreateContractorFees(c *gin.Context)
    CreateContractorPayouts(c *gin.Context)
    SendTaskOrderConfirmation(c *gin.Context) // NEW
}
```

## Test Scenarios

### Unit Tests

**TaskOrderLogService**:
1. QueryActiveDeploymentsByMonth - returns active deployments
2. QueryActiveDeploymentsByMonth - filters by discord username
3. QueryActiveDeploymentsByMonth - handles pagination
4. getClientInfo - extracts client name and country
5. getClientInfo - handles missing client relation

**Handler**:
1. SendTaskOrderConfirmation - sends emails to all contractors
2. SendTaskOrderConfirmation - filters by discord parameter
3. SendTaskOrderConfirmation - defaults to current month
4. SendTaskOrderConfirmation - validates month format
5. SendTaskOrderConfirmation - handles missing contractor email
6. SendTaskOrderConfirmation - continues on individual email failures

### Manual Test Cases

**Test Case 1: Send to Single Contractor**
- Query: `?month=2026-01&discord=johnd`
- Expected: 1 email sent to John Doe
- Response: emails_sent=1, emails_failed=0

**Test Case 2: Send to All Contractors**
- Query: `?month=2026-01`
- Expected: All active contractors receive emails
- Response: emails_sent=N, emails_failed=0

**Test Case 3: Default to Current Month**
- Query: (no parameters)
- Expected: Uses current month, sends emails
- Response: month=<current-month>, emails_sent=N

**Test Case 4: Invalid Month Format**
- Query: `?month=2026-1`
- Expected: HTTP 400 error
- Response: "invalid month format"

## Dependencies

### Internal Dependencies

- `pkg/service/notion/task_order_log.go`
- `pkg/service/googlemail/google_mail.go`
- `pkg/model/task_order_confirmation.go` (NEW)
- `pkg/templates/taskOrderConfirmation.tpl` (NEW)
- `pkg/handler/notion/interface.go`
- `pkg/routes/v1.go`

### External Dependencies

- `github.com/dstotijn/go-notion` - Notion API client
- `google.golang.org/api/gmail/v1` - Gmail API client
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/dwarvesf/fortress-api/pkg/logger` - Logging
- `github.com/dwarvesf/fortress-api/pkg/view` - Response formatting

## Future Enhancements

1. **Email Template Customization**: Store template in config or database
2. **Dry Run Mode**: Add query param to preview emails without sending
3. **Attachment Support**: Attach PDF work order documents
4. **Notification on Completion**: Send Discord message when cronjob completes
5. **Retry Logic**: Automatic retry for failed email sends
6. **Email Tracking**: Store sent emails in database for audit trail
