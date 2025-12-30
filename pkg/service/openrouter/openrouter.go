package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1/chat/completions"
	maxRetries        = 3
	initialBackoff    = 1 * time.Second
)

// Free models to rotate through if rate limits are hit
var freeModels = []string{
	"google/gemini-2.5-flash",
	"meta-llama/llama-3.3-70b-instruct:free",
	"deepseek/deepseek-r1:free",
	"qwen/qwen-2.5-72b-instruct:free",
	"google/gemini-2.0-flash-exp:free",
}

// OpenRouterService handles LLM summarization via OpenRouter API
type OpenRouterService struct {
	cfg    *config.Config
	logger logger.Logger
	client *http.Client
}

// NewOpenRouterService creates a new OpenRouter service
func NewOpenRouterService(cfg *config.Config, logger logger.Logger) *OpenRouterService {
	logger.Debug("creating new OpenRouterService")

	return &OpenRouterService{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatCompletionRequest represents the OpenRouter API request
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the OpenRouter API response
type ChatCompletionResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Message Message `json:"message"`
}

// APIError represents an API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ProofOfWorkEntry represents a single proof of work entry with hours
type ProofOfWorkEntry struct {
	Text  string
	Hours float64
}

// SummarizeProofOfWorks summarizes multiple proof of work entries into major work bullet points
// Hours are used as weight to prioritize more significant work
func (s *OpenRouterService) SummarizeProofOfWorks(ctx context.Context, entries []ProofOfWorkEntry) (string, error) {
	if len(entries) == 0 {
		s.logger.Debug("no entries to summarize, returning empty string")
		return "", nil
	}

	s.logger.Debug(fmt.Sprintf("summarizing %d proof of work entries", len(entries)))

	// Combine all entries into a single prompt with hours for weighting
	var combinedText strings.Builder
	var totalHours float64
	for i, entry := range entries {
		if entry.Text == "" {
			continue
		}
		totalHours += entry.Hours
		combinedText.WriteString(fmt.Sprintf("--- Entry %d (%.1f hours) ---\n%s\n\n", i+1, entry.Hours, entry.Text))
	}

	s.logger.Debug(fmt.Sprintf("total hours across all entries: %.1f", totalHours))

	if combinedText.Len() == 0 {
		s.logger.Debug("all texts are empty, returning empty string")
		return "", nil
	}

	prompt := combinedText.String()

	systemPrompt := `Role: You are a professional technical editor creating concise monthly work summaries for invoices.

Task: Group work logs by scope/category, then list key activities as comma-separated items.

Format: [Scope]: [activity 1], [activity 2], [activity 3]

Guidelines:
- Each entry includes hours spent - use hours as weight to prioritize more significant work
- Entries with more hours should be featured more prominently in the summary
- Group related tasks under clear scope labels (e.g., "Backend Infrastructure", "Invoice System", "API Development")
- List only 3-4 key activities per scope (most significant ones based on hours)
- Keep each activity to 2-3 words maximum
- Use professional, client-facing language
- Avoid technical jargon unless critical
- Maximum 2 scopes total

Example:
• Backend Infrastructure: Upload optimization, data retention, search capabilities
• Invoice System: USDC support, email routing, discount logic

Output: Bullet points only (use •), no introduction or headers.`

	// Call OpenRouter with retry
	summary, err := s.callWithRetry(ctx, systemPrompt, prompt)
	if err != nil {
		s.logger.Error(err, "failed to summarize proof of works, returning original text")
		// Return original text as fallback
		return combinedText.String(), nil
	}

	s.logger.Debug(fmt.Sprintf("successfully summarized %d entries into %d characters", len(entries), len(summary)))
	return summary, nil
}

// callWithRetry calls OpenRouter API with exponential backoff retry and model rotation
func (s *OpenRouterService) callWithRetry(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	var lastErr error
	backoff := initialBackoff

	// Try each free model in rotation
	for modelIdx, model := range freeModels {
		s.logger.Debug(fmt.Sprintf("trying model %d/%d: %s", modelIdx+1, len(freeModels), model))

		summary, err := s.callAPIWithModel(ctx, systemPrompt, userPrompt, model)
		if err == nil && strings.TrimSpace(summary) != "" {
			s.logger.Debug(fmt.Sprintf("successfully got response from model: %s", model))
			return summary, nil
		}

		lastErr = err
		if err != nil {
			s.logger.Debug(fmt.Sprintf("model %s failed: %v", model, err))
		} else {
			s.logger.Debug(fmt.Sprintf("model %s returned empty response", model))
		}

		// Wait before trying next model (except for last one)
		if modelIdx < len(freeModels)-1 {
			s.logger.Debug(fmt.Sprintf("waiting %v before trying next model", backoff))
			time.Sleep(backoff)
		}
	}

	if lastErr != nil {
		return "", fmt.Errorf("all models failed, last error: %w", lastErr)
	}
	return "", fmt.Errorf("all models returned empty responses")
}

// callAPIWithModel calls the OpenRouter API with a specific model
func (s *OpenRouterService) callAPIWithModel(ctx context.Context, systemPrompt, userPrompt, model string) (string, error) {
	if s.cfg.OpenRouter.APIKey == "" {
		return "", fmt.Errorf("OpenRouter API key not configured")
	}

	s.logger.Debug(fmt.Sprintf("calling OpenRouter API with model: %s", model))

	reqBody := ChatCompletionRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:   200,
		Temperature: 0,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.cfg.OpenRouter.APIKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("OpenRouter API response status: %d", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s (type: %s, code: %s)", result.Error.Message, result.Error.Type, result.Error.Code)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no completion choices returned")
	}

	summary := strings.TrimSpace(result.Choices[0].Message.Content)
	s.logger.Debug(fmt.Sprintf("received summary: %d characters", len(summary)))

	return summary, nil
}
