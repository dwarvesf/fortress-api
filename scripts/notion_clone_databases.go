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

// Database IDs created in dev - will be populated as we create them
var devDBs = map[string]string{
	"contractors":        "",
	"project":            "",
	"deployment_tracker": "",
	"invoice":            "",
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found: %v", err)
	}

	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		log.Fatal("NOTION_SECRET not set")
	}

	parentPageID := "2bfb69f8-f573-803e-a3e9-c2f0266aa683"
	if len(os.Args) > 1 {
		parentPageID = os.Args[1]
	}

	log.Printf("DEBUG: Using parent page ID: %s", parentPageID)
	log.Printf("DEBUG: Creating databases in order: Contractors -> Project -> Deployment Tracker -> Invoice")

	client := notion.NewClient(secret)
	ctx := context.Background()

	// Step 1: Create Contractors database
	log.Println("DEBUG: Step 1 - Creating Contractors database...")
	contractorsID, err := createContractorsDB(ctx, client, parentPageID)
	if err != nil {
		log.Fatalf("Failed to create Contractors: %v", err)
	}
	devDBs["contractors"] = contractorsID
	log.Printf("DEBUG: Contractors created: %s", contractorsID)

	// Step 2: Create Project database
	log.Println("DEBUG: Step 2 - Creating Project database...")
	projectID, err := createProjectDB(ctx, client, parentPageID)
	if err != nil {
		log.Fatalf("Failed to create Project: %v", err)
	}
	devDBs["project"] = projectID
	log.Printf("DEBUG: Project created: %s", projectID)

	// Step 3: Create Deployment Tracker database
	log.Println("DEBUG: Step 3 - Creating Deployment Tracker database...")
	deploymentID, err := createDeploymentTrackerDB(ctx, client, parentPageID, contractorsID, projectID)
	if err != nil {
		log.Fatalf("Failed to create Deployment Tracker: %v", err)
	}
	devDBs["deployment_tracker"] = deploymentID
	log.Printf("DEBUG: Deployment Tracker created: %s", deploymentID)

	// Step 4: Create Invoice database
	log.Println("DEBUG: Step 4 - Creating Invoice database...")
	invoiceID, err := createInvoiceDB(ctx, client, parentPageID, projectID)
	if err != nil {
		log.Fatalf("Failed to create Invoice: %v", err)
	}
	devDBs["invoice"] = invoiceID
	log.Printf("DEBUG: Invoice created: %s", invoiceID)

	// Step 5: Add self-referential relation to Invoice
	log.Println("DEBUG: Step 5 - Adding self-referential relation to Invoice...")
	if err := addInvoiceSelfRelation(ctx, client, invoiceID); err != nil {
		log.Fatalf("Failed to add self-relation to Invoice: %v", err)
	}
	log.Println("DEBUG: Self-referential relation added to Invoice")

	// Print summary
	fmt.Println("\n========== DEV DATABASE IDs ==========")
	fmt.Printf("Contractors:        %s\n", devDBs["contractors"])
	fmt.Printf("Project:            %s\n", devDBs["project"])
	fmt.Printf("Deployment Tracker: %s\n", devDBs["deployment_tracker"])
	fmt.Printf("Invoice:            %s\n", devDBs["invoice"])
	fmt.Println("=======================================")
	fmt.Println("\nNOTE: The following must be added manually in Notion UI:")
	fmt.Println("- status properties (use select as workaround)")
	fmt.Println("- formula properties")
	fmt.Println("- rollup properties")
	fmt.Println("- unique_id properties (auto-created)")
	fmt.Println("- button properties")
}

