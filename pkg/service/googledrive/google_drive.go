package googledrive

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/dwarvesf/fortress-api/pkg/config"
	"github.com/dwarvesf/fortress-api/pkg/logger"
	"github.com/dwarvesf/fortress-api/pkg/model"
)

const FullDriveAccessScope = "https://www.googleapis.com/auth/drive"

type googleService struct {
	config    *oauth2.Config
	token     *oauth2.Token
	service   *drive.Service
	appConfig *config.Config
	logger    logger.Logger
}

// New function return Google service
func New(config *oauth2.Config, appConfig *config.Config, l logger.Logger) IService {
	return &googleService{
		config:    config,
		appConfig: appConfig,
		logger:    l,
	}
}

func (g *googleService) UploadInvoicePDF(invoice *model.Invoice, dirName string) error {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	dir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), dirName)
	if err != nil {
		return err
	}

	_, err = g.newFile(fmt.Sprintf("#%s.pdf", invoice.Number), "application/pdf", bytes.NewReader(invoice.InvoiceFileContent), dir.Id)
	if err != nil {
		return err
	}

	return nil
}

func (g *googleService) MoveInvoicePDF(invoice *model.Invoice, fromDirName, toDirName string) error {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	fromDir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), fromDirName)
	if err != nil {
		return err
	}
	if fromDir == nil {
		return fmt.Errorf(`%v directory not found`, fromDirName)
	}

	toDir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), toDirName)
	if err != nil {
		return err
	}
	if toDir == nil {
		return fmt.Errorf(`%v directory not found`, toDirName)
	}

	invoicePdf, err := g.searchFile(fmt.Sprintf("#%s.pdf", invoice.Number), fromDir.Id, false)
	if err != nil {
		return err
	}
	if invoicePdf == nil {
		return ErrInvoicePDFNotFound
	}

	return g.updateInvoiceDir(invoicePdf.Id, fromDir.Id, toDir.Id)
}

func (g *googleService) updateInvoiceDir(fileID, oldDirID, newDirID string) error {
	_, err := g.service.Files.Update(fileID, nil).
		AddParents(newDirID).
		RemoveParents(oldDirID).
		SupportsAllDrives(true).
		Do()
	return err
}

func (g *googleService) findInvoiceDir(year, status string) (*drive.File, error) {
	yearDir, err := g.getDirID(year, g.appConfig.Invoice.DirID)
	if err != nil {
		return nil, err
	}

	return g.getDirID(status, yearDir.Id)
}

func (g *googleService) getDirID(dirName, parentDirID string) (*drive.File, error) {
	g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: searching for dirName=%s in parentDirID=%s", dirName, parentDirID))

	dir, err := g.searchFile(dirName, parentDirID, true)
	if err != nil {
		g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: search error: %v", err))
		return nil, err
	}

	if dir != nil {
		g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: found existing dir id=%s name=%s", dir.Id, dir.Name))
		return dir, nil
	}

	g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: dir not found, creating new dir: %s", dirName))
	newDir, err := g.newDir(dirName, parentDirID)
	if err != nil {
		g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: create dir error: %v", err))
		return nil, err
	}
	g.logger.Debug(fmt.Sprintf("[DEBUG] getDirID: created new dir id=%s name=%s", newDir.Id, newDir.Name))
	return newDir, nil
}

func (g *googleService) ensureToken(rToken string) error {
	token := &oauth2.Token{
		RefreshToken: rToken,
	}

	if !token.Valid() {
		tks := g.config.TokenSource(context.Background(), token)
		tok, err := tks.Token()
		if err != nil {
			return err
		}
		g.token = tok
	}

	return nil
}

func (g *googleService) prepareService() error {
	client := g.config.Client(context.Background(), g.token)
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return errors.New("Get Drive Service Failed " + err.Error())
	}
	g.service = service
	return nil
}

