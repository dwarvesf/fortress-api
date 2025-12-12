# Test Case Design: GetDiscordMentionFromUsername

**Function Under Test:** `handler.getDiscordMentionFromUsername(l logger.Logger, discordUsername string) string`

**Location:** `pkg/handler/webhook/notion_leave.go`

**Purpose:** Convert a Discord username to Discord mention format (`<@discord_id>`) by querying the fortress database.

## Test Strategy

### Approach
- **Pattern:** Table-driven tests with database query scenarios
- **Dependencies:** Mock database (GORM) and DiscordAccount store
- **Validation:** Assert correct mention format and database query behavior
- **Focus:** Database lookup, mention formatting, graceful degradation

### Test Data Requirements
- Mock `discord_accounts` table with test data
- Valid and invalid Discord usernames
- Database connection and query mocking

## Test Cases

### TC-4.1: Successfully Convert Username to Mention

**Given:**
- Discord username: `"johndoe"`
- Database contains matching record:
  - `discord_username`: `"johndoe"`
  - `discord_id`: `"123456789012345678"`

**When:**
- Call `getDiscordMentionFromUsername(logger, "johndoe")`

**Then:**
- Should call `store.DiscordAccount.OneByUsername(db, "johndoe")`
- Should retrieve Discord account with `discord_id = "123456789012345678"`
- Should format mention: `"<@123456789012345678>"`
- Should return: `"<@123456789012345678>"`
- Should log DEBUG: "discord username lookup: username=johndoe discord_id=123456789012345678"

**Database Record:**
```go
&model.DiscordAccount{
    ID:              "uuid-1",
    DiscordID:       "123456789012345678",
    DiscordUsername: "johndoe",
}
```

---

### TC-4.2: Username Not Found - Return Empty String

**Given:**
- Discord username: `"unknownuser"`
- Database query returns no matching record (nil)

**When:**
- Call `getDiscordMentionFromUsername(logger, "unknownuser")`

**Then:**
- Should call `store.DiscordAccount.OneByUsername(db, "unknownuser")`
- Should receive nil result (not found)
- Should return: `""` (empty string)
- Should log WARNING: "discord username not found in database: username=unknownuser"

**Rationale:** Not all Notion contractors have Discord accounts; graceful handling required

---

### TC-4.3: Database Error - Return Empty String

**Given:**
- Discord username: `"johndoe"`
- Database query returns error (connection error, query error)

**When:**
- Call `getDiscordMentionFromUsername(logger, "johndoe")`

**Then:**
- Should call `store.DiscordAccount.OneByUsername(db, "johndoe")`
- Should receive error from store
- Should return: `""` (empty string, graceful degradation)
- Should log ERROR: "failed to query discord account: username=johndoe error=%v"

**Mock Error:**
```go
errors.New("database connection lost")
```

**Rationale:** Don't fail webhook processing if Discord lookup fails

---

### TC-4.4: Empty Username - Return Empty String

**Given:**
- Discord username: `""`

**When:**
- Call `getDiscordMentionFromUsername(logger, "")`

**Then:**
- Should return: `""` (empty string immediately)
- Should NOT call database (short-circuit)
- Should log DEBUG: "empty discord username provided"

**Rationale:** Avoid unnecessary database queries for empty input

---

### TC-4.5: Username with Whitespace - Trim and Query

**Given:**
- Discord username: `"  johndoe  "` (with leading/trailing whitespace)
- Database contains record for `"johndoe"`

**When:**
- Call `getDiscordMentionFromUsername(logger, "  johndoe  ")`

**Then:**
- Should trim whitespace: `"johndoe"`
- Should query database with trimmed username
- Should return mention: `"<@123456789012345678>"`
- Should log DEBUG with trimmed username

**Rationale:** Handle input inconsistencies from Notion

---

### TC-4.6: Case Sensitivity - Exact Match Required

**Given:**
- Discord username input: `"JohnDoe"` (mixed case)
- Database contains: `"johndoe"` (lowercase)

**When:**
- Call `getDiscordMentionFromUsername(logger, "JohnDoe")`

**Then:**
- Should query database with: `"JohnDoe"` (exact as provided after trim)
- Should return: `""` if no exact match found
- Should log WARNING: "discord username not found"

