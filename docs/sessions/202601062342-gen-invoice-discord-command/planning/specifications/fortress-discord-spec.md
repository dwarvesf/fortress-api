# Specification: fortress-discord Changes for `?gen invoice` Command

## Overview
Implement a new Discord command `?gen invoice` (or `?gen inv`) that allows contractors to generate their own invoices via async webhook to fortress-api.

## Architecture Layers

### 1. Model Layer

#### Location
`/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/model/gen_invoice.go`

#### Structures

```go
package model

import "time"

// GenInvoiceRequest represents the user's command request
type GenInvoiceRequest struct {
    DiscordUsername string    // User who issued command
    Month           string    // YYYY-MM format (optional, defaults to current)
    DMChannelID     string    // DM channel for response
    DMMessageID     string    // Message to update with result
}

// GenInvoiceResponse represents the webhook response from fortress-api
type GenInvoiceResponse struct {
    Success       bool      `json:"success"`
    Message       string    `json:"message"`
    FileURL       string    `json:"file_url,omitempty"`
    Email         string    `json:"email,omitempty"`
    Month         string    `json:"month"`
    Error         string    `json:"error,omitempty"`
}
```

### 2. Adapter Layer

#### Location
`/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/gen_invoice.go`

#### Interface

```go
package fortress

import (
    "context"
    "github.com/dwarvesf/fortress-discord/pkg/discord/model"
)

type GenInvoiceAdapter interface {
    // GenerateInvoice posts webhook to fortress-api
    GenerateInvoice(ctx context.Context, req *model.GenInvoiceRequest) error
}
```

#### Implementation

```go
type genInvoiceAdapter struct {
    client     *http.Client
    webhookURL string // from config: FORTRESS_API_WEBHOOK_URL
}

func (a *genInvoiceAdapter) GenerateInvoice(ctx context.Context, req *model.GenInvoiceRequest) error {
    // POST to fortress-api webhook endpoint
    payload := map[string]string{
        "discord_username": req.DiscordUsername,
        "month":           req.Month,
        "dm_channel_id":   req.DMChannelID,
        "dm_message_id":   req.DMMessageID,
    }

    body, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal payload: %w", err)
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", a.webhookURL, bytes.NewBuffer(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    httpReq.Header.Set("Content-Type", "application/json")

    resp, err := a.client.Do(httpReq)
    if err != nil {
        return fmt.Errorf("failed to post webhook: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("webhook returned error %d: %s", resp.StatusCode, string(body))
    }

    return nil
}
```

#### Configuration
Add to config:
```yaml
fortress_api:
  webhook_url: "https://fortress-api.example.com/webhooks/discord/gen-invoice"
```

### 3. Service Layer

#### Location
`/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/gen_invoice.go`

#### Interface

```go
package service

import (
    "context"
    "github.com/dwarvesf/fortress-discord/pkg/discord/model"
)

type GenInvoiceService interface {
    // GenerateInvoice handles the command flow
    GenerateInvoice(ctx context.Context, username, month string) error
}
```

#### Implementation

```go
type genInvoiceService struct {
    session        *discordgo.Session
    fortressClient fortress.GenInvoiceAdapter
    view           view.GenInvoiceView
}

func (s *genInvoiceService) GenerateInvoice(ctx context.Context, username, month string) error {
    // 1. Validate month format
    if month != "" {
        if !isValidMonthFormat(month) {
            return fmt.Errorf("invalid month format, use YYYY-MM")
        }
    } else {
        // Default to current month
        month = time.Now().Format("2006-01")
    }

    // 2. Create DM channel with user
    user, err := s.getUserByUsername(username)
    if err != nil {
        return fmt.Errorf("failed to find user: %w", err)
    }

    dmChannel, err := s.session.UserChannelCreate(user.ID)
    if err != nil {
        return fmt.Errorf("failed to create DM channel (user may have DMs disabled): %w", err)
    }

    // 3. Send processing embed to DM
    processingEmbed := s.view.ProcessingEmbed(month)
    msg, err := s.session.ChannelMessageSendComplex(dmChannel.ID, &discordgo.MessageSend{
        Embeds: []*discordgo.MessageEmbed{processingEmbed},
    })
    if err != nil {
        return fmt.Errorf("failed to send DM: %w", err)
    }

    // 4. POST webhook to fortress-api
    req := &model.GenInvoiceRequest{
        DiscordUsername: username,
        Month:           month,
        DMChannelID:     dmChannel.ID,
        DMMessageID:     msg.ID,
    }

    err = s.fortressClient.GenerateInvoice(ctx, req)
    if err != nil {
        // Update DM with error before returning
        errorEmbed := s.view.ErrorEmbed("Failed to submit request", err.Error())
        s.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
            Channel: dmChannel.ID,
            ID:      msg.ID,
            Embeds:  []*discordgo.MessageEmbed{errorEmbed},
        })
        return err
    }

    return nil
}

func (s *genInvoiceService) getUserByUsername(username string) (*discordgo.User, error) {
    // Implementation to find user by username
    // May need to search guild members
    // For now, assume we can get from context or cache
    return nil, nil // TODO: implement
}

func isValidMonthFormat(month string) bool {
    // Validate YYYY-MM format
    _, err := time.Parse("2006-01", month)
    return err == nil
}
```