func (g *googleService) searchFile(name, parentId string, isFolder bool) (*drive.File, error) {
	var folderQuery string
	if isFolder {
		folderQuery = "mimeType='application/vnd.google-apps.folder' and "
	}

	var parentQuery string
	if parentId != "root" {
		parentQuery = fmt.Sprintf("'%s' in parents and ", parentId)
	}

	// Try exact match first
	r, err := g.service.Files.List().
		Q(parentQuery + folderQuery + fmt.Sprintf("name='%s'", name)).
		Fields("nextPageToken, files(id, name)").
		SupportsAllDrives(true).
		IncludeItemsFromAllDrives(true).
		Do()
	if err != nil {
		return nil, err
	}

	// If exact match found, return it
	if len(r.Files) > 0 {
		return r.Files[0], nil
	}

	// For folders, try case-insensitive search by listing all folders in parent and comparing
	if isFolder {
		r, err := g.service.Files.List().
			Q(parentQuery + folderQuery + "trashed=false").
			Fields("nextPageToken, files(id, name)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Do()
		if err != nil {
			return nil, err
		}

		// Compare names case-insensitively, and also try slugged comparison
		for _, file := range r.Files {
			if equalIgnoreCase(file.Name, name) {
				return file, nil
			}
			// Also try comparing slugged versions (to match "LE MINH QUANG" with "le-minh-quang")
			if slugContractorName(file.Name) == name {
				g.logger.Debug(fmt.Sprintf("[DEBUG] searchFile: found folder via slug match: '%s' (slugs to '%s') matches search '%s'",
					file.Name, slugContractorName(file.Name), name))
				return file, nil
			}
		}
	}

	return nil, nil
}

// equalIgnoreCase compares two strings case-insensitively
func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// slugContractorName converts contractor name to a consistent folder name format
// Preserves Unicode characters (e.g., Vietnamese names like "Trương Hồng Ngọc")
// Only removes characters that are invalid in Google Drive folder names
func slugContractorName(name string) string {
	slug := name

	// Remove characters that are not allowed in Google Drive folder names
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, ch := range invalidChars {
		slug = strings.ReplaceAll(slug, ch, "")
	}

	// Trim leading/trailing whitespace
	slug = strings.TrimSpace(slug)

	return slug
}

func (g *googleService) newDir(name string, parentId string) (*drive.File, error) {
	d := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentId},
	}

	return g.service.Files.Create(d).SupportsAllDrives(true).Do()
}

func (g *googleService) newFile(name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}

	return g.service.Files.Create(f).Media(content).SupportsAllDrives(true).Do()
}

// func (g *googleService) deleteFile(id string) error {
// 	return g.service.Files.Delete(id).Do()
// }

// func getDrivePreviewLink(fileID string) string {
// 	return fmt.Sprintf(`https://drive.google.com/file/d/%s/view`, fileID)
// }

// func getFileIDFromLink(url string) string {
// 	s := strings.Replace(url, "https://drive.google.com/file/d/", "", 1)
// 	return strings.Replace(s, "/view", "", 1)
// }

func (g *googleService) DownloadInvoicePDF(invoice *model.Invoice, dirName string) ([]byte, error) {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return nil, err
	}

	if err := g.prepareService(); err != nil {
		return nil, err
	}

	dir, err := g.findInvoiceDir(strconv.Itoa(invoice.Year), dirName)
	if err != nil {
		return nil, err
	}

	f, err := g.searchFile(fmt.Sprintf("#%s.pdf", invoice.Number), dir.Id, false)
	if err != nil {
		return nil, err
	}

	if f == nil {
		return nil, fmt.Errorf(`file not found`)
	}

	resp, err := g.service.Files.Get(f.Id).SupportsAllDrives(true).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// FindContractorInvoiceFileID finds a contractor invoice PDF in Google Drive