func createContractorsDB(ctx context.Context, client *notion.Client, parentPageID string) (string, error) {
	log.Println("DEBUG: Building Contractors properties...")

	// Position options
	positionOptions := []notion.SelectOptions{
		{Name: "Managing Director", Color: notion.ColorBrown},
		{Name: "CEO", Color: notion.ColorBlue},
		{Name: "CTO", Color: notion.ColorRed},
		{Name: "Head of Operations", Color: notion.ColorBlue},
		{Name: "Head of Business", Color: notion.ColorBlue},
		{Name: "Head of Engineering", Color: notion.ColorBlue},
		{Name: "DevOps Principal", Color: notion.ColorBlue},
		{Name: "Partner", Color: notion.ColorPurple},
		{Name: "Account Manager", Color: notion.ColorPurple},
		{Name: "Consultant", Color: notion.ColorGray},
		{Name: "Technical Lead", Color: notion.ColorGray},
		{Name: "Frontend Lead", Color: notion.ColorPurple},
		{Name: "Backend Lead", Color: notion.ColorPurple},
		{Name: "iOS Lead", Color: notion.ColorPurple},
		{Name: "Project Manager", Color: notion.ColorPurple},
		{Name: "Frontend", Color: notion.ColorRed},
		{Name: "Bubble", Color: notion.ColorRed},
		{Name: "Backend", Color: notion.ColorRed},
		{Name: "Web", Color: notion.ColorRed},
		{Name: "iOS", Color: notion.ColorRed},
		{Name: "Android", Color: notion.ColorRed},
		{Name: "Full-stack", Color: notion.ColorRed},
		{Name: "QC Automation", Color: notion.ColorRed},
		{Name: "QC Manual", Color: notion.ColorRed},
		{Name: "Blockchain", Color: notion.ColorRed},
		{Name: "Data Engineer", Color: notion.ColorRed},
		{Name: "Data Analyst", Color: notion.ColorRed},
		{Name: "Unity Game Client", Color: notion.ColorRed},
		{Name: "Java Engineer", Color: notion.ColorBrown},
		{Name: "Design Ops", Color: notion.ColorYellow},
		{Name: "Product Designer", Color: notion.ColorYellow},
		{Name: "UX Designer", Color: notion.ColorYellow},
		{Name: "UI Designer", Color: notion.ColorYellow},
		{Name: "Business Ops", Color: notion.ColorGreen},
		{Name: "Community Ops", Color: notion.ColorGreen},
		{Name: "Technical Recruiter", Color: notion.ColorGreen},
		{Name: "Marketing Executive", Color: notion.ColorGreen},
		{Name: "Operation Executive", Color: notion.ColorGreen},
		{Name: "Community Executive", Color: notion.ColorGreen},
		{Name: "Apprentices", Color: notion.ColorPurple},
		{Name: "Partnership", Color: notion.ColorGray},
		{Name: "Branding", Color: notion.ColorGray},
		{Name: "React Native", Color: notion.ColorDefault},
		{Name: "Product Data", Color: notion.ColorOrange},
		{Name: "Software Architect", Color: notion.ColorPurple},
		{Name: "DevOps", Color: notion.ColorPink},
	}

	// Stack options
	stackOptions := []notion.SelectOptions{
		{Name: "Laravel", Color: notion.ColorOrange},
		{Name: "ReactJS", Color: notion.ColorYellow},
		{Name: "Wordpress", Color: notion.ColorPink},
		{Name: "NodeJS", Color: notion.ColorBlue},
		{Name: "Blockchain", Color: notion.ColorDefault},
		{Name: "Android", Color: notion.ColorGreen},
		{Name: "Manual", Color: notion.ColorBrown},
		{Name: "Golang", Color: notion.ColorGray},
		{Name: "Java", Color: notion.ColorRed},
		{Name: "Automation", Color: notion.ColorPurple},
		{Name: "Designer", Color: notion.ColorDefault},
		{Name: "Unity", Color: notion.ColorGray},
		{Name: "WebGL", Color: notion.ColorBlue},
		{Name: "Elixir", Color: notion.ColorBlue},
		{Name: "Art", Color: notion.ColorBlue},
		{Name: "Discord Bot", Color: notion.ColorBlue},
		{Name: "Product Design", Color: notion.ColorOrange},
		{Name: "Angular", Color: notion.ColorGreen},
		{Name: "Swift", Color: notion.ColorOrange},
		{Name: "PHP", Color: notion.ColorGray},
		{Name: "Data", Color: notion.ColorRed},
		{Name: "Devops", Color: notion.ColorPink},
		{Name: "Modeling", Color: notion.ColorOrange},
		{Name: "React Native", Color: notion.ColorPink},
		{Name: "PM", Color: notion.ColorRed},
		{Name: "Operation Executive", Color: notion.ColorGreen},
		{Name: "Social Platform", Color: notion.ColorYellow},
		{Name: "Sales", Color: notion.ColorPink},
		{Name: "Community", Color: notion.ColorYellow},
		{Name: "Selenium", Color: notion.ColorRed},
		{Name: "Postman", Color: notion.ColorPurple},
		{Name: "Python", Color: notion.ColorRed},
		{Name: "Typescript", Color: notion.ColorGray},
	}

	// Status options (using select since status can't be created via API)
	statusOptions := []notion.SelectOptions{
		{Name: "Active", Color: notion.ColorGreen},
		{Name: "Left", Color: notion.ColorGray},
		{Name: "Probation", Color: notion.ColorPink},
		{Name: "Offered", Color: notion.ColorGreen},
		{Name: "Intern", Color: notion.ColorBlue},
		{Name: "Apprentices", Color: notion.ColorBlue},
	}

	// Blood Type options
	bloodTypeOptions := []notion.SelectOptions{
		{Name: "A", Color: notion.ColorBlue},
		{Name: "A+", Color: notion.ColorPink},
		{Name: "AB", Color: notion.ColorPurple},
		{Name: "B", Color: notion.ColorYellow},
		{Name: "O+", Color: notion.ColorGreen},
		{Name: "O", Color: notion.ColorDefault},
		{Name: "unknown", Color: notion.ColorBrown},
	}

	// Horoscope options
	horoscopeOptions := []notion.SelectOptions{
		{Name: "Aries", Color: notion.ColorRed},
		{Name: "Taurus", Color: notion.ColorBlue},
		{Name: "Gemini", Color: notion.ColorPurple},
		{Name: "Cancer", Color: notion.ColorOrange},
		{Name: "Leo", Color: notion.ColorPink},
		{Name: "Virgo", Color: notion.ColorGreen},
		{Name: "Libra", Color: notion.ColorBrown},
		{Name: "Scorpio", Color: notion.ColorRed},
		{Name: "Sagittarius", Color: notion.ColorYellow},
		{Name: "Capricorn", Color: notion.ColorBlue},
		{Name: "Aquarius", Color: notion.ColorGreen},
		{Name: "Pisces", Color: notion.ColorGray},
	}

	// MBTI options
	mbtiOptions := []notion.SelectOptions{
		{Name: "Advocate", Color: notion.ColorGreen},
		{Name: "Defender", Color: notion.ColorRed},
		{Name: "Logician", Color: notion.ColorYellow},
		{Name: "Virtuoso", Color: notion.ColorBrown},
		{Name: "Architect", Color: notion.ColorPink},
		{Name: "Logistician", Color: notion.ColorDefault},
		{Name: "Mediator", Color: notion.ColorGray},
		{Name: "Debater", Color: notion.ColorPurple},
		{Name: "Entrepreneur", Color: notion.ColorBlue},
		{Name: "Protagonist", Color: notion.ColorOrange},
		{Name: "Executive", Color: notion.ColorBlue},
		{Name: "Consul", Color: notion.ColorOrange},
		{Name: "Campaigner", Color: notion.ColorPink},
		{Name: "Entertainer", Color: notion.ColorYellow},
		{Name: "Explorer", Color: notion.ColorBlue},
		{Name: "Commander", Color: notion.ColorRed},
		{Name: "Adventurer", Color: notion.ColorBrown},
	}

	// MBTI Abbr options
	mbtiAbbrOptions := []notion.SelectOptions{
		{Name: "INFJ", Color: notion.ColorDefault},
		{Name: "INTP", Color: notion.ColorDefault},
		{Name: "INTP-T", Color: notion.ColorDefault},
		{Name: "INFP", Color: notion.ColorDefault},
		{Name: "INFP-A", Color: notion.ColorDefault},
		{Name: "INTJ", Color: notion.ColorDefault},
		{Name: "INTJ-A", Color: notion.ColorDefault},
		{Name: "ISTP", Color: notion.ColorDefault},
		{Name: "ISTJ-A", Color: notion.ColorDefault},
		{Name: "ISFJ", Color: notion.ColorDefault},
		{Name: "ISFJ-A", Color: notion.ColorDefault},
		{Name: "ENTP", Color: notion.ColorDefault},
		{Name: "ENFJ", Color: notion.ColorDefault},
		{Name: "ENFJ-A", Color: notion.ColorDefault},
		{Name: "ENFP-T", Color: notion.ColorDefault},
		{Name: "ENFP-A", Color: notion.ColorDefault},
		{Name: "ESTP", Color: notion.ColorDefault},
		{Name: "ESTJ-A", Color: notion.ColorDefault},
		{Name: "ESFJ-A", Color: notion.ColorDefault},
		{Name: "ESFP", Color: notion.ColorDefault},
		{Name: "INFJ-A", Color: notion.ColorDefault},
		{Name: "INFJ-T", Color: notion.ColorDefault},
		{Name: "ENTJ-A", Color: notion.ColorDefault},
		{Name: "ENFJ-T", Color: notion.ColorDefault},
		{Name: "INFP-T", Color: notion.ColorDefault},
		{Name: "ISFP-A", Color: notion.ColorDefault},
		{Name: "ISTP-A", Color: notion.ColorDefault},
		{Name: "ISFP", Color: notion.ColorDefault},
		{Name: "ISTJ-T", Color: notion.ColorDefault},
		{Name: "ISFJ-T", Color: notion.ColorDefault},
		{Name: "ENTJ", Color: notion.ColorDefault},
		{Name: "ENTJ-T", Color: notion.ColorDefault},
	}

	// Referred By options
	referredByOptions := []notion.SelectOptions{
		{Name: "Tiêu Quang Huy", Color: notion.ColorDefault},
		{Name: "Nguyễn Tường Vân", Color: notion.ColorDefault},
		{Name: "Trần Hữu Minh", Color: notion.ColorDefault},
		{Name: "Lê Quang Thịnh", Color: notion.ColorDefault},
		{Name: "Võ Hải Biên", Color: notion.ColorDefault},
		{Name: "Phạm Ngọc Tài", Color: notion.ColorDefault},
		{Name: "Trần Hoàng Nam", Color: notion.ColorGray},
		{Name: "Đào Tuấn", Color: notion.ColorGray},
		{Name: "Trần Khắc Vỹ", Color: notion.ColorGray},
		{Name: "Huỳnh Quang Lâm", Color: notion.ColorDefault},
		{Name: "Nguyễn Ngọc Huy Hoàng", Color: notion.ColorGray},
		{Name: "Trần Xuân Nhược Nam", Color: notion.ColorGray},
		{Name: "Lưu Quang Minh", Color: notion.ColorGray},
		{Name: "Nguyễn Hoàng Anh", Color: notion.ColorOrange},
		{Name: "Nguyễn Đăng Quỳnh", Color: notion.ColorPurple},
		{Name: "Nguyễn Văn Lê Tây", Color: notion.ColorYellow},
		{Name: "Mai Đức Chiến", Color: notion.ColorBlue},
	}

	// Local employment options
	localEmploymentOptions := []notion.SelectOptions{
		{Name: "Vietnam 🇻🇳", Color: notion.ColorGray},
	}

	properties := map[string]notion.DatabaseProperty{
		"Full Name":         {Type: notion.DBPropTypeTitle, Title: &notion.EmptyMetadata{}},
		"Team Email":        {Type: notion.DBPropTypeEmail, Email: &notion.EmptyMetadata{}},
		"Personal Email":    {Type: notion.DBPropTypeEmail, Email: &notion.EmptyMetadata{}},
		"Discord":           {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"First name":        {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Last name":         {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Status":            {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: statusOptions}},
		"Position":          {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: positionOptions}},
		"Stack":             {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: stackOptions}},
		"Onboard Date":      {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Offboard Date":     {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Birthday":          {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Mobile #":          {Type: notion.DBPropTypePhoneNumber, PhoneNumber: &notion.EmptyMetadata{}},
		"Referred By":       {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: referredByOptions}},
		"Local employment":  {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: localEmploymentOptions}},
		"Blood Type":        {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: bloodTypeOptions}},
		"Horoscope":         {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: horoscopeOptions}},
		"MBTI":              {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: mbtiOptions}},
		"MBTI Abbr.":        {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: mbtiAbbrOptions}},
		"ID No.":            {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"New CCCD No.":      {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Permanent Address": {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Shelter Address":   {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Note":              {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"CV":                {Type: notion.DBPropTypeFiles, Files: &notion.EmptyMetadata{}},
		"ID Image":          {Type: notion.DBPropTypeFiles, Files: &notion.EmptyMetadata{}},
		"New CCCD Image":    {Type: notion.DBPropTypeFiles, Files: &notion.EmptyMetadata{}},
		"Headshot":          {Type: notion.DBPropTypeFiles, Files: &notion.EmptyMetadata{}},
		"Linkedin":          {Type: notion.DBPropTypeURL, URL: &notion.EmptyMetadata{}},
		"Twitter":           {Type: notion.DBPropTypeURL, URL: &notion.EmptyMetadata{}},
		"Facebook":          {Type: notion.DBPropTypeURL, URL: &notion.EmptyMetadata{}},
		"GitHub":            {Type: notion.DBPropTypeURL, URL: &notion.EmptyMetadata{}},
	}

	log.Printf("DEBUG: Creating Contractors with %d properties", len(properties))

	db, err := client.CreateDatabase(ctx, notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{Type: notion.RichTextTypeText, Text: &notion.Text{Content: "Contractors"}},
		},
		Properties: properties,
	})
	if err != nil {
		return "", fmt.Errorf("create database: %w", err)
	}

	return db.ID, nil
}

