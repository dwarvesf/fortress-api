package nocodb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
)

// Service provides minimal access to NocoDB REST APIs.
type Service struct {
	client                 *http.Client
	baseURL                string
	token                  string
	invoiceTableID         string
	invoiceCommentsTableID string
	webhookSecret          string
}

var ErrNotFound = errors.New("nocodb: record not found")

// New creates a NocoDB service client. Returns nil if mandatory config is missing.
func New(cfg config.Noco) *Service {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" || cfg.Token == "" {
		return nil
	}

	return &Service{
		client:                 &http.Client{Timeout: 15 * time.Second},
		baseURL:                baseURL,
		token:                  cfg.Token,
		invoiceTableID:         cfg.InvoiceTableID,
		invoiceCommentsTableID: cfg.InvoiceCommentsTableID,
		webhookSecret:          cfg.InvoiceWebhookSecret,
	}
}

// BaseURL exposes the configured base API URL.
func (s *Service) BaseURL() string {
	if s == nil {
		return ""
	}
	return s.baseURL
}

// Token exposes the configured API token (mainly for downstream helpers).
func (s *Service) Token() string {
	if s == nil {
		return ""
	}
	return s.token
}

// InvoiceTableID returns the configured invoice table identifier.
func (s *Service) InvoiceTableID() string {
	if s == nil {
		return ""
	}
	return s.invoiceTableID
}

// InvoiceCommentsTableID returns the configured comments table identifier.
func (s *Service) InvoiceCommentsTableID() string {
	if s == nil {
		return ""
	}
	return s.invoiceCommentsTableID
}

// WebhookSecret returns the shared secret used to verify incoming payloads.
func (s *Service) WebhookSecret() string {
	if s == nil {
		return ""
	}
	return s.webhookSecret
}

func (s *Service) makeRequest(ctx context.Context, method, path string, query url.Values, body interface{}) (*http.Response, error) {
	if s == nil {
		return nil, errors.New("nocodb service is nil")
	}
	var buf io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewReader(payload)
	}
	endpoint := s.baseURL + path
	if len(query) > 0 {
		endpoint = endpoint + "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("xc-token", s.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.client.Do(req)
}

func (s *Service) findInvoiceRecord(ctx context.Context, invoiceNumber string) (string, map[string]interface{}, error) {
	if s.invoiceTableID == "" {
		return "", nil, errors.New("nocodb invoice table id is empty")
	}
	path := fmt.Sprintf("/tables/%s/records", s.invoiceTableID)
	query := url.Values{}
	query.Set("where", fmt.Sprintf("(invoice_number,eq,%s)", invoiceNumber))
	query.Set("limit", "1")
	resp, err := s.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", nil, ErrNotFound
	}
	if resp.StatusCode >= 300 {
		return "", nil, fmt.Errorf("nocodb list records failed: %s", resp.Status)
	}
	var out struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", nil, err
	}
	if len(out.List) == 0 {
		return "", nil, ErrNotFound
	}
	record := out.List[0]
	return extractRecordID(record), record, nil
}

func (s *Service) createInvoiceRecord(ctx context.Context, payload map[string]interface{}) (string, error) {
	if s.invoiceTableID == "" {
		return "", errors.New("nocodb invoice table id is empty")
	}
	path := fmt.Sprintf("/tables/%s/records", s.invoiceTableID)
	resp, err := s.makeRequest(ctx, http.MethodPost, path, nil, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("nocodb create record failed: %s - %s", resp.Status, string(body))
	}
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return extractRecordID(out), nil
}

func (s *Service) updateInvoiceRecord(ctx context.Context, id string, payload map[string]interface{}) error {
	if s.invoiceTableID == "" {
		return errors.New("nocodb invoice table id is empty")
	}
	path := fmt.Sprintf("/tables/%s/records/%s", s.invoiceTableID, id)
	resp, err := s.makeRequest(ctx, http.MethodPatch, path, nil, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nocodb update record failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func (s *Service) createInvoiceComment(ctx context.Context, payload map[string]interface{}) error {
	if s.invoiceCommentsTableID == "" {
		return nil
	}
	path := fmt.Sprintf("/tables/%s/records", s.invoiceCommentsTableID)
	resp, err := s.makeRequest(ctx, http.MethodPost, path, nil, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nocodb create comment failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func extractRecordID(record map[string]interface{}) string {
	if record == nil {
		return ""
	}
	if v, ok := record["Id"]; ok {
		return fmt.Sprintf("%v", v)
	}
	if v, ok := record["id"]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func (s *Service) UpsertInvoiceRecord(ctx context.Context, invoiceNumber string, payload map[string]interface{}) (string, error) {
	id, _, err := s.findInvoiceRecord(ctx, invoiceNumber)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return s.createInvoiceRecord(ctx, payload)
		}
		return "", err
	}
	if err := s.updateInvoiceRecord(ctx, id, payload); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) UpdateInvoiceStatus(ctx context.Context, id string, status string) error {
	if id == "" {
		return errors.New("missing record id")
	}
	return s.updateInvoiceRecord(ctx, id, map[string]interface{}{"status": status})
}

func (s *Service) CreateInvoiceComment(ctx context.Context, recordID string, author, message, msgType string) error {
	if s.invoiceCommentsTableID == "" {
		return nil
	}
	if recordID == "" {
		return errors.New("missing invoice record id")
	}
	payload := map[string]interface{}{
		"invoice_task_id": recordID,
		"author":          author,
		"message":         message,
		"type":            msgType,
	}
	return s.createInvoiceComment(ctx, payload)
}
