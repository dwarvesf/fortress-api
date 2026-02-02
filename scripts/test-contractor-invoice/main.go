package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

func main() {
	fmt.Println("Starting test...")

	// Initialize logger
	l := logger.NewLogrusLogger("debug")
	fmt.Println("Logger initialized")

	// Load config
	cfg := loadConfig()
	fmt.Println("Config loaded")

	// Create controller
	fmt.Println("Creating controller...")
	ctrl := controller.New(nil, nil, nil, nil, l, cfg)
	fmt.Println("Controller created")

	// Generate single invoice with combined line items
	l.Info("Generating contractor invoice...")
	testData := &invoice.ContractorInvoiceData{
		InvoiceNumber:     "INVC-202512-TEST-0001",
		ContractorName:    "Test Contractor",
		Month:             "2025-12",
		Date:              time.Now(),
		DueDate:           time.Now().AddDate(0, 0, 15),
		Description:       "Software Development Services for December 2025",
		BillingType:       "Hourly Rate",
		Currency:          "USD",
		Total:             120000000,
		TotalUSD:          4800,
		MonthlyFixed:      5000000,
		MonthlyFixedUSD:   2400,
		ExchangeRate:      1,
		BankAccountHolder: "LE MINH QUANG",
		BankName:          "ASIA COMMERCIAL JOINT STOCK BANK",
		BankAccountNumber: "260470189",
		BankSwiftBIC:      "ASCBVNVX",
		BankBranch:        "ACB - PGD BINH TAN",
		LineItems: []invoice.ContractorInvoiceLineItem{
			{
				Title:       "Project Alpha",
				Description: "• Backend Infrastructure: Concurrent uploads, batch operations, environment refinement\n• Data Management: Data retention, search optimization, automated cleanup",
				Hours:       40,
				Rate:        25,
				Amount:      1000,
				AmountUSD:   1000,
			},
			{
				Title:       "Project Beta",
				Description: "• Invoice System: Webhook handlers, USDC support, Discount logic, Invoice layout\n• Backend Infrastructure: Webhook refactoring, Data extraction, Calendar integration",
				Hours:       36,
				Rate:        25,
				Amount:      900,
				AmountUSD:   900,
			},
			{
				Title:       "Project Gamma",
				Description: "• Invoice System: Webhook handlers, USDC support, Discount logic, Invoice layout\n• Backend Infrastructure: Webhook refactoring, Data extraction, Calendar integration",
				Hours:       0,
				Rate:        0,
				Amount:      0,
				AmountUSD:   0,
			},
			{
				Title:       "Project Tetha",
				Description: "• System Design: Architecture planning, Database schema design\n• Code Review: Pull request reviews, Code quality improvements",
				Hours:       0,
				Rate:        0,
				Amount:      0,
				AmountUSD:   0,
			},
			{
				Title:       "Bonus",
				Description: "",
				Hours:       0,
				Rate:        0,
				Amount:      500,
				AmountUSD:   500,
			},
		},
	}

	pdfBytes, err := ctrl.Invoice.GenerateContractorInvoicePDF(l, testData)
	if err != nil {
		l.Error(err, "Failed to generate PDF")
		os.Exit(1)
	}

	outputFile := "contractor-invoice-test.pdf"
	if err := os.WriteFile(outputFile, pdfBytes, 0644); err != nil {
		l.Error(err, "Failed to write PDF file")
		os.Exit(1)
	}
	l.Info(fmt.Sprintf("PDF generated: %s (%d bytes)", outputFile, len(pdfBytes)))
}

func loadConfig() *config.Config {
	cfg := &config.Config{
		Env: "local",
	}

	// Get template path from current working directory
	cwd, _ := os.Getwd()
	cfg.Invoice.TemplatePath = cwd + "/pkg/templates"

	return cfg
}