### 4. View Layer

#### Location
`/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/gen_invoice.go`

#### Interface

```go
package view

import "github.com/bwmarrin/discordgo"

type GenInvoiceView interface {
    ProcessingEmbed(month string) *discordgo.MessageEmbed
    SuccessEmbed(month, fileURL, email string) *discordgo.MessageEmbed
    ErrorEmbed(title, message string) *discordgo.MessageEmbed
}
```

#### Implementation

```go
type genInvoiceView struct{}

func (v *genInvoiceView) ProcessingEmbed(month string) *discordgo.MessageEmbed {
    return &discordgo.MessageEmbed{
        Title:       "Generating Invoice",
        Description: fmt.Sprintf("Processing your invoice for %s...", formatMonthDisplay(month)),
        Color:       0x3498db, // Blue
        Footer: &discordgo.MessageEmbedFooter{
            Text: "This may take a few moments",
        },
        Timestamp: time.Now().Format(time.RFC3339),
    }
}

func (v *genInvoiceView) SuccessEmbed(month, fileURL, email string) *discordgo.MessageEmbed {
    return &discordgo.MessageEmbed{
        Title:       "Invoice Generated Successfully",
        Description: fmt.Sprintf("Your invoice for %s has been generated and shared to your email.", formatMonthDisplay(month)),
        Color:       0x2ecc71, // Green
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Month",
                Value:  formatMonthDisplay(month),
                Inline: true,
            },
            {
                Name:   "Email",
                Value:  email,
                Inline: true,
            },
            {
                Name:   "File",
                Value:  fmt.Sprintf("[View in Google Drive](%s)", fileURL),
                Inline: false,
            },
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: "Check your email for file access notification",
        },
        Timestamp: time.Now().Format(time.RFC3339),
    }
}

func (v *genInvoiceView) ErrorEmbed(title, message string) *discordgo.MessageEmbed {
    return &discordgo.MessageEmbed{
        Title:       title,
        Description: message,
        Color:       0xe74c3c, // Red
        Footer: &discordgo.MessageEmbedFooter{
            Text: "Contact support if this persists",
        },
        Timestamp: time.Now().Format(time.RFC3339),
    }
}

func formatMonthDisplay(month string) string {
    // Convert "2025-01" to "January 2025"
    t, err := time.Parse("2006-01", month)
    if err != nil {
        return month
    }
    return t.Format("January 2006")
}
```

### 5. Command Layer

#### Location
`/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/gen/gen.go`

#### Command Structure
Following the pattern of `base.TextCommander`:

```go
package gen

import (
    "github.com/bwmarrin/discordgo"
    "github.com/dwarvesf/fortress-discord/pkg/discord/base"
    "github.com/dwarvesf/fortress-discord/pkg/discord/service"
)

type GenCommand struct {
    service service.GenInvoiceService
}

func NewGenCommand(service service.GenInvoiceService) *GenCommand {
    return &GenCommand{
        service: service,
    }
}

// Command returns the command name
func (c *GenCommand) Command() string {
    return "gen"
}

// Help returns help text
func (c *GenCommand) Help() string {
    return "Generate invoices and other documents"
}

// Usage returns usage information
func (c *GenCommand) Usage() string {
    return `
Usage: ?gen <subcommand> [options]

Subcommands:
  invoice, inv <YYYY-MM>  Generate invoice for specified month (defaults to current month)

Examples:
  ?gen invoice          Generate invoice for current month
  ?gen inv              Short form
  ?gen invoice 2025-01  Generate invoice for January 2025
`
}

// Execute handles the command
func (c *GenCommand) Execute(ctx *base.CommandContext) error {
    args := ctx.Args

    // Require at least one argument (subcommand)
    if len(args) == 0 {
        return ctx.Reply("Missing subcommand. Use `?help gen` for usage.")
    }

    subcommand := args[0]

    switch subcommand {
    case "invoice", "inv":
        return c.handleInvoice(ctx, args[1:])
    default:
        return ctx.Reply(fmt.Sprintf("Unknown subcommand: %s", subcommand))
    }
}

func (c *GenCommand) handleInvoice(ctx *base.CommandContext, args []string) error {
    username := ctx.Author.Username

    // Optional month argument
    var month string
    if len(args) > 0 {
        month = args[0]
    }

    // Call service (will DM user with processing message and POST webhook)
    err := c.service.GenerateInvoice(ctx.Context, username, month)
    if err != nil {
        // Only reply in channel if DM failed
        if strings.Contains(err.Error(), "DM channel") {
            return ctx.Reply("Failed to send you a DM. Please enable DMs from server members.")
        }
        return ctx.Reply(fmt.Sprintf("Failed to generate invoice: %s", err.Error()))
    }

    // Success - user will get DM
    return ctx.Reply("Check your DMs for invoice generation status.")
}
```

