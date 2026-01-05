# Requirements: Add Email Content to InitTaskOrderLogs

## Overview
Enhance the existing `InitTaskOrderLogs` endpoint to generate email confirmation content and store it in the Order page body.

## Functional Requirements

### FR-1: Generate Email Confirmation Content
- After creating all Line Items for a contractor, generate plain text confirmation content
- Content should include:
  - Greeting with contractor name
  - Month and period information
  - List of clients (with headquarters/country info)
  - Confirmation request

### FR-2: Append Content to Order Page Body
- Store the generated content in the Notion Order page body (as block children)
- Content should be appended AFTER all Line Items are created for that contractor
- Use paragraph blocks for plain text format

### FR-3: Client Information Collection
- Collect client info from each deployment's project
- If client is in Vietnam, use "Dwarves LLC (USA)" instead
- Deduplicate clients by name+country

## Non-Functional Requirements

### NFR-1: Logging
- Add DEBUG level logging for content generation and append operations

### NFR-2: Error Handling
- Continue processing other contractors if content generation fails for one
- Log errors but don't fail the entire endpoint

## Constraints

- No new endpoint - modify existing `InitTaskOrderLogs`
- Plain text format only (no headings/bullets)
- Uses go-notion `AppendBlockChildren` API

## Clarified Requirements

1. Content generated AFTER all Line Items created (to have complete client info)
2. Store in page body, not a property
3. Plain text paragraphs
