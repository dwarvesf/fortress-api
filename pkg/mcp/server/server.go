package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/mcp/auth"
	"github.com/dwarvesf/fortress-api/pkg/mcp/tools/employee"
	"github.com/dwarvesf/fortress-api/pkg/mcp/tools/invoice"
	"github.com/dwarvesf/fortress-api/pkg/mcp/tools/payroll"
	"github.com/dwarvesf/fortress-api/pkg/mcp/tools/project"
	"github.com/dwarvesf/fortress-api/pkg/mcp/tools/workflow"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/store"
	"github.com/dwarvesf/fortress-api/pkg/store/agentactionlog"
	"github.com/dwarvesf/fortress-api/pkg/store/agentapikey"
)

// MCPServer represents the MCP server for fortress-api
type MCPServer struct {
	server         *server.MCPServer
	db             *gorm.DB
	cfg            *config.Config
	logger         logger.Logger
	authService    *auth.Service
	actionLogStore agentactionlog.IStore
	store          *store.Store
	repo           store.DBRepo

	// Service dependencies
	services *service.Service
}

// Config holds the configuration for the MCP server
type Config struct {
	Name    string
	Version string
	Port    int
}

// New creates a new MCP server instance
func New(cfg *config.Config, db *gorm.DB, logger logger.Logger, store *store.Store, repo store.DBRepo, services *service.Service) *MCPServer {
	// Initialize stores
	agentKeyStore := agentapikey.New(db)
	actionLogStore := agentactionlog.New(db)

	// Initialize auth service
	authService := auth.New(agentKeyStore)

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"Fortress API",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	return &MCPServer{
		server:         mcpServer,
		db:             db,
		cfg:            cfg,
		logger:         logger,
		authService:    authService,
		actionLogStore: actionLogStore,
		store:          store,
		repo:           repo,
		services:       services,
	}
}

// RegisterTools registers all MCP tools
func (s *MCPServer) RegisterTools() error {
	// Register employee management tools
	if err := s.registerEmployeeTools(); err != nil {
		return fmt.Errorf("failed to register employee tools: %w", err)
	}

	// Register project management tools
	if err := s.registerProjectTools(); err != nil {
		return fmt.Errorf("failed to register project tools: %w", err)
	}

	// Register invoice management tools
	if err := s.registerInvoiceTools(); err != nil {
		return fmt.Errorf("failed to register invoice tools: %w", err)
	}

	// Register payroll tools
	if err := s.registerPayrollTools(); err != nil {
		return fmt.Errorf("failed to register payroll tools: %w", err)
	}

	// Register workflow tools
	if err := s.registerWorkflowTools(); err != nil {
		return fmt.Errorf("failed to register workflow tools: %w", err)
	}

	return nil
}

// Serve starts the MCP server
func (s *MCPServer) Serve() error {
	s.logger.Info("Starting MCP server...")

	// Register all tools
	if err := s.RegisterTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	s.logger.Info("MCP tools registered successfully")
	s.logger.Info("MCP server ready to accept connections")

	// Serve using stdio transport
	httpServer := server.NewStreamableHTTPServer(s.server)
	s.logger.Info("HTTP server listening on :8080/mcp")
	if err := httpServer.Start(":8080"); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	return nil
}

