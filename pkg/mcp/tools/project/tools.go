package project

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/shopspring/decimal"

	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/store"
)

// Tools represents project-related MCP tools
type Tools struct {
	store *store.Store
	repo  store.DBRepo
}

// New creates a new project tools instance
func New(store *store.Store, repo store.DBRepo) *Tools {
	return &Tools{
		store: store,
		repo:  repo,
	}
}

// CreateProjectTool returns the MCP tool for creating a new project
func (t *Tools) CreateProjectTool() mcp.Tool {
	return mcp.NewTool(
		"create_project",
		mcp.WithDescription("Create a new project with basic information"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
		mcp.WithString("code", mcp.Required(), mcp.Description("Project code/identifier")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Project type (dwarves, fixed-cost, time-material)")),
		mcp.WithString("function", mcp.Required(), mcp.Description("Project function (development, learning, training, management)")),
		mcp.WithString("status", mcp.Description("Project status (on-boarding, active, paused, closed), default: on-boarding")),
		mcp.WithString("client_email", mcp.Description("Client email address")),
		mcp.WithString("start_date", mcp.Description("Project start date (YYYY-MM-DD format)")),
		mcp.WithString("end_date", mcp.Description("Project end date (YYYY-MM-DD format)")),
	)
}

// CreateProjectHandler handles the create_project tool execution
func (t *Tools) CreateProjectHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := req.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	code, err := req.RequireString("code")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	projectType, err := req.RequireString("type")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	function, err := req.RequireString("function")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status := req.GetString("status", "on-boarding")
	clientEmail := req.GetString("client_email", "")
	startDate := req.GetString("start_date", "")
	endDate := req.GetString("end_date", "")

	// Validate project type
	validTypes := []string{"dwarves", "fixed-cost", "time-material"}
	if !contains(validTypes, projectType) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project type: %s. Valid types: %v", projectType, validTypes)), nil
	}

	// Validate project function
	validFunctions := []string{"development", "learning", "training", "management"}
	if !contains(validFunctions, function) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project function: %s. Valid functions: %v", function, validFunctions)), nil
	}

	// Validate project status
	validStatuses := []string{"on-boarding", "active", "paused", "closed"}
	if !contains(validStatuses, status) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project status: %s. Valid statuses: %v", status, validStatuses)), nil
	}

	// Check if project code already exists
	exists, err := t.store.Project.IsExistByCode(t.repo.DB(), code)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to check project existence: %v", err)), nil
	}
	if exists {
		return mcp.NewToolResultError(fmt.Sprintf("project with code '%s' already exists", code)), nil
	}

	// Create project
	newProject := &model.Project{
		Name:        name,
		Code:        code,
		Type:        model.ProjectType(projectType),
		Function:    model.ProjectFunction(function),
		Status:      model.ProjectStatus(status),
		ClientEmail: clientEmail,
	}

	// Handle dates if provided
	if startDate != "" {
		if startTime, err := parseDate(startDate); err == nil {
			newProject.StartDate = &startTime
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("invalid start_date format: %v", err)), nil
		}
	}

	if endDate != "" {
		if endTime, err := parseDate(endDate); err == nil {
			newProject.EndDate = &endTime
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("invalid end_date format: %v", err)), nil
		}
	}

	err = t.store.Project.Create(t.repo.DB(), newProject)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create project: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Project created successfully: %s (ID: %s, Code: %s)", newProject.Name, newProject.ID, newProject.Code)), nil
}

// GetProjectDetailsTool returns the MCP tool for getting project details
func (t *Tools) GetProjectDetailsTool() mcp.Tool {
	return mcp.NewTool(
		"get_project_details",
		mcp.WithDescription("Retrieve comprehensive project information"),
		mcp.WithString("identifier", mcp.Required(), mcp.Description("Project ID (UUID) or project code")),
		mcp.WithString("include_members", mcp.Description("Include project members in response (true/false, default: true)")),
	)
}

