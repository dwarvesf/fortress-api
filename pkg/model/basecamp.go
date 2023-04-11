package model

import "time"

// BasecampRecording is a data structure define basecamp todo
type BasecampRecording struct {
	ID        int            `json:"id,omitempty"`
	Status    string         `json:"status,omitempty"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
	Title     string         `json:"title,omitempty"`
	URL       string         `json:"url,omitempty"`
	Parent    BasecampParent `json:"parent,omitempty"`
	Creator   BasecampUser   `json:"creator,omitempty"`
	Bucket    BasecampBucket `json:"bucket,omitempty"`
	Content   string         `json:"content,omitempty"`
}

type BasecampParent struct {
	ID    int64  `json:"id,omitempty"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

type BasecampUser struct {
	ID    int    `json:"id,omitempty"`
	Email string `json:"email_address,omitempty"`
	Name  string `json:"name,omitempty"`
}

// BasecampBucket is
type BasecampBucket struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}