func createProjectDB(ctx context.Context, client *notion.Client, parentPageID string) (string, error) {
	log.Println("DEBUG: Building Project properties...")

	// Status options (using select)
	statusOptions := []notion.SelectOptions{
		{Name: "New", Color: notion.ColorYellow},
		{Name: "Paused", Color: notion.ColorBlue},
		{Name: "Active", Color: notion.ColorGreen},
		{Name: "Done", Color: notion.ColorGreen},
		{Name: "Closed", Color: notion.ColorDefault},
		{Name: "Failed", Color: notion.ColorRed},
	}

	// Size options
	sizeOptions := []notion.SelectOptions{
		{Name: "Big", Color: notion.ColorBlue},
		{Name: "Medium +", Color: notion.ColorPink},
		{Name: "Medium -", Color: notion.ColorGreen},
		{Name: "Small", Color: notion.ColorBrown},
	}

	// Tags options
	tagsOptions := []notion.SelectOptions{
		{Name: "Partnership", Color: notion.ColorGreen},
		{Name: "Ventures", Color: notion.ColorGray},
		{Name: "Enterprise", Color: notion.ColorDefault},
		{Name: "MVP", Color: notion.ColorGreen},
		{Name: "Pre-seed", Color: notion.ColorBlue},
		{Name: "Seed", Color: notion.ColorYellow},
		{Name: "Series A", Color: notion.ColorOrange},
		{Name: "Series B", Color: notion.ColorBlue},
		{Name: "CSR", Color: notion.ColorPink},
	}

	// Tech Stack options
	techStackOptions := []notion.SelectOptions{
		{Name: "AI", Color: notion.ColorYellow},
		{Name: "Blockchain", Color: notion.ColorGreen},
		{Name: "Elixir", Color: notion.ColorPurple},
		{Name: "Golang", Color: notion.ColorBlue},
		{Name: "NodeJS", Color: notion.ColorDefault},
		{Name: "React", Color: notion.ColorPink},
		{Name: "JavaScript", Color: notion.ColorRed},
		{Name: "Ruby", Color: notion.ColorOrange},
		{Name: "React Native", Color: notion.ColorBrown},
		{Name: "Vue", Color: notion.ColorGray},
		{Name: "TypeScript", Color: notion.ColorGray},
		{Name: "Java", Color: notion.ColorPurple},
		{Name: "Web", Color: notion.ColorOrange},
	}

	// Sales options
	salesOptions := []notion.SelectOptions{
		{Name: "Inbound", Color: notion.ColorBlue},
		{Name: "Han (han@d.foundation)", Color: notion.ColorBlue},
		{Name: "Nikki (nikki@d.foundation)", Color: notion.ColorDefault},
		{Name: "Minh Le (minhla@d.foundation)", Color: notion.ColorPurple},
		{Name: "Matt (himattlock@gmail.com)", Color: notion.ColorBrown},
		{Name: "Huy Tieu (huytq@d.foundation)", Color: notion.ColorDefault},
		{Name: "Duc Nghiem", Color: notion.ColorGreen},
		{Name: "Khai Le", Color: notion.ColorOrange},
		{Name: "Son Le", Color: notion.ColorPink},
		{Name: "N/A", Color: notion.ColorGray},
	}

	// Deal Closing (Account Manager) options
	dealClosingOptions := []notion.SelectOptions{
		{Name: "Han (han@d.foundation)", Color: notion.ColorRed},
		{Name: "Minh Le (minhla@d.foundation)", Color: notion.ColorPurple},
		{Name: "Nikki (nikki@d.foundation)", Color: notion.ColorDefault},
		{Name: "Tom Nguyen (anhnx@d.foundation)", Color: notion.ColorYellow},
		{Name: "Thanh Pham (thanhpd@d.foundation)", Color: notion.ColorBlue},
		{Name: "An Tran (antt@d.foundation)", Color: notion.ColorGreen},
		{Name: "Huy Tieu (huytq@d.foundation)", Color: notion.ColorOrange},
	}

	// PM/Delivery (Technical Lead) options
	pmDeliveryOptions := []notion.SelectOptions{
		{Name: "Huy Nguyen (huy@d.foundation)", Color: notion.ColorOrange},
		{Name: "Thanh Pham (thanhpd@d.foundation)", Color: notion.ColorBlue},
		{Name: "Tom Nguyen (anhnx@d.foundation)", Color: notion.ColorYellow},
		{Name: "Lap Nguyen (alan@d.foundation)", Color: notion.ColorGreen},
		{Name: "An Tran (antt@d.foundation)", Color: notion.ColorGreen},
		{Name: "Minh Luu (leo@d.foundation)", Color: notion.ColorBrown},
		{Name: "N/A", Color: notion.ColorGray},
		{Name: "Huy Tieu (huytq@d.foundation)", Color: notion.ColorYellow},
		{Name: "GiangT (giangtht@d.foundation)", Color: notion.ColorYellow},
		{Name: "Hung Vong (hungvt@d.foundation)", Color: notion.ColorBrown},
		{Name: "Tay Nguyen", Color: notion.ColorPurple},
		{Name: "Khai Le", Color: notion.ColorRed},
		{Name: "Minh Le (minhla@d.foundation)", Color: notion.ColorDefault},
		{Name: "Quang Le (quang@d.foundation)", Color: notion.ColorPink},
	}

	properties := map[string]notion.DatabaseProperty{
		"Project":                      {Type: notion.DBPropTypeTitle, Title: &notion.EmptyMetadata{}},
		"Status":                       {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: statusOptions}},
		"Size":                         {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: sizeOptions}},
		"Tags":                         {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: tagsOptions}},
		"Tech Stack":                   {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: techStackOptions}},
		"Sales":                        {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: salesOptions}},
		"Deal Closing (Account Manager)": {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: dealClosingOptions}},
		"PM/Delivery (Technical Lead)": {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: pmDeliveryOptions}},
		"Recipient Emails":             {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Manual Codename":              {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Change log":                   {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Closed Date":                  {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
	}

	log.Printf("DEBUG: Creating Project with %d properties", len(properties))

	db, err := client.CreateDatabase(ctx, notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{Type: notion.RichTextTypeText, Text: &notion.Text{Content: "Projects"}},
		},
		Properties: properties,
	})
	if err != nil {
		return "", fmt.Errorf("create database: %w", err)
	}

	return db.ID, nil
}

