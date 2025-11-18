package basecamp

import (
	"github.com/dwarvesf/fortress-api/pkg/service/basecamp/model"
)

// ExpenseAdapter wraps the Basecamp service to implement the ExpenseProvider interface.
// This adapter delegates to the existing Todo service methods.
type ExpenseAdapter struct {
	svc *Service
}

// NewExpenseAdapter creates a new Basecamp expense provider adapter.
func NewExpenseAdapter(svc *Service) *ExpenseAdapter {
	return &ExpenseAdapter{svc: svc}
}

// GetAllInList fetches all expense todos in a specific list.
// Delegates to Basecamp Todo service, omitting optional query parameters.
func (e *ExpenseAdapter) GetAllInList(todolistID, projectID int) ([]model.Todo, error) {
	return e.svc.Todo.GetAllInList(todolistID, projectID)
}

// GetGroups fetches expense groups (e.g., "Out" group in accounting).
// Delegates to Basecamp Todo service.
func (e *ExpenseAdapter) GetGroups(todosetID, projectID int) ([]model.TodoGroup, error) {
	return e.svc.Todo.GetGroups(todosetID, projectID)
}

// GetLists fetches expense lists/views.
// Delegates to Basecamp Todo service.
func (e *ExpenseAdapter) GetLists(projectID, todosetID int) ([]model.TodoList, error) {
	return e.svc.Todo.GetLists(projectID, todosetID)
}
