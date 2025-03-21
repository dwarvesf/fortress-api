package webhook

import "time"

type n8nCalendarEvent struct {
	ID      string    `json:"id"`
	Kind    string    `json:"kind"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	Summary string    `json:"summary"`
	Creator struct {
		Email string `json:"email"`
	} `json:"creator"`
	Description string `json:"description"`
	HangoutLink string `json:"hangoutLink"`
	Organizer   struct {
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		Self        bool   `json:"self"`
	} `json:"organizer"`
	Start struct {
		DateTime time.Time `json:"dateTime"`
		Timezone string    `json:"timezone"`
	} `json:"start"`
	End struct {
		DateTime time.Time `json:"dateTime"`
		Timezone string    `json:"timezone"`
	} `json:"end"`
	Attendees []struct {
		Email          string `json:"email"`
		ResponseStatus string `json:"responseStatus"`
	} `json:"attendees"`
	ShouldSyncDiscord string `json:"shouldSyncDiscord"`
}

type n8nEvent struct {
	Kind              string            `json:"kind"`
	CalendarData      *n8nCalendarEvent `json:"calendarData"`
	ShouldSyncDiscord string            `json:"shouldSyncDiscord"`
}

type TransactionRequestEvent struct {
	ProfileID   string `json:"profile_id"`
	RequestCode string `json:"request_code"`
	Status      string `json:"status"`
	TxID        int    `json:"tx_id"`
	Description string `json:"description"`
	Timestamp   int64  `json:"timestamp"`
	Amount      string `json:"amount"`
	TokenName   string `json:"token_name"`
	TokenID     string `json:"token_id"`
}
