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

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

const (
	notionAPIURL        = "https://api.notion.com/v1"
	notionVersion       = "2022-06-28"
	contractorDBID      = "9d468753ebb44977a8dc156428398a6b"
	candidateDBID       = "2b764b29b84c802cb4e8fd4b4d9501cb"
	dwarvesGuildID      = "462663954813157376"
	avatarSize          = "128"
	rateLimitDelay      = 350 * time.Millisecond // Notion rate limit: ~3 requests/sec
)

var (
	dryRun         bool
	targetUsername string
	database       string
)

type NotionQueryResponse struct {
	Results    []NotionPage `json:"results"`
	HasMore    bool         `json:"has_more"`
	NextCursor *string      `json:"next_cursor"`
}

type NotionPage struct {
	ID         string                 `json:"id"`
	Properties map[string]interface{} `json:"properties"`
	Icon       interface{}            `json:"icon"`
}

type NotionPatchRequest struct {
	Icon       *NotionIcon `json:"icon,omitempty"`
	Properties interface{} `json:"properties,omitempty"`
}

type NotionIcon struct {
	Type     string              `json:"type"`
	External *NotionExternalFile `json:"external,omitempty"`
}

type NotionExternalFile struct {
	URL string `json:"url"`
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&dryRun, "dry-run", false, "Run without making changes to Notion")
	flag.StringVar(&targetUsername, "username", "", "Only process contractor with this Discord username")
	flag.StringVar(&database, "database", "contractors", "Database to update: 'contractors' or 'candidates'")
	flag.Parse()

	if dryRun {
		log.Printf("INFO: Running in DRY-RUN mode - no changes will be made")
	}
	if targetUsername != "" {
		log.Printf("INFO: Targeting specific Discord username: %s", targetUsername)
	}

	// Determine database ID
	var dbID string
	var dbName string
	switch database {
	case "contractors":
		dbID = contractorDBID
		dbName = "Contractors"
	case "candidates":
		dbID = candidateDBID
		dbName = "Candidates"
	default:
		log.Fatalf("Invalid database: %s (must be 'contractors' or 'candidates')", database)
	}
	log.Printf("INFO: Updating %s database", dbName)

	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found, using environment variables")
	}

	notionToken := os.Getenv("NOTION_SECRET")
	if notionToken == "" {
		log.Fatal("NOTION_SECRET environment variable is required")
	}

	discordToken := os.Getenv("DISCORD_SECRET_TOKEN")
	if discordToken == "" {
		log.Fatal("DISCORD_SECRET_TOKEN environment variable is required")
	}

	log.Printf("DEBUG: Creating Discord session")
	discord, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	log.Printf("DEBUG: Fetching records from Notion %s database", dbName)
	contractors, err := fetchAllContractors(notionToken, dbID)
	if err != nil {
		log.Fatalf("Error fetching records: %v", err)
	}
	log.Printf("DEBUG: Found %d records", len(contractors))

	var updated, skipped, notFound, alreadySet int

	for _, contractor := range contractors {
		name := getPageTitle(contractor)
		discordUsername := getDiscordUsername(contractor)

		if discordUsername == "" {
			log.Printf("DEBUG: [%s] No Discord username, skipping", name)
			skipped++
			continue
		}

		// Filter by target username if specified
		if targetUsername != "" && !strings.EqualFold(discordUsername, targetUsername) {
			continue
		}

		log.Printf("DEBUG: [%s] Processing Discord username: %s", name, discordUsername)

		avatarURL, err := getDiscordAvatarURL(discord, discordUsername)
		if err != nil {
			log.Printf("DEBUG: [%s] Error searching Discord: %v", name, err)
			notFound++
			continue
		}

		if avatarURL == "" {
			log.Printf("DEBUG: [%s] Discord user '%s' not found", name, discordUsername)
			notFound++
			continue
		}

		// Check if icon already set to this URL
		currentIconURL := getCurrentIconURL(contractor)
		if currentIconURL == avatarURL {
			log.Printf("DEBUG: [%s] Avatar already set", name)
			alreadySet++
			continue
		}

		log.Printf("DEBUG: [%s] Avatar needs update: %s", name, avatarURL)

		if dryRun {
			log.Printf("DRY-RUN: [%s] Would update avatar=%s", name, avatarURL)
			updated++
			continue
		}

		// Update Notion page
		log.Printf("DEBUG: [%s] Updating Notion page", name)
		if err := updateNotionPage(notionToken, contractor.ID, avatarURL); err != nil {
			log.Printf("ERROR: [%s] Failed to update: %v", name, err)
			continue
		}

		log.Printf("INFO: [%s] Updated successfully", name)
		updated++

		time.Sleep(rateLimitDelay)
	}

	log.Printf("\n=== Summary ===")
	if dryRun {
		log.Printf("Mode: DRY-RUN (no changes made)")
	}
	log.Printf("Total contractors: %d", len(contractors))
	log.Printf("Updated: %d", updated)
	log.Printf("Already set: %d", alreadySet)
	log.Printf("Not found: %d", notFound)
	log.Printf("Skipped (no Discord username): %d", skipped)
}

