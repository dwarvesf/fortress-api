package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type MemoLog struct {
	BaseModel

	Title       string
	URL         string
	Authors     MemoLogAuthors
	Tags        JSONArrayString
	Description string
	PublishedAt *time.Time
	Reward      decimal.Decimal
}

type MemoLogAuthors []MemoLogAuthor

// MemoLogAuthor is the author of the memo log
type MemoLogAuthor struct {
	EmployeeID string `json:"employeeID"`
	GithubID   string `json:"githubID"`
	DiscordID  string `json:"discordID"`
}

func (j MemoLogAuthors) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *MemoLogAuthors) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch t := value.(type) {
	case []uint8:
		jsonData := value.([]uint8)
		if string(jsonData) == "null" {
			return nil
		}
		return json.Unmarshal(jsonData, j)
	default:
		return fmt.Errorf("could not scan type %T into json", t)
	}
}