#### Command Registration
Add to command registry in main bot initialization:

```go
// pkg/discord/bot.go or wherever commands are registered

genInvoiceService := service.NewGenInvoiceService(
    session,
    fortressAdapter,
    view.NewGenInvoiceView(),
)

genCommand := gen.NewGenCommand(genInvoiceService)
commandRegistry.Register(genCommand)
```

## Testing Strategy

### Unit Tests

#### Service Layer Tests
`pkg/discord/service/gen_invoice_test.go`

```go
func TestGenerateInvoice_ValidMonth(t *testing.T) {
    // Test successful invoice generation with valid month
}

func TestGenerateInvoice_InvalidMonth(t *testing.T) {
    // Test error handling for invalid month format
}

func TestGenerateInvoice_DefaultMonth(t *testing.T) {
    // Test that empty month defaults to current month
}

func TestGenerateInvoice_DMFailed(t *testing.T) {
    // Test error handling when DM creation fails
}

func TestGenerateInvoice_WebhookFailed(t *testing.T) {
    // Test error handling when webhook POST fails
}
```

#### View Layer Tests
`pkg/discord/view/gen_invoice_test.go`

```go
func TestProcessingEmbed(t *testing.T) {
    // Test embed formatting and color
}

func TestSuccessEmbed(t *testing.T) {
    // Test success embed with all fields
}

func TestErrorEmbed(t *testing.T) {
    // Test error embed formatting
}

func TestFormatMonthDisplay(t *testing.T) {
    // Test month format conversion
}
```

#### Command Layer Tests
`pkg/discord/command/gen/gen_test.go`

```go
func TestGenCommand_Execute_Invoice(t *testing.T) {
    // Test invoice subcommand execution
}

func TestGenCommand_Execute_UnknownSubcommand(t *testing.T) {
    // Test error handling for unknown subcommands
}

func TestGenCommand_Execute_NoArgs(t *testing.T) {
    // Test error handling when no subcommand provided
}
```

## Error Handling

### User-Facing Errors

#### DM Disabled
```
"Failed to send you a DM. Please enable DMs from server members."
```

#### Invalid Month Format
```
"Invalid month format. Please use YYYY-MM (e.g., 2025-01)"
```

#### Webhook Failed
```
"Failed to submit request. Please try again later."
```

#### Rate Limited (from API)
```
"You have generated 3 invoices today. Limit resets at midnight UTC."
```

#### Not a Contractor (from API)
```
"You are not registered as an active contractor. Contact HR if this is incorrect."
```

### Logging
Log all errors with context:
```go
log.WithFields(log.Fields{
    "username": username,
    "month":    month,
    "error":    err,
}).Error("Failed to generate invoice")
```

## Configuration Requirements

Add to config file (e.g., `config.yaml`):

```yaml
fortress_api:
  webhook_url: "https://fortress-api.example.com/webhooks/discord/gen-invoice"
  timeout: 5s

discord:
  dm_timeout: 10s
```

## Dependencies

### New Dependencies (if not already present)
None - all required dependencies should already be in fortress-discord:
- `github.com/bwmarrin/discordgo` (Discord API)
- Standard library `net/http` (HTTP client)
- Standard library `encoding/json` (JSON marshaling)
- Standard library `time` (Date handling)

### Internal Dependencies
- `pkg/discord/base` (Command interface)
- `pkg/discord/model` (Data structures)
- `pkg/discord/service` (Service layer)
- `pkg/discord/view` (View layer)
- `pkg/adapter/fortress` (API adapter)

## Files to Create

1. `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/model/gen_invoice.go`
2. `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/adapter/fortress/gen_invoice.go`
3. `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/service/gen_invoice.go`
4. `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/view/gen_invoice.go`
5. `/Users/quang/workspace/dwarvesf/fortress-discord/pkg/discord/command/gen/gen.go`

## Files to Modify

1. `/Users/quang/workspace/dwarvesf/fortress-discord/config.yaml` (or equivalent config file)
   - Add fortress_api webhook configuration

2. Main bot initialization file (e.g., `cmd/bot/main.go` or `pkg/discord/bot.go`)
   - Register new GenCommand
   - Initialize GenInvoiceService with dependencies

## Deployment Considerations

### Environment Variables
```bash
FORTRESS_API_WEBHOOK_URL=https://fortress-api.example.com/webhooks/discord/gen-invoice
```

### Rollout Strategy
1. Deploy fortress-api webhook endpoint first
2. Test webhook endpoint manually
3. Deploy fortress-discord with new command
4. Test end-to-end flow in staging
5. Deploy to production
6. Monitor logs for errors

### Rollback Plan
If issues arise:
1. Remove command registration from bot
2. Redeploy without GenCommand
3. Users will get "Unknown command" error
4. No database cleanup needed (no persistence)
