# ADR-001: Async Discord Command Pattern with DM Message Updates

## Status
Proposed

## Context
The `?gen invoice` command needs to trigger invoice generation, which involves:
- Fetching contractor data from Notion API
- Generating invoice PDF (potentially slow)
- Uploading to Google Drive
- Sharing file with contractor's email

This process can take 5-15 seconds, but Discord has a hard 3-second timeout for interaction responses. We need an async pattern that provides immediate feedback while processing in the background.

### Constraints
- Discord webhook timeout: 3 seconds maximum
- User needs immediate acknowledgment that command was received
- User needs to know when processing completes (success or failure)
- Must not spam user with multiple messages
- Should maintain context of the original command

### Options Considered

#### Option 1: Channel Message + Follow-up DM (Rejected)
Post "Processing..." in the channel where command was issued, then DM the result.

**Pros:**
- Clear command acknowledgment visible to user
- Result in DM keeps invoice details private

**Cons:**
- Creates noise in public channels
- Two separate message threads (loses context)
- Channel message becomes stale/orphaned
- May confuse user about where to look for result

#### Option 2: Channel Message + Edit Same Message (Rejected)
Post "Processing..." in channel, then edit it with result.

**Pros:**
- Single message location
- Public visibility of command usage

**Cons:**
- Exposes invoice generation details in public channel
- Invoice information should be private
- Clutters channel history
- Not suitable for sensitive financial data

#### Option 3: DM Message + Edit Same Message (Selected)
Create DM immediately with "Processing..." embed, then edit that same message with result.

**Pros:**
- Private communication for financial data
- Single message maintains context
- Clean UX - user sees progress in one place
- No channel clutter
- Can include rich embed formatting
- User can reference the message later

**Cons:**
- Requires two API calls (create DM, send message)
- Slightly more complex message tracking
- User might miss DM notification

## Decision
We will use **Option 3: DM Message + Edit Same Message**.

### Implementation Pattern
```
1. User issues command in any channel: ?gen inv 2025-01
2. Bot creates/gets DM channel with user
3. Bot sends embed message: "Processing your invoice for 2025-01..."
4. Bot stores dmChannelID and dmMessageID
5. Bot POSTs webhook to fortress-api with:
   - discordUsername
   - month
   - dmChannelID
   - dmMessageID
6. Bot returns immediately (command completes < 3s)
7. fortress-api processes async in goroutine:
   - Generate invoice PDF
   - Upload to Google Drive
   - Share with contractor email
   - Call discord.UpdateChannelMessage(dmChannelID, dmMessageID, result)
8. User sees DM message update to success/failure
```

### Message Update Pattern
The same message ID is used for both states:

**Processing State:**
```
Embed:
  Title: "Generating Invoice"
  Description: "Processing your invoice for January 2025..."
  Color: Blue
  Footer: "This may take a few moments"
```

**Success State:**
```
Embed:
  Title: "Invoice Generated Successfully"
  Description: "Your invoice for January 2025 has been generated and shared to your email"
  Fields:
    - "Month": "January 2025"
    - "File": "Link to Google Drive file"
    - "Email": "contractor@example.com"
  Color: Green
  Footer: "Check your email for file access"
```

**Error State:**
```
Embed:
  Title: "Invoice Generation Failed"
  Description: "Failed to generate invoice: [error reason]"
  Color: Red
  Footer: "Contact support if this persists"
```

## Consequences

### Positive
- Clean, private user experience
- Maintains context in single message thread
- Suitable for sensitive financial data
- Scalable pattern for other async commands
- User has clear history of invoice generation attempts

### Negative
- Requires message tracking (dmChannelID + dmMessageID)
- Two Discord API calls per command (DM create + message send)
- Webhook payload slightly larger (includes message context)
- User must have DMs enabled (mitigated by error handling)

### Technical Requirements
- fortress-discord must implement DM channel creation
- fortress-discord must track message IDs to pass to webhook
- fortress-api must implement message update functionality
- fortress-api must handle DM permission errors gracefully

### Error Handling
- If DM creation fails (user has DMs disabled): Respond in channel with error
- If webhook POST fails: Update DM with error message before returning
- If fortress-api message update fails: Log error but don't retry (avoid spam)
- If user deletes DM message: Ignore (process completes silently)

## References
- Discord API: User DMs - https://discord.com/developers/docs/resources/user#create-dm
- Discord API: Edit Message - https://discord.com/developers/docs/resources/channel#edit-message
- fortress-discord service pattern: `pkg/discord/service/`
- fortress-api UpdateChannelMessage: `pkg/service/discord/discord.go:1080`
