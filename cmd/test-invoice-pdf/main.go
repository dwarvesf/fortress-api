package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/controller/invoice"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
	"github.com/dwarvesf/fortress-api/pkg/service"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
	"github.com/dwarvesf/fortress-api/pkg/store"
	invoiceStore "github.com/dwarvesf/fortress-api/pkg/store/invoice"
)

func main() {
	invoiceID := flag.String("invoice-id", "", "Invoice ID to generate PDF for (required)")
	outputPath := flag.String("output", "test-invoice.pdf", "Output PDF file path")
	flag.Parse()

	if *invoiceID == "" {
		fmt.Println("Error: --invoice-id is required")
		fmt.Println("\nUsage:")
		fmt.Println("  test-invoice-pdf --invoice-id=<uuid> [--output=test-invoice.pdf]")
		fmt.Println("\nExample:")
		fmt.Println("  test-invoice-pdf --invoice-id=123e4567-e89b-12d3-a456-426614174000 --output=invoice-test.pdf")
		os.Exit(1)
	}

	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger()

	log.Infof("Testing invoice PDF generation with patched wkhtmltopdf")
	log.Infof("Invoice ID: %s", *invoiceID)
	log.Infof("Output file: %s", *outputPath)

	v, err := vault.New(cfg)
	if err != nil {
		log.Error(err, "failed to init vault")
	}

	if v != nil {
		cfg = config.Generate(v)
	}

	s := store.New()
	repo := store.NewPostgresStore(cfg)

	svc, err := service.New(cfg, s, repo)
	if err != nil {
		log.Error(err, "failed to initialize service")
		os.Exit(1)
	}

	ctrl := invoice.New(s, repo, svc, nil, log, cfg)

	// Fetch existing invoice
	log.Infof("Fetching invoice with ID: %s", *invoiceID)
	inv, err := s.Invoice.One(repo.DB(), &invoiceStore.Query{
		ID: *invoiceID,
	})
	if err != nil {
		log.Errorf(err, "failed to fetch invoice", "invoice_id", *invoiceID)
		os.Exit(1)
	}

	// Load related data - Project with all nested relations
	if !inv.ProjectID.IsZero() {
		log.Info("Loading project and related data...")
		p, err := s.Project.One(repo.DB(), inv.ProjectID.String(), true)
		if err != nil {
			log.Errorf(err, "failed to load project", "project_id", inv.ProjectID)
			os.Exit(1)
		}
		inv.Project = p

		// Load client for the project
		if p != nil && !p.ClientID.IsZero() {
			client, err := s.Client.One(repo.DB(), p.ClientID.String())
			if err != nil {
				log.Errorf(err, "failed to load client", "client_id", p.ClientID)
				os.Exit(1)
			}
			inv.Project.Client = client
		}

		// Load bank account for the project
		if p != nil && p.BankAccount != nil {
			inv.Bank = p.BankAccount
		}
	}

	// Load bank account if not loaded from project
	if inv.Bank == nil && !inv.BankID.IsZero() {
		log.Info("Loading bank account...")
		b, err := s.BankAccount.One(repo.DB(), inv.BankID.String())
		if err != nil {
			log.Errorf(err, "failed to load bank account", "bank_id", inv.BankID)
			os.Exit(1)
		}
		inv.Bank = b
	}

	// Generate PDF
	log.Info("Generating PDF...")
	invoiceItems, err := model.GetInfoItems(inv.LineItems)
	if err != nil {
		log.Errorf(err, "failed to get info items", "invoice-lineItems", inv.LineItems)
		os.Exit(1)
	}

	if err := ctrl.GenerateInvoicePDFForTest(log, inv, invoiceItems); err != nil {
		log.Error(err, "failed to generate Invoice PDF")
		os.Exit(1)
	}

	// Write PDF to file
	if err := os.WriteFile(*outputPath, inv.InvoiceFileContent, 0644); err != nil {
		log.Errorf(err, "failed to write PDF to file", "path", *outputPath)
		os.Exit(1)
	}

	log.Infof("✅ PDF generated successfully: %s", *outputPath)
	log.Info("")
	log.Info("Please open the PDF and verify:")
	log.Info("  1. Text is selectable (try to highlight text)")
	log.Info("  2. Text is copyable (try Cmd+C / Ctrl+C)")
	log.Info("  3. Text is searchable (try Cmd+F / Ctrl+F)")
	log.Info("  4. Visual appearance matches original invoices")
}

