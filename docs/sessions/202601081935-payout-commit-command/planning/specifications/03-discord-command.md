# Specification: Discord Command for Payout Commit

## Overview

This specification defines the Discord command implementation for `?payout commit` in the fortress-discord repository. The command follows the established pattern: command → service → view → adapter.

## File Structure

```
fortress-discord/
├── pkg/
│   ├── discord/
│   │   ├── command/
│   │   │   ├── payout/
│   │   │   │   ├── command.go    [CREATE]
│   │   │   │   └── new.go        [CREATE]
│   │   │   └── command.go        [MODIFY - register command]
│   │   ├── service/
│   │   │   └── payout/
│   │   │       ├── service.go    [CREATE]
│   │   │       └── interface.go  [CREATE]
│   │   └── view/
│   │       └── payout/
│   │           └── payout.go     [CREATE]
│   ├── adapter/
│   │   └── fortress/
│   │       └── payout.go         [CREATE]
│   └── model/
│       └── payout.go             [CREATE]
```

## Command Flow

```
User: ?payout commit 2025-01 15
         |
         v
Command.Execute()
  - Parse arguments (month, batch)
  - Validate format
         |
         v
Service.PreviewCommit(month, batch)
  - Call API preview endpoint
  - Return preview data
         |
         v
View.ShowConfirmation(message, previewData)
  - Build Discord embed
  - Add Confirm/Cancel buttons
         |
         v
[User clicks Confirm]
         |
         v
Service.ExecuteCommit(month, batch)
  - Call API commit endpoint
  - Return result
         |
         v
View.ShowResult(message, result)
  - Show success/failure message
```

## 1. Command Implementation

### File: `pkg/discord/command/payout/command.go`

```go
package payout

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/dwarvesf/fortress-discord/pkg/config"
    "github.com/dwarvesf/fortress-discord/pkg/discord/service"
    "github.com/dwarvesf/fortress-discord/pkg/discord/view"
    "github.com/dwarvesf/fortress-discord/pkg/logger"
    "github.com/dwarvesf/fortress-discord/pkg/model"
    "github.com/dwarvesf/fortress-discord/pkg/utils/permutil"
)

type Command struct {
    cfg    *config.Config
    logger logger.Logger
    svc    service.IService
    view   view.IView
}

func (c *Command) Prefix() []string {
    return []string{"payout"}
}

func (c *Command) Execute(message *model.DiscordMessage) error {
    args := message.ContentArgs[1:] // Skip "payout"

    if len(args) == 0 {
        return c.Help(message)
    }

    subcommand := strings.ToLower(args[0])
    args = args[1:]

    switch subcommand {
    case "commit":
        return c.commit(message, args)
    case "help", "h":
        return c.Help(message)
    default:
        return c.Help(message)
    }
}

func (c *Command) Name() string {
    return "Payout Command"
}

func (c *Command) Help(message *model.DiscordMessage) error {
    return c.view.Payout().Help(message)
}

func (c *Command) DefaultCommand(message *model.DiscordMessage) error {
    return c.Help(message)
}

func (c *Command) PermissionCheck(message *model.DiscordMessage) (bool, []string) {
    // Require admin or ops role
    return permutil.AdminOrAbove(message.Roles)
}

// commit handles the "payout commit <month> <batch>" subcommand
func (c *Command) commit(message *model.DiscordMessage, args []string) error {
    l := c.logger.Fields(logger.Fields{
        "cmd":     "Payout Command",
        "subcmd":  "commit",
        "user":    message.Author.ID,
    })

    // Validate arguments
    if len(args) != 2 {
        return c.view.Error().Raise(message, "Usage: ?payout commit <month> <batch>\nExample: ?payout commit 2025-01 15")
    }

    month := strings.TrimSpace(args[0])
    batchStr := strings.TrimSpace(args[1])

    // Validate month format (YYYY-MM)
    if !isValidMonthFormat(month) {
        return c.view.Error().Raise(message, "Invalid month format. Use YYYY-MM (e.g., 2025-01)")
    }

    // Validate batch (must be 1 or 15)
    batch, err := strconv.Atoi(batchStr)
    if err != nil || (batch != 1 && batch != 15) {
        return c.view.Error().Raise(message, "Batch must be 1 or 15")
    }

    l.Debug(fmt.Sprintf("previewing commit for month=%s batch=%d", month, batch))

    // Call service to preview commit
    preview, err := c.svc.Payout().PreviewCommit(month, batch)
    if err != nil {
        l.Error(err, "failed to preview commit")
        return c.view.Error().Raise(message, err.Error())
    }

    // Handle empty result
    if preview.Count == 0 {
        return c.view.Payout().NoPayables(message, month, batch)
    }

    l.Debug(fmt.Sprintf("preview returned %d payables, total=%.2f", preview.Count, preview.TotalAmount))

    // Show confirmation dialog
    return c.view.Payout().ShowConfirmation(message, preview)
}

// ExecuteCommitConfirmation performs the actual commit after user confirms
func (c *Command) ExecuteCommitConfirmation(message *model.DiscordMessage, month string, batch int) error {
    l := c.logger.Fields(logger.Fields{
        "cmd":   "Payout Command",
        "month": month,
        "batch": batch,
        "user":  message.Author.ID,
    })

    l.Debug("executing confirmed payout commit")

    // Call service to execute commit
    result, err := c.svc.Payout().ExecuteCommit(month, batch)
    if err != nil {
        l.Error(err, "failed to execute commit")
        return c.view.Error().Raise(message, err.Error())
    }

    l.Info(fmt.Sprintf("commit complete: updated=%d failed=%d", result.Updated, result.Failed))

    // Display result
    return c.view.Payout().ShowResult(message, result)
}

// isValidMonthFormat validates month is in YYYY-MM format
func isValidMonthFormat(month string) bool {
    if len(month) != 7 {
        return false
    }
    if month[4] != '-' {
        return false
    }
    parts := strings.Split(month, "-")
    if len(parts) != 2 {
        return false
    }
    if len(parts[0]) != 4 || len(parts[1]) != 2 {
        return false
    }
    return true
}
```

