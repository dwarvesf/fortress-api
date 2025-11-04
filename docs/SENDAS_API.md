# Gmail SendAs Feature API Documentation

## Overview

The SendAs feature allows you to send emails from different email addresses (aliases) using Gmail API. This is essential for sending from group emails like `support@company.com` while authenticating with your personal account.

## Required OAuth Scopes

```go
scopes := []string{
    "https://www.googleapis.com/auth/gmail.send",
    "https://www.googleapis.com/auth/gmail.settings.basic",
    "https://www.googleapis.com/auth/gmail.settings.sharing",
}
```

## API Endpoints

### 1. List SendAs Aliases

**Endpoint:** `GET /gmail/v1/users/{userId}/settings/sendAs`

**Go Implementation:**
```go
func listSendAsAliases(srv *gmail.Service, userId string) ([]*gmail.SendAs, error) {
    result, err := srv.Users.Settings.SendAs.List(userId).Do()
    if err != nil {
        return nil, err
    }
    return result.SendAs, nil
}
```

**Response:**
```json
{
  "sendAs": [
    {
      "sendAsEmail": "user@company.com",
      "displayName": "John Doe",
      "isPrimary": true,
      "isDefault": true,
      "verificationStatus": "accepted"
    },
    {
      "sendAsEmail": "support@company.com",
      "displayName": "Support Team",
      "isPrimary": false,
      "isDefault": false,
      "verificationStatus": "accepted",
      "treatAsAlias": true
    }
  ]
}
```

### 2. Create SendAs Alias

**Endpoint:** `POST /gmail/v1/users/{userId}/settings/sendAs`

**Go Implementation:**
```go
func createSendAsAlias(srv *gmail.Service, userId string, email string, displayName string) (*gmail.SendAs, error) {
    sendAs := &gmail.SendAs{
        SendAsEmail:    email,
        DisplayName:    displayName,
        ReplyToAddress: email,
        TreatAsAlias:   true,
    }

    created, err := srv.Users.Settings.SendAs.Create(userId, sendAs).Do()
    if err != nil {
        return nil, err
    }

    return created, nil
}
```

**Request Body:**
```json
{
  "sendAsEmail": "support@company.com",
  "displayName": "Support Team",
  "replyToAddress": "support@company.com",
  "treatAsAlias": true
}
```

**Response:**
```json
{
  "sendAsEmail": "support@company.com",
  "displayName": "Support Team",
  "verificationStatus": "pending",
  "isPrimary": false,
  "isDefault": false
}
```

### 3. Verify SendAs Alias

**Endpoint:** `POST /gmail/v1/users/{userId}/settings/sendAs/{sendAsEmail}/verify`

**Go Implementation:**
```go
func verifySendAsAlias(srv *gmail.Service, userId string, email string) error {
    err := srv.Users.Settings.SendAs.Verify(userId, email).Do()
    return err
}
```

**Note:** This resends the verification email. The user must click the link in the email to complete verification.

### 4. Get SendAs Alias

**Endpoint:** `GET /gmail/v1/users/{userId}/settings/sendAs/{sendAsEmail}`

**Go Implementation:**
```go
func getSendAsAlias(srv *gmail.Service, userId string, email string) (*gmail.SendAs, error) {
    sendAs, err := srv.Users.Settings.SendAs.Get(userId, email).Do()
    if err != nil {
        return nil, err
    }
    return sendAs, nil
}
```

### 5. Update SendAs Alias

**Endpoint:** `PUT /gmail/v1/users/{userId}/settings/sendAs/{sendAsEmail}`

**Go Implementation:**
```go
func updateSendAsAlias(srv *gmail.Service, userId string, sendAs *gmail.SendAs) (*gmail.SendAs, error) {
    updated, err := srv.Users.Settings.SendAs.Update(userId, sendAs.SendAsEmail, sendAs).Do()
    if err != nil {
        return nil, err
    }
    return updated, nil
}
```

### 6. Delete SendAs Alias

**Endpoint:** `DELETE /gmail/v1/users/{userId}/settings/sendAs/{sendAsEmail}`

**Go Implementation:**
```go
func deleteSendAsAlias(srv *gmail.Service, userId string, email string) error {
    err := srv.Users.Settings.SendAs.Delete(userId, email).Do()
    return err
}
```

## SendAs Resource Fields

