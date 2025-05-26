# Specification: Consistent Notion Project Head Email Syncing

## Problem

The current implementation for fetching project head information from Notion in `pkg/service/notion/notion.go` (`GetProjectHeadDisplayNames`) extracts display names for Sales Person, Tech Lead, and Account Managers, but extracts emails for Deal Closing. This inconsistency can lead to issues when syncing this information to the internal database, as the dependent handler (`SyncProjectHeadsFromNotion` in `pkg/handler/project/project.go`) expects to work with names for some roles and emails for others. To ensure correct syncing and data integrity, all project head roles should be extracted as emails from Notion if available.

## Proposal

Modify the `GetProjectHeadDisplayNames` function in `pkg/service/notion/notion.go` to extract emails for the 'Source' (Sales Person), 'PM/Delivery' (Tech Lead), and 'Closing' (Account Managers) Notion properties, using similar logic to the existing extraction for 'Deal Closing' emails. This involves:

1.  Updating the function signature to return email strings instead of names (e.g., `salePersonEmail`, `techLeadEmail`, `accountManagerEmails`).
2.  Refactoring the extraction logic for 'Source', 'PM/Delivery', and 'Closing' properties to use the `extractEmailFromOptionName` helper function if these properties are Multi-Select type and contain emails in their option names. If the properties are not Multi-Select or do not contain emails in the expected format, the function should handle these cases gracefully (e.g., return an empty string or log a warning).

Subsequently, update the `SyncProjectHeadsFromNotion` handler in `pkg/handler/project/project.go` to consume the email addresses returned by the modified `GetProjectHeadDisplayNames` function and use them for syncing project heads in the database. This will involve:

1.  Updating the call to `notionService.GetProjectHeadDisplayNames`.
2.  Adjusting the logic that maps the returned email addresses to internal employee records for syncing with the `model.ProjectHead` table.

## Expected Outcome

After implementing these changes, the system will consistently extract email addresses for all supported project head roles from Notion. The `SyncProjectHeadsFromNotion` handler will then use these email addresses to accurately sync project head information to the database, improving data consistency and reliability.

## Dependencies

- Understanding the exact Notion property types for 'Source', 'PM/Delivery', and 'Closing' to ensure the email extraction logic is applicable. (This needs to be verified during the implementation phase).
