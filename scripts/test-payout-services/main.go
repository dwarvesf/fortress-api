package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/notion"
	"github.com/dwarvesf/fortress-api/pkg/service/wise"
)

func main() {
	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	l := logger.NewLogrusLogger("debug")

	// Support both CONTRACTOR_PAGE_ID and CONTRACTOR_DISCORD
	contractorPageID := os.Getenv("CONTRACTOR_PAGE_ID")
	contractorDiscord := os.Getenv("CONTRACTOR_DISCORD")
	month := os.Getenv("MONTH")
	generatePDF := os.Getenv("GENERATE_PDF") == "true"

	if contractorPageID == "" && contractorDiscord == "" {
		l.Error(nil, "Either CONTRACTOR_PAGE_ID or CONTRACTOR_DISCORD env var is required")
		os.Exit(1)
	}

	if month == "" {
		month = time.Now().Format("2006-01")
		l.Debugf("Using current month: %s", month)
	}

	ctx := context.Background()

	// If Discord provided, look up contractor page ID
	if contractorPageID == "" && contractorDiscord != "" {
		l.Debugf("Looking up contractor by Discord: %s", contractorDiscord)
		pageID, _, err := lookupContractorByDiscord(ctx, cfg, l, contractorDiscord)
		if err != nil {
			l.Error(err, "Failed to lookup contractor by Discord")
			os.Exit(1)
		}
		contractorPageID = pageID
	}

	l.Debugf("Testing with contractor page ID: %s", contractorPageID)

	payoutsSvc := notion.NewContractorPayoutsService(cfg, l)
	splitSvc := notion.NewInvoiceSplitService(cfg, l)
	wiseSvc := wise.New(cfg, l)

	// Test QueryPendingPayoutsByContractor
	l.Debug("Querying pending payouts...")
	entries, err := payoutsSvc.QueryPendingPayoutsByContractor(ctx, contractorPageID, "")
	if err != nil {
		l.Error(err, "QueryPendingPayoutsByContractor failed")
		os.Exit(1)
	}

	fmt.Printf("Found %d pending payouts\n", len(entries))

	var totalUSD float64

	for _, e := range entries {
		fmt.Printf("- %s: %.2f %s (Name: %s)\n", e.SourceType, e.Amount, e.Currency, e.Name)

		// Convert amount to USD and round to 2 decimal places
		amountUSD := e.Amount
		if e.Currency != "" && e.Currency != "USD" {
			l.Debugf("Converting %.2f %s to USD", e.Amount, e.Currency)
			converted, rate, err := wiseSvc.Convert(e.Amount, e.Currency, "USD")
			if err != nil {
				l.Debugf("Failed to convert currency: %v (using original amount)", err)
			} else {
				amountUSD = converted
				l.Debugf("Converted: %.2f %s = %.2f USD (rate: %.4f)", e.Amount, e.Currency, amountUSD, rate)
			}
		}
		// Round to 2 decimal places to avoid $0.01 differences
		amountUSD = roundTo2Decimals(amountUSD)

		// Test GetInvoiceSplitByID if split relation exists
		if e.InvoiceSplitID != "" {
			l.Debugf("Fetching invoice split: %s", e.InvoiceSplitID)
			split, err := splitSvc.GetInvoiceSplitByID(ctx, e.InvoiceSplitID)
			if err != nil {
				l.Error(err, "GetInvoiceSplitByID failed")
			} else {
				fmt.Printf("  Split: %s - %.2f %s\n", split.Role, split.Amount, split.Currency)
			}
		}

		totalUSD += amountUSD
	}

	l.Info("Test completed successfully")

	// Generate PDF if requested
	if generatePDF {
		l.Debug("Generating invoice PDF using GenerateContractorInvoice controller...")

		// Create controller for invoice generation with Wise service
		svc := &service.Service{
			Wise: wiseSvc,
		}
		ctrl := controller.New(nil, nil, svc, nil, l, cfg)

		// Use GenerateContractorInvoice to get properly sorted line items
		invoiceData, err := ctrl.Invoice.GenerateContractorInvoice(ctx, contractorDiscord, month)
		if err != nil {
			l.Error(err, "Failed to generate contractor invoice data")
			os.Exit(1)
		}

		// Print sorted line items for verification
		fmt.Printf("\n=== Sorted Line Items (by Type, then Amount ASC) ===\n")
		for i, item := range invoiceData.LineItems {
			fmt.Printf("%d. [%s] %s - $%.2f\n", i+1, item.Type, item.Title, item.Amount)
		}
		fmt.Printf("Total: $%.2f\n", invoiceData.Total)

		pdfBytes, err := ctrl.Invoice.GenerateContractorInvoicePDF(l, invoiceData)
		if err != nil {
			l.Error(err, "Failed to generate PDF")
			os.Exit(1)
		}

		outputFile := fmt.Sprintf("contractor-invoice-%s-%s.pdf", sanitizeFilename(invoiceData.ContractorName), month)
		if err := os.WriteFile(outputFile, pdfBytes, 0644); err != nil {
			l.Error(err, "Failed to write PDF file")
			os.Exit(1)
		}
		l.Info(fmt.Sprintf("PDF generated: %s (%d bytes)", outputFile, len(pdfBytes)))
	}
}

func lookupContractorByDiscord(ctx context.Context, cfg *config.Config, l logger.Logger, discord string) (string, string, error) {
	client := nt.NewClient(cfg.Notion.Secret)
	contractorDBID := cfg.Notion.Databases.Contractor

	l.Debugf("Querying contractor database: %s for Discord: %s", contractorDBID, discord)

	filter := &nt.DatabaseQueryFilter{
		Property: "Discord",
		DatabaseQueryPropertyFilter: nt.DatabaseQueryPropertyFilter{
			RichText: &nt.TextPropertyFilter{
				Equals: discord,
			},
		},
	}

	query := &nt.DatabaseQuery{
		Filter:   filter,
		PageSize: 1,
	}

	resp, err := client.QueryDatabase(ctx, contractorDBID, query)
	if err != nil {
		return "", "", fmt.Errorf("query contractor database failed: %w", err)
	}

	if len(resp.Results) == 0 {
		return "", "", fmt.Errorf("no contractor found with Discord: %s", discord)
	}

	page := resp.Results[0]
	pageID := page.ID

	// Extract contractor name
	var name string
	if props, ok := page.Properties.(nt.DatabasePageProperties); ok {
		if nameProp, ok := props["Name"]; ok && len(nameProp.Title) > 0 {
			for _, rt := range nameProp.Title {
				name += rt.PlainText
			}
		}
	}

	l.Debugf("Found contractor page ID: %s, Name: %s", pageID, name)

	return pageID, name, nil
}

func sanitizeFilename(name string) string {
	// Replace spaces and special characters with underscores
	result := strings.ReplaceAll(name, " ", "_")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.ReplaceAll(result, "\\", "_")
	return result
}

func roundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}
