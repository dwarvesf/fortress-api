# Notion Refund Webhook Payload Structure

## Endpoint

`POST /webhooks/notion/refund`

## Payload Structure

```json
{
  "source": {
    "type": "automation",
    "automation_id": "uuid",
    "action_id": "uuid",
    "event_id": "uuid",
    "attempt": 1
  },
  "data": {
    "object": "page",
    "id": "page-uuid",
    "created_time": "2025-12-11T08:05:00.000Z",
    "last_edited_time": "2025-12-11T08:05:00.000Z",
    "created_by": {
      "object": "user",
      "id": "user-uuid"
    },
    "last_edited_by": {
      "object": "user",
      "id": "user-uuid"
    },
    "parent": {
      "type": "data_source_id",
      "data_source_id": "uuid",
      "database_id": "uuid"
    },
    "archived": false,
    "in_trash": false,
    "is_locked": false,
    "properties": {
      "Status": {
        "id": "%3BWvf",
        "type": "status",
        "status": {
          "id": "kHV;",
          "name": "Pending",
          "color": "yellow"
        }
      },
      "Work Email": {
        "id": "WfZu",
        "type": "email",
        "email": "quang@d.foundation"
      },
      "Amount": {
        "id": "szt%3F",
        "type": "number",
        "number": 120000
      },
      "Currency": {
        "id": "crnR",
        "type": "select",
        "select": {
          "id": "uuid",
          "name": "VND",
          "color": "yellow"
        }
      },
      "Reason": {
        "id": "cl%5Bb",
        "type": "select",
        "select": {
          "id": "}ADC",
          "name": "Deduction Reversal",
          "color": "green"
        }
      },
      "Description": {
        "id": "jfja",
        "type": "rich_text",
        "rich_text": [
          {
            "type": "text",
            "text": {
              "content": "description text",
              "link": null
            },
            "plain_text": "description text"
          }
        ]
      },
      "Discord": {
        "id": "uHlT",
        "type": "rich_text",
        "rich_text": [
          {
            "type": "text",
            "text": {
              "content": "discord_username",
              "link": null
            },
            "plain_text": "discord_username"
          }
        ]
      },
      "Refund ID": {
        "id": "title",
        "type": "title",
        "title": [
          {
            "type": "text",
            "text": {
              "content": "RFD-2025-2NKC",
              "link": null
            },
            "plain_text": "RFD-2025-2NKC"
          }
        ]
      },
      "Date Requested": {
        "id": "%3B%5EjC",
        "type": "date",
        "date": {
          "start": "2025-12-11T15:05:00.000+07:00",
          "end": null,
          "time_zone": null
        }
      },
      "Date Approved": {
        "id": "KgbJ",
        "type": "date",
        "date": null
      },
      "Proof / Attachment": {
        "id": "Lo%5Er",
        "type": "files",
        "files": [
          {
            "name": "filename.png",
            "type": "file",
            "file": {
              "url": "https://prod-files-secure.s3.us-west-2.amazonaws.com/...",
              "expiry_time": "2025-12-11T09:05:11.125Z"
            }
          }
        ]
      },
      "Contractor": {
        "id": "ehxK",
        "type": "relation",
        "relation": [],
        "has_more": false
      },
      "Contractor Email": {
        "id": "i%5CF%5D",
        "type": "rollup",
        "rollup": {
          "type": "array",
          "array": [],
          "function": "show_original"
        }
      },
      "Project (optional)": {
        "id": "iwry",
        "type": "relation",
        "relation": [],
        "has_more": false
      },
      "ID": {
        "id": "W%5D%3EQ",
        "type": "unique_id",
        "unique_id": {
          "prefix": null,
          "number": 5
        }
      },
      "Auto Name": {
        "id": "Jkhr",
        "type": "formula",
        "formula": {
          "type": "string",
          "string": "RFD-2025-2NKC"
        }
      },
      "Payout": {
        "id": "HG%3FZ",
        "type": "rich_text",
        "rich_text": []
      },
      "Notes": {
        "id": "%5CTeg",
        "type": "rich_text",
        "rich_text": []
      }
    },
    "url": "https://www.notion.so/RFD-2025-2NKC-2c664b29b84c8101a59ce6278cad3b0f",
    "public_url": null,
    "request_id": "uuid"
  }
}
```

## Key Properties

| Property | Type | Description |
|----------|------|-------------|
| Status | status | Pending, Approved |
| Work Email | email | Employee work email |
| Amount | number | Refund amount |
| Currency | select | VND, USD, etc. |
| Reason | select | Deduction Reversal, etc. |
| Description | rich_text | Refund description |
| Discord | rich_text | Discord username |
| Refund ID | title | Auto-generated ID (RFD-YYYY-XXXX) |
| Date Requested | date | Request timestamp |
| Date Approved | date | Approval timestamp (null if pending) |
| Proof / Attachment | files | Supporting documents |
| Contractor | relation | Link to contractor page |
| Contractor Email | rollup | Rolled up from contractor |
| Project (optional) | relation | Link to project page |

## Headers

- `User-Agent`: NotionAutomation
- `Content-Type`: application/json
