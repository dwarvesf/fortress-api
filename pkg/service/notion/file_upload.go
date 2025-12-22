package notion

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	nt "github.com/dstotijn/go-notion"

	"github.com/dwarvesf/fortress-api/pkg/logger"
)

// FileUploadResponse represents the response from Notion file upload API
type FileUploadResponse struct {
	Object         string    `json:"object"`
	ID             string    `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	ExpiryTime     time.Time `json:"expiry_time"`
	UploadURL      string    `json:"upload_url"`
	Archived       bool      `json:"archived"`
	Status         string    `json:"status"`
	Filename       *string   `json:"filename"`
	ContentType    *string   `json:"content_type"`
	ContentLength  *int64    `json:"content_length"`
}

// UploadFile uploads a file to Notion and returns the file upload ID
// This is a 3-step process:
// 1. Create file upload object to get upload URL and ID
// 2. Send file content to the upload URL
// 3. Poll until status becomes "uploaded"
func (n *notionService) UploadFile(filename, contentType string, fileData []byte) (string, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "UploadFile",
	})

	l.Debug(fmt.Sprintf("uploading file to Notion: filename=%s, contentType=%s, size=%d bytes", filename, contentType, len(fileData)))

	// Step 1: Create file upload object
	fileUploadID, uploadURL, err := n.createFileUploadObject(filename, contentType)
	if err != nil {
		l.Error(err, "failed to create file upload object")
		return "", err
	}

	l.Debug(fmt.Sprintf("created file upload object: id=%s, uploadURL=%s", fileUploadID, uploadURL))

	// Step 2: Send file content to upload URL
	if err := n.sendFileContent(uploadURL, fileData); err != nil {
		l.Error(err, "failed to send file content")
		return "", err
	}

	l.Debug("file content sent successfully")

	// Step 3: Wait for upload to complete
	if err := n.waitForUploadComplete(fileUploadID); err != nil {
		l.Error(err, "failed waiting for upload to complete")
		return "", err
	}

	l.Info(fmt.Sprintf("file uploaded successfully: fileUploadID=%s", fileUploadID))

	return fileUploadID, nil
}

// createFileUploadObject creates a file upload object and returns the ID and upload URL
func (n *notionService) createFileUploadObject(filename, contentType string) (string, string, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "createFileUploadObject",
	})

	// Prepare request body
	requestBody := map[string]interface{}{}
	if filename != "" {
		requestBody["filename"] = filename
	}
	if contentType != "" {
		requestBody["content_type"] = contentType
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		l.Error(err, "failed to marshal request body")
		return "", "", err
	}

	l.Debug(fmt.Sprintf("creating file upload object with body: %s", string(bodyBytes)))

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/file_uploads", bytes.NewReader(bodyBytes))
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return "", "", err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request")
		return "", "", err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read response body")
		return "", "", err
	}

	l.Debug(fmt.Sprintf("received response: status=%d, body=%s", resp.StatusCode, string(respBody)))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "file upload object creation failed, body=%s", string(respBody))
		return "", "", fmt.Errorf("failed to create file upload object: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var uploadResp FileUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		l.Error(err, "failed to unmarshal response")
		return "", "", err
	}

	return uploadResp.ID, uploadResp.UploadURL, nil
}

// sendFileContent sends the file content to the upload URL using multipart/form-data
func (n *notionService) sendFileContent(uploadURL string, fileData []byte) error {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "sendFileContent",
	})

	l.Debug(fmt.Sprintf("sending file content to upload URL: size=%d bytes", len(fileData)))

	// Create multipart form data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file field with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="invoice.pdf"`)
	h.Set("Content-Type", "application/pdf")

	part, err := writer.CreatePart(h)
	if err != nil {
		l.Error(err, "failed to create form part")
		return err
	}

	_, err = part.Write(fileData)
	if err != nil {
		l.Error(err, "failed to write file data to form")
		return err
	}

	// Close writer to finalize multipart form
	err = writer.Close()
	if err != nil {
		l.Error(err, "failed to close multipart writer")
		return err
	}

	l.Debug(fmt.Sprintf("created multipart form data: content-type=%s, size=%d bytes", writer.FormDataContentType(), body.Len()))

	// Create HTTP request
	req, err := http.NewRequest("POST", uploadURL, &body)
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Notion-Version", "2022-06-28")

	l.Debug(fmt.Sprintf("sending request with headers: Content-Type=%s", writer.FormDataContentType()))

	// Send request
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request")
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read response body")
		return err
	}

	l.Debug(fmt.Sprintf("received response: status=%d, body=%s", resp.StatusCode, string(respBody)))

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		l.Errorf(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "file content upload failed, body=%s", string(respBody))
		return fmt.Errorf("failed to upload file content: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	l.Debug("file content uploaded successfully")

	return nil
}