func fetchAllContractors(token string, databaseID string) ([]NotionPage, error) {
	var allPages []NotionPage
	var cursor *string

	for {
		log.Printf("DEBUG: Fetching page of records (cursor: %v)", cursor)

		body := map[string]interface{}{
			"page_size": 100,
		}
		if cursor != nil {
			body["start_cursor"] = *cursor
		}

		jsonBody, _ := json.Marshal(body)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/databases/%s/query", notionAPIURL, databaseID), bytes.NewReader(jsonBody))
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
		log.Printf("DEBUG: Fetched %d records (total: %d)", len(result.Results), len(allPages))

		if !result.HasMore {
			break
		}
		cursor = result.NextCursor

		time.Sleep(rateLimitDelay)
	}

	return allPages, nil
}

func getPageTitle(page NotionPage) string {
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

func getDiscordUsername(page NotionPage) string {
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

func getCurrentIconURL(page NotionPage) string {
	if page.Icon == nil {
		return ""
	}

	iconMap, ok := page.Icon.(map[string]interface{})
	if !ok {
		return ""
	}

	iconType, _ := iconMap["type"].(string)
	if iconType != "external" {
		return ""
	}

	external, ok := iconMap["external"].(map[string]interface{})
	if !ok {
		return ""
	}

	url, _ := external["url"].(string)
	return url
}

func getDiscordAvatarURL(session *discordgo.Session, username string) (string, error) {
	log.Printf("DEBUG: Searching Discord guild %s for user: %s", dwarvesGuildID, username)

	members, err := session.GuildMembersSearch(dwarvesGuildID, username, 10)
	if err != nil {
		return "", fmt.Errorf("guild member search failed: %w", err)
	}

	log.Printf("DEBUG: Found %d members matching '%s'", len(members), username)

	// Find exact match
	for _, member := range members {
		log.Printf("DEBUG: Checking member: %s (ID: %s)", member.User.Username, member.User.ID)
		if strings.EqualFold(member.User.Username, username) {
			if member.User.Avatar == "" {
				log.Printf("DEBUG: User %s has no avatar set", username)
				return "", nil
			}
			avatarURL := member.User.AvatarURL(avatarSize)
			log.Printf("DEBUG: Found avatar URL: %s", avatarURL)
			return avatarURL, nil
		}
	}

	return "", nil
}

func updateNotionPage(token, pageID, avatarURL string) error {
	payload := NotionPatchRequest{
		Icon: &NotionIcon{
			Type: "external",
			External: &NotionExternalFile{
				URL: avatarURL,
			},
		},
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Printf("DEBUG: Updating Notion page %s with payload: %s", pageID, string(jsonBody))

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

	log.Printf("DEBUG: Notion page %s updated successfully", pageID)
	return nil
}