**Note:** Test both case-sensitive and case-insensitive scenarios depending on database collation

**Alternative (if using case-insensitive query):**
- Should find record and return mention
- Document whether lookup is case-sensitive or not

---

### TC-4.7: Multiple Records Found - Use First Result

**Given:**
- Discord username: `"johndoe"`
- Database query returns multiple records (data integrity issue)

**When:**
- Call `getDiscordMentionFromUsername(logger, "johndoe")`

**Then:**
- `OneByUsername` should return first record (GORM `First()` behavior)
- Should format mention from first record
- Should log WARNING: "multiple discord accounts found for username: johndoe (data issue)"

**Rationale:** Defensive programming; ideally username should be unique

---

### TC-4.8: Discord ID is Empty - Return Empty String

**Given:**
- Discord username: `"johndoe"`
- Database returns record but `discord_id` is empty string or null

**When:**
- Call `getDiscordMentionFromUsername(logger, "johndoe")`

**Then:**
- Should retrieve record successfully
- Should detect empty `discord_id`
- Should return: `""` (empty string)
- Should log WARNING: "discord account has empty discord_id: username=johndoe"

**Database Record:**
```go
&model.DiscordAccount{
    ID:              "uuid-1",
    DiscordID:       "", // Empty
    DiscordUsername: "johndoe",
}
```

---

### TC-4.9: Special Characters in Username

**Given:**
- Discord username: `"user_name-2024"`
- Database contains matching record with Discord ID

**When:**
- Call `getDiscordMentionFromUsername(logger, "user_name-2024")`

**Then:**
- Should query database correctly with special characters
- Should return mention: `"<@discord_id>"`
- Should handle underscores, hyphens correctly (no escaping issues)

---

### TC-4.10: Nil Logger - Handle Gracefully

**Given:**
- Discord username: `"johndoe"`
- Logger is nil

**When:**
- Call `getDiscordMentionFromUsername(nil, "johndoe")`

**Then:**
- Should NOT panic
- Should still perform database lookup
- Should return mention or empty string correctly
- Should skip logging (no-op if logger is nil)

**Note:** Defensive programming check

---

## Store Method: OneByUsername

**New Method to Implement:** `DiscordAccount.OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)`

**Location:** `pkg/store/discordaccount/discordaccount.go`

### Interface Addition
```go
// In pkg/store/discordaccount/interface.go
type IStore interface {
    // ... existing methods ...
    OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error)
}
```

### Test Cases for OneByUsername

#### TC-4.11: Store Method - Find Existing Account

**Given:**
- Database has record: `discord_username = "johndoe"`

**When:**
- Call `store.OneByUsername(db, "johndoe")`

**Then:**
- Should execute SQL: `SELECT * FROM discord_accounts WHERE discord_username = $1 LIMIT 1`
- Should return: `*model.DiscordAccount` with all fields populated
- Should return nil error

---

#### TC-4.12: Store Method - Account Not Found

**Given:**
- Database has no record for: `discord_username = "unknownuser"`

**When:**
- Call `store.OneByUsername(db, "unknownuser")`

**Then:**
- Should execute query
- Should return: `nil` (not found)
- Should return: `nil` error (graceful, not GORM error)
- Should detect `gorm.ErrRecordNotFound` and convert to nil, nil

**Implementation Pattern:**
```go
err := db.Where("discord_username = ?", username).First(&discordAccount).Error
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil // Graceful not found
    }
    return nil, err // Actual error
}
```

---

#### TC-4.13: Store Method - Database Error

**Given:**
- Database connection is down or query fails

**When:**
- Call `store.OneByUsername(db, "johndoe")`

**Then:**
- Should return: `nil` for account
- Should return: error from GORM
- Should NOT mask error (return actual error)

---

#### TC-4.14: Store Method - Empty Username

**Given:**
- Username parameter is empty string

**When:**
- Call `store.OneByUsername(db, "")`

**Then:**
- **Option 1:** Query database (likely returns nil)
- **Option 2:** Return error: "username is required"
- **Recommended:** Option 1 (let database handle, may return nothing)

---

## Mock Setup Requirements

