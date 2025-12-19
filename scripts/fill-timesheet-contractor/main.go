package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	notionAPIURL   = "https://api.notion.com/v1"
	notionVersion  = "2022-06-28"
	timesheetDBID  = "2c664b29b84c8089b304e9c5b5c70ac3"
	contractorDBID = "9d468753ebb44977a8dc156428398a6b"
	rateLimitDelay = 350 * time.Millisecond // Notion rate limit: ~3 requests/sec
)

var (
	dryRun         bool
	targetUsername string
)

type NotionQueryResponse struct {
	Results    []NotionPage `json:"results"`
	HasMore    bool         `json:"has_more"`
	NextCursor *string      `json:"next_cursor"`
}

type NotionPage struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	CreatedBy  NotionUser             `json:"created_by"`
}

type NotionUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

type NotionPatchRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

type NotionRelationProperty struct {
	Relation []NotionRelation `json:"relation"`
}

type NotionRelation struct {
	ID string `json:"id"`
}

type NotionSelectProperty struct {
	Select *NotionSelect `json:"select"`
}

type NotionSelect struct {
	Name string `json:"name"`
}

type ContractorInfo struct {
	PageID          string
	FullName        string
	DiscordUsername string
	PersonIDs       []string
	PersonNames     []string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&dryRun, "dry-run", false, "Run without making changes to Notion")
	flag.StringVar(&targetUsername, "username", "", "Only process timesheet entries created by this Notion username")
	flag.Parse()

	if dryRun {
		log.Printf("INFO: Running in DRY-RUN mode - no changes will be made")
	}
	if targetUsername != "" {
		log.Printf("INFO: Targeting specific Notion username: %s", targetUsername)
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found, using environment variables")
	}

	notionToken := os.Getenv("NOTION_SECRET")
	if notionToken == "" {
		log.Fatal("NOTION_SECRET environment variable is required")
	}

	// Step 1: Fetch all contractors and build person->contractor mapping
	log.Printf("DEBUG: Fetching contractors from Notion")
	contractors, err := fetchAllContractors(notionToken)
	if err != nil {
		log.Fatalf("Error fetching contractors: %v", err)
	}
	log.Printf("DEBUG: Found %d contractors", len(contractors))

	personToContractor := buildPersonToContractorMap(contractors)
	log.Printf("DEBUG: Built mapping for %d person IDs", len(personToContractor))

	// Get target person IDs if username is specified
	var targetPersonIDs []string
	if targetUsername != "" {
		targetPersonIDs = getPersonIDsByUsername(contractors, targetUsername)
		if len(targetPersonIDs) == 0 {
			log.Fatalf("ERROR: No contractor found with username: %s", targetUsername)
		}
		log.Printf("DEBUG: Found %d person IDs for username %s", len(targetPersonIDs), targetUsername)
	}

	// Step 2: Fetch all timesheet entries
	log.Printf("DEBUG: Fetching timesheet entries from Notion")
	timesheetEntries, err := fetchAllTimesheetEntries(notionToken)
	if err != nil {
		log.Fatalf("Error fetching timesheet entries: %v", err)
	}
	log.Printf("DEBUG: Found %d timesheet entries", len(timesheetEntries))

	var updated, skipped, notFound, alreadySet int

	for _, entry := range timesheetEntries {
		entryName := getPageTitle(entry)
		createdByID := entry.CreatedBy.ID

		if createdByID == "" {
			log.Printf("DEBUG: [%s] No created_by user ID, skipping", entryName)
			skipped++
			continue
		}

		// Filter by target username if specified
		if targetUsername != "" {
			found := false
			for _, targetID := range targetPersonIDs {
				if createdByID == targetID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		log.Printf("DEBUG: [%s] Processing entry created by: %s", entryName, createdByID)

		// Check if contractor and discord are already filled
		currentContractor := getCurrentContractor(entry)
		currentDiscord := getCurrentDiscord(entry)

		if currentContractor != "" && currentDiscord != "" {
			log.Printf("DEBUG: [%s] Contractor and Discord already set", entryName)
			alreadySet++
			continue
		}

		// Find matching contractor
		contractor, found := personToContractor[createdByID]
		if !found {
			log.Printf("DEBUG: [%s] No contractor found for person ID: %s", entryName, createdByID)
			notFound++
			continue
		}

		log.Printf("DEBUG: [%s] Found contractor: %s (Discord: %s)", entryName, contractor.FullName, contractor.DiscordUsername)

		// Determine what needs updating
		needsContractorUpdate := currentContractor == ""
		needsDiscordUpdate := currentDiscord == "" && contractor.DiscordUsername != ""

		if !needsContractorUpdate && !needsDiscordUpdate {
			log.Printf("DEBUG: [%s] No updates needed", entryName)
			alreadySet++
			continue
		}

		if dryRun {
			logMessage := fmt.Sprintf("DRY-RUN: [%s] Would update:", entryName)
			if needsContractorUpdate {
				logMessage += fmt.Sprintf(" contractor=%s", contractor.FullName)
			}
			if needsDiscordUpdate {
				logMessage += fmt.Sprintf(" discord=%s", contractor.DiscordUsername)
			}
			log.Printf("%s", logMessage)
			updated++
			continue
		}

		// Update Notion page
		log.Printf("DEBUG: [%s] Updating Notion page", entryName)
		if err := updateTimesheetEntry(notionToken, entry.ID, contractor, needsContractorUpdate, needsDiscordUpdate); err != nil {
			log.Printf("ERROR: [%s] Failed to update: %v", entryName, err)
			continue
		}

		log.Printf("INFO: [%s] Updated successfully", entryName)
		updated++

		time.Sleep(rateLimitDelay)
	}

	log.Printf("\n=== Summary ===")
	if dryRun {
		log.Printf("Mode: DRY-RUN (no changes made)")
	}
	log.Printf("Total timesheet entries: %d", len(timesheetEntries))
	log.Printf("Updated: %d", updated)
	log.Printf("Already set: %d", alreadySet)
	log.Printf("Not found: %d", notFound)
	log.Printf("Skipped (no created_by): %d", skipped)
}

func fetchAllContractors(token string) ([]NotionPage, error) {
	var allPages []NotionPage
	var cursor *string

	for {
		log.Printf("DEBUG: Fetching page of contractors (cursor: %v)", cursor)

		body := map[string]interface{}{
			"page_size": 100,
		}
		if cursor != nil {
			body["start_cursor"] = *cursor
		}

		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", notionAPIURL, contractorDBID), bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Notion-Version", notionVersion)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("notion API error: %s - %s", resp.Status, string(bodyBytes))
		}

		var result NotionQueryResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		allPages = append(allPages, result.Results...)
		log.Printf("DEBUG: Fetched %d contractors (total: %d)", len(result.Results), len(allPages))

		if !result.HasMore {
			break
		}
		cursor = result.NextCursor

		time.Sleep(rateLimitDelay)
	}

	return allPages, nil
}

