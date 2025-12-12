# Test Case Design: GetDiscordUsernameFromContractor

**Function Under Test:** `LeaveService.GetDiscordUsernameFromContractor(ctx context.Context, contractorPageID string) (string, error)`

**Location:** `pkg/service/notion/leave.go`

**Purpose:** Fetch a contractor page from Notion by ID and extract the Discord username from the "Discord" rich_text property.

## Test Strategy

### Approach
- **Pattern:** Table-driven tests with various property scenarios
- **Dependencies:** Mock Notion API client
- **Validation:** Assert correct username extraction and error handling
- **Focus:** Rich text concatenation, empty field handling, API error handling

### Test Data Requirements
- Mock Notion API responses for FindPageByID
- Contractor pages with various "Discord" property formats
- Edge cases: empty Discord field, multi-part rich text, special characters

## Test Cases

### TC-3.1: Extract Simple Discord Username

**Given:**
- Valid contractor page ID: `contractor-abc-123`
- Notion API returns contractor page with:
  - "Discord" property (rich_text): `["username123"]`

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, "contractor-abc-123")`

**Then:**
- Should call `client.FindPageByID(ctx, "contractor-abc-123")`
- Should extract "Discord" rich_text property
- Should return: `"username123"`
- Should return nil error
- Should log DEBUG: "extracted discord username: username123"

**Test Data:**
```go
props := nt.DatabasePageProperties{
    "Discord": nt.DatabasePageProperty{
        RichText: []nt.RichText{
            {
                Type:      "text",
                PlainText: "username123",
            },
        },
    },
}
```

---

### TC-3.2: Extract Username with Special Characters

**Given:**
- Contractor page with Discord username containing special characters:
  - Username: `user_name-2024.test`

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should return: `"user_name-2024.test"` (preserve special characters)
- Should return nil error
- Should handle underscores, hyphens, dots correctly

**Rationale:** Discord usernames can contain alphanumeric, underscores, hyphens, dots

---

### TC-3.3: Extract Multi-Part Rich Text - Concatenate

**Given:**
- Contractor page with "Discord" property containing multiple rich text parts:
  ```
  RichText[0].PlainText = "user"
  RichText[1].PlainText = "name"
  RichText[2].PlainText = "123"
  ```

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should concatenate all parts: `"username123"`
- Should join without separators (direct concatenation)
- Should return nil error

**Test Data:**
```go
"Discord": nt.DatabasePageProperty{
    RichText: []nt.RichText{
        {PlainText: "user"},
        {PlainText: "name"},
        {PlainText: "123"},
    },
}
```

---

### TC-3.4: Trim Whitespace from Username

**Given:**
- Contractor page with Discord username containing leading/trailing whitespace:
  - `"  username123  "`

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should return: `"username123"` (trimmed)
- Should use `strings.TrimSpace()` on final result
- Should return nil error

**Rationale:** Prevent whitespace issues in downstream Discord lookups

---

### TC-3.5: Empty Discord Field - Return Empty String

**Given:**
- Contractor page with "Discord" property that is empty:
  - `RichText: []` (empty array)

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should return: `""` (empty string)
- Should return nil error (graceful handling, not an error)
- Should log INFO: "discord field is empty for contractor: page_id=%s"

**Rationale:** Not all contractors have Discord usernames; this is acceptable

---

### TC-3.6: Discord Property Not Found - Return Empty String

**Given:**
- Contractor page without "Discord" property in properties map

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should return: `""` (empty string)
- Should return nil error (graceful handling)
- Should log INFO: "discord property not found for contractor: page_id=%s"

**Test Data:**
```go
props := nt.DatabasePageProperties{
    "Name": nt.DatabasePageProperty{...},
    // No "Discord" property
}
```

---

### TC-3.7: Notion API Error - Page Not Found

**Given:**
- Contractor page ID: `nonexistent-page-id`
- Notion API returns 404 error (page not found)

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, "nonexistent-page-id")`

**Then:**
- Should call `client.FindPageByID()`
- Should return: `""` (empty string)
- Should return error: "failed to fetch contractor page: page not found"
- Should log ERROR with page ID

**Mock Error:**
```go
errors.New("notion API: object not found")
```

---

### TC-3.8: Notion API Error - Rate Limit

**Given:**
- Valid contractor page ID
- Notion API returns rate limit error (429)

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should return: `""` (empty string)
- Should return error: "failed to fetch contractor page: rate limit exceeded"
- Should log ERROR for retry handling at caller level

---

### TC-3.9: Context Cancellation - Respect Context

**Given:**
- Valid contractor page ID
- Context is cancelled before API call completes

**When:**
- Call `GetDiscordUsernameFromContractor(cancelledCtx, contractorPageID)`

**Then:**
- Should attempt to call `client.FindPageByID()` with cancelled context
- Should return: `""` (empty string)
- Should return context.Canceled error or wrapped error
- Should not hang or block

---

### TC-3.10: Invalid Page Properties Type - Return Error

**Given:**
- Notion returns page where properties cannot be cast to `DatabasePageProperties`
  - e.g., page is a regular page, not a database page

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should attempt to cast properties
- Should return: `""` (empty string)
- Should return error: "failed to cast page properties"
- Should log ERROR with page ID

**Test Data:**
```go
page := nt.Page{
    ID:         contractorPageID,
    Properties: nt.PageProperties{}, // Not DatabasePageProperties
}
```