```go
type SendAs struct {
    SendAsEmail       string  // Email address
    DisplayName       string  // Display name
    ReplyToAddress    string  // Reply-to address
    Signature         string  // HTML signature
    IsPrimary         bool    // Is primary email (read-only)
    IsDefault         bool    // Is default sender
    TreatAsAlias      bool    // Treat as alias
    VerificationStatus string // "accepted" or "pending"
}
```

## Integration Examples

### Example 1: Check if SendAs Alias Exists

```go
func hasSendAsAlias(srv *gmail.Service, userId string, targetEmail string) (bool, error) {
    list, err := srv.Users.Settings.SendAs.List(userId).Do()
    if err != nil {
        return false, err
    }

    for _, sendAs := range list.SendAs {
        if sendAs.SendAsEmail == targetEmail {
            return true, nil
        }
    }

    return false, nil
}
```

### Example 2: Ensure SendAs Alias is Verified

```go
func ensureSendAsVerified(srv *gmail.Service, userId string, email string) error {
    sendAs, err := srv.Users.Settings.SendAs.Get(userId, email).Do()
    if err != nil {
        return fmt.Errorf("alias not found: %v", err)
    }

    if sendAs.VerificationStatus != "accepted" {
        return fmt.Errorf("alias not verified: status is %s", sendAs.VerificationStatus)
    }

    return nil
}
```

### Example 3: Send Email from Specific Alias

```go
func sendFromAlias(srv *gmail.Service, userId string, from string, to string, subject string, body string) error {
    // Verify alias is valid and verified
    err := ensureSendAsVerified(srv, userId, from)
    if err != nil {
        return err
    }

    // Construct email message
    var message strings.Builder
    message.WriteString(fmt.Sprintf("From: %s\r\n", from))
    message.WriteString(fmt.Sprintf("To: %s\r\n", to))
    message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
    message.WriteString("MIME-Version: 1.0\r\n")
    message.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
    message.WriteString("\r\n")
    message.WriteString(body)

    // Encode message
    encoded := base64.URLEncoding.EncodeToString([]byte(message.String()))

    // Send email
    gmailMessage := &gmail.Message{Raw: encoded}
    _, err = srv.Users.Messages.Send(userId, gmailMessage).Do()

    return err
}
```

### Example 4: Complete SendAs Setup Workflow

```go
func setupSendAsAlias(srv *gmail.Service, userId string, email string, displayName string) error {
    // Check if already exists
    exists, err := hasSendAsAlias(srv, userId, email)
    if err != nil {
        return err
    }

    if exists {
        // Check verification status
        sendAs, err := srv.Users.Settings.SendAs.Get(userId, email).Do()
        if err != nil {
            return err
        }

        if sendAs.VerificationStatus == "accepted" {
            return nil // Already set up
        }

        // Resend verification
        return srv.Users.Settings.SendAs.Verify(userId, email).Do()
    }

    // Create new alias
    sendAs := &gmail.SendAs{
        SendAsEmail:    email,
        DisplayName:    displayName,
        ReplyToAddress: email,
        TreatAsAlias:   true,
    }

    created, err := srv.Users.Settings.SendAs.Create(userId, sendAs).Do()
    if err != nil {
        return err
    }

    if created.VerificationStatus == "pending" {
        return fmt.Errorf("verification required: check email at %s", email)
    }

    return nil
}
```

## Verification Status

| Status | Description | Action Required |
|--------|-------------|-----------------|
| `accepted` | Alias is verified and ready to use | None |
| `pending` | Verification email sent, awaiting confirmation | User must click verification link in email |

## Common Errors

### Error: 403 Insufficient Permission

```
googleapi: Error 403: Request had insufficient authentication scopes.
Reason: ACCESS_TOKEN_SCOPE_INSUFFICIENT
```

**Solution:** Regenerate refresh token with correct scopes:
- `https://www.googleapis.com/auth/gmail.settings.basic`
- `https://www.googleapis.com/auth/gmail.settings.sharing`

### Error: Alias Not Verified

```
googleapi: Error 400: SendAs alias not verified
```

**Solution:**
1. Check verification status: `GET /settings/sendAs/{email}`
2. Resend verification: `POST /settings/sendAs/{email}/verify`
3. Click link in verification email

### Error: Alias Does Not Exist

```
googleapi: Error 404: SendAs alias not found
```

**Solution:** Create the alias first using `POST /settings/sendAs`

## Best Practices

1. **Always verify before sending:**
   ```go
   if err := ensureSendAsVerified(srv, userId, from); err != nil {
       return err
   }
   ```