func createDeploymentTrackerDB(ctx context.Context, client *notion.Client, parentPageID, contractorsID, projectID string) (string, error) {
	log.Println("DEBUG: Building Deployment Tracker properties...")

	// Deployment Status options (using select)
	deploymentStatusOptions := []notion.SelectOptions{
		{Name: "Not started", Color: notion.ColorDefault},
		{Name: "Active", Color: notion.ColorBlue},
		{Name: "Done", Color: notion.ColorGreen},
	}

	// Type options
	typeOptions := []notion.SelectOptions{
		{Name: "Official", Color: notion.ColorBlue},
		{Name: "Part-time", Color: notion.ColorPurple},
		{Name: "Shadow", Color: notion.ColorYellow},
		{Name: "Not started", Color: notion.ColorDefault},
		{Name: "Done", Color: notion.ColorGreen},
	}

	// Position options
	positionOptions := []notion.SelectOptions{
		{Name: "AI Engineer", Color: notion.ColorPink},
		{Name: "Quality Engineer", Color: notion.ColorPurple},
		{Name: "Frontend Engineer", Color: notion.ColorBrown},
		{Name: "Backend Engineer", Color: notion.ColorBlue},
		{Name: "Fullstack Engineer", Color: notion.ColorGreen},
		{Name: "Technical Lead", Color: notion.ColorOrange},
		{Name: "Web3 Developer", Color: notion.ColorDefault},
	}

	properties := map[string]notion.DatabaseProperty{
		"Name":              {Type: notion.DBPropTypeTitle, Title: &notion.EmptyMetadata{}},
		"Deployment Status": {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: deploymentStatusOptions}},
		"Type":              {Type: notion.DBPropTypeMultiSelect, MultiSelect: &notion.SelectMetadata{Options: typeOptions}},
		"Position":          {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: positionOptions}},
		"Start Date":        {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"End Date":          {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Charges":           {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Contractor": {
			Type: notion.DBPropTypeRelation,
			Relation: &notion.RelationMetadata{
				DatabaseID:     contractorsID,
				Type:           notion.RelationTypeSingleProperty,
				SingleProperty: &struct{}{},
			},
		},
		"Project": {
			Type: notion.DBPropTypeRelation,
			Relation: &notion.RelationMetadata{
				DatabaseID:     projectID,
				Type:           notion.RelationTypeSingleProperty,
				SingleProperty: &struct{}{},
			},
		},
		"Upsell Person": {
			Type: notion.DBPropTypeRelation,
			Relation: &notion.RelationMetadata{
				DatabaseID:     contractorsID,
				Type:           notion.RelationTypeSingleProperty,
				SingleProperty: &struct{}{},
			},
		},
	}

	log.Printf("DEBUG: Creating Deployment Tracker with %d properties", len(properties))

	db, err := client.CreateDatabase(ctx, notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{Type: notion.RichTextTypeText, Text: &notion.Text{Content: "Deployment Tracker"}},
		},
		Properties: properties,
	})
	if err != nil {
		return "", fmt.Errorf("create database: %w", err)
	}

	return db.ID, nil
}