### File: `pkg/discord/command/payout/new.go`

```go
package payout

import (
    "github.com/dwarvesf/fortress-discord/pkg/config"
    "github.com/dwarvesf/fortress-discord/pkg/discord/service"
    "github.com/dwarvesf/fortress-discord/pkg/discord/view"
    "github.com/dwarvesf/fortress-discord/pkg/logger"
)

func New(cfg *config.Config, logger logger.Logger, svc service.IService, view view.IView) *Command {
    return &Command{
        cfg:    cfg,
        logger: logger,
        svc:    svc,
        view:   view,
    }
}
```

---

## 2. Service Layer

### File: `pkg/discord/service/payout/interface.go`

```go
package payout

import "github.com/dwarvesf/fortress-discord/pkg/model"

type IService interface {
    PreviewCommit(month string, batch int) (*model.PayoutPreview, error)
    ExecuteCommit(month string, batch int) (*model.PayoutCommitResult, error)
}
```

### File: `pkg/discord/service/payout/service.go`

```go
package payout

import (
    "fmt"

    "github.com/dwarvesf/fortress-discord/pkg/adapter/fortress"
    "github.com/dwarvesf/fortress-discord/pkg/config"
    "github.com/dwarvesf/fortress-discord/pkg/logger"
    "github.com/dwarvesf/fortress-discord/pkg/model"
)

type service struct {
    cfg     *config.Config
    logger  logger.Logger
    adapter fortress.IAdapter
}

func New(cfg *config.Config, logger logger.Logger, adapter fortress.IAdapter) IService {
    return &service{
        cfg:     cfg,
        logger:  logger,
        adapter: adapter,
    }
}

func (s *service) PreviewCommit(month string, batch int) (*model.PayoutPreview, error) {
    l := s.logger.Fields(logger.Fields{
        "service": "payout",
        "method":  "PreviewCommit",
        "month":   month,
        "batch":   batch,
    })

    l.Debug("calling API preview endpoint")

    preview, err := s.adapter.Payout().PreviewCommit(month, batch)
    if err != nil {
        l.Error(err, "failed to call preview endpoint")
        return nil, fmt.Errorf("failed to preview commit: %w", err)
    }

    l.Debug(fmt.Sprintf("preview returned: count=%d total=%.2f", preview.Count, preview.TotalAmount))

    return preview, nil
}

func (s *service) ExecuteCommit(month string, batch int) (*model.PayoutCommitResult, error) {
    l := s.logger.Fields(logger.Fields{
        "service": "payout",
        "method":  "ExecuteCommit",
        "month":   month,
        "batch":   batch,
    })

    l.Debug("calling API commit endpoint")

    result, err := s.adapter.Payout().ExecuteCommit(month, batch)
    if err != nil {
        l.Error(err, "failed to call commit endpoint")
        return nil, fmt.Errorf("failed to execute commit: %w", err)
    }

    l.Debug(fmt.Sprintf("commit returned: updated=%d failed=%d", result.Updated, result.Failed))

    return result, nil
}
```