2. **Cache SendAs list:**
   ```go
   // Cache for 5 minutes
   var cachedAliases []*gmail.SendAs
   var cacheTime time.Time

   if time.Since(cacheTime) > 5*time.Minute {
       cachedAliases, _ = srv.Users.Settings.SendAs.List(userId).Do()
       cacheTime = time.Now()
   }
   ```

3. **Handle verification gracefully:**
   ```go
   if created.VerificationStatus == "pending" {
       log.Printf("Verification email sent to %s", email)
       // Store pending verification in database
       // Set up periodic check or webhook
   }
   ```

4. **Use service accounts for automation:**
   - For domain-wide delegation
   - Bypasses verification for domain emails
   - Requires Workspace Admin setup

## Complete Integration Code

```go
package main

import (
    "context"
    "encoding/base64"
    "fmt"
    "strings"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
    "google.golang.org/api/option"
)

type GmailSendAsClient struct {
    service *gmail.Service
    userId  string
}

func NewGmailSendAsClient(clientID, clientSecret, refreshToken, userId string) (*GmailSendAsClient, error) {
    ctx := context.Background()

    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        Endpoint:     google.Endpoint,
        Scopes: []string{
            gmail.GmailSendScope,
            "https://www.googleapis.com/auth/gmail.settings.basic",
            "https://www.googleapis.com/auth/gmail.settings.sharing",
        },
    }

    token := &oauth2.Token{RefreshToken: refreshToken}
    srv, err := gmail.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, token)))
    if err != nil {
        return nil, err
    }

    return &GmailSendAsClient{
        service: srv,
        userId:  userId,
    }, nil
}

func (c *GmailSendAsClient) ListAliases() ([]*gmail.SendAs, error) {
    result, err := c.service.Users.Settings.SendAs.List(c.userId).Do()
    if err != nil {
        return nil, err
    }
    return result.SendAs, nil
}

func (c *GmailSendAsClient) CreateAlias(email, displayName string) (*gmail.SendAs, error) {
    sendAs := &gmail.SendAs{
        SendAsEmail:    email,
        DisplayName:    displayName,
        ReplyToAddress: email,
        TreatAsAlias:   true,
    }
    return c.service.Users.Settings.SendAs.Create(c.userId, sendAs).Do()
}

func (c *GmailSendAsClient) VerifyAlias(email string) error {
    return c.service.Users.Settings.SendAs.Verify(c.userId, email).Do()
}

func (c *GmailSendAsClient) IsAliasVerified(email string) (bool, error) {
    sendAs, err := c.service.Users.Settings.SendAs.Get(c.userId, email).Do()
    if err != nil {
        return false, err
    }
    return sendAs.VerificationStatus == "accepted", nil
}

func (c *GmailSendAsClient) SendEmail(from, to, subject, body string) error {
    verified, err := c.IsAliasVerified(from)
    if err != nil {
        return fmt.Errorf("alias check failed: %v", err)
    }
    if !verified {
        return fmt.Errorf("alias %s is not verified", from)
    }

    var message strings.Builder
    message.WriteString(fmt.Sprintf("From: %s\r\n", from))
    message.WriteString(fmt.Sprintf("To: %s\r\n", to))
    message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
    message.WriteString("MIME-Version: 1.0\r\n")
    message.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
    message.WriteString("\r\n")
    message.WriteString(body)

    encoded := base64.URLEncoding.EncodeToString([]byte(message.String()))
    gmailMessage := &gmail.Message{Raw: encoded}

    _, err = c.service.Users.Messages.Send(c.userId, gmailMessage).Do()
    return err
}
```

## Usage in Your Application

```go
// Initialize client
client, err := NewGmailSendAsClient(
    os.Getenv("GOOGLE_CLIENT_ID"),
    os.Getenv("GOOGLE_CLIENT_SECRET"),
    os.Getenv("GOOGLE_REFRESH_TOKEN"),
    os.Getenv("EMAIL_ID"),
)
if err != nil {
    log.Fatal(err)
}

// List aliases
aliases, err := client.ListAliases()
for _, alias := range aliases {
    fmt.Printf("%s (%s)\n", alias.SendAsEmail, alias.VerificationStatus)
}

// Send from group email
err = client.SendEmail(
    "support@company.com",
    "customer@example.com",
    "Your Support Ticket",
    "We received your inquiry...",
)
```
