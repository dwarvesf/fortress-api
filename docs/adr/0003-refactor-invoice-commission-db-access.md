## Architectural Decision Record: Refactor Invoice Commission Database Access

**Status:** Accepted

**Context:**

The `CalculateCommissions` handler function in `pkg/handler/invoice/invoice.go` was directly accessing the database using `h.repo.DB()`. This approach violates the established project architecture, which dictates that database operations should be confined to the `pkg/store/` layer and business logic, including transaction management, should reside within the `pkg/controller/` layer.

**Decision:**

To adhere to the project's architectural principles and improve maintainability and testability, we will refactor the invoice commission calculation logic.

Database access operations previously in the handler will be moved to the appropriate store methods (`pkg/store/employeecommission/`, `pkg/store/inboundfundtransaction/`, `pkg/store/invoice/`).

A new controller method, `ProcessCommissions`, will be introduced in `pkg/controller/invoice/commission.go`. This method will orchestrate the database operations by calling the store methods and manage the database transaction for the commission calculation and saving process.

The `CalculateCommissions` handler will be updated to call the new `ProcessCommissions` controller method, removing all direct database interaction from the handler.

**Consequences:**

*   **Positive:**
    *   Improved adherence to the project's layered architecture.
    *   Enhanced separation of concerns, making the code easier to understand and maintain.
    *   Increased testability of the business logic (in the controller) and database operations (in the stores).
    *   Centralized transaction management within the controller.

*   **Negative:**
    *   Requires modifying existing code in the handler, controller, and store layers.
    *   Requires updating the controller interface.