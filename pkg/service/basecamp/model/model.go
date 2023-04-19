package model

import (
	"encoding/json"
	"time"
)

const (
	CompanyID              = "4108948"
	GetBasecampUserInfoURL = "https://launchpad.37signals.com/authorization.json"
	BasecampEndpoint       = "https://launchpad.37signals.com/authorization"
	BasecampAPIEndpoint    = "https://3.basecampapi.com"
)

type Person struct {
	ID             int    `json:"id"`
	AttachableSgID string `json:"attachable_sgid"`
	Name           string `json:"name"`
	EmailAddress   string `json:"email_address"`
	Title          string `json:"title"`
	Bio            string `json:"bio"`
}

// UserInfo fully define basecamp user info struct
type UserInfo struct {
	ExpiresAt time.Time `json:"expires_at"`
	Identity  Identity  `json:"identity"`
}

// Identity define User Identity
type Identity struct {
	ID           int    `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
}

// AuthenticationResponse define basecamp auth response
type AuthenticationResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// Basecamp define basecamp config struct
type Basecamp struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type SubscriptionList struct {
	Subscriptions   []int `json:"subscriptions"`
	Unsubscriptions []int `json:"unsubscriptions"`
}

type TodoList struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Title           string  `json:"title"`
	Type            string  `json:"type"`
	CreatedAt       string  `json:"created_at"`
	TodosURL        string  `json:"todos_url"`
	UpdatedAt       string  `json:"updated_at"`
	Parent          *Parent `json:"parent"`
	SubscriptionURL string  `json:"subscription_url"`
}

type TodoGroup struct {
	ID             int       `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Title          string    `json:"title"`
	InheritsStatus bool      `json:"inherits_status"`
	Type           string    `json:"type"`
	Parent         *Parent   `json:"parent"`
	Completed      bool      `json:"completed"`
	CompletedRatio string    `json:"completed_ratio"`
	SubscriberIDs  []int     `json:"subscriber_ids"`
	Name           string    `json:"name"`
}

type Todo struct {
	ID                    int          `json:"id"`
	Title                 string       `json:"title"`
	Type                  string       `json:"type"`
	Assignees             []Assignee   `json:"assignees"`
	AssigneeIDs           []int        `json:"assignee_ids"`
	CompletionSubscribers []Subscriber `json:"completion_subscribers"`
	Completed             bool         `json:"completed"`
	AppURL                string       `json:"app_url"`
	Content               string       `json:"content"`
	CommentsURL           string       `json:"comments_url"`
	Description           string       `json:"description"`
	DueOn                 string       `json:"due_on"`
	InheritsStatus        bool         `json:"inherits_status"`
	StartsOn              string       `json:"starts_on"`
	Status                string       `json:"status"`
	Parent                *Parent      `json:"parent"`
	SubscriptionURL       string       `json:"subscription_url"`
	CreatedAt             string       `json:"created_at"`
	UpdatedAt             string       `json:"updated_at"`
	Notify                bool         `json:"notify"`
	Bucket                Bucket       `json:"bucket"`
}

type Subscriber struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	EmailAddress   string    `json:"email_address"`
	PersonableType string    `json:"personable_type"`
	Title          string    `json:"title"`
	Bio            string    `json:"bio"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Admin          bool      `json:"admin"`
	Owner          bool      `json:"owner"`
	TimeZone       string    `json:"time_zone"`
	AvatarURL      string    `json:"avatar_url"`
}

type Comment struct {
	ID             int64     `json:"id"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Title          string    `json:"title"`
	InheritsStatus bool      `json:"inherits_status"`
	Type           string    `json:"type"`
	URL            string    `json:"url"`
	AppURL         string    `json:"app_url"`
	BookmarkURL    string    `json:"bookmark_url"`
	Parent         Parent    `json:"parent"`
	Bucket         Bucket    `json:"bucket"`
	Creator        Assignee  `json:"creator"`
	Content        string    `json:"content"`
}

type Parent struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
	URL   string `json:"url"`
}