// GetProjectDetailsHandler handles the get_project_details tool execution
func (t *Tools) GetProjectDetailsHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	identifier, err := req.RequireString("identifier")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	includeMembersStr := req.GetString("include_members", "true")
	includeMembers := includeMembersStr == "true"

	// Try to get project by ID first, then by code
	project, err := t.store.Project.One(t.repo.DB(), identifier, includeMembers)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
	}

	result := map[string]interface{}{
		"id":            project.ID,
		"name":          project.Name,
		"code":          project.Code,
		"type":          project.Type,
		"function":      project.Function,
		"status":        project.Status,
		"client_email":  project.ClientEmail,
		"start_date":    project.StartDate,
		"end_date":      project.EndDate,
		"organization":  project.Organization,
		"company_info":  project.CompanyInfo,
		"bank_account":  project.BankAccount,
		"account_rating": project.AccountRating,
		"delivery_rating": project.DeliveryRating,
		"lead_rating":   project.LeadRating,
	}

	if includeMembers {
		result["members"] = project.ProjectMembers
		result["heads"] = project.Heads
	}

	return mcp.NewToolResultText(fmt.Sprintf("Project details: %+v", result)), nil
}

// AssignProjectMemberTool returns the MCP tool for assigning an employee to a project
func (t *Tools) AssignProjectMemberTool() mcp.Tool {
	return mcp.NewTool(
		"assign_project_member",
		mcp.WithDescription("Assign employee to project role"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
		mcp.WithString("employee_id", mcp.Required(), mcp.Description("Employee ID (UUID)")),
		mcp.WithString("deployment_type", mcp.Description("Deployment type (official, shadow, part-time), default: official")),
		mcp.WithString("status", mcp.Description("Member status (pending, on-boarding, active, inactive), default: pending")),
		mcp.WithNumber("rate", mcp.Description("Hourly rate for this assignment")),
		mcp.WithString("start_date", mcp.Description("Assignment start date (YYYY-MM-DD format)")),
		mcp.WithString("end_date", mcp.Description("Assignment end date (YYYY-MM-DD format)")),
	)
}

// AssignProjectMemberHandler handles the assign_project_member tool execution
func (t *Tools) AssignProjectMemberHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIDStr, err := req.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	employeeIDStr, err := req.RequireString("employee_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}


	deploymentType := req.GetString("deployment_type", "official")
	status := req.GetString("status", "pending")
	rate := req.GetFloat("rate", 0.0)
	startDate := req.GetString("start_date", "")
	endDate := req.GetString("end_date", "")

	// Validate UUIDs
	projectID, err := model.UUIDFromString(projectIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project_id format: %v", err)), nil
	}

	employeeID, err := model.UUIDFromString(employeeIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid employee_id format: %v", err)), nil
	}

	// Validate deployment type
	validDeploymentTypes := []string{"official", "shadow", "part-time"}
	if !contains(validDeploymentTypes, deploymentType) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid deployment_type: %s. Valid types: %v", deploymentType, validDeploymentTypes)), nil
	}

	// Validate status
	validStatuses := []string{"pending", "on-boarding", "active", "inactive"}
	if !contains(validStatuses, status) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s. Valid statuses: %v", status, validStatuses)), nil
	}

	// Verify project exists
	exists, err := t.store.Project.IsExist(t.repo.DB(), projectID.String())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to verify project existence: %v", err)), nil
	}
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("project with ID %s does not exist", projectID)), nil
	}

	// Verify employee exists
	_, err = t.store.Employee.One(t.repo.DB(), employeeID.String(), false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("employee with ID %s does not exist: %v", employeeID, err)), nil
	}

	// Create project member assignment
	newMember := &model.ProjectMember{
		ProjectID:      projectID,
		EmployeeID:     employeeID,
		DeploymentType: model.DeploymentType(deploymentType),
		Status:         model.ProjectMemberStatus(status),
		Rate:           decimal.NewFromFloat(rate),
	}

	// Handle dates if provided
	if startDate != "" {
		if startTime, err := parseDate(startDate); err == nil {
			newMember.StartDate = &startTime
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("invalid start_date format: %v", err)), nil
		}
	}

	if endDate != "" {
		if endTime, err := parseDate(endDate); err == nil {
			newMember.EndDate = &endTime
		} else {
			return mcp.NewToolResultError(fmt.Sprintf("invalid end_date format: %v", err)), nil
		}
	}

	err = t.store.ProjectMember.Create(t.repo.DB(), newMember)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to assign project member: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Employee %s assigned to project %s", employeeID, projectID)), nil
}

