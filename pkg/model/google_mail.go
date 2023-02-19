package model

// GoogleMailThread --
type GoogleMailThread struct {
	ID       string              `json:"id"`
	Messages []GoogleMailMessage `json:"messages"`
}

// GoogleMailMessage --
type GoogleMailMessage struct {
	ID       string   `json:"id"`
	ThreadID string   `json:"threadId"`
	Payload  *Payload `json:"payload"`
}

type Payload struct {
	Headers []Header `json:"headers"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