func createInvoiceDB(ctx context.Context, client *notion.Client, parentPageID, projectID string) (string, error) {
	log.Println("DEBUG: Building Invoice properties...")

	// Type options
	typeOptions := []notion.SelectOptions{
		{Name: "Invoice", Color: notion.ColorOrange},
		{Name: "Line Item", Color: notion.ColorGreen},
	}

	// Status options (using select)
	statusOptions := []notion.SelectOptions{
		{Name: "Draft", Color: notion.ColorGray},
		{Name: "Sent", Color: notion.ColorBlue},
		{Name: "Overdue", Color: notion.ColorRed},
		{Name: "Paid", Color: notion.ColorGreen},
		{Name: "Cancelled", Color: notion.ColorOrange},
	}

	// Currency options
	currencyOptions := []notion.SelectOptions{
		{Name: "USD", Color: notion.ColorGreen},
		{Name: "EUR", Color: notion.ColorBlue},
		{Name: "GBP", Color: notion.ColorPurple},
		{Name: "JPY", Color: notion.ColorRed},
		{Name: "CNY", Color: notion.ColorOrange},
		{Name: "SGD", Color: notion.ColorYellow},
		{Name: "CAD", Color: notion.ColorPink},
		{Name: "AUD", Color: notion.ColorBrown},
	}

	// Discount Type options
	discountTypeOptions := []notion.SelectOptions{
		{Name: "None", Color: notion.ColorDefault},
		{Name: "Percentage", Color: notion.ColorGreen},
		{Name: "Fixed Amount", Color: notion.ColorBlue},
		{Name: "Bulk Discount", Color: notion.ColorPurple},
		{Name: "Seasonal", Color: notion.ColorOrange},
		{Name: "Loyalty", Color: notion.ColorPink},
		{Name: "Early Payment", Color: notion.ColorYellow},
	}

	// Payment Method options
	paymentMethodOptions := []notion.SelectOptions{
		{Name: "Bank Transfer", Color: notion.ColorBlue},
		{Name: "Credit Card", Color: notion.ColorPurple},
		{Name: "Cash", Color: notion.ColorGreen},
		{Name: "Check", Color: notion.ColorGray},
		{Name: "PayPal", Color: notion.ColorBlue},
		{Name: "Venmo", Color: notion.ColorPink},
		{Name: "Crypto", Color: notion.ColorOrange},
		{Name: "Mobile Payment", Color: notion.ColorYellow},
	}

	properties := map[string]notion.DatabaseProperty{
		// Identity
		"Invoice Number": {Type: notion.DBPropTypeTitle, Title: &notion.EmptyMetadata{}},
		"Type":           {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: typeOptions}},

		// Status & Dates
		"Status":     {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: statusOptions}},
		"Issue Date": {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Due Date":   {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},
		"Paid Date":  {Type: notion.DBPropTypeDate, Date: &notion.EmptyMetadata{}},

		// Financial - Line Item Level
		"Quantity":   {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatNumber}},
		"Unit Price": {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatNumber}},
		"Tax Rate":   {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent}},
		"Currency":   {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: currencyOptions}},

		// Discounts
		"Discount Type":  {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: discountTypeOptions}},
		"Discount Value": {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatNumberWithCommas}},

		// Commission Percentages
		"% Sales":           {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent}},
		"% Account Mgr":     {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent}},
		"% Delivery Lead":   {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent}},
		"% Hiring Referral": {Type: notion.DBPropTypeNumber, Number: &notion.NumberMetadata{Format: notion.NumberFormatPercent}},

		// Text Fields
		"Description": {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Role":        {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Notes":       {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},
		"Sent by":     {Type: notion.DBPropTypeRichText, RichText: &notion.EmptyMetadata{}},

		// Payment
		"Payment Method": {Type: notion.DBPropTypeSelect, Select: &notion.SelectMetadata{Options: paymentMethodOptions}},

		// Other
		"Splits Generated": {Type: notion.DBPropTypeCheckbox, Checkbox: &notion.EmptyMetadata{}},

		// Relations
		"Project": {
			Type: notion.DBPropTypeRelation,
			Relation: &notion.RelationMetadata{
				DatabaseID:     projectID,
				Type:           notion.RelationTypeSingleProperty,
				SingleProperty: &struct{}{},
			},
		},
	}

	log.Printf("DEBUG: Creating Invoice with %d properties", len(properties))

	db, err := client.CreateDatabase(ctx, notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{Type: notion.RichTextTypeText, Text: &notion.Text{Content: "Invoices"}},
		},
		Properties: properties,
	})
	if err != nil {
		return "", fmt.Errorf("create database: %w", err)
	}

	return db.ID, nil
}

func addInvoiceSelfRelation(ctx context.Context, client *notion.Client, invoiceID string) error {
	log.Println("DEBUG: Adding self-referential relation (Parent item / Line Item)...")

	_, err := client.UpdateDatabase(ctx, invoiceID, notion.UpdateDatabaseParams{
		Properties: map[string]*notion.DatabaseProperty{
			"Parent item": {
				Type: notion.DBPropTypeRelation,
				Relation: &notion.RelationMetadata{
					DatabaseID: invoiceID,
					Type:       notion.RelationTypeDualProperty,
					DualProperty: &notion.DualPropertyRelation{
						SyncedPropName: "Line Item",
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("update database: %w", err)
	}

	return nil
}