func fetchAllTimesheetEntries(token string) ([]NotionPage, error) {
	var allPages []NotionPage
	var cursor *string

	for {
		log.Printf("DEBUG: Fetching page of timesheet entries (cursor: %v)", cursor)

		body := map[string]interface{}{
			"page_size": 100,
		}
		if cursor != nil {
			body["start_cursor"] = *cursor
		}

		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", notionAPIURL, timesheetDBID), bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Notion-Version", notionVersion)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("notion API error: %s - %s", resp.Status, string(bodyBytes))
		}

		var result NotionQueryResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		allPages = append(allPages, result.Results...)
		log.Printf("DEBUG: Fetched %d timesheet entries (total: %d)", len(result.Results), len(allPages))

		if !result.HasMore {
			break
		}
		cursor = result.NextCursor

		time.Sleep(rateLimitDelay)
	}

	return allPages, nil
}

func buildPersonToContractorMap(contractors []NotionPage) map[string]ContractorInfo {
	mapping := make(map[string]ContractorInfo)

	for _, contractor := range contractors {
		fullName := getContractorFullName(contractor)
		discordUsername := getContractorDiscord(contractor)
		personIDs, personNames := getContractorPersonInfo(contractor)

		info := ContractorInfo{
			PageID:          contractor.ID,
			FullName:        fullName,
			DiscordUsername: discordUsername,
			PersonIDs:       personIDs,
			PersonNames:     personNames,
		}

		for i, personID := range personIDs {
			mapping[personID] = info
			personName := "unknown"
			if i < len(personNames) {
				personName = personNames[i]
			}
			log.Printf("DEBUG: Mapped person %s (%s) to contractor %s", personID, personName, fullName)
		}
	}

	return mapping
}

