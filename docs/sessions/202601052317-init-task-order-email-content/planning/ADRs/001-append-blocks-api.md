# ADR-001: Use Notion AppendBlockChildren API for Page Content

## Status
Accepted

## Context
Need to store email confirmation content in the Order page body. Options:
1. Use a rich text property on the page
2. Use page body (block children)

## Decision
Use Notion's `AppendBlockChildren` API to add content as paragraph blocks to the Order page body.

## Rationale
- User explicitly requested page body storage
- More flexible for content formatting
- Visible when viewing the page in Notion
- go-notion library supports this via `AppendBlockChildren` method

## Consequences
- Need to add new service method for appending blocks
- Content appears in page body, not properties panel
- Paragraph blocks with rich text for plain text content
