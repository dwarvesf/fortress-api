package model

import "time"

type Issue struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Rootcause    string     `json:"rootcause"`
	Resolution   string     `json:"resolution"`
	Scope        string     `json:"scope"`
	Priority     string     `json:"priority"`
	Severity     string     `json:"severity"`
	IncidentDate *time.Time `json:"incident_date"`
	SolvedDate   *time.Time `json:"solve_date"`
	PIC          string     `json:"pic"`
	Projects     []string   `json:"projects"`
	Status       string     `json:"status"`
	Source       string     `json:"source"`
	Profile      string     `json:"profile"`
}
