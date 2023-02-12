package model

// ProjectChangelogPage -- notion project changelog page
type ProjectChangelogPage struct {
	RowID        string `json:"row_id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	ChangelogURL string `json:"changelog_url"`
}