// Searches in the contractor's subfolder under the contractor invoice parent folder
// Returns the file ID if found, empty string if not found
func (g *googleService) FindContractorInvoiceFileID(contractorName, invoiceID string) (string, error) {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return "", err
	}

	if err := g.prepareService(); err != nil {
		return "", err
	}

	// Slug the contractor name to match folder naming convention
	folderName := slugContractorName(contractorName)

	g.logger.Debug(fmt.Sprintf("[DEBUG] FindContractorInvoiceFileID: searching for invoiceID=%s in contractorName=%s (folderName=%s)",
		invoiceID, contractorName, folderName))

	// Find contractor's subfolder
	contractorDir, err := g.searchFile(folderName, g.appConfig.Invoice.ContractorInvoiceDirID, true)
	if err != nil {
		return "", fmt.Errorf("failed to search contractor directory: %w", err)
	}
	if contractorDir == nil {
		g.logger.Debug(fmt.Sprintf("[DEBUG] FindContractorInvoiceFileID: contractor folder not found: %s", folderName))
		return "", nil
	}

	g.logger.Debug(fmt.Sprintf("[DEBUG] FindContractorInvoiceFileID: found contractor folder id=%s name=%s", contractorDir.Id, contractorDir.Name))

	// Search for the PDF file in contractor's folder
	fileName := invoiceID + ".pdf"
	file, err := g.searchFile(fileName, contractorDir.Id, false)
	if err != nil {
		return "", fmt.Errorf("failed to search invoice file: %w", err)
	}
	if file == nil {
		g.logger.Debug(fmt.Sprintf("[DEBUG] FindContractorInvoiceFileID: invoice file not found: %s", fileName))
		return "", nil
	}

	g.logger.Debug(fmt.Sprintf("[DEBUG] FindContractorInvoiceFileID: found file id=%s name=%s", file.Id, file.Name))
	return file.Id, nil
}

// ShareFileWithEmail shares a Google Drive file with the specified email address
// Google automatically sends a notification email to the recipient
// Uses spawn@d.foundation (TeamGoogleRefreshToken) for sharing
func (g *googleService) ShareFileWithEmail(fileID, email string) error {
	if err := g.ensureToken(g.appConfig.Google.TeamGoogleRefreshToken); err != nil {
		return err
	}

	if err := g.prepareService(); err != nil {
		return err
	}

	permission := &drive.Permission{
		Type:         "user",
		Role:         "reader",
		EmailAddress: email,
	}

	_, err := g.service.Permissions.Create(fileID, permission).
		SendNotificationEmail(true).
		EmailMessage("Your invoice has been generated and is ready for review.").
		SupportsAllDrives(true).
		Do()

	if err != nil {
		return fmt.Errorf("failed to share file with email %s: %w", email, err)
	}

	return nil
}

// UploadContractorInvoicePDF uploads a contractor invoice PDF to Google Drive
// It creates a subfolder for the contractor if it doesn't exist
// Returns the public URL of the uploaded file
func (g *googleService) UploadContractorInvoicePDF(contractorName, fileName string, pdfBytes []byte) (string, error) {
	if err := g.ensureToken(g.appConfig.Google.AccountingGoogleRefreshToken); err != nil {
		return "", err
	}

	if err := g.prepareService(); err != nil {
		return "", err
	}

	// Slug the contractor name for consistent folder naming
	folderName := slugContractorName(contractorName)

	// Debug logging
	g.logger.Debug(fmt.Sprintf("[DEBUG] UploadContractorInvoicePDF: contractorName=%s folderName=%s parentDirID=%s",
		contractorName, folderName, g.appConfig.Invoice.ContractorInvoiceDirID))

	// Get or create contractor subfolder
	contractorDir, err := g.getDirID(folderName, g.appConfig.Invoice.ContractorInvoiceDirID)
	if err != nil {
		return "", fmt.Errorf("failed to get contractor directory: %w", err)
	}

	g.logger.Debug(fmt.Sprintf("[DEBUG] UploadContractorInvoicePDF: got contractorDir id=%s name=%s", contractorDir.Id, contractorDir.Name))

	// Upload the PDF file
	file, err := g.newFile(fileName, "application/pdf", bytes.NewReader(pdfBytes), contractorDir.Id)
	if err != nil {
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	// Return the Google Drive file URL
	return fmt.Sprintf("https://drive.google.com/file/d/%s/view", file.Id), nil
}
