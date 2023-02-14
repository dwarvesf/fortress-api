package model

type TechRadar struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Assign     string   `json:"assign"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Quadrant   string   `json:"quadrant"`
	Ring       string   `json:"ring"`
}
