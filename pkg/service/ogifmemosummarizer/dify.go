package ogifmemosummarizer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
)

type difyService struct {
	appConfig *config.Config
}

// New function return dify service
func New(appConfig *config.Config) IService {
	return &difyService{
		appConfig: appConfig,
	}
}

// BaseEvent represents the common fields in the event
type BaseEvent struct {
	Event          string `json:"event,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	MessageID      string `json:"message_id,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	TaskID         string `json:"task_id,omitempty"`
	ID             string `json:"id,omitempty"`
	Position       int    `json:"position,omitempty"`
}

// AgentThought represents the specific fields for agent_thought events
type AgentThought struct {
	BaseEvent
	Thought      string      `json:"thought,omitempty"`
	Observation  string      `json:"observation,omitempty"`
	Tool         string      `json:"tool,omitempty"`
	ToolLabels   interface{} `json:"tool_labels,omitempty"`
	ToolInput    string      `json:"tool_input,omitempty"`
	MessageFiles interface{} `json:"message_files,omitempty"`
}

// AgentMessage represents the specific fields for agent_message events
type AgentMessage struct {
	BaseEvent
	Answer string `json:"answer,omitempty"`
}

func (d *difyService) SummarizeOGIFMemo(youtubeURL string) (content string, err error) {
	// Define the URL and request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"inputs":          map[string]interface{}{},
		"query":           youtubeURL,
		"response_mode":   "streaming",
		"conversation_id": "",
		"user":            "abc-123",
	})
	if err != nil {
		return "", err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", d.appConfig.Dify.URL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", nil
	}

	// Set the required headers
	req.Header.Set("Authorization", "Bearer "+d.appConfig.Dify.Token)
	req.Header.Set("Content-Type", "application/json")

	// Send the request using http.Client
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Handle streaming response
	fmt.Println("Streaming response...")
	thoughts := []AgentThought{}
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}

		// Remove the "data: " prefix
		line = bytes.TrimPrefix(line, []byte("data: "))
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse the JSON event to a map to determine the event type
		var rawEvent map[string]interface{}
		err = json.Unmarshal(line, &rawEvent)
		if err != nil {
			fmt.Printf("Error unmarshal: %v\n", string(line))
			continue
		}

		eventType, ok := rawEvent["event"].(string)
		if !ok {
			fmt.Println("Error: event type is missing or not a string")
			continue
		}

		// Process specific event types
		switch eventType {
		case "agent_thought":
			var event AgentThought
			err = json.Unmarshal(line, &event)
			if err != nil {
				fmt.Printf("Error parsing agent_thought JSON: %v\n", err)
				continue
			}
			thoughts = append(thoughts, event)
		case "agent_message":
			// just ignore agent_message event
			// var event AgentMessage
			// err = json.Unmarshal(line, &event)
			// if err != nil {
			// 	fmt.Printf("Error parsing agent_message JSON: %v\n", err)
			// 	continue
			// }
			// fmt.Printf("Agent Message: %s\n", event.Answer)
		default:
			// fmt.Printf("Unknown event type: %s\n", eventType)
		}
	}

	// get the last event
	if len(thoughts) == 0 {
		return "", fmt.Errorf("no thought found")
	}

	fmt.Println("Processing thoughts...")
	for i := len(thoughts) - 1; i >= 0; i-- {
		if thoughts[i].Thought != "" {
			content = thoughts[i].Thought
			break
		}
	}

	fmt.Println("Completed processing thoughts")
	return content, nil
}
