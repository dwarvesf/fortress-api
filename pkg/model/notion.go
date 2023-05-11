package model

import "time"

// ProjectChangelogPage -- notion project changelog page
type ProjectChangelogPage struct {
	RowID        string `json:"row_id"`
	Name         string `json:"name"`
	Title        string `json:"title"`
	ChangelogURL string `json:"changelog_url"`
}

type NotionMemo struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tags      []string  `json:"tags"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type NotionUpdate struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Audience  string    `json:"audience"`
	CreatedAt time.Time `json:"created_at"`
}

type NotionEarn struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Reward   int          `json:"reward"`
	Progress int          `json:"progress"`
	Priority string       `json:"priority"`
	Tags     []string     `json:"tags"`
	PICs     []Employee   `json:"pics"`
	Status   string       `json:"status"`
	Function []string     `json:"function"`
	SubItems []NotionEarn `json:"sub_items"`
	ParentID string       `json:"-"`
	DueDate  *time.Time   `json:"due_date"`
}

type NotionAudience struct {
	ID        string    `json:"id"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	Sources   []string  `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

type NotionTechRadar struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Assign     string   `json:"assign"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
	Quadrant   string   `json:"quadrant"`
	Ring       string   `json:"ring"`
}

type NotionDigest struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type NotionStaffingDemand struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Request string `json:"request"`
}

type NotionIssue struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	RootCause    string     `json:"rootcause"`
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

type HiringType string

const (
	HiringTypeDirect     HiringType = "direct"
	HiringTypeReferral   HiringType = "referral"
	HiringTypeInternship HiringType = "internship"
)

type NotionHiringRelationship struct {
	BaseModel

	UserID     UUID       `json:"user_id"`
	SupplierID UUID       `json:"supplier_id"`
	HiringType HiringType `json:"hiring_type"`

	Employee Employee `json:"user"`
}

type NotionHiringPosition struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Projects  []string  `json:"project"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type NotionProjectMilestone struct {
	ID            string                    `json:"id"`
	Project       string                    `json:"-"`
	Name          string                    `json:"name"`
	StartDate     time.Time                 `json:"start_date"`
	EndDate       time.Time                 `json:"end_date"`
	SubMilestones []*NotionProjectMilestone `json:"sub_milestones"`
}

func (o *NotionProjectMilestone) GetSubMilestones() []*NotionProjectMilestone {
	return o.SubMilestones
}