---

## 3. View Layer

### File: `pkg/discord/view/payout/payout.go`

```go
package payout

import (
    "fmt"

    "github.com/bwmarrin/discordgo"

    "github.com/dwarvesf/fortress-discord/pkg/logger"
    "github.com/dwarvesf/fortress-discord/pkg/model"
    "github.com/dwarvesf/fortress-discord/pkg/utils/interactionutil"
)

const (
    ColorGreen  = 0x2ECC71
    ColorOrange = 0xE67E22
    ColorRed    = 0xE74C3C
    ColorBlue   = 0x3498DB
)

type View struct {
    logger  logger.Logger
    session *discordgo.Session
}

func New(logger logger.Logger, session *discordgo.Session) *View {
    return &View{
        logger:  logger,
        session: session,
    }
}

// Help shows the help message for payout command
func (v *View) Help(message *model.DiscordMessage) error {
    embed := &discordgo.MessageEmbed{
        Title:       "Payout Command",
        Description: "Commit contractor payables from Pending to Paid status",
        Color:       ColorBlue,
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:  "Usage",
                Value: "`?payout commit <month> <batch>`",
            },
            {
                Name:  "Parameters",
                Value: "`month`: Billing period in YYYY-MM format (e.g., 2025-01)\n`batch`: PayDay batch, must be 1 or 15",
            },
            {
                Name:  "Example",
                Value: "`?payout commit 2025-01 15`",
            },
            {
                Name:  "Permissions",
                Value: "Requires admin or ops role",
            },
        },
    }

    _, err := v.session.ChannelMessageSendEmbed(message.ChannelID, embed)
    return err
}

// NoPayables shows message when no pending payables found
func (v *View) NoPayables(message *model.DiscordMessage, month string, batch int) error {
    embed := &discordgo.MessageEmbed{
        Title:       "No Pending Payables",
        Description: fmt.Sprintf("No pending payables found for **%s** batch **%d**", month, batch),
        Color:       ColorOrange,
    }

    _, err := v.session.ChannelMessageSendEmbed(message.ChannelID, embed)
    return err
}

// ShowConfirmation shows the preview and asks for confirmation
func (v *View) ShowConfirmation(message *model.DiscordMessage, preview *model.PayoutPreview) error {
    // Build contractor list (limit to first 10 to avoid embed size limits)
    contractorList := ""
    displayCount := preview.Count
    if displayCount > 10 {
        displayCount = 10
    }

    for i := 0; i < displayCount; i++ {
        contractor := preview.Contractors[i]
        contractorList += fmt.Sprintf("• %s - %.2f %s\n", contractor.Name, contractor.Amount, contractor.Currency)
    }

    if preview.Count > 10 {
        contractorList += fmt.Sprintf("\n... and %d more", preview.Count-10)
    }

    embed := &discordgo.MessageEmbed{
        Title:       "Confirm Payout Commit",
        Description: fmt.Sprintf("You are about to commit payables for **%s** batch **%d**", preview.Month, preview.Batch),
        Color:       ColorOrange,
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:  "Count",
                Value: fmt.Sprintf("%d payables", preview.Count),
            },
            {
                Name:  "Total Amount",
                Value: fmt.Sprintf("$%.2f", preview.TotalAmount),
            },
            {
                Name:  "Contractors",
                Value: contractorList,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: "This will update all related records (Payables, Payouts, Invoice Splits, Refunds)",
        },
    }

    // Create Confirm/Cancel buttons
    components := []discordgo.MessageComponent{
        discordgo.ActionsRow{
            Components: []discordgo.MessageComponent{
                discordgo.Button{
                    Label:    "Confirm",
                    Style:    discordgo.SuccessButton,
                    CustomID: fmt.Sprintf("payout_commit_confirm:%s:%d", preview.Month, preview.Batch),
                },
                discordgo.Button{
                    Label:    "Cancel",
                    Style:    discordgo.DangerButton,
                    CustomID: "payout_commit_cancel",
                },
            },
        },
    }

    _, err := v.session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
        Embed:      embed,
        Components: components,
    })

    return err
}

// ShowResult shows the result of the commit operation
func (v *View) ShowResult(message *model.DiscordMessage, result *model.PayoutCommitResult) error {
    var embed *discordgo.MessageEmbed

    // Success (no failures)
    if result.Failed == 0 {
        embed = &discordgo.MessageEmbed{
            Title:       "Payout Commit Successful",
            Description: fmt.Sprintf("Successfully committed **%d** payables for **%s** batch **%d**", result.Updated, result.Month, result.Batch),
            Color:       ColorGreen,
        }
    } else {
        // Partial failure
        embed = &discordgo.MessageEmbed{
            Title:       "Payout Commit Partially Successful",
            Description: fmt.Sprintf("Committed **%d** payables for **%s** batch **%d**\n**%d** failed - check logs for details", result.Updated, result.Month, result.Batch, result.Failed),
            Color:       ColorOrange,
        }

        // Add error details if available (limit to first 5)
        if len(result.Errors) > 0 {
            errorList := ""
            displayCount := len(result.Errors)
            if displayCount > 5 {
                displayCount = 5
            }

            for i := 0; i < displayCount; i++ {
                err := result.Errors[i]
                errorList += fmt.Sprintf("• `%s`: %s\n", err.PayableID, err.Error)
            }

            if len(result.Errors) > 5 {
                errorList += fmt.Sprintf("\n... and %d more errors", len(result.Errors)-5)
            }

            embed.Fields = []*discordgo.MessageEmbedField{
                {
                    Name:  "Errors",
                    Value: errorList,
                },
            }
        }
    }

    _, err := v.session.ChannelMessageSendEmbed(message.ChannelID, embed)
    return err
}
```

