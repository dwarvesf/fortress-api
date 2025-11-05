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
	invoiceID := flag.String("invoice-id", "", "Invoice ID to send (required)")
	toEmail := flag.String("to", "", "Recipient email (optional, uses invoice email if not specified)")
	flag.Parse()

	if *invoiceID == "" {
		fmt.Println("Error: --invoice-id is required")
		fmt.Println("\nUsage:")
		fmt.Println("  test-send-invoice --invoice-id=<uuid> [--to=email@example.com]")
		fmt.Println("\nExample:")
		fmt.Println("  test-send-invoice --invoice-id=123e4567-e89b-12d3-a456-426614174000")
		fmt.Println("  test-send-invoice --invoice-id=123e4567-e89b-12d3-a456-426614174000 --to=test@example.com")
		os.Exit(1)
	}

	fmt.Println("=== Testing Invoice Email Sending ===")
	fmt.Printf("Invoice ID: %s\n", *invoiceID)
	if *toEmail != "" {
		fmt.Printf("Recipient: %s\n", *toEmail)
	}
	fmt.Println()

	cfg := config.LoadConfig(config.DefaultConfigLoaders())
	log := logger.NewLogrusLogger()

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
	fmt.Printf("📋 Fetching invoice with ID: %s\n", *invoiceID)
	inv, err := s.Invoice.One(repo.DB(), &invoiceStore.Query{
		ID: *invoiceID,
	})
	if err != nil {
		log.Errorf(err, "failed to fetch invoice", "invoice_id", *invoiceID)
		os.Exit(1)
	}

	// Override recipient email if specified
	originalEmail := inv.Email
	if *toEmail != "" {
		inv.Email = *toEmail
		fmt.Printf("✏️  Overriding recipient from %s to %s\n", originalEmail, *toEmail)
	}

	// Load related data - Project with all nested relations
	if !inv.ProjectID.IsZero() {
		fmt.Println("📦 Loading project and related data...")
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
		fmt.Println("💳 Loading bank account...")
		b, err := s.BankAccount.One(repo.DB(), inv.BankID.String())
		if err != nil {
			log.Errorf(err, "failed to load bank account", "bank_id", inv.BankID)
			os.Exit(1)
		}
		inv.Bank = b
	}

	// Generate PDF if not exists
	if len(inv.InvoiceFileContent) == 0 {
		fmt.Println("📄 Generating PDF...")
		invoiceItems, err := model.GetInfoItems(inv.LineItems)
		if err != nil {
			log.Errorf(err, "failed to get info items", "invoice-lineItems", inv.LineItems)
			os.Exit(1)
		}

		if err := ctrl.GenerateInvoicePDFForTest(log, inv, invoiceItems); err != nil {
			log.Error(err, "failed to generate Invoice PDF")
			os.Exit(1)
		}
		fmt.Println("✅ PDF generated")
	} else {
		fmt.Println("✅ Using existing PDF")
	}

	// Send email
	fmt.Println()
	fmt.Printf("📧 Sending invoice email to: %s\n", inv.Email)
	threadID, err := svc.GoogleMail.SendInvoiceMail(inv)
	if err != nil {
		log.Errorf(err, "failed to send invoice email")
		fmt.Println()
		fmt.Println("❌ Failed to send invoice email")
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ Invoice email sent successfully!")
	fmt.Printf("📨 Gmail Thread ID: %s\n", threadID)
	fmt.Println()
	fmt.Println("Please check:")
	fmt.Printf("  1. Email inbox at %s\n", inv.Email)
	fmt.Println("  2. Email has PDF attachment")
	fmt.Println("  3. Email sender is accounting@d.foundation")
	fmt.Println("  4. Email subject matches invoice details")
}
