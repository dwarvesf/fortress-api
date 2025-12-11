//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dstotijn/go-notion"
	"github.com/joho/godotenv"
)

// Dev database IDs
const (
	devInvoiceDBID            = "2c3b69f8-f573-8124-8e45-e3a9060ffa47"
	devDeploymentTrackerDBID  = "2c3b69f8-f573-811a-97f6-e750268414a4"
	devProjectDBID            = "2c3b69f8-f573-818d-aff0-d63f386a5e27"
)

// Parent invoice IDs
var parentInvoices = map[string]string{
	"INV-202512-NGHENHAN-TP8Q": "2c3b69f8-f573-816b-ae9f-fc1c5b8184fe",
	"INV-202512-KAFI-001":      "2c3b69f8-f573-81ff-9328-c2e6cb8d9223",
	"INV-202512-ASCENDA-001":   "2c3b69f8-f573-812d-bb62-edfc4f079af2",
	"INV-202512-MUDAH-001":     "2c3b69f8-f573-81c0-b6ab-e9dd4d1201a1",
	"INV-202512-INLOOP-001":    "2c3b69f8-f573-81e8-bd99-f68f1e6d7113",
}

// Line item IDs with their parent mapping
var lineItems = map[string]struct {
	ID       string
	ParentID string
}{
	"NGHENHAN :: dev1 / Nov 2025":      {ID: "2c3b69f8-f573-8101-9598-efc77975d283", ParentID: "2c3b69f8-f573-816b-ae9f-fc1c5b8184fe"},
	"NGHENHAN :: dev2 / Nov 2025":      {ID: "2c3b69f8-f573-81e1-b95a-f944e4b90246", ParentID: "2c3b69f8-f573-816b-ae9f-fc1c5b8184fe"},
	"KAFI :: minhth / Nov 2025":        {ID: "2c3b69f8-f573-81ce-8db6-cf0b65ac4591", ParentID: "2c3b69f8-f573-81ff-9328-c2e6cb8d9223"},
	"KAFI :: dev3 / Nov 2025":          {ID: "2c3b69f8-f573-811a-8e1d-f2e82b80af5c", ParentID: "2c3b69f8-f573-81ff-9328-c2e6cb8d9223"},
	"ASCENDA :: john.dev / Nov 2025":   {ID: "2c3b69f8-f573-81a3-9a48-c9262d3903a5", ParentID: "2c3b69f8-f573-812d-bb62-edfc4f079af2"},
	"ASCENDA :: sarah.qa / Nov 2025":   {ID: "2c3b69f8-f573-8181-934c-ebb3a1b008da", ParentID: "2c3b69f8-f573-812d-bb62-edfc4f079af2"},
	"MUDAH :: alex.fe / Nov 2025":      {ID: "2c3b69f8-f573-8170-a972-cb3bed69da71", ParentID: "2c3b69f8-f573-81c0-b6ab-e9dd4d1201a1"},
	"INLOOP :: monotykamary / Nov 2025": {ID: "2c3b69f8-f573-8122-892d-d6c6aa17e8fe", ParentID: "2c3b69f8-f573-81e8-bd99-f68f1e6d7113"},
}

// Deployment tracker IDs
var deploymentTrackers = []string{
	"2c3b69f8-f573-8131-b6f0-df25d623a8c4", // Kafi :: minhth
	"2c3b69f8-f573-8178-8eec-d217e5d9b70d", // inloop.studio :: monotykamary
	"2c3b69f8-f573-8146-8ed8-cab09941c555", // Ascenda :: john.dev
	"2c3b69f8-f573-8135-ad22-c257812a6c38", // Mudah :: sarah.qa
	"2c3b69f8-f573-8126-a588-d5416ebe6af2", // nghenhan.trade :: alex.fe
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found: %v", err)
	}

	// Use dev token - check for NOTION_DEV_SECRET first, fallback to NOTION_SECRET
	secret := os.Getenv("NOTION_DEV_SECRET")
	if secret == "" {
		secret = os.Getenv("NOTION_SECRET")
	}
	if secret == "" {
		log.Fatal("NOTION_SECRET or NOTION_DEV_SECRET not set")
	}

	log.Printf("DEBUG: Starting Notion dev data update")

	client := notion.NewClient(secret)
	ctx := context.Background()

	// Step 1: Update parent invoices - set Type=Invoice, Status=Sent
	log.Println("DEBUG: Step 1 - Updating parent invoices...")
	for name, id := range parentInvoices {
		log.Printf("DEBUG: Updating parent invoice: %s (%s)", name, id)
		if err := updateParentInvoice(ctx, client, id); err != nil {
			log.Printf("ERROR: Failed to update parent invoice %s: %v", name, err)
		} else {
			log.Printf("DEBUG: Successfully updated parent invoice: %s", name)
		}
	}

	// Step 2: Update line items - set Parent item relation, Type=Line Item
	log.Println("DEBUG: Step 2 - Updating line items with parent relations...")
	for name, item := range lineItems {
		log.Printf("DEBUG: Updating line item: %s (%s) -> parent: %s", name, item.ID, item.ParentID)
		if err := updateLineItem(ctx, client, item.ID, item.ParentID); err != nil {
			log.Printf("ERROR: Failed to update line item %s: %v", name, err)
		} else {
			log.Printf("DEBUG: Successfully updated line item: %s", name)
		}
	}

	// Step 3: Update deployment trackers - set Deployment Status=Active
	log.Println("DEBUG: Step 3 - Updating deployment trackers...")
	for _, id := range deploymentTrackers {
		log.Printf("DEBUG: Updating deployment tracker: %s", id)
		if err := updateDeploymentTracker(ctx, client, id); err != nil {
			log.Printf("ERROR: Failed to update deployment tracker %s: %v", id, err)
		} else {
			log.Printf("DEBUG: Successfully updated deployment tracker: %s", id)
		}
	}

	fmt.Println("\n========== UPDATE COMPLETE ==========")
	fmt.Printf("Parent invoices updated: %d\n", len(parentInvoices))
	fmt.Printf("Line items linked: %d\n", len(lineItems))
	fmt.Printf("Deployment trackers updated: %d\n", len(deploymentTrackers))
	fmt.Println("======================================")
}

