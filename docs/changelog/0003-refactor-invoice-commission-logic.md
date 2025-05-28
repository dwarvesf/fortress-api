## Refactor Invoice Commission Calculation Logic

**Summary:**

Refactored the internal implementation of the invoice commission calculation endpoint to improve code structure, maintainability, and adherence to the project's architectural patterns.

**Details:**

Previously, the logic for calculating and saving invoice commissions had direct database access calls within the handler layer. This update moves all database interaction into the designated store layer and centralizes the business logic and transaction management within the controller layer.

This change is primarily an internal code quality improvement and does not introduce new features or alter the external behavior of the API endpoint (`POST /invoices/{id}/calculate-commissions`). It enhances the codebase's robustness and makes future development and testing easier.

**Impact:**

No direct impact on API users. Developers working on the invoice commission logic will benefit from a cleaner, more structured codebase.