// wrapToolWithAuth wraps a tool handler with authentication and logging
func (s *MCPServer) wrapToolWithAuth(toolName string, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		// For now, use the existing test agent key for MCP context
		// In production, implement proper MCP authentication mechanism
		
		// Use the existing test agent key ID from database
		agentKey := &model.AgentAPIKey{
			BaseModel: model.BaseModel{ID: model.MustGetUUIDFromString("5dc80b39-9203-46c9-87ab-40a0f9c752cc")},
			Name:      "test-mcp-agent",
		}

		// Add agent to context
		ctx = auth.SetAgentInContext(ctx, agentKey)

		// Execute the tool
		result, err := handler(ctx, req)

		// Log the action
		duration := time.Since(startTime)
		status := model.AgentActionLogStatusSuccess
		var errorMessage *string

		if err != nil {
			status = model.AgentActionLogStatusError
			errMsg := err.Error()
			errorMessage = &errMsg
		}

		// Create action log
		actionLog := &model.AgentActionLog{
			AgentKeyID:   agentKey.ID,
			ToolName:     toolName,
			Status:       status,
			DurationMs:   func() *int { d := int(duration.Milliseconds()); return &d }(),
			ErrorMessage: errorMessage,
		}

		// Store input and output data
		// TODO: Convert request and result to JSON for storage

		// Save the log (ignore errors to not break the tool execution)
		if _, logErr := s.actionLogStore.Create(context.Background(), actionLog); logErr != nil {
			s.logger.Fields(logger.Fields{"error": logErr}).Error(logErr, "Failed to create action log")
		}

		// Log the tool execution
		if err != nil {
			s.logger.Fields(logger.Fields{"tool": toolName, "duration": duration}).Error(err, "Tool execution failed")
		} else {
			s.logger.Fields(logger.Fields{"tool": toolName, "duration": duration}).Info("Tool executed successfully")
		}

		return result, err
	}
}

// Tool registration methods
func (s *MCPServer) registerEmployeeTools() error {
	// Create employee tools instance
	employeeTools := employee.New(s.store, s.repo)

	// Register get_employee tool
	getEmployeeTool := employeeTools.GetEmployeeTool()
	getEmployeeHandler := s.wrapToolWithAuth("get_employee", employeeTools.GetEmployeeHandler)
	s.server.AddTool(getEmployeeTool, getEmployeeHandler)

	// Register list_available_employees tool
	listEmployeesTool := employeeTools.ListAvailableEmployeesTool()
	listEmployeesHandler := s.wrapToolWithAuth("list_available_employees", employeeTools.ListAvailableEmployeesHandler)
	s.server.AddTool(listEmployeesTool, listEmployeesHandler)

	// Register update_employee_status tool
	updateStatusTool := employeeTools.UpdateEmployeeStatusTool()
	updateStatusHandler := s.wrapToolWithAuth("update_employee_status", employeeTools.UpdateEmployeeStatusHandler)
	s.server.AddTool(updateStatusTool, updateStatusHandler)

	// Register get_employee_skills tool
	getSkillsTool := employeeTools.GetEmployeeSkillsTool()
	getSkillsHandler := s.wrapToolWithAuth("get_employee_skills", employeeTools.GetEmployeeSkillsHandler)
	s.server.AddTool(getSkillsTool, getSkillsHandler)

	s.logger.Info("Employee tools registered successfully")
	return nil
}

func (s *MCPServer) registerProjectTools() error {
	// Create project tools instance
	projectTools := project.New(s.store, s.repo)

	// Register create_project tool
	createProjectTool := projectTools.CreateProjectTool()
	createProjectHandler := s.wrapToolWithAuth("create_project", projectTools.CreateProjectHandler)
	s.server.AddTool(createProjectTool, createProjectHandler)

	// Register get_project_details tool
	getProjectDetailsTool := projectTools.GetProjectDetailsTool()
	getProjectDetailsHandler := s.wrapToolWithAuth("get_project_details", projectTools.GetProjectDetailsHandler)
	s.server.AddTool(getProjectDetailsTool, getProjectDetailsHandler)

	// Register assign_project_member tool
	assignMemberTool := projectTools.AssignProjectMemberTool()
	assignMemberHandler := s.wrapToolWithAuth("assign_project_member", projectTools.AssignProjectMemberHandler)
	s.server.AddTool(assignMemberTool, assignMemberHandler)

	// Register get_project_members tool
	getMembersTool := projectTools.GetProjectMembersTool()
	getMembersHandler := s.wrapToolWithAuth("get_project_members", projectTools.GetProjectMembersHandler)
	s.server.AddTool(getMembersTool, getMembersHandler)

	// Register update_project_status tool
	updateStatusTool := projectTools.UpdateProjectStatusTool()
	updateStatusHandler := s.wrapToolWithAuth("update_project_status", projectTools.UpdateProjectStatusHandler)
	s.server.AddTool(updateStatusTool, updateStatusHandler)

	s.logger.Info("Project tools registered successfully")
	return nil
}

