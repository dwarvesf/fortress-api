# Configuration Specification

## Overview

This document specifies the configuration requirements for the Contractor Invoice Email Listener feature. All configuration follows the existing patterns in `pkg/config/config.go` using environment variables.

## Configuration Structure

### File: `pkg/config/config.go`

### New Config Struct

Add `InvoiceEmailListener` struct to main `Config`:

```go
type Config struct {
    // ... existing fields
    InvoiceEmailListener InvoiceEmailListener
}

type InvoiceEmailListener struct {
    Enabled        bool          // Enable/disable the feature
    EmailAddress   string        // Email address to monitor (e.g., "bill@d.foundation")
    RefreshToken   string        // Google OAuth refresh token for inbox access
    PollInterval   time.Duration // How often to check inbox (e.g., 5 minutes)
    ProcessedLabel string        // Gmail label for processed emails
    MaxMessages    int64         // Max messages to process per run
    PDFMaxSizeMB   int           // Max PDF size to download (MB)
}
```

## Environment Variables

### Required Variables

| Variable Name | Type | Description | Example Value |
|---------------|------|-------------|---------------|
| `INVOICE_LISTENER_ENABLED` | bool | Enable invoice email listener | `true` |
| `INVOICE_LISTENER_EMAIL` | string | Email address to monitor | `bill@d.foundation` |
| `INVOICE_LISTENER_REFRESH_TOKEN` | string | Google OAuth refresh token | `1//abc123...` |

### Optional Variables (with defaults)

| Variable Name | Type | Default | Description |
|---------------|------|---------|-------------|
| `INVOICE_LISTENER_POLL_INTERVAL` | duration | `5m` | Poll interval (e.g., `5m`, `10m`, `1h`) |
| `INVOICE_LISTENER_LABEL` | string | `fortress-api/processed` | Gmail label for processed emails |
| `INVOICE_LISTENER_MAX_MESSAGES` | int | `50` | Max messages per polling run |
| `INVOICE_LISTENER_PDF_MAX_SIZE_MB` | int | `5` | Max PDF download size in MB |

## Configuration Loading

### File: `pkg/config/config.go`

Update `Generate` function:

```go
func Generate(v ENV) *Config {
    // ... existing code

    // Parse poll interval with default
    pollInterval := getStringWithDefault(v, "INVOICE_LISTENER_POLL_INTERVAL", "5m")
    pollDuration, err := time.ParseDuration(pollInterval)
    if err != nil {
        // Invalid format, use default
        pollDuration = 5 * time.Minute
    }

    return &Config{
        // ... existing fields

        InvoiceEmailListener: InvoiceEmailListener{
            Enabled:        getBoolWithDefault(v, "INVOICE_LISTENER_ENABLED", false),
            EmailAddress:   v.GetString("INVOICE_LISTENER_EMAIL"),
            RefreshToken:   v.GetString("INVOICE_LISTENER_REFRESH_TOKEN"),
            PollInterval:   pollDuration,
            ProcessedLabel: getStringWithDefault(v, "INVOICE_LISTENER_LABEL", "fortress-api/processed"),
            MaxMessages:    int64(getIntWithDefault(v, "INVOICE_LISTENER_MAX_MESSAGES", 50)),
            PDFMaxSizeMB:   getIntWithDefault(v, "INVOICE_LISTENER_PDF_MAX_SIZE_MB", 5),
        },
    }
}
```

## Configuration Validation

### Validation Function

Add validation at service initialization:

```go
// pkg/service/invoiceemail/processor.go

func NewProcessorService(
    cfg *config.Config,
    logger logger.Logger,
    gmailService googlemail.IService,
    notionService *notion.ContractorPayablesService,
) (IService, error) {
    // Validate configuration
    if err := validateConfig(cfg); err != nil {
        logger.Error(err, "[ERROR] invoice_email: invalid configuration")
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }

    return &ProcessorService{
        cfg:           cfg,
        logger:        logger,
        gmailService:  gmailService,
        notionService: notionService,
    }, nil
}

func validateConfig(cfg *config.Config) error {
    if !cfg.InvoiceEmailListener.Enabled {
        // Feature disabled, skip validation
        return nil
    }

    if cfg.InvoiceEmailListener.EmailAddress == "" {
        return errors.New("INVOICE_LISTENER_EMAIL is required when feature is enabled")
    }

    if cfg.InvoiceEmailListener.RefreshToken == "" {
        return errors.New("INVOICE_LISTENER_REFRESH_TOKEN is required when feature is enabled")
    }

    if cfg.InvoiceEmailListener.PollInterval < time.Minute {
        return errors.New("INVOICE_LISTENER_POLL_INTERVAL must be at least 1 minute")
    }

    if cfg.InvoiceEmailListener.MaxMessages < 1 || cfg.InvoiceEmailListener.MaxMessages > 500 {
        return errors.New("INVOICE_LISTENER_MAX_MESSAGES must be between 1 and 500")
    }

    if cfg.InvoiceEmailListener.PDFMaxSizeMB < 1 || cfg.InvoiceEmailListener.PDFMaxSizeMB > 50 {
        return errors.New("INVOICE_LISTENER_PDF_MAX_SIZE_MB must be between 1 and 50")
    }

    return nil
}
```

