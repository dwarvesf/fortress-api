package employee

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/dwarvesf/fortress-api/pkg/mcp/view"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/employee"
)

// Tools represents employee-related MCP tools
type Tools struct {
	store *store.Store
	repo  store.DBRepo
}

// New creates a new employee tools instance
func New(store *store.Store, repo store.DBRepo) *Tools {
	return &Tools{
		store: store,
		repo:  repo,
	}
}

// GetEmployeeTool returns the MCP tool for getting an employee by ID
func (t *Tools) GetEmployeeTool() mcp.Tool {
	return mcp.NewTool(
		"get_employee",
		mcp.WithDescription("Retrieve employee details by ID"),
		mcp.WithString("employee_id", mcp.Required(), mcp.Description("UUID of the employee to retrieve")),
	)
}

// GetEmployeeHandler handles the get_employee tool execution
func (t *Tools) GetEmployeeHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	employeeIDStr, err := req.RequireString("employee_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	employeeID, err := model.UUIDFromString(employeeIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid employee_id format: %v", err)), nil
	}

	employee, err := t.store.Employee.One(t.repo.DB(), employeeID.String(), false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get employee: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Employee: %s (%s)", employee.FullName, employee.ID)), nil
}

// ListAvailableEmployeesTool returns the MCP tool for listing available employees by role
func (t *Tools) ListAvailableEmployeesTool() mcp.Tool {
	return mcp.NewTool(
		"list_available_employees",
		mcp.WithDescription("Find available employees by role and status"),
		mcp.WithString("role", mcp.Description("Employee role/position to filter by (optional)")),
		mcp.WithString("status", mcp.Description("Employee working status (optional, default: full-time)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of employees to return (optional, default: 10)")),
	)
}

// ListAvailableEmployeesHandler handles the list_available_employees tool execution
func (t *Tools) ListAvailableEmployeesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Set default values
	role := req.GetString("role", "")
	status := req.GetString("status", "full-time")
	limit := int(req.GetFloat("limit", 10.0))

	// Build filter with available options
	filter := employee.EmployeeFilter{
		WorkingStatuses: []string{status},
	}

	// Get employees from store
	employees, _, err := t.store.Employee.All(t.repo.DB(), filter, model.Pagination{Page: 1, Size: int64(limit)})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list employees: %v", err)), nil
	}

	// Apply role filter if specified (client-side filtering for now)
	var filteredEmployees []*model.Employee
	for _, emp := range employees {
		if role == "" || t.hasRole(emp, role) {
			filteredEmployees = append(filteredEmployees, emp)
			if len(filteredEmployees) >= limit {
				break
			}
		}
	}

	// Format and return result using utility function
	return view.FormatJSONResponse(filteredEmployees)
}

// UpdateEmployeeStatusTool returns the MCP tool for updating employee status
func (t *Tools) UpdateEmployeeStatusTool() mcp.Tool {
	return mcp.NewTool(
		"update_employee_status",
		mcp.WithDescription("Update employee working status"),
		mcp.WithString("employee_id", mcp.Required(), mcp.Description("UUID of the employee")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New working status (full-time, part-time, contractor, probation, left, on-boarding)")),
	)
}

// UpdateEmployeeStatusHandler handles the update_employee_status tool execution
func (t *Tools) UpdateEmployeeStatusHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	employeeIDStr, err := req.RequireString("employee_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	employeeID, err := model.UUIDFromString(employeeIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid employee_id format: %v", err)), nil
	}

	// Validate status
	validStatuses := []string{"full-time", "part-time", "contractor", "probation", "left", "on-boarding"}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s. Valid statuses: %v", status, validStatuses)), nil
	}

	// Get current employee
	employee, err := t.store.Employee.One(t.repo.DB(), employeeID.String(), false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get employee: %v", err)), nil
	}

	// Update status
	employee.WorkingStatus = model.WorkingStatus(status)

	// Save changes
	updatedEmployee, err := t.store.Employee.Update(t.repo.DB(), employee)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update employee status: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Employee status updated to %s for %s", status, updatedEmployee.FullName)), nil
}

// GetEmployeeSkillsTool returns the MCP tool for getting employee skills/stacks
func (t *Tools) GetEmployeeSkillsTool() mcp.Tool {
	return mcp.NewTool(
		"get_employee_skills",
		mcp.WithDescription("Retrieve employee technology stacks and skills"),
		mcp.WithString("employee_id", mcp.Required(), mcp.Description("UUID of the employee")),
	)
}

// GetEmployeeSkillsHandler handles the get_employee_skills tool execution
func (t *Tools) GetEmployeeSkillsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	employeeIDStr, err := req.RequireString("employee_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	employeeID, err := model.UUIDFromString(employeeIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid employee_id format: %v", err)), nil
	}

	// Get employee with stacks
	employee, err := t.store.Employee.One(t.repo.DB(), employeeID.String(), true) // true to include stacks
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get employee: %v", err)), nil
	}

	result := map[string]interface{}{
		"employee_id": employee.ID,
		"name":        employee.FullName,
		"stacks":      employee.EmployeeStacks,
		"positions":   employee.Positions,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Employee skills for %s: %+v", employee.FullName, result)), nil
}

// Helper function to check if employee has a specific role
func (t *Tools) hasRole(employee *model.Employee, role string) bool {
	for _, position := range employee.Positions {
		if position.Name == role || position.Code == role {
			return true
		}
	}
	return false
}
