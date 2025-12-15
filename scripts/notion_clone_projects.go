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

const (
	prodProjectDBID = "0ddadba5-bbf2-440c-a286-9f607eca88db"
	devProjectDBID  = "2c3b69f8-f573-818d-aff0-d63f386a5e27"
)

type ProjectData struct {
	Name       string
	Status     string
	Size       string
	Tags       []string
	TechStack  []string
	Sales      string
	AccountMgr []string
	DeliveryLd []string
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found: %v", err)
	}

	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		log.Fatal("NOTION_SECRET not set")
	}

	log.Printf("DEBUG: Starting project clone from prod to dev")

	client := notion.NewClient(secret)
	ctx := context.Background()

	// Step 1: Fetch projects from production
	log.Println("DEBUG: Step 1 - Fetching projects from production...")
	projects, err := fetchProdProjects(ctx, client)
	if err != nil {
		log.Fatalf("Failed to fetch prod projects: %v", err)
	}
	log.Printf("DEBUG: Fetched %d projects from production", len(projects))

	// Step 2: Clear existing dev projects (optional - skip for now)
	// Step 3: Create projects in dev
	log.Println("DEBUG: Step 2 - Creating projects in dev...")
	created := 0
	for _, proj := range projects {
		log.Printf("DEBUG: Creating project: %s", proj.Name)
		if err := createDevProject(ctx, client, proj); err != nil {
			log.Printf("ERROR: Failed to create project %s: %v", proj.Name, err)
		} else {
			log.Printf("DEBUG: Successfully created project: %s", proj.Name)
			created++
		}
	}

	fmt.Println("\n========== CLONE COMPLETE ==========")
	fmt.Printf("Projects fetched from prod: %d\n", len(projects))
	fmt.Printf("Projects created in dev: %d\n", created)
	fmt.Println("=====================================")
}

func fetchProdProjects(ctx context.Context, client *notion.Client) ([]ProjectData, error) {
	var projects []ProjectData

	query := &notion.DatabaseQuery{
		PageSize: 20,
	}

	resp, err := client.QueryDatabase(ctx, prodProjectDBID, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	log.Printf("DEBUG: Got %d results from prod", len(resp.Results))

	for _, page := range resp.Results {
		proj := ProjectData{}

		// Get title
		if titleProp, ok := page.Properties.(notion.DatabasePageProperties)["Project"]; ok {
			if titleProp.Title != nil && len(titleProp.Title) > 0 {
				proj.Name = titleProp.Title[0].PlainText
			}
		}

		// Get Status
		if statusProp, ok := page.Properties.(notion.DatabasePageProperties)["Status"]; ok {
			if statusProp.Status != nil {
				proj.Status = statusProp.Status.Name
			}
		}

		// Get Size
		if sizeProp, ok := page.Properties.(notion.DatabasePageProperties)["Size"]; ok {
			if sizeProp.Select != nil {
				proj.Size = sizeProp.Select.Name
			}
		}

		// Get Tags
		if tagsProp, ok := page.Properties.(notion.DatabasePageProperties)["Tags"]; ok {
			if tagsProp.MultiSelect != nil {
				for _, tag := range tagsProp.MultiSelect {
					proj.Tags = append(proj.Tags, tag.Name)
				}
			}
		}

		// Get Tech Stack
		if techProp, ok := page.Properties.(notion.DatabasePageProperties)["Tech Stack"]; ok {
			if techProp.MultiSelect != nil {
				for _, tech := range techProp.MultiSelect {
					proj.TechStack = append(proj.TechStack, tech.Name)
				}
			}
		}

		// Get Sales
		if salesProp, ok := page.Properties.(notion.DatabasePageProperties)["Sales"]; ok {
			if salesProp.Select != nil {
				proj.Sales = salesProp.Select.Name
			}
		}

		// Get Account Manager
		if amProp, ok := page.Properties.(notion.DatabasePageProperties)["Deal Closing (Account Manager)"]; ok {
			if amProp.MultiSelect != nil {
				for _, am := range amProp.MultiSelect {
					proj.AccountMgr = append(proj.AccountMgr, am.Name)
				}
			}
		}

		// Get Delivery Lead
		if dlProp, ok := page.Properties.(notion.DatabasePageProperties)["PM/Delivery (Technical Lead)"]; ok {
			if dlProp.MultiSelect != nil {
				for _, dl := range dlProp.MultiSelect {
					proj.DeliveryLd = append(proj.DeliveryLd, dl.Name)
				}
			}
		}

		if proj.Name != "" {
			projects = append(projects, proj)
			log.Printf("DEBUG: Parsed project: %s (Status: %s, Size: %s)", proj.Name, proj.Status, proj.Size)
		}
	}

	return projects, nil
}

func createDevProject(ctx context.Context, client *notion.Client, proj ProjectData) error {
	props := notion.DatabasePageProperties{
		"Project": notion.DatabasePageProperty{
			Title: []notion.RichText{
				{Text: &notion.Text{Content: proj.Name}},
			},
		},
	}

	// Add Status (as select in dev)
	if proj.Status != "" {
		props["Status"] = notion.DatabasePageProperty{
			Select: &notion.SelectOptions{Name: proj.Status},
		}
	}

	// Add Size
	if proj.Size != "" {
		props["Size"] = notion.DatabasePageProperty{
			Select: &notion.SelectOptions{Name: proj.Size},
		}
	}

	// Add Tags
	if len(proj.Tags) > 0 {
		var tags []notion.SelectOptions
		for _, tag := range proj.Tags {
			tags = append(tags, notion.SelectOptions{Name: tag})
		}
		props["Tags"] = notion.DatabasePageProperty{
			MultiSelect: tags,
		}
	}

	// Add Tech Stack
	if len(proj.TechStack) > 0 {
		var techStack []notion.SelectOptions
		for _, tech := range proj.TechStack {
			techStack = append(techStack, notion.SelectOptions{Name: tech})
		}
		props["Tech Stack"] = notion.DatabasePageProperty{
			MultiSelect: techStack,
		}
	}

	// Add Sales
	if proj.Sales != "" {
		props["Sales"] = notion.DatabasePageProperty{
			Select: &notion.SelectOptions{Name: proj.Sales},
		}
	}

	// Add Account Manager
	if len(proj.AccountMgr) > 0 {
		var managers []notion.SelectOptions
		for _, am := range proj.AccountMgr {
			managers = append(managers, notion.SelectOptions{Name: am})
		}
		props["Deal Closing (Account Manager)"] = notion.DatabasePageProperty{
			MultiSelect: managers,
		}
	}

	// Add Delivery Lead
	if len(proj.DeliveryLd) > 0 {
		var leads []notion.SelectOptions
		for _, dl := range proj.DeliveryLd {
			leads = append(leads, notion.SelectOptions{Name: dl})
		}
		props["PM/Delivery (Technical Lead)"] = notion.DatabasePageProperty{
			MultiSelect: leads,
		}
	}

	params := notion.CreatePageParams{
		ParentType:             notion.ParentTypeDatabase,
		ParentID:               devProjectDBID,
		DatabasePageProperties: &props,
	}

	_, err := client.CreatePage(ctx, params)
	return err
}
