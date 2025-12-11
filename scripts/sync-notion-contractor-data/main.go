package main

import (
	"bytes"
	"database/sql"
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
	_ "github.com/lib/pq"
)

const (
	notionAPIURL   = "https://api.notion.com/v1"
	notionVersion  = "2022-06-28"
	contractorDBID = "9d468753ebb44977a8dc156428398a6b"
	dwarvesGuildID = "462663954813157376"
	avatarSize     = "128"
	rateLimitDelay = 350 * time.Millisecond // Notion rate limit: ~3 requests/sec
)

var (
	dryRun         bool
	targetUsername string
	updateAvatar   bool
	updateOnboard  bool
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

type NotionDateProperty struct {
	Date *NotionDate `json:"date"`
}

type NotionDate struct {
	Start    string  `json:"start"`
	End      *string `json:"end,omitempty"`
	TimeZone *string `json:"time_zone,omitempty"`
}

type EmployeeData struct {
	FullName        string
	DisplayName     string
	JoinedDate      *time.Time
	DiscordUsername string
	WorkingStatus   string
	TeamEmail       string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&dryRun, "dry-run", false, "Run without making changes to Notion")
	flag.StringVar(&targetUsername, "username", "", "Only process contractor with this Discord username")
	flag.BoolVar(&updateAvatar, "avatar", true, "Update Discord avatar icon (default: true)")
	flag.BoolVar(&updateOnboard, "onboard", true, "Update onboard date from database (default: true)")
	flag.Parse()

	if dryRun {
		log.Printf("INFO: Running in DRY-RUN mode - no changes will be made")
	}
	if targetUsername != "" {
		log.Printf("INFO: Targeting specific Discord username: %s", targetUsername)
	}
	log.Printf("INFO: Update avatar: %v, Update onboard: %v", updateAvatar, updateOnboard)

	if err := godotenv.Load(); err != nil {
		log.Printf("DEBUG: No .env file found, using environment variables")
	}

	notionToken := os.Getenv("NOTION_SECRET")
	if notionToken == "" {
		log.Fatal("NOTION_SECRET environment variable is required")
	}

	var discord *discordgo.Session
	var db *sql.DB
	var err error

	// Initialize Discord if avatar update is enabled
	if updateAvatar {
		discordToken := os.Getenv("DISCORD_SECRET_TOKEN")
		if discordToken == "" {
			log.Fatal("DISCORD_SECRET_TOKEN environment variable is required for avatar updates")
		}

		log.Printf("DEBUG: Creating Discord session")
		discord, err = discordgo.New("Bot " + discordToken)
		if err != nil {
			log.Fatalf("Error creating Discord session: %v", err)
		}
	}

	// Initialize database if onboard date update is enabled
	if updateOnboard {
		dbDSN := os.Getenv("DATABASE_URL")
		if dbDSN == "" {
			log.Fatal("DATABASE_URL environment variable is required for onboard date updates")
		}

		log.Printf("DEBUG: Connecting to database")
		db, err = sql.Open("postgres", dbDSN)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Fatalf("Error pinging database: %v", err)
		}
	}

	log.Printf("DEBUG: Fetching contractors from Notion")
	contractors, err := fetchAllContractors(notionToken)
	if err != nil {
		log.Fatalf("Error fetching contractors: %v", err)
	}
	log.Printf("DEBUG: Found %d contractors", len(contractors))

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

		var avatarURL string
		var employeeData *EmployeeData
		var needsUpdate bool

		// Get Discord avatar if enabled
		if updateAvatar {
			avatarURL, err = getDiscordAvatarURL(discord, discordUsername)
			if err != nil {
				log.Printf("DEBUG: [%s] Error searching Discord: %v", name, err)
			} else if avatarURL == "" {
				log.Printf("DEBUG: [%s] Discord user '%s' not found", name, discordUsername)
			} else {
				// Check if icon already set to this URL
				currentIconURL := getCurrentIconURL(contractor)
				if currentIconURL != avatarURL {
					needsUpdate = true
					log.Printf("DEBUG: [%s] Avatar needs update: %s", name, avatarURL)
				} else {
					log.Printf("DEBUG: [%s] Avatar already set", name)
				}
			}
		}

		// Get employee data from database if enabled
		if updateOnboard {
			employeeData, err = getEmployeeDataByDiscord(db, discordUsername)
			if err != nil {
				log.Printf("DEBUG: [%s] Error fetching employee data: %v", name, err)
			} else if employeeData == nil {
				log.Printf("DEBUG: [%s] No employee data found for Discord username: %s", name, discordUsername)
			} else {
				// Check if onboard date needs update
				currentOnboardDate := getCurrentOnboardDate(contractor)
				if employeeData.JoinedDate != nil {
					expectedDate := employeeData.JoinedDate.Format("2006-01-02")
					if currentOnboardDate != expectedDate {
						needsUpdate = true
						log.Printf("DEBUG: [%s] Onboard date needs update: %s (current: %s)", name, expectedDate, currentOnboardDate)
					} else {
						log.Printf("DEBUG: [%s] Onboard date already set: %s", name, currentOnboardDate)
					}
				}
			}
		}

		// Skip if nothing to update
		if !needsUpdate {
			if avatarURL == "" && employeeData == nil {
				notFound++
			} else {
				alreadySet++
			}
			continue
		}

		if dryRun {
			logMessage := fmt.Sprintf("DRY-RUN: [%s] Would update:", name)
			if avatarURL != "" && updateAvatar {
				logMessage += fmt.Sprintf(" avatar=%s", avatarURL)
			}
			if employeeData != nil && employeeData.JoinedDate != nil && updateOnboard {
				logMessage += fmt.Sprintf(" onboard=%s", employeeData.JoinedDate.Format("2006-01-02"))
			}
			log.Printf("%s", logMessage)
			updated++
			continue
		}

		// Update Notion page
		log.Printf("DEBUG: [%s] Updating Notion page", name)
		if err := updateNotionPage(notionToken, contractor.ID, avatarURL, employeeData, updateAvatar, updateOnboard); err != nil {
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

func getCurrentOnboardDate(page NotionPage) string {
	onboardProp, ok := page.Properties["Onboard Date"].(map[string]interface{})
	if !ok {
		return ""
	}

	dateProp, ok := onboardProp["date"]
	if !ok || dateProp == nil {
		return ""
	}

	dateMap, ok := dateProp.(map[string]interface{})
	if !ok {
		return ""
	}

	start, _ := dateMap["start"].(string)
	return start
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

func getEmployeeDataByDiscord(db *sql.DB, discordUsername string) (*EmployeeData, error) {
	query := `
		SELECT
			e.full_name,
			e.display_name,
			e.joined_date,
			da.discord_username,
			e.working_status,
			e.team_email
		FROM employees e
		LEFT JOIN discord_accounts da ON e.discord_account_id = da.id
		WHERE LOWER(da.discord_username) = LOWER($1)
		AND e.deleted_at IS NULL
		LIMIT 1
	`

	var emp EmployeeData
	var joinedDate sql.NullTime
	var workingStatus sql.NullString
	var teamEmail sql.NullString

	err := db.QueryRow(query, discordUsername).Scan(
		&emp.FullName,
		&emp.DisplayName,
		&joinedDate,
		&emp.DiscordUsername,
		&workingStatus,
		&teamEmail,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if joinedDate.Valid {
		emp.JoinedDate = &joinedDate.Time
	}
	if workingStatus.Valid {
		emp.WorkingStatus = workingStatus.String
	}
	if teamEmail.Valid {
		emp.TeamEmail = teamEmail.String
	}

	log.Printf("DEBUG: Found employee data: %s, joined: %v", emp.FullName, emp.JoinedDate)
	return &emp, nil
}

func updateNotionPage(token, pageID, avatarURL string, employeeData *EmployeeData, doUpdateAvatar, doUpdateOnboard bool) error {
	payload := NotionPatchRequest{}

	// Add icon if avatar update is enabled and URL is provided
	if doUpdateAvatar && avatarURL != "" {
		payload.Icon = &NotionIcon{
			Type: "external",
			External: &NotionExternalFile{
				URL: avatarURL,
			},
		}
	}

	// Add onboard date if update is enabled and data is available
	if doUpdateOnboard && employeeData != nil && employeeData.JoinedDate != nil {
		dateStr := employeeData.JoinedDate.Format("2006-01-02")
		payload.Properties = map[string]interface{}{
			"Onboard Date": NotionDateProperty{
				Date: &NotionDate{
					Start: dateStr,
				},
			},
		}
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
