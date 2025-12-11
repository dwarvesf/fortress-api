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

const devProjectDBID = "2c3b69f8-f573-818d-aff0-d63f386a5e27"

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

// Production project data extracted from MCP query (Active projects)
var prodProjects = []ProjectData{
	{
		Name:       "inloop.studio",
		Status:     "Active",
		Size:       "Medium -",
		Tags:       []string{"MVP"},
		TechStack:  []string{"Ruby", "AI"},
		Sales:      "Inbound",
		AccountMgr: []string{"Nikki (nikki@d.foundation)", "Tom Nguyen (anhnx@d.foundation)"},
		DeliveryLd: []string{"Thanh Pham (thanhpd@d.foundation)", "Tom Nguyen (anhnx@d.foundation)"},
	},
	{
		Name:       "Kafi",
		Status:     "Active",
		Size:       "Medium +",
		Tags:       []string{"Enterprise"},
		TechStack:  []string{"Elixir", "Golang"},
		Sales:      "Nikki (nikki@d.foundation)",
		AccountMgr: []string{"Nikki (nikki@d.foundation)"},
		DeliveryLd: []string{"Huy Nguyen (huy@d.foundation)"},
	},
	{
		Name:      "ScoutQA",
		Status:    "Active",
		Size:      "Medium -",
		Tags:      []string{"Enterprise", "MVP"},
		TechStack: []string{"AI", "React", "Java"},
		Sales:     "Huy Tieu (huytq@d.foundation)",
	},
	{
		Name:       "Dai-ichi Insight",
		Status:     "Active",
		Size:       "Small",
		Tags:       []string{"Enterprise"},
		TechStack:  []string{"Web"},
		Sales:      "Inbound",
		AccountMgr: []string{"Nikki (nikki@d.foundation)", "Minh Le (minhla@d.foundation)"},
	},
	{
		Name:       "Botts Ingest",
		Status:     "Active",
		Size:       "Small",
		Tags:       []string{"MVP"},
		TechStack:  []string{"AI"},
		Sales:      "Han (han@d.foundation)",
		AccountMgr: []string{"Minh Le (minhla@d.foundation)"},
	},
	{
		Name:       "Renaiss",
		Status:     "Active",
		Size:       "Medium -",
		Tags:       []string{"Pre-seed"},
		Sales:      "Minh Le (minhla@d.foundation)",
		AccountMgr: []string{"Nikki (nikki@d.foundation)"},
	},
	{
		Name:       "Ascenda",
		Status:     "Active",
		Size:       "Medium +",
		Tags:       []string{"Series B"},
		TechStack:  []string{"AI", "Ruby", "Golang"},
		Sales:      "Han (han@d.foundation)",
		AccountMgr: []string{"Nikki (nikki@d.foundation)"},
		DeliveryLd: []string{"Huy Nguyen (huy@d.foundation)"},
	},
	{
		Name:       "nghenhan.trade",
		Status:     "Active",
		Size:       "Medium -",
		TechStack:  []string{"Elixir", "React", "React Native"},
		Sales:      "Han (han@d.foundation)",
		AccountMgr: []string{"Han (han@d.foundation)"},
		DeliveryLd: []string{"An Tran (antt@d.foundation)"},
	},
	{
		Name:       "Mudah",
		Status:     "Active",
		Size:       "Medium -",
		Tags:       []string{"Enterprise"},
		TechStack:  []string{"Golang", "React"},
		Sales:      "Han (han@d.foundation)",
		AccountMgr: []string{"Han (han@d.foundation)"},
		DeliveryLd: []string{"Hung Vong (hungvt@d.foundation)"},
	},
	{
		Name:       "Plot",
		Status:     "Active",
		Size:       "Medium -",
		Tags:       []string{"Series A"},
		TechStack:  []string{"AI", "NodeJS", "React", "React Native"},
		Sales:      "Nikki (nikki@d.foundation)",
		AccountMgr: []string{"Nikki (nikki@d.foundation)"},
		DeliveryLd: []string{"Lap Nguyen (alan@d.foundation)"},
	},
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found: %v", err)
	}

	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		log.Fatal("NOTION_SECRET not set")
	}

	log.Printf("DEBUG: Starting project seed to dev database")
	log.Printf("DEBUG: Target database: %s", devProjectDBID)

	client := notion.NewClient(secret)
	ctx := context.Background()

	created := 0
	for _, proj := range prodProjects {
		log.Printf("DEBUG: Creating project: %s (Status: %s, Size: %s)", proj.Name, proj.Status, proj.Size)
		if err := createDevProject(ctx, client, proj); err != nil {
			log.Printf("ERROR: Failed to create project %s: %v", proj.Name, err)
		} else {
			log.Printf("DEBUG: Successfully created project: %s", proj.Name)
			created++
		}
	}

	fmt.Println("\n========== SEED COMPLETE ==========")
	fmt.Printf("Projects to create: %d\n", len(prodProjects))
	fmt.Printf("Projects created: %d\n", created)
	fmt.Println("====================================")
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