### Mock Store
```go
type MockDiscordAccountStore struct {
    OneByUsernameFunc func(db *gorm.DB, username string) (*model.DiscordAccount, error)
    callCount         int
    capturedUsernames []string
}

func (m *MockDiscordAccountStore) OneByUsername(db *gorm.DB, username string) (*model.DiscordAccount, error) {
    m.callCount++
    m.capturedUsernames = append(m.capturedUsernames, username)
    if m.OneByUsernameFunc != nil {
        return m.OneByUsernameFunc(db, username)
    }
    return nil, nil
}
```

### Mock Assertions
- Verify `OneByUsername` called with correct username
- Verify trimming logic applied before query
- Verify mention format is correct: `<@discord_id>`
- Verify empty string returned on errors/not found

## Mention Format Specification

### Discord Mention Format
- **Pattern:** `<@DISCORD_ID>`
- **Example:** `<@123456789012345678>`
- **Discord ID Format:** 18-digit numeric string (snowflake ID)

### Format Validation
- Must start with `<@`
- Must end with `>`
- Must contain valid Discord ID between `<@` and `>`
- No spaces or special characters

### Test Mention Formatting
```go
func formatDiscordMention(discordID string) string {
    if discordID == "" {
        return ""
    }
    return fmt.Sprintf("<@%s>", discordID)
}
```

## Error Handling Strategy

### Return Empty String (Graceful Degradation)
1. Username not found in database
2. Database query error
3. Empty username input
4. Empty Discord ID in record

### Do Not Fail Webhook
- This function is called during webhook processing
- Failing here would block entire leave request notification
- Better to send notification without mentions than fail completely

## Logging Assertions

### DEBUG Level
- "querying discord account by username: %s"
- "discord username lookup: username=%s discord_id=%s"

### WARNING Level
- "discord username not found in database: username=%s"
- "discord account has empty discord_id: username=%s"
- "multiple discord accounts found for username: %s"

### ERROR Level
- "failed to query discord account: username=%s error=%v"

## Configuration Dependencies

**None** - Uses handler's existing store and database connection

## Performance Considerations

### Database Query Performance
- **Expected:** < 50ms per query
- **Index Required:** Create index on `discord_username` column
- **Query Pattern:** Simple WHERE clause with LIMIT 1

### Index Recommendation
```sql
CREATE INDEX idx_discord_accounts_username ON discord_accounts(discord_username);
```

### Optimization
- Current implementation: Sequential queries for each username
- Future optimization: Batch query with IN clause for multiple usernames
- Expected load: 1-5 usernames per leave request (acceptable without batching)

## Integration Notes

### Caller Expectations
- Caller: `handleNotionLeaveCreated` after collecting Discord usernames from Notion
- Caller will iterate over multiple usernames
- Caller expects empty string as valid response
- Caller will filter out empty mentions before sending to Discord

### Downstream Usage
- Returned mentions will be joined into Discord message
- Discord API will parse mentions and notify users
- Invalid mention format will appear as plain text (graceful degradation)

## Database Schema

### DiscordAccount Model
```go
type DiscordAccount struct {
    ID              string // UUID
    DiscordID       string // Discord snowflake ID (18 digits)
    DiscordUsername string // Discord username
    // ... other fields
}
```

### Table: discord_accounts
- **Column:** `discord_username` (string, indexed)
- **Column:** `discord_id` (string, Discord snowflake)
- **Constraint:** `discord_username` should be unique (business logic)

## Test Implementation Checklist

- [ ] Test successful username to mention conversion (happy path)
- [ ] Test username not found (graceful return)
- [ ] Test database error (graceful return)
- [ ] Test empty username (short-circuit)
- [ ] Test username with whitespace (trim before query)
- [ ] Test case sensitivity (document behavior)
- [ ] Test empty Discord ID in record
- [ ] Test special characters in username
- [ ] Test nil logger handling
- [ ] Implement `OneByUsername` store method
- [ ] Test store method with existing account
- [ ] Test store method with not found
- [ ] Test store method with database error
- [ ] Verify mention format is correct
- [ ] Verify logging at appropriate levels
- [ ] Verify database query executed correctly
- [ ] Test with real database (integration test boundary)