---

### TC-3.11: Empty Contractor Page ID - Input Validation

**Given:**
- Empty contractor page ID: `""`

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, "")`

**Then:**
- Should return: `""` (empty string)
- Should return error: "contractor page ID is required"
- Should NOT call Notion API
- Should log WARNING about invalid input

**Rationale:** Prevent unnecessary API calls

---

### TC-3.12: Nil Context - Handle Gracefully

**Given:**
- Valid contractor page ID
- Context is nil

**When:**
- Call `GetDiscordUsernameFromContractor(nil, contractorPageID)`

**Then:**
- Should return error: "context is required"
- Should NOT proceed with API call

**Note:** Defensive programming check

---

### TC-3.13: Rich Text with Empty PlainText - Skip Empty Parts

**Given:**
- Contractor page with "Discord" containing rich text with empty parts:
  ```
  RichText[0].PlainText = "user"
  RichText[1].PlainText = ""
  RichText[2].PlainText = "name"
  ```

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should concatenate non-empty parts: `"username"`
- Or simply concatenate all (result: `"username"` after trim)
- Should return nil error

**Implementation Note:** Both approaches acceptable; document chosen behavior

---

### TC-3.14: Rich Text with Only Whitespace - Return Empty After Trim

**Given:**
- Contractor page with "Discord" containing only whitespace:
  ```
  RichText[0].PlainText = "   "
  RichText[1].PlainText = " "
  ```

**When:**
- Call `GetDiscordUsernameFromContractor(ctx, contractorPageID)`

**Then:**
- Should concatenate and trim: `""`
- Should return empty string
- Should return nil error
- Should log INFO: "discord field is empty (whitespace only)"

---

## Property Extraction Helper Function

### Helper: extractRichText
**Signature:** `extractRichText(props nt.DatabasePageProperties, propName string) string`

**Test Coverage:**
- Single rich text part → return plain text
- Multiple rich text parts → concatenate
- Empty rich text array → return empty string
- Property not found → return empty string
- Whitespace handling → trim final result

**Implementation Pattern:**
```go
func (s *LeaveService) extractRichText(props nt.DatabasePageProperties, propName string) string {
    if prop, ok := props[propName]; ok && len(prop.RichText) > 0 {
        var parts []string
        for _, rt := range prop.RichText {
            parts = append(parts, rt.PlainText)
        }
        return strings.TrimSpace(strings.Join(parts, ""))
    }
    return ""
}
```

## Mock Setup Requirements

### Mock Notion Client
```go
type MockNotionClient struct {
    FindPageByIDFunc func(ctx context.Context, pageID string) (*nt.Page, error)
    callCount        int
}
```

### Mock Assertions
- Verify `FindPageByID` called with correct page ID
- Verify context passed correctly
- Verify error handling propagates correctly

## Error Handling Strategy

### Error Return Conditions
1. **API Errors:** Wrap and return (page not found, rate limit, network error)
2. **Invalid Page Properties:** Return error if cannot cast
3. **Input Validation Errors:** Empty page ID

### Graceful Degradation Conditions
1. **Empty Discord Field:** Return empty string, nil error
2. **Property Not Found:** Return empty string, nil error

### Error Messages
- Include contractor page ID in all error messages for debugging
- Use descriptive error messages that indicate the failure reason

## Logging Assertions

### DEBUG Level
- "fetching contractor page: page_id=%s"
- "extracted discord username: %s from contractor: %s"

### INFO Level
- "discord field is empty for contractor: page_id=%s"
- "discord property not found for contractor: page_id=%s"

### ERROR Level
- "failed to fetch contractor page: page_id=%s error=%v"
- "failed to cast page properties: page_id=%s"

## Configuration Dependencies

**None** - Uses Notion client already configured in LeaveService

## Performance Considerations

### Expected Behavior
- Single API call per contractor
- Typical response time: < 500ms (Notion API)
- May be called multiple times in sequence for multiple stakeholders

### Optimization Opportunities (Future)
- **Batch Fetching:** Fetch multiple contractor pages in parallel (if Notion API supports)
- **Caching:** Cache contractor Discord usernames (short TTL)
- **Current Implementation:** No optimization needed for expected load (5-10 stakeholders per leave request)

## Integration Notes

### Caller Expectations
- Caller: `handleNotionLeaveCreated` after extracting stakeholders
- Caller will iterate over multiple contractor page IDs
- Caller expects empty string as valid response (some contractors have no Discord)
- Caller will skip empty usernames in downstream processing

### Downstream Usage
- Returned Discord username will be passed to `GetDiscordMentionFromUsername`
- Username must match the format in `discord_accounts.discord_username` table

## Test Implementation Checklist

- [ ] Test simple username extraction (happy path)
- [ ] Test multi-part rich text concatenation
- [ ] Test whitespace trimming
- [ ] Test empty Discord field (graceful return)
- [ ] Test missing Discord property (graceful return)
- [ ] Test Notion API errors (page not found, rate limit)
- [ ] Test context cancellation
- [ ] Test invalid page properties type
- [ ] Test input validation (empty page ID)
- [ ] Test special characters in username
- [ ] Test rich text with empty parts
- [ ] Test whitespace-only Discord field
- [ ] Verify helper function `extractRichText` works correctly
- [ ] Verify logging at appropriate levels
- [ ] Verify error messages include page ID
