//go:build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dstotijn/go-notion"
	"github.com/joho/godotenv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/notion_create_invoice_db.go <parent_page_id> [project_db_id]")
		fmt.Println("  parent_page_id: Page where database will be created")
		fmt.Println("  project_db_id: (optional) Project database ID for relation")
		os.Exit(1)
	}

	_ = godotenv.Load()

	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		fmt.Println("NOTION_SECRET not set")
		os.Exit(1)
	}

	parentPageID := os.Args[1]
	var projectDBID string
	if len(os.Args) > 2 {
		projectDBID = os.Args[2]
	}

	client := notion.NewClient(secret)
	ctx := context.Background()

	fmt.Printf("DEBUG: Creating invoice database under page %s\n", parentPageID)

	// Step 1: Create database with basic properties
	properties := map[string]notion.DatabaseProperty{
		// Title
		"Invoice Number": {
			Type:  notion.DBPropTypeTitle,
			Title: &notion.EmptyMetadata{},
		},
		// Type select
		"Type": {
			Type: notion.DBPropTypeSelect,
			Select: &notion.SelectMetadata{
				Options: []notion.SelectOptions{
					{Name: "Invoice", Color: notion.ColorOrange},
					{Name: "Line Item", Color: notion.ColorGreen},
				},
			},
		},
		// Status (select - status type cannot be created via API)
		"Status": {
			Type: notion.DBPropTypeSelect,
			Select: &notion.SelectMetadata{
				Options: []notion.SelectOptions{
					{Name: "Draft", Color: notion.ColorDefault},
					{Name: "Sent", Color: notion.ColorBlue},
					{Name: "Overdue", Color: notion.ColorRed},
					{Name: "Paid", Color: notion.ColorGreen},
					{Name: "Cancelled", Color: notion.ColorOrange},
				},
			},
		},
		// Currency
		"Currency": {
			Type: notion.DBPropTypeSelect,
			Select: &notion.SelectMetadata{
				Options: []notion.SelectOptions{
					{Name: "USD", Color: notion.ColorGreen},
					{Name: "EUR", Color: notion.ColorBlue},
					{Name: "GBP", Color: notion.ColorPurple},
					{Name: "JPY", Color: notion.ColorRed},
					{Name: "CNY", Color: notion.ColorOrange},
					{Name: "SGD", Color: notion.ColorYellow},
					{Name: "CAD", Color: notion.ColorPink},
					{Name: "AUD", Color: notion.ColorBrown},
				},
			},
		},
		// Dates
		"Issue Date": {
			Type: notion.DBPropTypeDate,
			Date: &notion.EmptyMetadata{},
		},
		"Due Date": {
			Type: notion.DBPropTypeDate,
			Date: &notion.EmptyMetadata{},
		},
		"Paid Date": {
			Type: notion.DBPropTypeDate,
			Date: &notion.EmptyMetadata{},
		},
		// Numbers - Line Item
		"Quantity": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatNumber},
		},
		"Unit Price": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatNumber},
		},
		"Tax Rate": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent},
		},
		// Discount
		"Discount Type": {
			Type: notion.DBPropTypeSelect,
			Select: &notion.SelectMetadata{
				Options: []notion.SelectOptions{
					{Name: "None", Color: notion.ColorDefault},
					{Name: "Percentage", Color: notion.ColorGreen},
					{Name: "Fixed Amount", Color: notion.ColorBlue},
					{Name: "Bulk Discount", Color: notion.ColorPurple},
					{Name: "Seasonal", Color: notion.ColorOrange},
					{Name: "Loyalty", Color: notion.ColorPink},
					{Name: "Early Payment", Color: notion.ColorYellow},
				},
			},
		},
		"Discount Value": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatNumberWithCommas},
		},
		// Commission percentages
		"% Sales": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent},
		},
		"% Account Mgr": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent},
		},
		"% Delivery Lead": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent},
		},
		"% Hiring Referral": {
			Type:   notion.DBPropTypeNumber,
			Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent},
		},
		// Text fields
		"Description": {
			Type:     notion.DBPropTypeRichText,
			RichText: &notion.EmptyMetadata{},
		},
		"Role": {
			Type:     notion.DBPropTypeRichText,
			RichText: &notion.EmptyMetadata{},
		},
		"Notes": {
			Type:     notion.DBPropTypeRichText,
			RichText: &notion.EmptyMetadata{},
		},
		"Sent by": {
			Type:     notion.DBPropTypeRichText,
			RichText: &notion.EmptyMetadata{},
		},
		// Payment
		"Payment Method": {
			Type: notion.DBPropTypeSelect,
			Select: &notion.SelectMetadata{
				Options: []notion.SelectOptions{
					{Name: "Bank Transfer", Color: notion.ColorBlue},
					{Name: "Credit Card", Color: notion.ColorPurple},
					{Name: "Cash", Color: notion.ColorGreen},
					{Name: "Check", Color: notion.ColorDefault},
					{Name: "PayPal", Color: notion.ColorBlue},
					{Name: "Venmo", Color: notion.ColorPink},
					{Name: "Crypto", Color: notion.ColorOrange},
					{Name: "Mobile Payment", Color: notion.ColorYellow},
				},
			},
		},
		// Checkbox
		"Splits Generated": {
			Type:     notion.DBPropTypeCheckbox,
			Checkbox: &notion.EmptyMetadata{},
		},
	}

	// Add Project relation if project DB ID provided
	if projectDBID != "" {
		fmt.Printf("DEBUG: Adding Project relation to database %s\n", projectDBID)
		properties["Project"] = notion.DatabaseProperty{
			Type: notion.DBPropTypeRelation,
			Relation: &notion.RelationMetadata{
				DatabaseID:     projectDBID,
				Type:           notion.RelationTypeSingleProperty,
				SingleProperty: &struct{}{},
			},
		}
	}

	db, err := client.CreateDatabase(ctx, notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{Text: &notion.Text{Content: "Invoices"}},
		},
		Properties: properties,
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to create database: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DEBUG: Database created with ID: %s\n", db.ID)

	// Step 2: Add self-referential relation (Parent item / Line Item)
	fmt.Printf("DEBUG: Adding self-referential relation (Parent item <-> Line Item)\n")

	updatedDB, err := client.UpdateDatabase(ctx, db.ID, notion.UpdateDatabaseParams{
		Properties: map[string]*notion.DatabaseProperty{
			"Parent item": {
				Type: notion.DBPropTypeRelation,
				Relation: &notion.RelationMetadata{
					DatabaseID: db.ID,
					Type:       notion.RelationTypeDualProperty,
					DualProperty: &notion.DualPropertyRelation{
						SyncedPropName: "Line Item",
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to add self-referential relation: %v\n", err)
		fmt.Println("You may need to add this manually in Notion UI")
	} else {
		fmt.Printf("DEBUG: Self-referential relation added successfully\n")
		db = updatedDB
	}

	fmt.Println("\n========== DATABASE CREATED ==========")
	fmt.Printf("Database ID: %s\n", db.ID)
	fmt.Printf("URL: %s\n", db.URL)
	fmt.Println("\n========== MANUAL SETUP REQUIRED ==========")
	fmt.Println("The following must be created manually in Notion UI:")
	fmt.Println("")
	fmt.Println("1. Convert Status from 'select' to 'status' type (optional)")
	fmt.Println("")
	fmt.Println("2. Add Unique ID property:")
	fmt.Println("   - Name: ID")
	fmt.Println("   - Type: ID")
	fmt.Println("")
	fmt.Println("3. Add Formula properties:")
	fmt.Println("   - Auto Name")
	fmt.Println("   - Line Total: Quantity × Unit Price × (1 + Tax Rate) - Discount")
	fmt.Println("   - Subtotal: Quantity × Unit Price")
	fmt.Println("   - Final Total: Total Amount - Discount Amount with currency")
	fmt.Println("   - Discount Amount")
	fmt.Println("   - Discount Display")
	fmt.Println("   - Sales Amount")
	fmt.Println("   - Account Amount")
	fmt.Println("   - Delivery Lead Amount")
	fmt.Println("   - Hiring Referral Amount")
	fmt.Println("   - Total Commission Paid")
	fmt.Println("   - Account Manager (from Deployment Tracker)")
	fmt.Println("   - Delivery Lead (from Deployment Tracker)")
	fmt.Println("   - Hiring Referral (from Deployment Tracker)")
	fmt.Println("")
	fmt.Println("4. Add Rollup properties:")
	fmt.Println("   - Total Amount: sum of Line Item → Line Total")
	fmt.Println("   - Code: from Project → Codename")
	fmt.Println("   - Client: from Project → Client")
	fmt.Println("   - Recipients: from Project → Recipient Emails")
	fmt.Println("   - Redacted Codename: from Project → Redacted Code")
	fmt.Println("   - All Sales Amounts: sum of Line Item → Sales Amount")
	fmt.Println("   - All AM Amounts: sum of Line Item → Account Amount")
	fmt.Println("   - All DL Amounts: sum of Line Item → Delivery Lead Amount")
	fmt.Println("   - All Hiring Ref Amounts: sum of Line Item → Hiring Referral Amount")
	fmt.Println("")
	fmt.Println("5. Add Relations (if not provided):")
	if projectDBID == "" {
		fmt.Println("   - Project → Projects database")
	}
	fmt.Println("   - Google Drive File → Google Drive database")
	fmt.Println("   - Bank Account → Bank Accounts database")
	fmt.Println("   - Deployment Tracker → Deployments database")
	fmt.Println("   - Splits → Splits database")
	fmt.Println("")
	fmt.Println("6. Add Button property:")
	fmt.Println("   - Generate Splits")
	fmt.Println("==========================================")

	out, _ := json.MarshalIndent(db, "", "  ")
	fmt.Println("\nFull database JSON:")
	fmt.Println(string(out))
}
