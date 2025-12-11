# Notion API - Database Creation

## Overview

Notion API supports creating databases programmatically via `POST /v1/databases`.

## Key Constraints

| Constraint | Detail |
|------------|--------|
| **Parent** | Must be a Notion page or wiki database |
| **Inline** | Use `is_inline: true` for embedded, `false` for full-page |
| **Properties** | Must define schema at creation time |

## Supported Property Types

- `title`
- `rich_text`
- `number`
- `select`
- `multi_select`
- `date`
- `people`
- `files`
- `checkbox`
- `url`
- `email`
- `phone_number`
- `created_time`
- `created_by`
- `last_edited_time`
- `last_edited_by`

## Example Request

```json
POST /v1/databases
{
  "parent": { "page_id": "your-page-id" },
  "title": [{ "text": { "content": "My Database" } }],
  "is_inline": false,
  "properties": {
    "Name": { "title": {} },
    "Status": {
      "select": {
        "options": [
          { "name": "Todo", "color": "red" },
          { "name": "Done", "color": "green" }
        ]
      }
    },
    "Amount": { "number": { "format": "dollar" } }
  }
}
```

## References

- [Create a database](https://developers.notion.com/reference/create-a-database)
- [Working with Databases](https://developers.notion.com/docs/working-with-databases)
- [Create a page](https://developers.notion.com/reference/post-page)
