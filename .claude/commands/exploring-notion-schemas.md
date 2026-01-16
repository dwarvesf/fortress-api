# Efficiently Exploring Notion Schemas

## Problem

Direct database retrieval via `API-retrieve-a-data-source` requires knowing the exact database ID. When a database is renamed or when searching for databases, this approach fails with 404 errors and consumes tokens unnecessarily.

## Solution: Use Search API with Filters

### Method

Use `API-post-search` with the `filter` parameter to search for databases (data sources) by name:

```json
{
  "query": "Database Name",
  "filter": {
    "property": "object",
    "value": "data_source"
  }
}
```

### Benefits

1. **Single API Call**: Find databases by title without knowing the exact ID
2. **Full Schema Returned**: The search results include complete database properties, status options, and all metadata
3. **Token Efficient**: One comprehensive response vs. multiple trial-and-error retrieve calls
4. **Discover Changes**: Quickly identify renamed databases, new database IDs, or schema updates

### Example

**Scenario**: Database previously named "Timesheet" was renamed to "Project Updates"

**Approach**:
```javascript
// Instead of guessing the ID:
API-retrieve-a-data-source({ data_source_id: "old-id" }) // ❌ 404 error

// Use search with filter:
API-post-search({
  query: "Project Updates",
  filter: { property: "object", value: "data_source" }
}) // ✅ Returns full schema in one call
```

### What You Get

The search result includes:
- Correct database ID
- All property definitions (name, type, ID, description)
- Status/select options with colors
- Formula expressions
- Relation targets
- Created/modified timestamps
- Database description and icon

### Token Savings Comparison

| Approach | API Calls | Approximate Tokens |
|----------|-----------|-------------------|
| Trial-and-error retrieve | 3-5+ (failures + success) | 15,000-25,000+ |
| Search with filter | 1 | 8,000-12,000 |

**Savings**: ~50-60% token reduction

### Best Practices

1. **Search First**: Always use search when you don't have the exact database ID
2. **Use Specific Queries**: Search for exact database titles rather than partial matches
3. **Filter by Type**: Always include `filter: {property: "object", value: "data_source"}` to limit results to databases only
4. **Extract What You Need**: The response is comprehensive - extract only the properties you need to document

### Common Use Cases

- Database renamed: Search by new name
- Multiple databases: Search returns all matches - review to find the right one
- Schema exploration: Get complete property list without additional API calls
- Validation: Verify database structure matches documentation

## Related APIs

- `API-post-search`: Find databases and pages by title
- `API-retrieve-a-data-source`: Get database by exact ID (use when ID is known)
- `API-query-data-source`: Query database entries (requires known database ID)