## Environment File Examples

### Development (.env.local)

```bash
# Invoice Email Listener
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=bill@d.foundation
INVOICE_LISTENER_REFRESH_TOKEN=1//0abcdefg...  # From OAuth flow
INVOICE_LISTENER_POLL_INTERVAL=5m
INVOICE_LISTENER_LABEL=fortress-api/processed
INVOICE_LISTENER_MAX_MESSAGES=50
INVOICE_LISTENER_PDF_MAX_SIZE_MB=5
```

### Production (.env.prod)

```bash
# Invoice Email Listener
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=bill@d.foundation
INVOICE_LISTENER_REFRESH_TOKEN=${VAULT_INVOICE_LISTENER_TOKEN}  # From Vault
INVOICE_LISTENER_POLL_INTERVAL=5m
INVOICE_LISTENER_LABEL=fortress-api/processed
INVOICE_LISTENER_MAX_MESSAGES=100
INVOICE_LISTENER_PDF_MAX_SIZE_MB=10
```

### Disabled (Default)

```bash
# Invoice Email Listener - Disabled
INVOICE_LISTENER_ENABLED=false
```

## OAuth Refresh Token Setup

### Obtaining Refresh Token

The refresh token must be obtained through the Google OAuth2 flow with appropriate scopes.

**Required Scopes**:
- `https://www.googleapis.com/auth/gmail.readonly`
- `https://www.googleapis.com/auth/gmail.modify`
- `https://www.googleapis.com/auth/gmail.labels`

**Setup Steps**:

1. **Use Existing OAuth Credentials**:
   - Client ID: `GOOGLE_API_CLIENT_ID` (already configured)
   - Client Secret: `GOOGLE_API_CLIENT_SECRET` (already configured)

2. **Generate Refresh Token**:
   ```bash
   # Use existing fortress-api OAuth flow or create standalone script
   # Example: pkg/scripts/get_gmail_token.go
   go run pkg/scripts/get_gmail_token.go
   ```

3. **Store Refresh Token**:
   - Development: Store in `.env` file
   - Production: Store in Vault or secret manager
   - Set `INVOICE_LISTENER_REFRESH_TOKEN` environment variable

### Token Rotation

- Refresh tokens don't expire unless revoked
- Access tokens auto-refresh via `oauth2.Config.TokenSource`
- Monitor for auth errors and alert if token becomes invalid

## Configuration by Environment

### Local Development

**Purpose**: Test with personal Gmail account

```bash
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=test-billing@yourdomain.com
INVOICE_LISTENER_REFRESH_TOKEN=<your-test-token>
INVOICE_LISTENER_POLL_INTERVAL=1m  # Faster for testing
INVOICE_LISTENER_MAX_MESSAGES=10   # Smaller batch for testing
```

### Staging/Dev Environment

**Purpose**: Test with staging Gmail account

```bash
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=bill-staging@d.foundation
INVOICE_LISTENER_REFRESH_TOKEN=${STAGING_GMAIL_TOKEN}
INVOICE_LISTENER_POLL_INTERVAL=5m
INVOICE_LISTENER_MAX_MESSAGES=50
```

### Production Environment

**Purpose**: Production billing operations

```bash
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=bill@d.foundation
INVOICE_LISTENER_REFRESH_TOKEN=${PRODUCTION_GMAIL_TOKEN}
INVOICE_LISTENER_POLL_INTERVAL=5m
INVOICE_LISTENER_MAX_MESSAGES=100
INVOICE_LISTENER_PDF_MAX_SIZE_MB=10
```

## Feature Toggle

### Graceful Degradation

If `INVOICE_LISTENER_ENABLED=false`:
- Service initialization skipped
- Cron job not started
- No API calls made
- Log message: `[INFO] invoice_email: listener disabled, skipping initialization`

### Runtime Check

```go
// pkg/worker/worker.go
func (w *Worker) handleInvoiceEmailCheck(l logger.Logger, payload interface{}) error {
    if !w.service.Config.InvoiceEmailListener.Enabled {
        l.Info("[INFO] invoice_email: listener disabled, skipping check")
        return nil
    }

    // ... proceed with processing
}
```

## Monitoring Configuration

### Logging Configuration

Inherit from existing `LOG_LEVEL` environment variable:

- `LOG_LEVEL=debug` → Log all invoice email processing steps
- `LOG_LEVEL=info` → Log only summary and errors
- `LOG_LEVEL=error` → Log only errors