type Bucket struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Assignee struct {
	Admin          bool        `json:"admin"`
	AttachableSgid string      `json:"attachable_sgid"`
	AvatarURL      string      `json:"avatar_url"`
	Bio            interface{} `json:"bio"`
	CreatedAt      string      `json:"created_at"`
	EmailAddress   string      `json:"email_address"`
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Owner          bool        `json:"owner"`
	PersonableType string      `json:"personable_type"`
	TimeZone       string      `json:"time_zone"`
	Title          string      `json:"title"`
	UpdatedAt      string      `json:"updated_at"`
}

type Project struct {
	CreatedAt   string        `json:"created_at"`
	Description string        `json:"description"`
	Dock        []ProjectDock `json:"dock"`
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Purpose     string        `json:"purpose"`
	Status      string        `json:"status"`
	UpdatedAt   string        `json:"updated_at"`
	URL         string        `json:"url"`
}

type ProjectDock struct {
	AppURL   string `json:"app_url"`
	Enabled  bool   `json:"enabled"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Title    string `json:"title"`
	URL      string `json:"url"`
}

type ScheduleEntry struct {
	ID                 json.Number         `json:"id"`
	Summary            string              `json:"summary"`
	Description        string              `json:"description"`
	ParticipantIDs     []int               `json:"participant_ids"`
	Participants       []*Assignee         `json:"participants"`
	AllDay             bool                `json:"all_day"`
	Notify             bool                `json:"notify"`
	AppUrl             string              `json:"app_url"`
	StartsAt           string              `json:"starts_at"`
	EndsAt             string              `json:"ends_at"`
	RecurrenceSchedule *RecurrenceSchedule `json:"recurrence_schedule"`
	SubscriptionUrl    string              `json:"subscription_url"`
}

type RecurrenceSchedule struct {
	Days         []int  `json:"days"`
	WeekInstance int    `json:"week_instance"`
	Frequency    string `json:"frequency"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
}

type PeopleCreate struct {
	Name         string `json:"name"`
	EmailAddress string `json:"email_address"`
	CompanyName  string `json:"company_name"`
}

type PeopleEntry struct {
	Grant  []int64        `json:"grant"`
	Revoke []int64        `json:"revoke"`
	Create []PeopleCreate `json:"create"`
}

type CampfireLine struct {
	Content string `json:"content"`
}

type Hook struct {
	ID         int       `json:"id"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	PayloadURL string    `json:"payload_url"`
	Types      []string  `json:"types"`
	URL        string    `json:"url"`
	AppURL     string    `json:"app_url"`
}

type Message struct {
	ID          int        `json:"id"`
	Subject     string     `json:"subject"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	AppURL      string     `json:"app_url"`
	CommentsURL string     `json:"comments_url"`
	CreatedAt   *time.Time `json:"created_at"`
}

type Recording struct {
	ID               int       `json:"id"`
	Status           string    `json:"status"`
	VisibleToClients bool      `json:"visible_to_clients"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Title            string    `json:"title"`
	InheritsStatus   bool      `json:"inherits_status"`
	Type             string    `json:"type"`
	URL              string    `json:"url"`
	AppURL           string    `json:"app_url"`
	BookmarkURL      string    `json:"bookmark_url"`
	SubscriptionURL  string    `json:"subscription_url"`
	CommentsCount    int       `json:"comments_count"`
	CommentsURL      string    `json:"comments_url"`
	Position         int       `json:"position,omitempty"`
	Parent           Parent    `json:"parent"`
	Bucket           Bucket    `json:"bucket"`
	Creator          Person    `json:"creator"`
	Description      string    `json:"description"`
	Completed        bool      `json:"completed"`
	Content          string    `json:"content"`
	StartsOn         string    `json:"starts_on"`
	DueOn            string    `json:"due_on"`
}

type Event struct {
	ID          int       `json:"id"`
	RecordingID int       `json:"recording_id"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	Creator     Person    `json:"creator"`
}

// below is list of message model, use with caution
const (
	BasecampCommentMsg    string = "basecamp_comment"
	BasecampTodoMsg       string = "basecamp_todo"
	BasecampHiringTodoMsg string = "basecamp_todo_hiring"
)

const (
	CommentMsgTypeFailed    string = "failed"
	CommentMsgTypeCompleted string = "completed"
)

// BasecampCommentMessage is use for worker to create a basecamp comment
type BasecampCommentMessage struct {
	ProjectID   int
	RecordingID int
	Payload     *Comment
}

type BasecampTodoMessageModel struct {
	ProjectID int
	ListID    int
	Payload   Todo
}