---

## 4. Adapter Layer

### File: `pkg/adapter/fortress/payout.go`

```go
package fortress

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/dwarvesf/fortress-discord/pkg/config"
    "github.com/dwarvesf/fortress-discord/pkg/logger"
    "github.com/dwarvesf/fortress-discord/pkg/model"
)

type payoutAdapter struct {
    cfg    *config.Config
    logger logger.Logger
    client *http.Client
}

func newPayoutAdapter(cfg *config.Config, logger logger.Logger, client *http.Client) *payoutAdapter {
    return &payoutAdapter{
        cfg:    cfg,
        logger: logger,
        client: client,
    }
}

// PreviewCommit calls the API to preview commit
func (a *payoutAdapter) PreviewCommit(month string, batch int) (*model.PayoutPreview, error) {
    url := fmt.Sprintf("%s/api/v1/contractor-payables/preview-commit?month=%s&batch=%d",
        a.cfg.FortressAPIURL, month, batch)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+a.cfg.FortressAPIKey)

    resp, err := a.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call API: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var apiResp struct {
        Data *model.PayoutPreview `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return apiResp.Data, nil
}

// ExecuteCommit calls the API to execute commit
func (a *payoutAdapter) ExecuteCommit(month string, batch int) (*model.PayoutCommitResult, error) {
    url := fmt.Sprintf("%s/api/v1/contractor-payables/commit", a.cfg.FortressAPIURL)

    body := map[string]interface{}{
        "month": month,
        "batch": batch,
    }

    bodyBytes, err := json.Marshal(body)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+a.cfg.FortressAPIKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := a.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call API: %w", err)
    }
    defer resp.Body.Close()

    // Accept both 200 OK and 207 Multi-Status
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusMultiStatus {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var apiResp struct {
        Data *model.PayoutCommitResult `json:"data"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return apiResp.Data, nil
}
```

Update `pkg/adapter/fortress/fortress.go` to include payout adapter:

```go
type IAdapter interface {
    // ... existing methods
    Payout() IPayoutAdapter
}

type IPayoutAdapter interface {
    PreviewCommit(month string, batch int) (*model.PayoutPreview, error)
    ExecuteCommit(month string, batch int) (*model.PayoutCommitResult, error)
}

type adapter struct {
    // ... existing fields
    payout *payoutAdapter
}

func (a *adapter) Payout() IPayoutAdapter {
    return a.payout
}
```

---

## 5. Model Definitions

### File: `pkg/model/payout.go`

```go
package model

// PayoutPreview contains preview data before committing
type PayoutPreview struct {
    Month       string              `json:"month"`
    Batch       int                 `json:"batch"`
    Count       int                 `json:"count"`
    TotalAmount float64             `json:"total_amount"`
    Contractors []ContractorPreview `json:"contractors"`
}

// ContractorPreview contains preview data for a single contractor
type ContractorPreview struct {
    Name      string  `json:"name"`
    Amount    float64 `json:"amount"`
    Currency  string  `json:"currency"`
    PayableID string  `json:"payable_id"`
}

// PayoutCommitResult contains the result of a commit operation
type PayoutCommitResult struct {
    Month   string        `json:"month"`
    Batch   int           `json:"batch"`
    Updated int           `json:"updated"`
    Failed  int           `json:"failed"`
    Errors  []CommitError `json:"errors,omitempty"`
}

// CommitError contains error details for failed updates
type CommitError struct {
    PayableID string `json:"payable_id"`
    Error     string `json:"error"`
}
```

---

## 6. Command Registration

### File: `pkg/discord/command/command.go` (MODIFY)

Add payout command to the list of commands:

```go
import (
    // ... existing imports
    payoutCmd "github.com/dwarvesf/fortress-discord/pkg/discord/command/payout"
)

func NewCommands(
    cfg *config.Config,
    logger logger.Logger,
    svc service.IService,
    view view.IView,
) []ICommand {
    return []ICommand{
        // ... existing commands
        payoutCmd.New(cfg, logger, svc, view),
    }
}
```

---

## 7. Button Interaction Handler

The confirmation buttons need to be handled by the Discord interaction handler. This is typically in the main bot event handler or a dedicated interaction handler.

### Example Integration (in main bot handler):

```go
// Handle button interactions
func handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if i.Type != discordgo.InteractionMessageComponent {
        return
    }

    customID := i.MessageComponentData().CustomID

    // Handle payout commit confirmation
    if strings.HasPrefix(customID, "payout_commit_confirm:") {
        parts := strings.Split(customID, ":")
        if len(parts) != 3 {
            return
        }

        month := parts[1]
        batch, err := strconv.Atoi(parts[2])
        if err != nil {
            return
        }

        // Acknowledge the interaction
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseDeferredMessageUpdate,
        })

        // Execute the commit
        message := &model.DiscordMessage{
            ChannelID: i.ChannelID,
            Author:    i.Member.User,
        }

        cmd := payoutCmd.New(cfg, logger, svc, view)
        if err := cmd.ExecuteCommitConfirmation(message, month, batch); err != nil {
            // Handle error
        }

        return
    }

    // Handle payout commit cancel
    if customID == "payout_commit_cancel" {
        // Acknowledge and delete the message
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseUpdateMessage,
            Data: &discordgo.InteractionResponseData{
                Content:    "Payout commit cancelled.",
                Embeds:     []*discordgo.MessageEmbed{},
                Components: []discordgo.MessageComponent{},
            },
        })
        return
    }
}
```

---

## Testing Considerations

### Unit Tests

1. **Command Tests**:
   - Test argument parsing
   - Test validation (month format, batch values)
   - Test permission checking
   - Mock service layer

2. **Service Tests**:
   - Mock adapter layer
   - Test error handling
   - Test response parsing

3. **View Tests**:
   - Test embed formatting
   - Test button creation
   - Test message sending (mock Discord session)

### Integration Tests

1. Test full command flow with test API
2. Test button interactions
3. Test error scenarios
4. Test empty results

### Manual Testing Checklist

- [ ] `?payout help` shows help message
- [ ] `?payout commit` without args shows error
- [ ] `?payout commit invalid-month 15` shows format error
- [ ] `?payout commit 2025-01 99` shows batch error
- [ ] `?payout commit 2025-01 15` with no payables shows info message
- [ ] `?payout commit 2025-01 15` with payables shows confirmation
- [ ] Click Cancel button dismisses confirmation
- [ ] Click Confirm button executes commit and shows result
- [ ] Partial failure shows error details
- [ ] Permission check works (non-admin can't run)

---

## Configuration

Add to `pkg/config/config.go`:

```go
type Config struct {
    // ... existing fields
    FortressAPIURL string `env:"FORTRESS_API_URL" envDefault:"https://api.example.com"`
    FortressAPIKey string `env:"FORTRESS_API_KEY" envRequired:"true"`
}
```

---

## Error Handling

**User-Facing Errors**:
- Invalid month format: "Invalid month format. Use YYYY-MM (e.g., 2025-01)"
- Invalid batch: "Batch must be 1 or 15"
- No payables: "No pending payables found for 2025-01 batch 15"
- API error: "Failed to preview commit: <error message>"

**Logging**:
- Log all user commands
- Log API calls and responses
- Log button interactions
- Log errors with context

---

## Security Considerations

1. **Permission Check**: Enforce admin/ops role requirement
2. **Input Validation**: Validate all user inputs before API calls
3. **API Authentication**: Use Bearer token for API calls
4. **Button Security**: Include month/batch in button CustomID to prevent tampering

---

## Performance Considerations

1. **Embed Size Limits**: Limit contractor list to 10 entries in preview
2. **API Timeout**: Add timeout to HTTP client (e.g., 30 seconds)
3. **Button Expiration**: Discord buttons expire after 15 minutes (document this)

---

## User Experience

**Happy Path**:
```
User: ?payout commit 2025-01 15

Bot: [Embed]
     Confirm Payout Commit
     You are about to commit payables for 2025-01 batch 15

     Count: 3 payables
     Total Amount: $15,000.00

     Contractors:
     • John Doe - 5000.00 USD
     • Jane Smith - 7500.00 USD
     • Bob Wilson - 2500.00 USD

     [Confirm] [Cancel]

User: [clicks Confirm]

Bot: [Embed]
     Payout Commit Successful
     Successfully committed 3 payables for 2025-01 batch 15
```

**No Payables Path**:
```
User: ?payout commit 2025-01 1

Bot: [Embed]
     No Pending Payables
     No pending payables found for 2025-01 batch 1
```

**Partial Failure Path**:
```
User: [clicks Confirm]

Bot: [Embed]
     Payout Commit Partially Successful
     Committed 2 payables for 2025-01 batch 15
     1 failed - check logs for details

     Errors:
     • page-id-3: failed to update payout: network timeout
```

---

## References

- Existing Command Pattern: `pkg/discord/command/invoice/command.go`
- Button Interaction: Discord.js/discordgo documentation
- API Spec: `/docs/sessions/202601081935-payout-commit-command/planning/specifications/01-api-endpoints.md`
- ADR: `/docs/sessions/202601081935-payout-commit-command/planning/ADRs/ADR-001-cascade-status-update.md`
