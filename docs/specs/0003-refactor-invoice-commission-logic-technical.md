## Specification: Technical Details of Invoice Commission Logic Refactoring

**1. Overview**

This specification details the technical implementation of the refactoring applied to the invoice commission calculation logic within the Fortress API. The primary goal is to align the implementation with the project's layered architecture by relocating database access from the handler to the store and controller layers.

**2. Motivation**

The original implementation in the `CalculateCommissions` handler (`pkg/handler/invoice/invoice.go`) directly interacted with the database (`h.repo.DB()`), bypassing the `store` and `controller` layers for certain operations. This tightly coupled the handler to the database implementation, reducing maintainability, testability, and violating the principle of separation of concerns.

**3. Technical Changes**

-   **Store Layer (`pkg/store/...`):**
    -   Introduced `DeleteUnpaidByInvoiceID(db *gorm.DB, invoiceID string) error` methods in `pkg/store/employeecommission/employee_commission.go` and `pkg/store/inboundfundtransaction/inbound_fund_transaction.go` to handle the deletion of unpaid commissions and inbound fund transactions associated with a specific invoice within a given transaction.
    -   Ensured existing `Create` methods in the store layer can accept a `*gorm.DB` instance to participate in transactions managed by the controller.
    -   Confirmed the `One` method in `pkg/store/invoice/invoice.go` supports the necessary preloads (`Project`, `Project.Heads`, `Project.BankAccount`, `Project.BankAccount.Currency`) for fetching invoice data required for commission calculation.

-   **Controller Layer (`pkg/controller/invoice/commission.go`):**
    -   Created a new public method `ProcessCommissions(invoiceID string, dryRun bool, l logger.Logger) ([]model.EmployeeCommission, error)`. This method encapsulates the core business logic and transaction management.
    -   Inside `ProcessCommissions`:
        -   A database transaction is initiated using `c.repo.DB().Begin()`.
        -   The invoice is fetched using the `store.Invoice.One` method with the transaction instance.
        -   The commission calculation logic (`calculateCommissionFromInvoice`) is called.
        -   If `dryRun` is false:
            -   Existing unpaid commissions and inbound fund transactions for the invoice are deleted using the new store methods (`DeleteUnpaidByInvoiceID`) within the transaction.
            -   Inbound fund transactions are created using the `store.InboundFundTransaction.Create` method with the transaction.
            -   Employee commissions (excluding inbound fund ones) are created using the `store.EmployeeCommission.Create` method with the transaction.
        -   The transaction is committed on success or rolled back on error.
    -   The `IController` interface (`pkg/controller/invoice/new.go`) was updated to include the `ProcessCommissions` method.

-   **Handler Layer (`pkg/handler/invoice/invoice.go`):**
    -   The `CalculateCommissions` handler method was updated to delegate the entire process to the `controller.Invoice.ProcessCommissions` method.
    -   All direct database access code (fetching invoice, beginning/committing transactions, deleting/creating records) was removed from the handler.

**4. Alternatives Considered**

-   Leaving the database access in the handler: Rejected due to violation of architectural principles and reduced maintainability/testability.
-   Moving only parts of the database access to the store: Rejected to ensure complete separation of concerns and centralize transaction management in the controller.