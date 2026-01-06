# Gmail Reply Detection

## Overview

This document describes how to detect when an email is replied to using Google Gmail API with OAuth refresh tokens.

## Prerequisites

- Google Cloud project with Gmail API enabled
- OAuth 2.0 credentials (Client ID, Client Secret)
- Refresh token with appropriate scopes
- Required scopes: `gmail.readonly` or `gmail.modify`

## Approaches

### 1. Push Notifications (Recommended)

Use Gmail's `watch` API with Google Cloud Pub/Sub to receive real-time notifications.

#### Setup Steps

1. Create a Pub/Sub topic in Google Cloud Console
2. Grant Gmail publish permissions to the topic
3. Create a subscription for the topic
4. Call the `watch` API to start receiving notifications

#### API Call

```
POST https://gmail.googleapis.com/gmail/v1/users/me/watch

{
  "topicName": "projects/PROJECT_ID/topics/TOPIC_NAME",
  "labelIds": ["INBOX"],
  "labelFilterBehavior": "INCLUDE"
}
```

#### Response

```json
{
  "historyId": "12345",
  "expiration": "1234567890000"
}
```

#### Handling Notifications

When a notification arrives:
1. Decode the Pub/Sub message
2. Extract `historyId` from the notification
3. Call `history.list` to get changes since last `historyId`
4. Check if new messages belong to tracked threads

```
GET https://gmail.googleapis.com/gmail/v1/users/me/history?startHistoryId={historyId}
```

### 2. Polling

Periodically query for new messages in a specific thread.

#### Get Thread Messages

```
GET https://gmail.googleapis.com/gmail/v1/users/me/threads/{threadId}
```

#### Response

```json
{
  "id": "thread123",
  "historyId": "12345",
  "messages": [
    {
      "id": "msg1",
      "threadId": "thread123",
      "labelIds": ["INBOX", "UNREAD"],
      "snippet": "Original message...",
      "internalDate": "1234567890000"
    },
    {
      "id": "msg2",
      "threadId": "thread123",
      "labelIds": ["INBOX", "UNREAD"],
      "snippet": "Reply message...",
      "internalDate": "1234567891000"
    }
  ]
}
```

#### Detection Logic

1. Store the original `messageId` and `threadId` when sending
2. Periodically fetch the thread
3. Compare message count or check for new `messageId`s
4. New messages in the thread indicate replies

## Key Concepts

### Thread ID

- Every email belongs to a thread (`threadId`)
- Replies share the same `threadId` as the original email
- Use `threadId` to track conversation state

### History ID

- Monotonically increasing identifier for mailbox state
- Use with `history.list` to get incremental changes
- More efficient than full thread polling

### Message Headers

To identify the sender of a reply:

```
GET https://gmail.googleapis.com/gmail/v1/users/me/messages/{messageId}?format=metadata&metadataHeaders=From&metadataHeaders=Subject
```

## Implementation Considerations

### Push Notifications

**Pros:**
- Real-time detection
- Lower API quota usage
- More efficient for high-volume scenarios

**Cons:**
- Requires Pub/Sub setup
- Watch expires after 7 days (must renew)
- More complex infrastructure

### Polling

**Pros:**
- Simpler implementation
- No additional infrastructure needed
- Works without Pub/Sub access

**Cons:**
- Delayed detection (depends on poll interval)
- Higher API quota usage
- Not suitable for real-time requirements

## Refresh Token Usage

The refresh token is used to obtain new access tokens when they expire:

```go
func getAccessToken(refreshToken string) (string, error) {
    resp, err := http.PostForm("https://oauth2.googleapis.com/token", url.Values{
        "client_id":     {clientID},
        "client_secret": {clientSecret},
        "refresh_token": {refreshToken},
        "grant_type":    {"refresh_token"},
    })
    // Parse response for access_token
}
```

## Rate Limits

- Gmail API: 250 quota units per user per second
- `threads.get`: 5 quota units
- `history.list`: 2 quota units
- `watch`: 100 quota units

## References

- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [Push Notifications](https://developers.google.com/gmail/api/guides/push)
- [Threads](https://developers.google.com/gmail/api/reference/rest/v1/users.threads)
- [History](https://developers.google.com/gmail/api/reference/rest/v1/users.history)
