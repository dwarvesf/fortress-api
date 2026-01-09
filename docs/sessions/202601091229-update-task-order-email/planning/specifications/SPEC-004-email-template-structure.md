# SPEC-004: Email Template Structure

## Overview
Replace the task order confirmation email template with new invoice reminder format.

## Implementation Location

**File**: `pkg/templates/taskOrderConfirmation.tpl`

## Template Strategy

**Approach**: Complete replacement (not incremental edits)
- Easier to review (clear before/after)
- Simpler rollback
- Cleaner diff in version control

## New Template Structure

### MIME Headers
```
Mime-Version: 1.0
From: "Spawn @ Dwarves LLC" <spawn@d.foundation>
To: {{.TeamEmail}}
Subject: Quick update for {{formattedMonth}} – Invoice reminder & client milestones
Content-Type: multipart/mixed; boundary=main
```

### Email Sections

#### 1. Greeting
```html
<p>Hi {{contractorLastName}},</p>
```

#### 2. Opening Message
```html
<p>Hope you're having a great start to {{formattedMonth}}!</p>
```

#### 3. Invoice Reminder
```html
<p>Just a quick note:</p>

<p>Your regular monthly invoice for {{formattedMonth}} services is due by <b>{{invoiceDueDay}}</b>. As usual, please use the standard template and send to <a href="mailto:billing@dwarves.llc">billing@dwarves.llc</a>.</p>
```

#### 4. Client Milestones
```html
<p>Upcoming client milestones (for awareness):</p>
<ul>
    {{range .Milestones}}
    <li>{{.}}</li>
    {{end}}
</ul>
```

#### 5. Encouragement
```html
<p>You're continuing to do excellent work on the embedded team – clients are very happy with your contributions.</p>
```

#### 6. Support Offer
```html
<p>If anything comes up or you need support, just ping me anytime.</p>
```

#### 7. Closing & Signature
```html
<p>Best,</p>

<div><br></div>-- <br>
{{ template "signature.tpl" }}
```

## Template Functions Required

Must be defined in `pkg/service/googlemail/utils.go`:

```go
"formattedMonth": func() string {
    // "January 2026"
}

"contractorLastName": func() string {
    // Extract last name from full name
}

"invoiceDueDay": func() string {
    // Return data.InvoiceDueDay
}
```

## Data Bindings

| Template Variable | Source | Example |
|-------------------|--------|---------|
| `{{.TeamEmail}}` | `data.TeamEmail` | "john@example.com" |
| `{{formattedMonth}}` | Template function | "January 2026" |
| `{{contractorLastName}}` | Template function | "Smith" |
| `{{invoiceDueDay}}` | Template function | "10th" |
| `{{range .Milestones}}` | `data.Milestones` | Array iteration |

## MIME Format

### Boundary Structure
```
--main
Content-Type: text/html; charset="UTF-8"
Content-Transfer-Encoding: quoted-printable

[HTML CONTENT]

--main--
```

### Encoding
- Content-Transfer-Encoding: quoted-printable
- Charset: UTF-8
- Maintains compatibility with Gmail API

## HTML Styling

### Inline Styles
- No external CSS
- Use inline styles for Gmail compatibility
- Bold tags for emphasis: `<b>{{invoiceDueDay}}</b>`
- Links with mailto: `<a href="mailto:billing@dwarves.llc">`

### Structure
- Simple paragraph tags: `<p>`
- Unordered lists: `<ul><li>`
- Divs for spacing: `<div><br></div>`

## Signature Integration

### Template Include
```html
{{ template "signature.tpl" }}
```

### Signature Template
- Location: `pkg/templates/signature.tpl`
- Unchanged structure (HTML table with contact info)
- Only signature content changes (name/title via template functions)

## Edge Cases

### Empty Milestones
If `data.Milestones` is empty or nil:
- The `{{range}}` produces no output
- Empty `<ul>` tag is rendered
- Browser handles gracefully (no visual issue)

### Missing Data
- All fields are required to be populated by handler
- Template assumes all data is valid
- No null checks needed

## Testing Verification

### Visual Checks
1. Subject line correct
2. Greeting uses last name
3. Month formatted correctly
4. Due date shows ordinal (10th/25th)
5. Milestones display as bullets
6. Links are clickable
7. Signature displays correctly

### HTML Validation
1. Well-formed HTML
2. Proper MIME boundaries
3. Quoted-printable encoding valid
4. No broken tags

## Rollback

### Quick Rollback
```bash
git checkout HEAD~1 -- pkg/templates/taskOrderConfirmation.tpl
```

Result: Old email content with new signature (if signature was updated separately)

## Migration Notes

### Removed Template Functions
These functions from old template are NO LONGER NEEDED:
- `periodEndDay` - Not used in new template
- `monthName` - Not used in new template  
- `year` - Not used in new template

### Removed Content
- Client list with countries - Replaced with milestones
- "Please reply to confirm" - Removed
- "All tasks tracked in Notion/Jira" - Removed
- "Period: 01 – XX Month, Year" - Removed