### Metrics Configuration (Future)

Placeholder for potential metrics:

```bash
# Optional: Future metrics endpoint
INVOICE_LISTENER_METRICS_ENABLED=false
INVOICE_LISTENER_METRICS_INTERVAL=1m
```

## Configuration Documentation

### README Update

Add to `README.md`:

```markdown
### Invoice Email Listener

Monitor Gmail inbox for contractor invoice submissions and auto-update Notion payables.

**Environment Variables**:
- `INVOICE_LISTENER_ENABLED` - Enable feature (default: false)
- `INVOICE_LISTENER_EMAIL` - Email address to monitor
- `INVOICE_LISTENER_REFRESH_TOKEN` - Google OAuth refresh token
- `INVOICE_LISTENER_POLL_INTERVAL` - Poll interval (default: 5m)
- `INVOICE_LISTENER_LABEL` - Gmail label for processed emails (default: fortress-api/processed)

**Setup**:
1. Obtain Google OAuth refresh token with Gmail scopes
2. Set environment variables
3. Enable feature: `INVOICE_LISTENER_ENABLED=true`
4. Restart application or trigger cron job
```

## Security Considerations

### Secrets Management

1. **Refresh Token Storage**:
   - **Never commit** refresh token to repository
   - Use `.env` file (gitignored) for local development
   - Use Vault or AWS Secrets Manager for production
   - Rotate tokens periodically (manual process)

2. **Access Control**:
   - Refresh token grants access to entire Gmail account
   - Use dedicated Gmail account for billing (e.g., `bill@d.foundation`)
   - Don't reuse personal Gmail accounts

3. **Audit Logging**:
   - Log all configuration loading at startup
   - Mask sensitive values in logs (show only first 8 chars):
     ```
     [INFO] config: INVOICE_LISTENER_REFRESH_TOKEN=1//0abcd...***
     ```

### Validation on Startup

Log configuration status on application startup:

```go
// cmd/server/main.go
func main() {
    // ... existing initialization

    if cfg.InvoiceEmailListener.Enabled {
        logger.Info(fmt.Sprintf("[INFO] config: invoice email listener ENABLED email=%s interval=%s",
            cfg.InvoiceEmailListener.EmailAddress,
            cfg.InvoiceEmailListener.PollInterval))
    } else {
        logger.Info("[INFO] config: invoice email listener DISABLED")
    }
}
```

## Testing Configuration

### Test Configuration

For unit tests, use `LoadTestConfig()` pattern:

```go
// pkg/config/config_test.go
func LoadTestInvoiceListenerConfig() InvoiceEmailListener {
    return InvoiceEmailListener{
        Enabled:        true,
        EmailAddress:   "test@example.com",
        RefreshToken:   "test-refresh-token",
        PollInterval:   1 * time.Minute,
        ProcessedLabel: "fortress-api-test/processed",
        MaxMessages:    10,
        PDFMaxSizeMB:   5,
    }
}
```

### Integration Test Configuration

For integration tests with real Gmail:

```bash
# .env.test
INVOICE_LISTENER_ENABLED=true
INVOICE_LISTENER_EMAIL=fortress-test@yourdomain.com
INVOICE_LISTENER_REFRESH_TOKEN=${TEST_GMAIL_TOKEN}
INVOICE_LISTENER_POLL_INTERVAL=30s
INVOICE_LISTENER_LABEL=fortress-api-test/processed
INVOICE_LISTENER_MAX_MESSAGES=5
```

## Migration Path

### Phase 1: Add Configuration (No Behavior Change)

- Add config struct and environment variables
- Load configuration in `Generate()`
- Default `Enabled=false` (no runtime changes)
- Deploy and verify logs show "listener DISABLED"

### Phase 2: Enable in Development

- Set `INVOICE_LISTENER_ENABLED=true` in dev environment
- Test end-to-end flow
- Monitor logs for errors

### Phase 3: Enable in Production

- Set `INVOICE_LISTENER_ENABLED=true` in production
- Monitor logs and metrics
- Adjust `PollInterval` and `MaxMessages` based on load

## Rollback Plan

### Quick Disable

If issues arise in production:

```bash
# Set environment variable and restart
INVOICE_LISTENER_ENABLED=false
```

### Gradual Rollback

1. Disable feature: `INVOICE_LISTENER_ENABLED=false`
2. Verify no processing occurs (check logs)
3. Keep configuration in place for future re-enable
4. No code deployment needed

## References

- ADR: `/docs/sessions/202601100842-contractor-invoice-email-listener/planning/ADRs/001-email-listener-architecture.md`
- Existing Config: `/pkg/config/config.go`
- Environment Variables: `/docs/environment-variables.md` (create if not exists)
- OAuth Setup Guide: `/docs/oauth-setup.md` (create if not exists)
