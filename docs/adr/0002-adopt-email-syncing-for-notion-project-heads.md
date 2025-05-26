## Architectural Decision Record: Adopt Email-Based Syncing for Notion Project Heads

### Context
The system currently synchronizes project head information from Notion, but handles different roles inconsistently. Sales Person, Tech Lead, and Account Managers are synced using their display names, while Deal Closing is synced using email addresses. This inconsistency makes reliable mapping to internal user records challenging and error-prone, as our internal user identification primarily relies on email addresses.

### Decision
We have decided to standardize the synchronization process for all relevant Notion project head roles (Sales Person, Tech Lead, Account Managers, and Deal Closing) to use email addresses as the primary identifier. The Notion service layer will be updated to extract email addresses from the respective properties whenever available, using the same logic currently applied to Deal Closing.

### Reasoning
Using email addresses provides a consistent and reliable key for identifying individuals across Notion and our internal database. This approach simplifies the mapping process, reduces ambiguity, and ensures better data integrity compared to relying on potentially non-unique or changing display names. It aligns with our internal user identification strategy.

### Consequences
- The `GetProjectHeadDisplayNames` function in the Notion service needs to be updated to extract emails for all relevant properties and its return signature modified.
- The `SyncProjectHeadsFromNotion` handler needs to be updated to consume email addresses instead of names for Sales Person, Tech Lead, and Account Managers.
- Potential edge cases where email addresses might not be available in Notion properties need to be handled gracefully (e.g., by logging warnings or skipping the sync for that specific role).
- This change improves the accuracy and reliability of project head information in the internal database.