// GetProjectMembersTool returns the MCP tool for listing project members
func (t *Tools) GetProjectMembersTool() mcp.Tool {
	return mcp.NewTool(
		"get_project_members",
		mcp.WithDescription("List all project team members"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
		mcp.WithString("status", mcp.Description("Filter by member status (pending, on-boarding, active, inactive)")),
		mcp.WithNumber("limit", mcp.Description("Maximum number of members to return (default: 50)")),
	)
}

// GetProjectMembersHandler handles the get_project_members tool execution
func (t *Tools) GetProjectMembersHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIDStr, err := req.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status := req.GetString("status", "")
	limit := int(req.GetFloat("limit", 50.0))

	// Validate UUID
	projectID, err := model.UUIDFromString(projectIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project_id format: %v", err)), nil
	}

	// Verify project exists
	exists, err := t.store.Project.IsExist(t.repo.DB(), projectID.String())
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to verify project existence: %v", err)), nil
	}
	if !exists {
		return mcp.NewToolResultError(fmt.Sprintf("project with ID %s does not exist", projectID)), nil
	}

	// Get project with members
	project, err := t.store.Project.One(t.repo.DB(), projectID.String(), true)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project members: %v", err)), nil
	}

	// Filter members by status if specified
	var filteredMembers []model.ProjectMember
	for _, member := range project.ProjectMembers {
		if status == "" || string(member.Status) == status {
			filteredMembers = append(filteredMembers, member)
			if len(filteredMembers) >= limit {
				break
			}
		}
	}

	result := map[string]interface{}{
		"project_id":     project.ID,
		"project_name":   project.Name,
		"members_count":  len(filteredMembers),
		"members":        filteredMembers,
	}

	return mcp.NewToolResultText(fmt.Sprintf("Project members for %s: %+v", project.Name, result)), nil
}

// UpdateProjectStatusTool returns the MCP tool for updating project status
func (t *Tools) UpdateProjectStatusTool() mcp.Tool {
	return mcp.NewTool(
		"update_project_status",
		mcp.WithDescription("Change project status"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID (UUID)")),
		mcp.WithString("status", mcp.Required(), mcp.Description("New project status (on-boarding, active, paused, closed)")),
	)
}

// UpdateProjectStatusHandler handles the update_project_status tool execution
func (t *Tools) UpdateProjectStatusHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectIDStr, err := req.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status, err := req.RequireString("status")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Validate UUID
	projectID, err := model.UUIDFromString(projectIDStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid project_id format: %v", err)), nil
	}

	// Validate status
	validStatuses := []string{"on-boarding", "active", "paused", "closed"}
	if !contains(validStatuses, status) {
		return mcp.NewToolResultError(fmt.Sprintf("invalid status: %s. Valid statuses: %v", status, validStatuses)), nil
	}

	// Get current project
	project, err := t.store.Project.One(t.repo.DB(), projectID.String(), false)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get project: %v", err)), nil
	}

	// Update status
	project.Status = model.ProjectStatus(status)

	_, err = t.store.Project.UpdateSelectedFieldsByID(t.repo.DB(), projectID.String(), *project, "status")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update project status: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Project status updated to %s for %s (%s)", status, project.Name, project.Code)), nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseDate(dateStr string) (time.Time, error) {
	// Parse date in YYYY-MM-DD format
	parsedTime, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be in YYYY-MM-DD format: %v", err)
	}
	return parsedTime, nil
}