func getPersonIDsByUsername(contractors []NotionPage, username string) []string {
	for _, contractor := range contractors {
		personIDs, personNames := getContractorPersonInfo(contractor)

		// Check if any of the person names match the username
		for _, personName := range personNames {
			if strings.EqualFold(personName, username) {
				fullName := getContractorFullName(contractor)
				log.Printf("DEBUG: Found contractor %s with Notion username %s, person IDs: %v", fullName, personName, personIDs)
				return personIDs
			}
		}
	}
	return nil
}

func getPageTitle(page NotionPage) string {
	// Try to get auto-generated title
	titleProp, ok := page.Properties["(auto) Timesheet Entry"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	titleArray, ok := titleProp["title"].([]interface{})
	if !ok || len(titleArray) == 0 {
		return "Unknown"
	}

	firstTitle, ok := titleArray[0].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	plainText, ok := firstTitle["plain_text"].(string)
	if !ok {
		return "Unknown"
	}

	return plainText
}

func getContractorFullName(page NotionPage) string {
	titleProp, ok := page.Properties["Full Name"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	titleArray, ok := titleProp["title"].([]interface{})
	if !ok || len(titleArray) == 0 {
		return "Unknown"
	}

	firstTitle, ok := titleArray[0].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	plainText, ok := firstTitle["plain_text"].(string)
	if !ok {
		return "Unknown"
	}

	return plainText
}

func getContractorDiscord(page NotionPage) string {
	discordProp, ok := page.Properties["Discord"].(map[string]interface{})
	if !ok {
		return ""
	}

	richText, ok := discordProp["rich_text"].([]interface{})
	if !ok || len(richText) == 0 {
		return ""
	}

	firstText, ok := richText[0].(map[string]interface{})
	if !ok {
		return ""
	}

	plainText, ok := firstText["plain_text"].(string)
	if !ok {
		return ""
	}

	return strings.TrimSpace(plainText)
}

func getContractorPersonInfo(page NotionPage) ([]string, []string) {
	personProp, ok := page.Properties["Person"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	peopleArray, ok := personProp["people"].([]interface{})
	if !ok {
		return nil, nil
	}

	var personIDs []string
	var personNames []string
	for _, person := range peopleArray {
		personMap, ok := person.(map[string]interface{})
		if !ok {
			continue
		}

		id, ok := personMap["id"].(string)
		if ok && id != "" {
			personIDs = append(personIDs, id)
		}

		name, ok := personMap["name"].(string)
		if ok && name != "" {
			personNames = append(personNames, name)
		}
	}

	return personIDs, personNames
}

func getCurrentContractor(page NotionPage) string {
	contractorProp, ok := page.Properties["Contractor"].(map[string]interface{})
	if !ok {
		return ""
	}

	relation, ok := contractorProp["relation"].([]interface{})
	if !ok || len(relation) == 0 {
		return ""
	}

	firstRelation, ok := relation[0].(map[string]interface{})
	if !ok {
		return ""
	}

	id, ok := firstRelation["id"].(string)
	if !ok {
		return ""
	}

	return id
}

func getCurrentDiscord(page NotionPage) string {
	discordProp, ok := page.Properties["Discord"].(map[string]interface{})
	if !ok {
		return ""
	}

	selectProp, ok := discordProp["select"]
	if !ok || selectProp == nil {
		return ""
	}

	selectMap, ok := selectProp.(map[string]interface{})
	if !ok {
		return ""
	}

	name, ok := selectMap["name"].(string)
	if !ok {
		return ""
	}

	return name
}

func updateTimesheetEntry(token, pageID string, contractor ContractorInfo, updateContractor, updateDiscord bool) error {
	properties := make(map[string]interface{})

	if updateContractor {
		properties["Contractor"] = NotionRelationProperty{
			Relation: []NotionRelation{
				{ID: contractor.PageID},
			},
		}
	}

	if updateDiscord && contractor.DiscordUsername != "" {
		properties["Discord"] = NotionSelectProperty{
			Select: &NotionSelect{
				Name: contractor.DiscordUsername,
			},
		}
	}

	payload := NotionPatchRequest{
		Properties: properties,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("DEBUG: Updating timesheet entry %s with payload: %s", pageID, string(jsonBody))

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s/pages/%s", notionAPIURL, pageID), bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", notionVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notion API error: %s - %s", resp.Status, string(bodyBytes))
	}

	log.Printf("DEBUG: Timesheet entry %s updated successfully", pageID)
	return nil
}