func (s *MCPServer) registerInvoiceTools() error {
	// Create invoice tools instance
	invoiceTools := invoice.New(s.store, s.repo)
	
	// Register generate_invoice tool
	generateInvoiceTool := invoiceTools.GenerateInvoiceTool()
	generateInvoiceHandler := s.wrapToolWithAuth("generate_invoice", invoiceTools.GenerateInvoiceHandler)
	s.server.AddTool(generateInvoiceTool, generateInvoiceHandler)
	
	// Register get_invoice_status tool
	getInvoiceStatusTool := invoiceTools.GetInvoiceStatusTool()
	getInvoiceStatusHandler := s.wrapToolWithAuth("get_invoice_status", invoiceTools.GetInvoiceStatusHandler)
	s.server.AddTool(getInvoiceStatusTool, getInvoiceStatusHandler)
	
	// Register update_invoice_status tool
	updateInvoiceStatusTool := invoiceTools.UpdateInvoiceStatusTool()
	updateInvoiceStatusHandler := s.wrapToolWithAuth("update_invoice_status", invoiceTools.UpdateInvoiceStatusHandler)
	s.server.AddTool(updateInvoiceStatusTool, updateInvoiceStatusHandler)
	
	// Register calculate_commission tool
	calculateCommissionTool := invoiceTools.CalculateCommissionTool()
	calculateCommissionHandler := s.wrapToolWithAuth("calculate_commission", invoiceTools.CalculateCommissionHandler)
	s.server.AddTool(calculateCommissionTool, calculateCommissionHandler)
	
	s.logger.Info("Invoice tools registered successfully")
	return nil
}

func (s *MCPServer) registerPayrollTools() error {
	// Create payroll tools instance
	payrollTools := payroll.New(s.store, s.repo)
	
	// Register calculate_payroll tool
	calculatePayrollTool := payrollTools.CalculatePayrollTool()
	calculatePayrollHandler := s.wrapToolWithAuth("calculate_payroll", payrollTools.CalculatePayrollHandler)
	s.server.AddTool(calculatePayrollTool, calculatePayrollHandler)
	
	// Register process_salary_advance tool
	processSalaryAdvanceTool := payrollTools.ProcessSalaryAdvanceTool()
	processSalaryAdvanceHandler := s.wrapToolWithAuth("process_salary_advance", payrollTools.ProcessSalaryAdvanceHandler)
	s.server.AddTool(processSalaryAdvanceTool, processSalaryAdvanceHandler)
	
	// Register get_payroll_summary tool
	getPayrollSummaryTool := payrollTools.GetPayrollSummaryTool()
	getPayrollSummaryHandler := s.wrapToolWithAuth("get_payroll_summary", payrollTools.GetPayrollSummaryHandler)
	s.server.AddTool(getPayrollSummaryTool, getPayrollSummaryHandler)
	
	s.logger.Info("Payroll tools registered successfully")
	return nil
}

func (s *MCPServer) registerWorkflowTools() error {
	// Create workflow tools instance
	workflowTools := workflow.New(s.cfg, s.store, s.repo, s.services.Wise)
	
	// Register calculate_monthly_payroll tool
	calculateMonthlyPayrollTool := workflowTools.CalculateMonthlyPayrollTool()
	calculateMonthlyPayrollHandler := s.wrapToolWithAuth("calculate_monthly_payroll", workflowTools.CalculateMonthlyPayrollHandler)
	s.server.AddTool(calculateMonthlyPayrollTool, calculateMonthlyPayrollHandler)
	
	// Register generate_financial_report tool
	generateFinancialReportTool := workflowTools.GenerateFinancialReportTool()
	generateFinancialReportHandler := s.wrapToolWithAuth("generate_financial_report", workflowTools.GenerateFinancialReportHandler)
	s.server.AddTool(generateFinancialReportTool, generateFinancialReportHandler)
	
	s.logger.Info("Workflow tools registered successfully")
	return nil
}
