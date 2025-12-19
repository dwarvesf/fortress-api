# Fill Timesheet Contractor Script

This script automatically fills the `Contractor` and `Discord` fields in the Timesheet database by matching the `Created By` field with contractors from the Contractor database.

## How It Works

1. **Fetches all contractors** from the Contractor database (ID: `9d468753ebb44977a8dc156428398a6b`)
2. **Builds a mapping** from Notion Person IDs (from the `Person` field) to contractor information
3. **Fetches all timesheet entries** from the Timesheet database (ID: `2c664b29b84c8089b304e9c5b5c70ac3`)
4. **For each timesheet entry**:
   - Gets the `Created By` user ID
   - Looks up the corresponding contractor using the Person mapping
   - Updates the `Contractor` relation field with the contractor's page ID
   - Updates the `Discord` select field with the contractor's Discord username

## Prerequisites

- Go 1.16 or higher
- `NOTION_SECRET` environment variable set (can be in `.env` file)
- Notion integration must have access to both Timesheet and Contractor databases

## Usage

### Dry Run (Preview Changes)

```bash
# Preview all changes without making updates
go run main.go -dry-run

# Preview changes for a specific username
go run main.go -dry-run -username=vincent
```

### Fill All Missing Data

```bash
# Fill contractor and discord data for all timesheet entries
go run main.go

# Fill data only for entries created by a specific username
go run main.go -username=vincent
```

## Command Line Flags

- `-dry-run`: Run without making changes to Notion (default: `false`)
- `-username`: Only process timesheet entries created by this Notion username (optional)

## Examples

```bash
# Dry run to see what would be updated
go run main.go -dry-run

# Update all timesheet entries missing contractor/discord data
go run main.go

# Update only entries created by a specific username
go run main.go -username=vincent

# Dry run for a specific username
go run main.go -dry-run -username=phucld
```

## Output

The script logs:
- **DEBUG**: Detailed information about the process
- **INFO**: Successful updates
- **ERROR**: Failed updates

### Summary Statistics

At the end of execution, the script displays:
- Total timesheet entries processed
- Updated: Number of entries that were updated
- Already set: Number of entries that already had data
- Not found: Number of entries where no matching contractor was found
- Skipped: Number of entries without a `Created By` user

## Matching Logic

The script matches timesheet entries to contractors using:
1. **Timesheet `Created By`** → User ID from the `created_by` field
2. **Contractor `Person`** → Array of Notion Users (ID + name) in the `Person` field

When filtering by username:
- Searches for Notion username in the Contractor's `Person` field
- Returns all person IDs associated with that Notion user

When a match is found:
- **Contractor field**: Set to the contractor's page ID (creates a relation)
- **Discord field**: Set to the contractor's Discord username (if available)

## Notes

- The script respects Notion's rate limits (~3 requests/sec)
- Only updates entries where Contractor or Discord fields are empty
- Skips entries that already have both fields filled
- Safe to run multiple times (idempotent)
