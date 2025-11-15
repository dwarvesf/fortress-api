package nocodb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dwarvesf/fortress-api/pkg/config"
)

// Service provides minimal access to NocoDB REST APIs.
type Service struct {
	client                        *http.Client
	baseURL                       string
	token                         string
	workspaceID                   string
	baseID                        string
	invoiceTableID                string
	invoiceCommentsTableID        string
	webhookSecret                 string
	accountingTodosTableID        string
	accountingTransactionsTableID string
}

// AttachmentUploadResult captures response payload from NocoDB's storage upload API.
type AttachmentUploadResult struct {
	Title      string `json:"title"`
	MIMEType   string `json:"mimetype"`
	Size       int64  `json:"size"`
	URL        string `json:"url"`
	SignedURL  string `json:"signedUrl"`
	Path       string `json:"path"`
	SignedPath string `json:"signedPath"`
}

var ErrNotFound = errors.New("nocodb: record not found")

// New creates a NocoDB service client. Returns nil if mandatory config is missing.
func New(cfg config.Noco) *Service {
	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	if baseURL == "" || cfg.Token == "" {
		return nil
	}

	return &Service{
		client:                        &http.Client{Timeout: 15 * time.Second},
		baseURL:                       baseURL,
		token:                         cfg.Token,
		workspaceID:                   cfg.WorkspaceID,
		baseID:                        cfg.BaseID,
		invoiceTableID:                cfg.InvoiceTableID,
		invoiceCommentsTableID:        cfg.InvoiceCommentsTableID,
		webhookSecret:                 cfg.WebhookSecret,
		accountingTodosTableID:        cfg.AccountingTodosTableID,
		accountingTransactionsTableID: cfg.AccountingTransactionsTableID,
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

// AccountingTodosTableID returns the configured table identifier for accounting todos.
func (s *Service) AccountingTodosTableID() string {
	if s == nil {
		return ""
	}
	return s.accountingTodosTableID
}

// AccountingTransactionsTableID returns the configured transactions table identifier.
func (s *Service) AccountingTransactionsTableID() string {
	if s == nil {
		return ""
	}
	return s.accountingTransactionsTableID
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
	if payload == nil {
		payload = map[string]interface{}{}
	}
	assignRecordID(payload, id)
	path := fmt.Sprintf("/tables/%s/records", s.invoiceTableID)
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

func (s *Service) UploadInvoiceAttachment(ctx context.Context, fileName, contentType string, content []byte) (*AttachmentUploadResult, error) {
	if s == nil {
		return nil, errors.New("nocodb service is nil")
	}
	if len(content) == 0 {
		return nil, errors.New("missing attachment content")
	}
	if fileName == "" {
		fileName = fmt.Sprintf("invoice-%d.pdf", time.Now().Unix())
	}
	sanitized := sanitizeFileName(fileName)
	path := s.buildInvoiceAttachmentPath(sanitized)
	return s.uploadAttachment(ctx, path, sanitized, contentType, content)
}

func (s *Service) CreateAccountingTodo(ctx context.Context, payload map[string]interface{}) (string, error) {
	if s.accountingTodosTableID == "" {
		return "", errors.New("nocodb accounting todos table id is empty")
	}
	path := fmt.Sprintf("/tables/%s/records", s.accountingTodosTableID)
	resp, err := s.makeRequest(ctx, http.MethodPost, path, nil, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("nocodb create accounting todo failed: %s - %s", resp.Status, string(body))
	}
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return extractRecordID(out), nil
}

func (s *Service) GetAccountingTodo(ctx context.Context, recordID string) (map[string]interface{}, error) {
	if s.accountingTodosTableID == "" {
		return nil, errors.New("nocodb accounting todos table id is empty")
	}
	if recordID == "" {
		return nil, errors.New("missing accounting todo id")
	}
	path := fmt.Sprintf("/tables/%s/records", s.accountingTodosTableID)
	query := url.Values{}
	query.Set("where", fmt.Sprintf("(Id,eq,%s)", recordID))
	query.Set("limit", "1")
	resp, err := s.makeRequest(ctx, http.MethodGet, path, query, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("nocodb get accounting todo failed: %s - %s", resp.Status, string(body))
	}
	var out struct {
		List []map[string]interface{} `json:"list"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.List) == 0 {
		return nil, ErrNotFound
	}
	return out.List[0], nil
}

func (s *Service) UpdateAccountingTodo(ctx context.Context, recordID string, payload map[string]interface{}) error {
	if s.accountingTodosTableID == "" {
		return errors.New("nocodb accounting todos table id is empty")
	}
	if recordID == "" {
		return errors.New("missing accounting todo id")
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	assignRecordID(payload, recordID)
	path := fmt.Sprintf("/tables/%s/records", s.accountingTodosTableID)
	resp, err := s.makeRequest(ctx, http.MethodPatch, path, nil, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nocodb update accounting todo failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func (s *Service) uploadAttachment(ctx context.Context, path, fileName, contentType string, content []byte) (*AttachmentUploadResult, error) {
	if s == nil {
		return nil, errors.New("nocodb service is nil")
	}
	if path == "" {
		return nil, errors.New("missing attachment path")
	}
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("data", fileName)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(content); err != nil {
		return nil, err
	}
	_ = writer.WriteField("title", fileName)
	if contentType != "" {
		_ = writer.WriteField("mimetype", contentType)
	}
	_ = writer.WriteField("size", fmt.Sprintf("%d", len(content)))
	if err := writer.Close(); err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/storage/upload", s.baseURL)
	query := url.Values{}
	query.Set("path", path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+query.Encode(), &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("xc-token", s.token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nocodb upload attachment failed: %s - %s", resp.Status, string(respBody))
	}
	if res, err := parseAttachmentUploadResult(respBody); err == nil && res != nil {
		return res, nil
	}
	return nil, fmt.Errorf("nocodb upload attachment: unexpected response %s", string(respBody))
}

func parseAttachmentUploadResult(body []byte) (*AttachmentUploadResult, error) {
	var single AttachmentUploadResult
	if err := json.Unmarshal(body, &single); err == nil && single != (AttachmentUploadResult{}) {
		return &single, nil
	}
	var list []AttachmentUploadResult
	if err := json.Unmarshal(body, &list); err == nil {
		if len(list) > 0 {
			return &list[0], nil
		}
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported attachment response")
}

func (s *Service) buildInvoiceAttachmentPath(fileName string) string {
	workspace := firstNonEmpty(s.workspaceID, "workspace")
	base := firstNonEmpty(s.baseID, s.invoiceTableID, "base")
	return fmt.Sprintf("download/noco/%s/%s/invoice_tasks/%d-%s", workspace, base, time.Now().UnixNano(), fileName)
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "" {
		name = "invoice.pdf"
	}
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	sanitized := strings.Trim(b.String(), "-_")
	if sanitized == "" {
		sanitized = "invoice"
	}
	if filepath.Ext(sanitized) == "" {
		sanitized += ".pdf"
	}
	return sanitized
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func (r *AttachmentUploadResult) AccessibleURL(baseURL string) string {
	if r == nil {
		return ""
	}
	if strings.TrimSpace(r.URL) != "" {
		return r.URL
	}
	if strings.TrimSpace(r.SignedURL) != "" {
		return r.SignedURL
	}
	path := strings.TrimPrefix(r.Path, "/")
	if path == "" {
		path = strings.TrimPrefix(r.SignedPath, "/")
	}
	if path == "" {
		return ""
	}
	host := baseURL
	for _, suffix := range []string{"/api/v2", "/api/v1"} {
		host = strings.TrimSuffix(host, suffix)
	}
	host = strings.TrimSuffix(host, "/")
	if host == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", host, path)
}

func (r *AttachmentUploadResult) ToMap() map[string]any {
	if r == nil {
		return nil
	}
	out := map[string]any{}
	if r.Title != "" {
		out["title"] = r.Title
	}
	if r.MIMEType != "" {
		out["mimetype"] = r.MIMEType
	}
	if r.Size != 0 {
		out["size"] = r.Size
	}
	if r.URL != "" {
		out["url"] = r.URL
	}
	if r.SignedURL != "" {
		out["signedUrl"] = r.SignedURL
	}
	if r.Path != "" {
		out["path"] = r.Path
	}
	if r.SignedPath != "" {
		out["signedPath"] = r.SignedPath
	}
	return out
}

func assignRecordID(payload map[string]interface{}, id string) {
	if payload == nil {
		return
	}
	if i, err := strconv.Atoi(id); err == nil {
		payload["Id"] = i
		return
	}
	payload["Id"] = id
}