func updateParentInvoice(ctx context.Context, client *notion.Client, pageID string) error {
	// Invoice Status options from dev schema:
	// - Draft (gray) - 1d269fb0-7439-4e17-83a2-9809a463f84c
	// - Sent (blue) - dd3b533c-4cd2-48a0-b2da-5ae84bd5e453
	// - Overdue (red) - 54ac5915-795e-4f2a-88bb-bf3d3440a992
	// - Paid (green) - 60d88bc8-ab65-47ec-8068-6a57021256f8
	// - Cancelled (orange) - 9d928372-8385-4dac-a8ea-d742ae16b70a

	// Invoice Type options from dev schema:
	// - Invoice (orange) - c9422b4c-cab4-4d6d-9d68-458d28e9e7ae
	// - Line Item (green) - 7682dea5-602e-450a-abac-7ddf8596f037

	// Currency options:
	// - USD (green) - b4351c6f-7780-493d-aeca-e4c954fedabd

	params := notion.UpdatePageParams{
		DatabasePageProperties: notion.DatabasePageProperties{
			"Type": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Invoice",
				},
			},
			"Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Sent",
				},
			},
			"Currency": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "USD",
				},
			},
		},
	}

	_, err := client.UpdatePage(ctx, pageID, params)
	return err
}

func updateLineItem(ctx context.Context, client *notion.Client, pageID, parentID string) error {
	params := notion.UpdatePageParams{
		DatabasePageProperties: notion.DatabasePageProperties{
			"Type": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Line Item",
				},
			},
			"Parent item": notion.DatabasePageProperty{
				Relation: []notion.Relation{
					{ID: parentID},
				},
			},
			"Currency": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "USD",
				},
			},
		},
	}

	_, err := client.UpdatePage(ctx, pageID, params)
	return err
}

func updateDeploymentTracker(ctx context.Context, client *notion.Client, pageID string) error {
	// Deployment Status options from dev schema:
	// - Not started (default) - 49ee6abf-06d5-451c-9194-9027d478a563
	// - Active (blue) - e5c5b9e1-182c-4547-b136-767cd7ad1a79
	// - Done (green) - b10a3aff-0402-4e33-95b2-af8709896309

	// Type options:
	// - Official (blue) - 11567287-0684-4818-8007-322d1e7ea725
	// - Part-time (purple) - ea6fccac-64a2-449f-ae62-a9f1d922dfba

	// Position options:
	// - AI Engineer (pink) - d99f831f-dba2-48e8-860d-3612fcf16b4b
	// - Fullstack Engineer (green) - f618787c-d24b-439b-bcc8-f1435d5b17e3
	// - Backend Engineer (blue) - 40944ae9-3d37-43dd-90cd-08214b3a64f6

	params := notion.UpdatePageParams{
		DatabasePageProperties: notion.DatabasePageProperties{
			"Deployment Status": notion.DatabasePageProperty{
				Select: &notion.SelectOptions{
					Name: "Active",
				},
			},
			"Type": notion.DatabasePageProperty{
				MultiSelect: []notion.SelectOptions{
					{Name: "Official"},
				},
			},
		},
	}

	_, err := client.UpdatePage(ctx, pageID, params)
	return err
}
