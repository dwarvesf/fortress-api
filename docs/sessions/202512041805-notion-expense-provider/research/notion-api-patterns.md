# Notion API Patterns and Best Practices

## Overview

This document provides research findings on using the Notion API with the `go-notion` library (`github.com/dstotijn/go-notion`) for implementing the Expense Provider functionality.

## Notion API Resources

- **Official Documentation**: [Notion API Reference](https://developers.notion.com/)
- **Go Library**: [github.com/dstotijn/go-notion](https://github.com/dstotijn/go-notion)
- **Go Packages**: [pkg.go.dev/github.com/dstotijn/go-notion](https://pkg.go.dev/github.com/dstotijn/go-notion)

## 1. Querying Databases with Status Filter

### Status Property Filter Syntax

The Notion API supports filtering by status property type with the following operators:

```go
filter := &notion.DatabaseQueryFilter{
    Property: "Status",
    DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
        Status: &notion.StatusDatabaseQueryFilter{
            Equals: "Approved",  // Filter for specific status name
        },
    },
}
```

**Supported Status Filter Operators**:
- `Equals`: Match exact status name (e.g., "Approved", "Paid")
- `DoesNotEqual`: Exclude specific status
- `IsEmpty`: Status has no value
- `IsNotEmpty`: Status has a value

### Example: Query Approved Expenses

```go
ctx := context.Background()
client := notion.NewClient("secret-token")

query := &notion.DatabaseQuery{
    Filter: &notion.DatabaseQueryFilter{
        Property: "Status",
        DatabaseQueryPropertyFilter: notion.DatabaseQueryPropertyFilter{
            Status: &notion.StatusDatabaseQueryFilter{
                Equals: "Approved",
            },
        },
    },
    PageSize: 100,
}

result, err := client.QueryDatabase(ctx, databaseID, query)
```

**Key Points**:
- Status filter expects the **status option name** as a string (e.g., "Approved", not status ID)
- The status property type is different from select property type (cannot use `Select` filter on status)
- Case-sensitive matching of status names

**Sources**:
- [Notion API: Filter database entries](https://developers.notion.com/reference/post-database-query-filter)
- [Stack Overflow: Trying to query with filter Status Notion API Database](https://stackoverflow.com/questions/74317823/trying-to-query-with-filter-status-notion-api-database)
- [Stack Overflow: Cannot filter by status in a notion database](https://stackoverflow.com/questions/74845676/cannot-filter-by-status-in-a-notion-database-using-the-notion-js-client)

## 2. Updating Page Properties (Status)

### Update Page API

The Notion API allows updating page properties, including status properties:

```go
ctx := context.Background()
client := notion.NewClient("secret-token")

updateParams := notion.UpdatePageParams{
    Properties: notion.DatabasePageProperties{
        "Status": notion.DatabasePageProperty{
            Type: notion.DBPropTypeStatus,
            Status: &notion.SelectOptions{
                Name: "Paid",  // New status name
            },
        },
    },
}

updatedPage, err := client.UpdatePage(ctx, pageID, updateParams)
```

**Key Points**:
- Use `DBPropTypeStatus` for status property type
- Provide the status option **name** (not ID) in `SelectOptions.Name`
- Cannot update status property schema (option names, colors) via API - must be done in Notion UI
- Page ID is a UUID string (e.g., "2bfb69f8-f573-81cb-a2da-f06d28896390")

**Important Limitations**:
1. Cannot modify status property **schema** (add/remove/rename options) via API
2. Can only set status to existing option names defined in Notion UI
3. Must have "update content" capabilities for the integration

**Sources**:
- [Notion API: Update page](https://developers.notion.com/reference/patch-page)
- [Stack Overflow: Cannot update status on page property using notion api](https://stackoverflow.com/questions/72986295/cannot-update-status-on-page-property-using-notion-api)
- [Stack Overflow: How to update the Status column via the Notion API](https://stackoverflow.com/questions/79147367/how-to-update-the-status-column-via-the-notion-api)

## 3. Handling Relation Properties

### Relation Property in Queries

Relation properties link pages between databases. The Notion API returns relation data as an array of page references:

```go
// Reading relation property
relationProp := page.Properties.(notion.DatabasePageProperties)["Requestor"]
if relationProp.Type == notion.DBPropTypeRelation {
    relatedPageIDs := relationProp.Relation
    // relatedPageIDs is []notion.PageReference
    for _, ref := range relatedPageIDs {
        relatedPageID := ref.ID  // UUID of related page
    }
}
```

### Querying Related Database for Details

To get actual data from a related page (e.g., contractor email):

```go
// Option 1: Use rollup property (recommended if available)
emailProp := page.Properties.(notion.DatabasePageProperties)["Email"]
if emailProp.Type == notion.DBPropTypeRollup {
    // Rollup aggregates data from relation
    rollupValue := emailProp.Rollup
    // Extract value based on rollup type
}

// Option 2: Fetch related page directly
relatedPage, err := client.FindPageByID(ctx, relatedPageID)
contractorEmail := relatedPage.Properties.(notion.DatabasePageProperties)["Email"].Email
```

**Key Points**:
- Relations store only page references (IDs), not the actual data
- Use **rollup properties** to aggregate data from relations (more efficient)
- Rollup properties are read-only and cannot be updated via API
- Relation queries require multiple API calls if not using rollups

**Sources**:
- [Notion API: Page properties](https://developers.notion.com/reference/page-property-values)
- [github.com/dstotijn/go-notion](https://github.com/dstotijn/go-notion)

## 4. Handling Rollup Properties

### Rollup Property Types

Rollup properties aggregate data from related pages. They can compute:
- Arrays (list of values)
- Dates (earliest, latest, date range)
- Numbers (sum, average, count, etc.)
- Text (show original, show unique)

### Reading Rollup Email Values

For the Expense Request use case (Email rollup from Requestor relation):

```go
emailProp := page.Properties.(notion.DatabasePageProperties)["Email"]
if emailProp.Type == notion.DBPropTypeRollup {
    rollup := emailProp.Rollup

    // Check rollup computation type
    switch rollup.Type {
    case notion.RollupTypeArray:
        // Email is likely in array format
        if len(rollup.Array) > 0 {
            // Array contains property values
            if emailVal, ok := rollup.Array[0].(notion.DatabasePageProperty); ok {
                email := emailVal.Email  // Extract email from first array item
            }
        }
    case notion.RollupTypeNumber:
        // Not applicable for email
    case notion.RollupTypeDate:
        // Not applicable for email
    }
}
```

**Key Points**:
- Rollup properties are **read-only** via API
- Rollup type depends on configuration in Notion UI
- For email rollups, expect `RollupTypeArray` with email property values
- Cannot update rollup values directly - they are computed automatically

**Sources**:
- [Notion API: Filter database entries (Rollup section)](https://developers.notion.com/reference/post-database-query-filter)
- [Notion API: Update page (Rollup limitation note)](https://developers.notion.com/reference/patch-page)

## 5. Pagination Best Practices

### Handling Large Result Sets

The Notion API uses cursor-based pagination:

```go
var allPages []notion.Page
var startCursor string

for {
    query := &notion.DatabaseQuery{
        Filter:      filter,
        PageSize:    100,  // Max 100
        StartCursor: startCursor,
    }

    result, err := client.QueryDatabase(ctx, databaseID, query)
    if err != nil {
        return nil, err
    }

    allPages = append(allPages, result.Results...)

    if !result.HasMore {
        break
    }

    startCursor = *result.NextCursor
}
```

**Key Points**:
- Maximum page size is 100 entries per request
- Use `HasMore` and `NextCursor` for pagination
- Start cursor is optional for first request (defaults to beginning)
- Pagination is consistent for sorted queries

**Sources**:
- [go-notion/database.go](https://github.com/dstotijn/go-notion/blob/main/database.go)
- [Notion API Reference](https://developers.notion.com/reference/post-database-query)

## 6. Property Type Mapping

### Notion Property Types → Go Types

| Notion Property Type | go-notion Type                  | Notes                                      |
|----------------------|---------------------------------|--------------------------------------------|
| title                | `notion.TitleProperty`          | Array of rich text                         |
| rich_text            | `notion.RichTextProperty`       | Array of rich text                         |
| number               | `notion.NumberProperty`         | Float64 value                              |
| select               | `notion.SelectProperty`         | Single option with name                    |
| status               | `notion.StatusProperty`         | Single status option with name (like select)|
| multi_select         | `notion.MultiSelectProperty`    | Array of options                           |
| date                 | `notion.DateProperty`           | Start and optional end date                |
| people               | `notion.PeopleProperty`         | Array of user references                   |
| files                | `notion.FilesProperty`          | Array of file/attachment objects           |
| checkbox             | `notion.CheckboxProperty`       | Boolean value                              |
| url                  | `notion.URLProperty`            | String URL                                 |
| email                | `notion.EmailProperty`          | String email                               |
| phone_number         | `notion.PhoneNumberProperty`    | String phone                               |
| formula              | `notion.FormulaProperty`        | Computed value (read-only)                 |
| relation             | `notion.RelationProperty`       | Array of page references                   |
| rollup               | `notion.RollupProperty`         | Aggregated value (read-only)               |
| created_time         | `notion.CreatedTimeProperty`    | ISO 8601 timestamp (read-only)             |
| created_by           | `notion.CreatedByProperty`      | User reference (read-only)                 |
| last_edited_time     | `notion.LastEditedTimeProperty` | ISO 8601 timestamp (read-only)             |
| last_edited_by       | `notion.LastEditedByProperty`   | User reference (read-only)                 |

**Sources**:
- [go-notion package documentation](https://pkg.go.dev/github.com/dstotijn/go-notion)
- [Notion API: Page property values](https://developers.notion.com/reference/page-property-values)

## 7. Error Handling Best Practices

### Common Notion API Errors

```go
result, err := client.QueryDatabase(ctx, databaseID, query)
if err != nil {
    // Check for specific Notion API errors
    if notionErr, ok := err.(*notion.Error); ok {
        switch notionErr.Code {
        case notion.ErrorCodeObjectNotFound:
            // Database or page not found
        case notion.ErrorCodeUnauthorized:
            // Invalid or missing API token
        case notion.ErrorCodeRestrictedResource:
            // Integration lacks required permissions
        case notion.ErrorCodeValidationError:
            // Invalid request parameters (e.g., wrong filter type)
        case notion.ErrorCodeConflictError:
            // Concurrent modification conflict
        case notion.ErrorCodeRateLimited:
            // Rate limit exceeded (implement backoff)
        case notion.ErrorCodeInternalServerError:
            // Notion service error (retry with backoff)
        }
    }
}
```

**Best Practices**:
1. Always check for `notion.Error` type for API-specific errors
2. Implement exponential backoff for rate limits and server errors
3. Log full error details for debugging (status code, message, request ID)
4. Validate database/page IDs before making requests
5. Handle permission errors gracefully (log and skip, don't crash)

**Sources**:
- [go-notion error handling](https://github.com/dstotijn/go-notion)
- [Notion API: Status codes](https://developers.notion.com/reference/status-codes)

## 8. Testing Strategies

### Unit Testing with go-notion

```go
// Mock Notion client interface for testing
type NotionClient interface {
    QueryDatabase(ctx context.Context, id string, query *notion.DatabaseQuery) (notion.DatabaseQueryResponse, error)
    UpdatePage(ctx context.Context, id string, params notion.UpdatePageParams) (notion.Page, error)
    FindPageByID(ctx context.Context, id string) (notion.Page, error)
}

// Test with mocked responses
func TestFetchApprovedExpenses(t *testing.T) {
    mockClient := &MockNotionClient{
        QueryDatabaseFunc: func(ctx context.Context, id string, query *notion.DatabaseQuery) (notion.DatabaseQueryResponse, error) {
            return notion.DatabaseQueryResponse{
                Results: []notion.Page{
                    // Mock page data
                },
                HasMore: false,
            }, nil
        },
    }

    service := NewExpenseService(mockClient, cfg, store, repo, logger)
    todos, err := service.GetAllInList(0, 0)

    require.NoError(t, err)
    assert.Len(t, todos, 1)
}
```

**Testing Recommendations**:
1. Use interface abstraction for Notion client (enables mocking)
2. Test pagination logic with mock multi-page responses
3. Test error handling for all Notion API error codes
4. Validate property transformation logic with real Notion response structures
5. Use table-driven tests for various status/filter combinations

**Sources**:
- Go testing best practices
- [go-notion test examples](https://github.com/dstotijn/go-notion/blob/main/client_test.go)

## Summary

### Key Takeaways

1. **Status Filters**: Use `StatusDatabaseQueryFilter` with `Equals` operator and status option name
2. **Status Updates**: Use `UpdatePage` with `DBPropTypeStatus` and `SelectOptions.Name`
3. **Relations**: Use rollup properties for efficient data access from relations
4. **Rollups**: Read-only, type varies based on computation (array, number, date)
5. **Pagination**: Implement cursor-based pagination for large result sets (max 100/page)
6. **Error Handling**: Check for `notion.Error` type and handle rate limits with backoff
7. **Page IDs**: UUIDs (not integers like NocoDB), format: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"

### Migration Considerations (NocoDB → Notion)

| Aspect                | NocoDB                     | Notion                         |
|-----------------------|----------------------------|--------------------------------|
| Record ID Type        | Integer                    | UUID string                    |
| API Client            | HTTP REST                  | go-notion library              |
| Status Property       | Text field ("approved")    | Status type ("Approved")       |
| Query Filter          | `where` query param        | `DatabaseQueryFilter` struct   |
| Update Syntax         | PATCH with Id field        | UpdatePage with page ID        |
| Relation Access       | Direct field extraction    | Rollup or secondary query      |
| Pagination            | `limit` + `offset`         | Cursor-based with `HasMore`    |

### References

- [Notion API Reference](https://developers.notion.com/)
- [go-notion GitHub Repository](https://github.com/dstotijn/go-notion)
- [go-notion Package Documentation](https://pkg.go.dev/github.com/dstotijn/go-notion)
- [Notion API Filter Database Entries](https://developers.notion.com/reference/post-database-query-filter)
- [Notion API Update Page](https://developers.notion.com/reference/patch-page)
- [Stack Overflow: Notion Status Filter Questions](https://stackoverflow.com/questions/74317823/trying-to-query-with-filter-status-notion-api-database)
