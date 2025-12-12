# Research Status

**Session:** 202512120930-notion-leave-webhook-amdl
**Phase:** Research
**Status:** ✅ COMPLETED
**Date:** 2025-12-12

## Research Objectives

Research the current implementation patterns for Notion webhooks to inform the AM/DL lookup implementation.

### Focus Areas

1. ✅ How Notion API queries work (filtering by relation, status fields)
2. ✅ How rollup and formula fields are extracted from Notion pages
3. ✅ Current patterns for Discord mention lookup (Discord username → Discord ID)
4. ✅ How relation fields are updated on Notion pages

## Files Analyzed

- ✅ `/pkg/handler/webhook/notion_leave.go` - Event webhooks, validation, Discord notifications
- ✅ `/pkg/handler/webhook/notion_refund.go` - Automation webhooks, auto-fill relations
- ✅ `/pkg/service/notion/leave.go` - Leave service, data source queries, property extraction
- ✅ `/pkg/service/notion/expense.go` - Expense service, multi-source databases
- ✅ `/pkg/handler/webhook/nocodb_leave.go` - Discord mention lookup reference

## Key Findings

### 1. Notion API Query Patterns

- **Standard Database Query:** Uses go-notion client with filters (status, email, select)
- **Data Source Query:** Requires raw HTTP API with `Notion-Version: 2025-09-03` header
- **Pagination:** Implemented with cursor-based pagination (PageSize: 100)
- **Email Filter:** Uses `Email: &nt.TextPropertyFilter{ Equals: email }`
- **Status Filter:** Uses `Status: &nt.StatusDatabaseQueryFilter{ Equals: "value" }`

### 2. Property Extraction Patterns

Documented patterns for:
- ✅ Title (rich text array)
- ✅ Rich Text (with cascading fallbacks)
- ✅ Select (single option)
- ✅ Multi-Select (array of options)
- ✅ Email (pointer to string)
- ✅ Date (Start/End with time.Time)
- ✅ Number (pointer to float64)
- ✅ Relation (array of page IDs)
- ✅ Status (similar to select)

**Note:** Rollup and formula extraction patterns are inferred but not yet implemented in codebase.

### 3. Discord Mention Lookup

**Current Pattern (Email-based):**
```
Email → Employee.OneByEmail() → DiscordAccount.One(accountID) → <@discord_id>
```

**Proposed Pattern (Username-based):**
```
Notion Discord field → DiscordAccount.OneByUsername() → <@discord_id>
```

**Key Insight:** Need new store method `DiscordAccount.OneByUsername()` for AM/DL lookup.

### 4. Relation Update Patterns

- **Single Relation:** `Relation: []nt.Relation{{ ID: pageID }}`
- **Multiple Fields:** Build `DatabasePageProperties` map with all updates
- **Atomic Updates:** All field updates in single API call
- **Error Handling:** Log errors but continue processing (graceful degradation)

## Deliverables

- ✅ `notion-patterns.md` - Comprehensive research document with:
  - 10 major sections covering all patterns
  - Code examples from production code
  - Best practices and common patterns
  - Data flow diagrams
  - Key takeaways for AM/DL implementation

## Next Steps

1. Proceed to planning phase
2. Design AM/DL lookup architecture based on research findings
3. Create ADR for design decisions
4. Write technical specification

## Notes

- All patterns are production-tested (extracted from existing codebase)
- go-notion library does NOT support data source queries (requires raw HTTP)
- Notion API version `2025-09-03` required for data sources
- Discord mention format: `<@discord_id>` (not username)
- Signature verification uses HMAC-SHA256 with constant-time comparison

---

**Research Completed:** 2025-12-12
**Researcher:** @agent-researcher
**Next Phase:** Planning
