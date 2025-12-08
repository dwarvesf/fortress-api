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
		fmt.Println("Usage: go run scripts/notion_schema.go <database_id> [query]")
		os.Exit(1)
	}

	_ = godotenv.Load()

	secret := os.Getenv("NOTION_SECRET")
	if secret == "" {
		fmt.Println("NOTION_SECRET not set")
		os.Exit(1)
	}

	databaseID := os.Args[1]
	client := notion.NewClient(secret)

	// If "query" arg provided, fetch data instead of schema
	if len(os.Args) > 2 && os.Args[2] == "query" {
		res, err := client.QueryDatabase(context.Background(), databaseID, &notion.DatabaseQuery{
			PageSize: 10,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		out, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(out))
		return
	}

	db, err := client.FindDatabaseByID(context.Background(), databaseID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DEBUG: Database Title: %s\n", db.Title[0].PlainText)
	fmt.Printf("DEBUG: Database ID: %s\n", db.ID)
	fmt.Println("\n=== Relations Found ===")
	for name, prop := range db.Properties {
		if prop.Type == notion.DBPropTypeRelation {
			fmt.Printf("  %s -> %s\n", name, prop.Relation.DatabaseID)
		}
	}
	fmt.Println("")

	out, _ := json.MarshalIndent(db, "", "  ")
	fmt.Println(string(out))
}