// waitForUploadComplete polls the file upload status until it becomes "uploaded"
func (n *notionService) waitForUploadComplete(fileUploadID string) error {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "waitForUploadComplete",
	})

	l.Debug(fmt.Sprintf("waiting for file upload to complete: fileUploadID=%s", fileUploadID))

	maxAttempts := 30
	retryInterval := 2 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		l.Debug(fmt.Sprintf("checking upload status (attempt %d/%d)", attempt, maxAttempts))

		status, err := n.getFileUploadStatus(fileUploadID)
		if err != nil {
			l.Error(err, "failed to get file upload status")
			return err
		}

		l.Debug(fmt.Sprintf("current status: %s", status))

		if status == "uploaded" {
			l.Info("file upload completed successfully")
			return nil
		}

		if status == "archived" {
			l.Error(errors.New("file upload archived"), "file upload was archived before completion")
			return errors.New("file upload was archived (may have expired)")
		}

		// Wait before next attempt
		if attempt < maxAttempts {
			time.Sleep(retryInterval)
		}
	}

	return fmt.Errorf("file upload did not complete within timeout (%d attempts)", maxAttempts)
}

// getFileUploadStatus retrieves the current status of a file upload
func (n *notionService) getFileUploadStatus(fileUploadID string) (string, error) {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "getFileUploadStatus",
	})

	// Note: This assumes the Notion API provides a GET endpoint for file uploads
	// If not available, we may need to use a different approach
	url := fmt.Sprintf("https://api.notion.com/v1/file_uploads/%s", fileUploadID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request")
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read response body")
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "failed to get file upload status, body=%s", string(respBody))
		return "", fmt.Errorf("failed to get file upload status: status=%d", resp.StatusCode)
	}

	var uploadResp FileUploadResponse
	if err := json.Unmarshal(respBody, &uploadResp); err != nil {
		l.Error(err, "failed to unmarshal response")
		return "", err
	}

	return uploadResp.Status, nil
}

// UpdatePageProperties updates the properties of a Notion page
// Special handling for file uploads using raw HTTP since go-notion doesn't support file_upload type yet
func (n *notionService) UpdatePageProperties(pageID string, properties nt.UpdatePageParams) error {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "UpdatePageProperties",
	})

	l.Debug(fmt.Sprintf("updating page properties: pageID=%s", pageID))

	// Marshal properties to JSON for raw HTTP request
	bodyBytes, err := json.Marshal(properties)
	if err != nil {
		l.Error(err, "failed to marshal properties")
		return err
	}

	l.Debug(fmt.Sprintf("updating page with properties: %s", string(bodyBytes)))

	// Create HTTP request
	url := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request")
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read response body")
		return err
	}

	l.Debug(fmt.Sprintf("received response: status=%d, body=%s", resp.StatusCode, string(respBody)))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "page update failed, body=%s", string(respBody))
		return fmt.Errorf("failed to update page properties: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	l.Info(fmt.Sprintf("page properties updated successfully: pageID=%s", pageID))

	return nil
}

// UpdatePagePropertiesWithFileUpload updates a page property with a file upload
// This method constructs the proper JSON structure for file_upload type that go-notion doesn't support yet
func (n *notionService) UpdatePagePropertiesWithFileUpload(pageID, propertyName, fileUploadID, filename string) error {
	l := n.l.Fields(logger.Fields{
		"service": "notion",
		"method":  "UpdatePagePropertiesWithFileUpload",
	})

	l.Debug(fmt.Sprintf("updating page %s property %s with file upload: fileID=%s, filename=%s", pageID, propertyName, fileUploadID, filename))

	// Construct the proper JSON structure for file_upload type
	// {"properties": {"PropertyName": {"files": [{"type": "file_upload", "file_upload": {"id": "..."}, "name": "..."}]}}}
	requestBody := map[string]interface{}{
		"properties": map[string]interface{}{
			propertyName: map[string]interface{}{
				"files": []map[string]interface{}{
					{
						"type": "file_upload",
						"file_upload": map[string]string{
							"id": fileUploadID,
						},
						"name": filename,
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		l.Error(err, "failed to marshal request body")
		return err
	}

	l.Debug(fmt.Sprintf("request body: %s", string(bodyBytes)))

	// Create HTTP request
	url := fmt.Sprintf("https://api.notion.com/v1/pages/%s", pageID)
	req, err := http.NewRequest("PATCH", url, bytes.NewReader(bodyBytes))
	if err != nil {
		l.Error(err, "failed to create HTTP request")
		return err
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+n.secret)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		l.Error(err, "failed to send HTTP request")
		return err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error(err, "failed to read response body")
		return err
	}

	l.Debug(fmt.Sprintf("received response: status=%d, body=%s", resp.StatusCode, string(respBody)))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		l.Errorf(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "page update failed, body=%s", string(respBody))
		return fmt.Errorf("failed to attach file to page: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	l.Info(fmt.Sprintf("file attached to page successfully: pageID=%s, property=%s, fileID=%s", pageID, propertyName, fileUploadID))

	return nil
}
