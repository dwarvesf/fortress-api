package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/service/googlemail"
	"github.com/dwarvesf/fortress-api/pkg/service/vault"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	action := flag.String("action", "list", "Action to perform: list, check, verify, create")
	userID := flag.String("user-id", "", "User ID (email ID from config, defaults to accounting)")
	email := flag.String("email", "", "Email alias to check/verify/create")
	displayName := flag.String("display-name", "", "Display name for create action")
	flag.Parse()

	cfg := config.LoadConfig(config.DefaultConfigLoaders())

	fmt.Println("=== Testing SendAs functionality ===")
	fmt.Printf("Action: %s\n\n", *action)

	v, _ := vault.New(cfg)

	if v != nil {
		cfg = config.Generate(v)
	}

	// Initialize Gmail service directly without database
	mailConfig := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			gmail.MailGoogleComScope,
			"https://www.googleapis.com/auth/gmail.settings.basic",
			"https://www.googleapis.com/auth/gmail.settings.sharing",
		},
	}

	gmailSvc := googlemail.New(mailConfig, cfg)

	// Default to accounting email ID if not specified
	userIDToUse := *userID
	if userIDToUse == "" {
		userIDToUse = cfg.Google.AccountingEmailID
		fmt.Printf("Using default Accounting Email ID: %s\n\n", userIDToUse)
	}

	switch *action {
	case "list":
		// List all SendAs aliases
		fmt.Printf("📋 Listing all SendAs aliases for user: %s\n", userIDToUse)
		aliases, err := gmailSvc.ListSendAsAliases(userIDToUse)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: failed to list SendAs aliases: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n✅ Found %d aliases:\n\n", len(aliases))
		for i, alias := range aliases {
			fmt.Printf("%d. %s\n", i+1, alias.SendAsEmail)
			if alias.DisplayName != "" {
				fmt.Printf("   Display Name: %s\n", alias.DisplayName)
			}
			switch alias.VerificationStatus {
			case "accepted":
				fmt.Printf("   Status: ✅ VERIFIED\n")
			case "pending":
				fmt.Printf("   Status: ⏳ PENDING\n")
			default:
				fmt.Printf("   Status: %s\n", alias.VerificationStatus)
			}
			if alias.IsPrimary {
				fmt.Printf("   Primary: ✅\n")
			}
			if alias.IsDefault {
				fmt.Printf("   Default: ✅\n")
			}
			fmt.Println()
		}

	case "check":
		// Check if specific alias is verified
		if *email == "" {
			fmt.Println("Error: --email is required for check action")
			fmt.Println("\nUsage:")
			fmt.Println("  test-sendas --action=check --email=accounting@d.foundation")
			os.Exit(1)
		}

		fmt.Printf("🔍 Checking verification status for: %s\n\n", *email)
		verified, err := gmailSvc.IsAliasVerified(userIDToUse, *email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: failed to check verification status: %v\n", err)
			os.Exit(1)
		}

		if verified {
			fmt.Printf("✅ %s is VERIFIED\n\n", *email)
		} else {
			fmt.Printf("❌ %s is NOT VERIFIED\n\n", *email)
		}

		// Get detailed info
		alias, err := gmailSvc.GetSendAsAlias(userIDToUse, *email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: failed to get alias details: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Alias Details:")
		fmt.Printf("  Display Name: %s\n", alias.DisplayName)
		fmt.Printf("  Verification Status: %s\n", alias.VerificationStatus)
		fmt.Printf("  Reply To: %s\n", alias.ReplyToAddress)
		fmt.Printf("  Treat As Alias: %v\n", alias.TreatAsAlias)

	case "verify":
		// Resend verification email
		if *email == "" {
			fmt.Println("Error: --email is required for verify action")
			fmt.Println("\nUsage:")
			fmt.Println("  test-sendas --action=verify --email=accounting@d.foundation")
			os.Exit(1)
		}

		fmt.Printf("📧 Resending verification email for: %s\n\n", *email)
		err := gmailSvc.VerifySendAsAlias(userIDToUse, *email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: failed to send verification email: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Verification email sent to %s\n", *email)
		fmt.Println("📬 Please check the inbox and click the verification link")

	case "create":
		// Create new SendAs alias
		if *email == "" || *displayName == "" {
			fmt.Println("Error: --email and --display-name are required for create action")
			fmt.Println("\nUsage:")
			fmt.Println("  test-sendas --action=create --email=accounting@d.foundation --display-name=\"Accounting @ Dwarves Foundation\"")
			os.Exit(1)
		}

		fmt.Printf("➕ Creating SendAs alias: %s (%s)\n\n", *email, *displayName)
		alias, err := gmailSvc.CreateSendAsAlias(userIDToUse, *email, *displayName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: failed to create SendAs alias: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ SendAs alias created successfully")
		fmt.Printf("  Email: %s\n", alias.SendAsEmail)
		fmt.Printf("  Display Name: %s\n", alias.DisplayName)
		fmt.Printf("  Verification Status: %s\n", alias.VerificationStatus)

		if alias.VerificationStatus == "pending" {
			fmt.Println()
			fmt.Println("⚠️  Verification required!")
			fmt.Printf("📬 Please check the inbox at %s and click the verification link\n", *email)
		}

	default:
		fmt.Printf("Error: unknown action '%s'\n", *action)
		fmt.Println("\nAvailable actions:")
		fmt.Println("  list     - List all SendAs aliases")
		fmt.Println("  check    - Check if specific alias is verified")
		fmt.Println("  verify   - Resend verification email for alias")
		fmt.Println("  create   - Create new SendAs alias")
		fmt.Println("\nExamples:")
		fmt.Println("  test-sendas --action=list")
		fmt.Println("  test-sendas --action=check --email=accounting@d.foundation")
		fmt.Println("  test-sendas --action=verify --email=accounting@d.foundation")
		fmt.Println("  test-sendas --action=create --email=accounting@d.foundation --display-name=\"Accounting @ Dwarves Foundation\"")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("✅ SendAs test completed successfully")
}
